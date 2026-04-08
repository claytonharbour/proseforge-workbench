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
	// image_generate
	s.AddTool(
		tool("image_generate",
			mcp.WithDescription("Generate an AI image via platform providers. Costs 2 credits. Async (202) — poll with image_get until status='completed'. Optionally scope to a story/section for context-aware generation."),
			mcp.WithString("story_id", mcp.Description("Story ID to generate image for (provides context to the AI)")),
			mcp.WithString("section_id", mcp.Description("Section ID for section-specific imagery. Requires story_id.")),
			mcp.WithString("user_prompt", mcp.Description("Prompt describing the desired image")),
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

			body := gen.HandlersGenerateImageRequest{}
			if v := optionalArg(req, "story_id"); v != "" {
				body.StoryId = &v
			}
			if v := optionalArg(req, "section_id"); v != "" {
				body.SectionId = &v
			}
			if v := optionalArg(req, "user_prompt"); v != "" {
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

	// image_upload
	s.AddTool(
		tool("image_upload",
			mcp.WithDescription("Upload a pre-made image (BYOAI path). For images generated externally (ComfyUI, DALL-E, etc). Max 10MB, jpeg/png/webp. Requires file_path on the local filesystem."),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the image file on disk")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			filePath, err := requireArg(req, "file_path")
			if err != nil {
				return toolError(err), nil
			}

			contentType, body, err := buildMultipartUpload(filePath)
			if err != nil {
				return toolError(err), nil
			}

			result, err := svc.UploadImage(ctx, contentType, body)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// image_get
	s.AddTool(
		tool("image_get",
			mcp.WithDescription("Get image details and generation status. Poll after image_generate or image_regenerate until status='completed'. Requires image_id."),
			mcp.WithString("image_id", mcp.Required(), mcp.Description("Image ID from image_generate, image_upload, or image_list")),
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

	// image_list
	s.AddTool(
		tool("image_list",
			mcp.WithDescription("List the user's image library. Supports pagination with limit/offset."),
			mcp.WithNumber("limit", mcp.Description("Max results (default 50, max 100)")),
			mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			params := &gen.GetImagesParams{}
			if v := optionalIntArg(req, "limit", 0); v > 0 {
				params.Limit = &v
			}
			if v := optionalIntArg(req, "offset", 0); v > 0 {
				params.Offset = &v
			}

			result, err := svc.ListImages(ctx, params)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// image_regenerate
	s.AddTool(
		tool("image_regenerate",
			mcp.WithDescription("Re-roll an existing image with optional new prompt. Generates a new image for the same slot — does NOT delete the original. Costs 2 credits. Async (202) — poll with image_get. Requires image_id from image_get or image_list."),
			mcp.WithString("image_id", mcp.Required(), mcp.Description("Image ID to regenerate")),
			mcp.WithString("user_prompt", mcp.Description("Optional new prompt for the re-roll")),
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
			if v := optionalArg(req, "user_prompt"); v != "" {
				body.UserPrompt = &v
			}

			result, err := svc.RegenerateImage(ctx, id, body)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// story_images
	s.AddTool(
		tool("story_images",
			mcp.WithDescription("List images attached to a story. Requires story_id from story_list or story_get."),
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

	// story_image_attach
	s.AddTool(
		tool("story_image_attach",
			mcp.WithDescription("Attach an image to a story. Requires story_id and image_id. Get image_id from image_generate, image_upload, or image_list."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("image_id", mcp.Required(), mcp.Description("Image ID to attach")),
			mcp.WithBoolean("is_primary", mcp.Description("Set as primary/cover image on attach (default false)")),
			mcp.WithNumber("position", mcp.Description("Display position/order for the image")),
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
			imageID, err := requireArg(req, "image_id")
			if err != nil {
				return toolError(err), nil
			}

			body := gen.HandlersAddToStoryRequest{}
			if optionalBoolArg(req, "is_primary") {
				t := true
				body.IsPrimary = &t
			}
			if v := optionalIntArg(req, "position", 0); v > 0 {
				body.Position = &v
			}

			if err := svc.AttachImageToStory(ctx, storyID, imageID, body); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Image attached."), nil
		},
	)

	// story_image_cover
	s.AddTool(
		tool("story_image_cover",
			mcp.WithDescription("Set an attached image as the story's cover/primary image. Image must already be attached — use story_image_attach first. Requires story_id and image_id from story_images."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("image_id", mcp.Required(), mcp.Description("Image ID (must be already attached to the story)")),
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
			imageID, err := requireArg(req, "image_id")
			if err != nil {
				return toolError(err), nil
			}

			if err := svc.SetStoryImageCover(ctx, storyID, imageID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Cover image set."), nil
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
