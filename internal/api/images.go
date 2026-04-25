package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// GenerateImage triggers async AI image generation. Returns 202 with generation info.
// Poll with GetImage until status='completed'.
func (c *Client) GenerateImage(ctx context.Context, req gen.HandlersGenerateImageRequest) (json.RawMessage, error) {
	resp, err := c.raw.PostImagesGenerate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("generate image: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("generate image: %w", err)
	}
	return json.RawMessage(body), nil
}

// UploadImage uploads a pre-made image via multipart form.
// The caller provides the content type (e.g. "multipart/form-data; boundary=...")
// and the encoded body reader.
func (c *Client) UploadImage(ctx context.Context, contentType string, body io.Reader) (json.RawMessage, error) {
	resp, err := c.raw.PostImagesUploadWithBody(ctx, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("upload image: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("upload image: %w", err)
	}
	return json.RawMessage(respBody), nil
}

// GetImage returns image details and generation status.
func (c *Client) GetImage(ctx context.Context, id string) (json.RawMessage, error) {
	resp, err := c.raw.GetImageId(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get image %s: %w", id, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get image %s: %w", id, err)
	}
	return json.RawMessage(body), nil
}

// ListImages returns the user's image library with pagination.
func (c *Client) ListImages(ctx context.Context, params *gen.GetImagesParams) (json.RawMessage, error) {
	resp, err := c.raw.GetImages(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	return json.RawMessage(body), nil
}

// RegenerateImage re-rolls an existing image with an optional new prompt.
// Returns 202 — poll with GetImage until status='completed'.
func (c *Client) RegenerateImage(ctx context.Context, id string, req gen.HandlersRegenerateRequest) (json.RawMessage, error) {
	resp, err := c.raw.PostImageIdRegenerate(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("regenerate image %s: %w", id, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("regenerate image %s: %w", id, err)
	}
	return json.RawMessage(body), nil
}

// ListStoryImages returns images attached to a story.
func (c *Client) ListStoryImages(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdImages(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("list images for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list images for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// AttachImageToStory attaches an image to a story.
func (c *Client) AttachImageToStory(ctx context.Context, storyID, imageID string, req gen.HandlersAddToStoryRequest) error {
	resp, err := c.raw.PostStoryIdImagesImageId(ctx, storyID, imageID, req)
	if err != nil {
		return fmt.Errorf("attach image %s to story %s: %w", imageID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("attach image %s to story %s: %w", imageID, storyID, err)
	}
	return nil
}

// SetStoryImageCover sets an attached image as the story's primary/cover image.
func (c *Client) SetStoryImageCover(ctx context.Context, storyID, imageID string) error {
	resp, err := c.raw.PutStoryIdImagesImageIdPrimary(ctx, storyID, imageID)
	if err != nil {
		return fmt.Errorf("set cover image %s for story %s: %w", imageID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("set cover image %s for story %s: %w", imageID, storyID, err)
	}
	return nil
}
