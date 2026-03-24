package api

import (
	"context"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// CreateStory creates a new story.
func (c *Client) CreateStory(ctx context.Context, req CreateStoryRequest) (*Story, error) {
	resp, err := c.raw.PostStories(ctx, gen.PostStoriesJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("creating story: %w", err)
	}
	defer resp.Body.Close()

	var result Story
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("creating story: %w", err)
	}
	return &result, nil
}

// UpdateStory updates a story's metadata.
func (c *Client) UpdateStory(ctx context.Context, id string, req UpdateStoryRequest) error {
	resp, err := c.raw.PutStoryId(ctx, id, gen.PutStoryIdJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("updating story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("updating story %s: %w", id, err)
	}
	return nil
}

// PublishStory publishes a story.
func (c *Client) PublishStory(ctx context.Context, id string) error {
	resp, err := c.raw.PostStoryIdPublish(ctx, id)
	if err != nil {
		return fmt.Errorf("publishing story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("publishing story %s: %w", id, err)
	}
	return nil
}

// UnpublishStory unpublishes a story.
func (c *Client) UnpublishStory(ctx context.Context, id string) error {
	resp, err := c.raw.PostStoryIdUnpublish(ctx, id)
	if err != nil {
		return fmt.Errorf("unpublishing story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("unpublishing story %s: %w", id, err)
	}
	return nil
}

// ListStories returns the authenticated user's stories.
func (c *Client) ListStories(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error) {
	resp, err := c.raw.GetStories(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("listing stories: %w", err)
	}
	defer resp.Body.Close()

	var result StoryList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("listing stories: %w", err)
	}
	return &result, nil
}

// GetStory returns a single story by ID.
func (c *Client) GetStory(ctx context.Context, id string) (*Story, error) {
	resp, err := c.raw.GetStoryId(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting story %s: %w", id, err)
	}
	defer resp.Body.Close()

	var result Story
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("getting story %s: %w", id, err)
	}
	return &result, nil
}

// DownloadStory downloads a story in the given format (json, markdown).
func (c *Client) DownloadStory(ctx context.Context, id string, format string) (string, error) {
	resp, err := c.raw.GetStoryIdDownloadFormat(ctx, id, gen.GetStoryIdDownloadFormatParamsFormat(format))
	if err != nil {
		return "", fmt.Errorf("downloading story %s as %s: %w", id, format, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return "", fmt.Errorf("downloading story %s as %s: %w", id, format, err)
	}
	return string(body), nil
}

// ListStoriesWithReviewStatus returns the author's stories with review status info.
func (c *Client) ListStoriesWithReviewStatus(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error) {
	resp, err := c.raw.GetStoriesMyReviewStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("listing stories with review status: %w", err)
	}
	defer resp.Body.Close()

	var result StoriesWithReview
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("listing stories with review status: %w", err)
	}
	return &result, nil
}
