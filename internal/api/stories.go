package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// contentFalse is used to request thin responses without section content.
var contentFalse = false

// excludeContent is a request editor that appends ?content=false to the URL.
// Used for list endpoints whose params struct doesn't have a Content field yet.
func excludeContent(_ context.Context, req *http.Request) error {
	q := req.URL.Query()
	q.Set("content", "false")
	req.URL.RawQuery = q.Encode()
	return nil
}

// CreateStory creates a new story.
func (c *Client) CreateStory(ctx context.Context, req CreateStoryRequest) (*Story, error) {
	resp, err := c.raw.PostStories(ctx, gen.PostStoriesJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("create story: %w", err)
	}
	defer resp.Body.Close()

	var result Story
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create story: %w", err)
	}
	return &result, nil
}

// UpdateStory updates a story's metadata.
func (c *Client) UpdateStory(ctx context.Context, id string, req UpdateStoryRequest) error {
	resp, err := c.raw.PutStoryId(ctx, id, gen.PutStoryIdJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("update story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update story %s: %w", id, err)
	}
	return nil
}

// PublishStory publishes a story with optional visibility.
// Pass "" to use the server default (public).
func (c *Client) PublishStory(ctx context.Context, id string, visibility string) error {
	body := gen.PostStoryIdPublishJSONRequestBody{}
	if visibility != "" {
		body["visibility"] = visibility
	}
	resp, err := c.raw.PostStoryIdPublish(ctx, id, body)
	if err != nil {
		return fmt.Errorf("publish story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("publish story %s: %w", id, err)
	}
	return nil
}

// UnpublishStory unpublishes a story.
func (c *Client) UnpublishStory(ctx context.Context, id string) error {
	resp, err := c.raw.PostStoryIdUnpublish(ctx, id)
	if err != nil {
		return fmt.Errorf("unpublish story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("unpublish story %s: %w", id, err)
	}
	return nil
}

// UpdateVisibility changes the visibility of a published story.
// visibility must be "public" or "members".
func (c *Client) UpdateVisibility(ctx context.Context, id string, visibility string) error {
	body := gen.PutStoryIdVisibilityJSONRequestBody{"visibility": visibility}
	resp, err := c.raw.PutStoryIdVisibility(ctx, id, body)
	if err != nil {
		return fmt.Errorf("update visibility for story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update visibility for story %s: %w", id, err)
	}
	return nil
}

// ListStories returns the authenticated user's stories.
// Section content is excluded by default — use Export or GetSection for content.
func (c *Client) ListStories(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error) {
	resp, err := c.raw.GetStories(ctx, params, excludeContent)
	if err != nil {
		return nil, fmt.Errorf("list stories: %w", err)
	}
	defer resp.Body.Close()

	var result StoryList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list stories: %w", err)
	}
	return &result, nil
}

// GetStory returns a single story by ID.
// Section content is excluded by default — use Export or GetSection for content.
func (c *Client) GetStory(ctx context.Context, id string) (*Story, error) {
	params := &gen.GetStoryIdParams{Content: &contentFalse}
	resp, err := c.raw.GetStoryId(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("get story %s: %w", id, err)
	}
	defer resp.Body.Close()

	var result Story
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get story %s: %w", id, err)
	}
	return &result, nil
}

// ResolveVanityURL resolves a vanity URL (@handle/slug) to story metadata.
func (c *Client) ResolveVanityURL(ctx context.Context, handle, slug string) (json.RawMessage, error) {
	resp, err := c.raw.GetResolveAuthorHandleStorySlug(ctx, handle, slug)
	if err != nil {
		return nil, fmt.Errorf("resolve @%s/%s: %w", handle, slug, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("resolve @%s/%s: %w", handle, slug, err)
	}
	return json.RawMessage(body), nil
}

// GetStoryWithContent returns a story with full section content included.
func (c *Client) GetStoryWithContent(ctx context.Context, id string) (*Story, error) {
	contentTrue := true
	params := &gen.GetStoryIdParams{Content: &contentTrue}
	resp, err := c.raw.GetStoryId(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("get story %s with content: %w", id, err)
	}
	defer resp.Body.Close()

	var result Story
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get story %s with content: %w", id, err)
	}
	return &result, nil
}

// DownloadStory downloads a story in the given format (json, markdown).
func (c *Client) DownloadStory(ctx context.Context, id string, format string) (string, error) {
	resp, err := c.raw.GetStoryIdDownloadFormat(ctx, id, gen.GetStoryIdDownloadFormatParamsFormat(format))
	if err != nil {
		return "", fmt.Errorf("download story %s as %s: %w", id, format, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return "", fmt.Errorf("download story %s as %s: %w", id, format, err)
	}
	return string(body), nil
}

// ListVersions returns version history (git commits) for a story.
func (c *Client) ListVersions(ctx context.Context, storyID string, params *gen.GetStoryIdVersionsParams) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdVersions(ctx, storyID, params)
	if err != nil {
		return nil, fmt.Errorf("list versions for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list versions for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// GetVersion returns story content at a specific version (git SHA).
func (c *Client) GetVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdVersionsSha(ctx, storyID, sha)
	if err != nil {
		return nil, fmt.Errorf("get version %s for story %s: %w", sha, storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get version %s for story %s: %w", sha, storyID, err)
	}
	return json.RawMessage(body), nil
}

// DiffVersions returns the diff between two story versions.
func (c *Client) DiffVersions(ctx context.Context, storyID, fromSha, toSha string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdVersionsFromShaToShaDiff(ctx, storyID, fromSha, toSha)
	if err != nil {
		return nil, fmt.Errorf("diff versions %s..%s for story %s: %w", fromSha, toSha, storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("diff versions %s..%s for story %s: %w", fromSha, toSha, storyID, err)
	}
	return json.RawMessage(body), nil
}

// ListStoriesWithReviewStatus returns the author's stories with review status info.
func (c *Client) ListStoriesWithReviewStatus(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error) {
	resp, err := c.raw.GetStoriesMyReviewStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list stories with review status: %w", err)
	}
	defer resp.Body.Close()

	var result StoriesWithReview
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list stories with review status: %w", err)
	}
	return &result, nil
}
