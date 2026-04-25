package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// === story image (group) ===

func newStoryImageGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Image generation and management",
	}
	cmd.AddCommand(
		newImageGenerateCmd(),
		newImageUploadCmd(),
		newImageGetCmd(),
		newImageListCmd(),
		newImageRegenerateCmd(),
		newStoryImagesCmd(),
		newStoryImageAttachCmd(),
		newStoryImageCoverCmd(),
	)
	return cmd
}

// === image generate ===

func newImageGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate an AI image (async, costs 2 credits)",
		RunE:  runImageGenerate,
	}
	cmd.Flags().String("story-id", "", "Story ID for context-aware generation")
	cmd.Flags().String("section-id", "", "Section ID for section-specific imagery")
	cmd.Flags().String("prompt", "", "Prompt describing the desired image")
	cmd.Flags().String("template-id", "", "Image template ID")
	cmd.Flags().Int("width", 0, "Image width in pixels")
	cmd.Flags().Int("height", 0, "Image height in pixels")
	return cmd
}

func runImageGenerate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	req := gen.HandlersGenerateImageRequest{}
	if v, _ := cmd.Flags().GetString("story-id"); v != "" {
		req.StoryId = &v
	}
	if v, _ := cmd.Flags().GetString("section-id"); v != "" {
		req.SectionId = &v
	}
	if v, _ := cmd.Flags().GetString("prompt"); v != "" {
		req.UserPrompt = &v
	}
	if v, _ := cmd.Flags().GetString("template-id"); v != "" {
		req.TemplateId = &v
	}
	if v, _ := cmd.Flags().GetInt("width"); v > 0 {
		req.Width = &v
	}
	if v, _ := cmd.Flags().GetInt("height"); v > 0 {
		req.Height = &v
	}

	result, err := svc.GenerateImage(cmd.Context(), req)
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === image upload ===

func newImageUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a pre-made image (BYOAI path, max 10MB)",
		RunE:  runImageUpload,
	}
	cmd.Flags().String("file", "", "Path to image file (jpeg/png/webp)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runImageUpload(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	filePath, _ := cmd.Flags().GetString("file")

	contentType, body, err := buildMultipartUpload(filePath)
	if err != nil {
		return err
	}

	result, err := svc.UploadImage(cmd.Context(), contentType, body)
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === image get ===

func newImageGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <image-id>",
		Short: "Get image details and generation status",
		Args:  cobra.ExactArgs(1),
		RunE:  runImageGet,
	}
}

func runImageGet(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.GetImage(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === image list ===

func newImageListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user's image library",
		RunE:  runImageList,
	}
	cmd.Flags().Int("limit", 50, "Max results (1-100)")
	cmd.Flags().Int("offset", 0, "Pagination offset")
	return cmd
}

func runImageList(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")

	params := &gen.GetImagesParams{}
	if limit > 0 {
		params.Limit = &limit
	}
	if offset > 0 {
		params.Offset = &offset
	}

	result, err := svc.ListImages(cmd.Context(), params)
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === image regenerate ===

func newImageRegenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate <image-id>",
		Short: "Re-roll an existing image (async, costs 2 credits)",
		Args:  cobra.ExactArgs(1),
		RunE:  runImageRegenerate,
	}
	cmd.Flags().String("prompt", "", "Optional new prompt for the re-roll")
	return cmd
}

func runImageRegenerate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	req := gen.HandlersRegenerateRequest{}
	if v, _ := cmd.Flags().GetString("prompt"); v != "" {
		req.UserPrompt = &v
	}

	result, err := svc.RegenerateImage(cmd.Context(), args[0], req)
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === story images (list attached) ===

func newStoryImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "story-images <story-id>",
		Short: "List images attached to a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryImages,
	}
}

func runStoryImages(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.ListStoryImages(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === story image attach ===

func newStoryImageAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <story-id> <image-id>",
		Short: "Attach an image to a story",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryImageAttach,
	}
	cmd.Flags().Bool("primary", false, "Set as primary/cover image on attach")
	cmd.Flags().Int("position", 0, "Display position/order")
	return cmd
}

func runStoryImageAttach(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	req := gen.HandlersAddToStoryRequest{}
	if v, _ := cmd.Flags().GetBool("primary"); v {
		req.IsPrimary = &v
	}
	if v, _ := cmd.Flags().GetInt("position"); v > 0 {
		req.Position = &v
	}

	if err := svc.AttachImageToStory(cmd.Context(), args[0], args[1], req); err != nil {
		return err
	}

	fmt.Println("Image attached.")
	return nil
}

// === story image cover ===

func newStoryImageCoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cover <story-id> <image-id>",
		Short: "Set an attached image as the story cover",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryImageCover,
	}
}

func runStoryImageCover(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.SetStoryImageCover(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}

	fmt.Println("Cover image set.")
	return nil
}

// buildMultipartUpload reads a file from disk and encodes it as a multipart form upload.
func buildMultipartUpload(filePath string) (string, *bytes.Buffer, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Use CreatePart with explicit Content-Type so the API can detect the file type
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
