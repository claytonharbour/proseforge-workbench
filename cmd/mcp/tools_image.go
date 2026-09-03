package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/claytonharbour/proseforge-workbench/internal/story"
)

func registerImageTools(s *server.MCPServer, r *clientResolver) {
	// story_image_generate — Generate an image for a story
	s.AddTool(
		tool("story_image_generate",
			mcp.WithDescription(
				"Generate an AI image for a story. The image is automatically attached to the "+
					"story when generation completes — no separate attach call needed.\n\n"+
					"Pass story_id alone for a cover image. Pass story_id + section_id for a "+
					"section-specific image. Pass prompt for custom imagery, or omit it to let "+
					"the platform generate from story/section context.\n\n"+
					"Costs 2 credits. Generation is async — the image will appear on the story "+
					"when ready. Use story_images to check what's attached."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story to generate image for (auto-attaches)")),
			mcp.WithString("section_id", mcp.Description("Section for context-aware generation")),
			mcp.WithString("prompt", mcp.Description("Image prompt. If omitted, platform generates from story/section context")),
			mcp.WithString("template_id", mcp.Description("Image template ID for guided generation")),
			mcp.WithNumber("width", mcp.Description("Image width in pixels")),
			mcp.WithNumber("height", mcp.Description("Image height in pixels")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}

			body := gen.HandlersGenerateImageRequest{
				StoryId: &storyID,
			}
			if v := optionalArg(req, "section_id"); v != "" {
				body.SectionId = &v
			}
			if v := optionalArg(req, "prompt"); v != "" {
				body.UserPrompt = &v
			}
			if v := optionalArg(req, "template_id"); v != "" {
				body.TemplateId = &v
			}
			if w := optionalIntArg(req, "width", 0); w > 0 {
				body.Width = &w
			}
			if h := optionalIntArg(req, "height", 0); h > 0 {
				body.Height = &h
			}

			result, err := svc.GenerateImage(ctx, body)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// story_image_upload — Upload and attach a pre-made image
	s.AddTool(
		tool("story_image_upload",
			mcp.WithDescription(
				"Upload a pre-made image and attach it to a story. For images generated "+
					"externally (ComfyUI, DALL-E, etc). Max 10 MiB (10,485,760 bytes — binary, not 10,000,000), jpeg/png/webp.\n\n"+
					"Requires file_path on the local filesystem and story_id."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story to attach the image to")),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the image file on disk")),
			mcp.WithBoolean("cover", mcp.Description("Set as cover image after attaching (default false)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			filePath, err := requireArg(req, "file_path")
			if err != nil {
				return toolError(err), nil
			}

			cover := optionalBoolArg(req, "cover")

			contentType, body, err := buildMultipartUpload(filePath)
			if err != nil {
				return toolError(err), nil
			}

			result, err := svc.UploadAndAttachImage(ctx, storyID, contentType, body, cover)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// story_images — List images attached to a story
	s.AddTool(
		tool("story_images",
			mcp.WithDescription(
				"List images attached to a story. Each entry includes the image id, "+
					"position, isPrimary (cover) flag, and the embedded image record with "+
					"url, model (which AI model produced it), status, and a "+
					"promptExpansionId reference. For the full expanded prompt text "+
					"(the actual prompt sent to the model after platform wrapping), "+
					"call image_get with the image id.\n\n"+
					"By default, returns story-level (cover-tier) images only. Pass "+
					"include_sections=true to also return images attached to the story's "+
					"sections in the same response — useful for one-call audits across a "+
					"whole story instead of N per-section calls."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithBoolean("include_sections", mcp.Description("If true, also include images attached to the story's sections (default false)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			includeSections := optionalBoolArg(req, "include_sections")

			result, err := svc.ListStoryImages(ctx, storyID, includeSections)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// image_regenerate — Re-roll an existing image
	s.AddTool(
		tool("image_regenerate",
			mcp.WithDescription(
				"Re-roll an existing image with an optional new prompt. Generates a new "+
					"version for the same slot — does NOT delete the original. Costs 2 credits. "+
					"Use when an attached image needs iteration without changing the attachment."),
			mcp.WithString("image_id", mcp.Required(), mcp.Description("Image ID to regenerate")),
			mcp.WithString("prompt", mcp.Description("Optional new prompt for the re-roll")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			id, err := requireArg(req, "image_id")
			if err != nil {
				return toolError(err), nil
			}

			body := gen.HandlersRegenerateRequest{}
			if v := optionalArg(req, "prompt"); v != "" {
				body.UserPrompt = &v
			}

			result, err := svc.RegenerateImage(ctx, id, body)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// image_get — Get image details (status, url, model, expanded prompt)
	s.AddTool(
		tool("image_get",
			mcp.WithDescription(
				"Get image details by ID. Returns status (pending/processing/completed/failed), "+
					"url (the CDN URL once ready), model (which AI model produced it), and "+
					"promptExpansion (the actual prompt sent to the model after platform "+
					"wrapping). Use to poll an in-flight generation or to audit a finished image."),
			mcp.WithString("image_id", mcp.Required(), mcp.Description("Image ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			id, err := requireArg(req, "image_id")
			if err != nil {
				return toolError(err), nil
			}

			result, err := svc.GetImage(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// story_image_generate_and_wait — One-call generate + poll + return URL
	s.AddTool(
		tool("story_image_generate_and_wait",
			mcp.WithDescription(
				"Generate an AI image for a story, wait for it to finish, and return the "+
					"completed image record with URL. Same parameters as story_image_generate "+
					"plus an optional timeout. Use this when you need the URL in one call "+
					"instead of generate-then-poll. Costs 2 credits.\n\n"+
					"On timeout, returns the most recent image record and an image_id you can "+
					"keep polling with image_get — generation is not cancelled."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story to generate image for (auto-attaches)")),
			mcp.WithString("section_id", mcp.Description("Section for context-aware generation")),
			mcp.WithString("prompt", mcp.Description("Image prompt. If omitted, platform generates from story/section context")),
			mcp.WithString("template_id", mcp.Description("Image template ID for guided generation")),
			mcp.WithNumber("width", mcp.Description("Image width in pixels")),
			mcp.WithNumber("height", mcp.Description("Image height in pixels")),
			mcp.WithNumber("timeout_seconds", mcp.Description("Max seconds to wait for completion (default 120)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}

			body := buildImageRequest(req, storyID)
			timeoutSec := optionalIntArg(req, "timeout_seconds", 120)

			result, imageID, err := svc.GenerateImageAndWait(ctx, body, time.Duration(timeoutSec)*time.Second)
			if err != nil {
				// Return the partial result alongside the error so the caller has
				// the image_id and can keep polling.
				partial := map[string]any{
					"error":    err.Error(),
					"image_id": imageID,
				}
				if len(result) > 0 {
					var lastBody json.RawMessage = result
					partial["last_status"] = lastBody
				}
				out, _ := json.Marshal(partial)
				return mcp.NewToolResultText(string(out)), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// story_images_generate_batch — Fire N image generations in parallel
	s.AddTool(
		tool("story_images_generate_batch",
			mcp.WithDescription(
				"Fire N image generations for a story in parallel. Useful for per-section "+
					"regeneration across a whole story. Each item accepts section_id and "+
					"prompt; story_id is shared across the batch.\n\n"+
					"Returns initial generation responses indexed in input order. Each entry "+
					"is either a generate response (with image_id you can poll) or an error. "+
					"Costs 2 credits per item."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story all images attach to")),
			mcp.WithString("items", mcp.Required(), mcp.Description(
				"JSON array of {section_id?, prompt?, template_id?, width?, height?} objects, "+
					"e.g. '[{\"section_id\":\"abc\",\"prompt\":\"...\"},{\"section_id\":\"def\",\"prompt\":\"...\"}]'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			itemsJSON, err := requireArg(req, "items")
			if err != nil {
				return toolError(err), nil
			}

			var items []batchImageItem
			if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
				return toolError(fmt.Errorf("parse items JSON: %w", err)), nil
			}
			if len(items) == 0 {
				return toolError(fmt.Errorf("items array is empty")), nil
			}

			requests := make([]gen.HandlersGenerateImageRequest, len(items))
			for i, it := range items {
				sid := storyID
				requests[i] = gen.HandlersGenerateImageRequest{StoryId: &sid}
				if it.SectionID != "" {
					s := it.SectionID
					requests[i].SectionId = &s
				}
				if it.Prompt != "" {
					p := it.Prompt
					requests[i].UserPrompt = &p
				}
				if it.TemplateID != "" {
					t := it.TemplateID
					requests[i].TemplateId = &t
				}
				if it.Width > 0 {
					w := it.Width
					requests[i].Width = &w
				}
				if it.Height > 0 {
					h := it.Height
					requests[i].Height = &h
				}
			}

			results, errs := svc.GenerateImageBatch(ctx, requests)

			out := make([]map[string]any, len(items))
			for i := range items {
				entry := map[string]any{"index": i}
				if errs[i] != nil {
					entry["error"] = errs[i].Error()
				} else {
					var raw json.RawMessage = results[i]
					entry["response"] = raw
				}
				out[i] = entry
			}
			body, _ := json.Marshal(map[string]any{"results": out})
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// batchImageItem is one entry in the story_images_generate_batch input.
type batchImageItem struct {
	SectionID  string `json:"section_id"`
	Prompt     string `json:"prompt"`
	TemplateID string `json:"template_id"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

// buildImageRequest constructs a HandlersGenerateImageRequest from the
// optional MCP arguments shared between story_image_generate and
// story_image_generate_and_wait.
func buildImageRequest(req mcp.CallToolRequest, storyID string) gen.HandlersGenerateImageRequest {
	body := gen.HandlersGenerateImageRequest{StoryId: &storyID}
	if v := optionalArg(req, "section_id"); v != "" {
		body.SectionId = &v
	}
	if v := optionalArg(req, "prompt"); v != "" {
		body.UserPrompt = &v
	}
	if v := optionalArg(req, "template_id"); v != "" {
		body.TemplateId = &v
	}
	if w := optionalIntArg(req, "width", 0); w > 0 {
		body.Width = &w
	}
	if h := optionalIntArg(req, "height", 0); h > 0 {
		body.Height = &h
	}
	return body
}

// buildMultipartUpload reads a file from disk and encodes it as a multipart form upload.
// Returns the content type (with boundary) and the encoded body.
func buildMultipartUpload(filePath string) (string, io.Reader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	mimeType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	}
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath))},
		"Content-Type":        {mimeType},
	})
	if err != nil {
		return "", nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return "", nil, fmt.Errorf("copy file data: %w", err)
	}

	if err := w.Close(); err != nil {
		return "", nil, fmt.Errorf("close multipart writer: %w", err)
	}

	return w.FormDataContentType(), &buf, nil
}
