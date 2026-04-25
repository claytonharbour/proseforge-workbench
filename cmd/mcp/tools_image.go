package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

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
					"externally (ComfyUI, DALL-E, etc). Max 10MB, jpeg/png/webp.\n\n"+
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
			mcp.WithDescription("List images attached to a story. Shows cover and section images with URLs and metadata."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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

			result, err := svc.ListStoryImages(ctx, storyID)
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
