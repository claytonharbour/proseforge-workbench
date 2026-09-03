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
	resp, err := c.raw.CreateStory(ctx, gen.CreateStoryJSONRequestBody(req))
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
	resp, err := c.raw.UpdateStory(ctx, id, gen.UpdateStoryJSONRequestBody(req))
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
	body := gen.PublishStoryJSONRequestBody{}
	if visibility != "" {
		body["visibility"] = visibility
	}
	resp, err := c.raw.PublishStory(ctx, id, body)
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
	resp, err := c.raw.UnpublishStory(ctx, id)
	if err != nil {
		return fmt.Errorf("unpublish story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("unpublish story %s: %w", id, err)
	}
	return nil
}

// CreatePitch creates a new story in pitch status (pre-writing idea).
// Takes the same body as CreateStory but uses the /stories/pitch endpoint.
func (c *Client) CreatePitch(ctx context.Context, req CreateStoryRequest) (*Story, error) {
	// Generated client, not a hand-built URL (#455).
	//
	// No wire-format hazard here, unlike the other body-carrying calls:
	// CreateStoryRequest is an alias for gen.HandlersCreateStoryRequest, which
	// IS CreateStoriesPitchJSONRequestBody. The caller already owns the
	// pointer-fielded struct, so the bytes are unchanged by construction.
	resp, err := c.raw.CreateStoriesPitch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create pitch: %w", err)
	}
	defer resp.Body.Close()

	var result Story
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create pitch: %w", err)
	}
	return &result, nil
}

// PromoteStory promotes a pitch to draft status.
func (c *Client) PromoteStory(ctx context.Context, id string) error {
	// Uses the GENERATED client rather than a hand-built URL (#455). A hand-built
	// path is invisible to `make generate-stale`, so an upstream rename or
	// deletion produces a runtime 404 instead of a compile error — which is how
	// `pfw author bookshelf` broke on 2026-08-29 with the guard still green.
	resp, err := c.raw.PromoteStory(ctx, id)
	if err != nil {
		return fmt.Errorf("promote story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("promote story %s: %w", id, err)
	}
	return nil
}

// UpsertStoryMeta writes story planning data (creates if missing, updates if present).
// metaType is one of: "story", "characters", "plot".
func (c *Client) UpsertStoryMeta(ctx context.Context, storyID, metaType, content string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	//
	// ⚠️ The generated body uses *string with omitempty, where the hand-built
	// payload used plain string. ALWAYS take the address — passing nil would
	// OMIT the field, which the previous code never did: it sent "content":""
	// for empty content. Addressing the value preserves the wire bytes exactly.
	//
	// metaType stays a plain string on this signature: the generated enum type
	// is a string underneath and the client never validates it, so converting
	// keeps an unknown value reaching the server exactly as it did before.
	resp, err := c.raw.UpdateStoryMeta(ctx, storyID,
		gen.UpdateStoryMetaParamsMetaType(metaType),
		gen.UpdateStoryMetaJSONRequestBody{Content: &content})
	if err != nil {
		return nil, fmt.Errorf("upsert meta %s for story %s: %w", metaType, storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("upsert meta %s for story %s: %w", metaType, storyID, err)
	}
	return json.RawMessage(body), nil
}

// UpdateVisibility changes the visibility of a published story.
// visibility must be "public" or "members".
func (c *Client) UpdateVisibility(ctx context.Context, id string, visibility string) error {
	body := gen.UpdateStoryVisibilityJSONRequestBody{"visibility": visibility}
	resp, err := c.raw.UpdateStoryVisibility(ctx, id, body)
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
func (c *Client) ListStories(ctx context.Context, params *gen.ListStoriesParams) (*StoryList, error) {
	resp, err := c.raw.ListStories(ctx, params, excludeContent)
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
	params := &gen.GetStoryParams{Content: &contentFalse}
	resp, err := c.raw.GetStory(ctx, id, params)
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
	resp, err := c.raw.GetResolveAuthorStory(ctx, handle, slug)
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
	params := &gen.GetStoryParams{Content: &contentTrue}
	resp, err := c.raw.GetStory(ctx, id, params)
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
	resp, err := c.raw.GetStoryDownload(ctx, id, gen.GetStoryDownloadParamsFormat(format))
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
func (c *Client) ListVersions(ctx context.Context, storyID string, params *gen.ListStoryVersionsParams) (json.RawMessage, error) {
	resp, err := c.raw.ListStoryVersions(ctx, storyID, params)
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
	resp, err := c.raw.GetStoryVersion(ctx, storyID, sha)
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
	// Route reshaped upstream: .../versions/{from}/{to}/diff became
	// .../versions/{from}/diff/{to}. Same arguments in the same order, so the
	// only signal was the generated method vanishing — regenerating turned a
	// would-be 404 into this compile error.
	resp, err := c.raw.GetStoryVersionsDiff(ctx, storyID, fromSha, toSha)
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

// DeleteStory permanently deletes a story.
func (c *Client) DeleteStory(ctx context.Context, id string) error {
	// Generated client, not a hand-built URL (#455). No body, so no
	// omitempty hazard.
	resp, err := c.raw.DeleteStory(ctx, id)
	if err != nil {
		return fmt.Errorf("delete story %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete story %s: %w", id, err)
	}
	return nil
}

// DeleteSection deletes a section from a story.
func (c *Client) DeleteSection(ctx context.Context, storyID, sectionID string) error {
	// Generated client, not a hand-built URL (#455). No body, so no
	// omitempty hazard.
	resp, err := c.raw.DeleteStorySection(ctx, storyID, sectionID)
	if err != nil {
		return fmt.Errorf("delete section %s in story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete section %s in story %s: %w", sectionID, storyID, err)
	}
	return nil
}

// RestoreVersion restores a story to a previous version.
func (c *Client) RestoreVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	//
	// nil params is deliberate and safe: the generated request builder guards
	// with `if params != nil`, so this produces no query string — byte-identical
	// to the hand-built URL. The optional sectionId filter is not exposed on
	// this signature and was never sent.
	resp, err := c.raw.RestoreStoryVersion(ctx, storyID, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("restore version %s for story %s: %w", sha, storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("restore version %s for story %s: %w", sha, storyID, err)
	}
	return json.RawMessage(body), nil
}

// GetMetaStale returns sections affected by meta changes.
func (c *Client) GetMetaStale(ctx context.Context, storyID string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.GetStoryMetaStale(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get meta stale for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get meta stale for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// AcknowledgeMetaStale dismisses all meta staleness warnings.
func (c *Client) AcknowledgeMetaStale(ctx context.Context, storyID string) error {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.AcknowledgeStoryMeta(ctx, storyID)
	if err != nil {
		return fmt.Errorf("acknowledge meta stale for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("acknowledge meta stale for story %s: %w", storyID, err)
	}
	return nil
}

// RegenerateTagline queues AI tagline regeneration.
func (c *Client) RegenerateTagline(ctx context.Context, storyID string) error {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.CreateStoryRegenerateTagline(ctx, storyID)
	if err != nil {
		return fmt.Errorf("regenerate tagline for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("regenerate tagline for story %s: %w", storyID, err)
	}
	return nil
}

// RegenerateTitle queues AI title regeneration.
func (c *Client) RegenerateTitle(ctx context.Context, storyID string) error {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.CreateStoryRegenerateTitle(ctx, storyID)
	if err != nil {
		return fmt.Errorf("regenerate title for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("regenerate title for story %s: %w", storyID, err)
	}
	return nil
}

// RegenerateStaleNarration auto-detects and regenerates stale narration chapters.
func (c *Client) RegenerateStaleNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.RegenerateStoryNarration(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("regenerate stale narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("regenerate stale narration for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// AcknowledgeNarrationStale dismisses narration staleness in bulk.
func (c *Client) AcknowledgeNarrationStale(ctx context.Context, storyID string) error {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.AcknowledgeStoryNarration(ctx, storyID)
	if err != nil {
		return fmt.Errorf("acknowledge narration stale for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("acknowledge narration stale for story %s: %w", storyID, err)
	}
	return nil
}

// ListStoriesWithReviewStatus returns the author's stories with review status info.
func (c *Client) ListStoriesWithReviewStatus(ctx context.Context, params *gen.ListMyStoriesReviewStatusParams) (*StoriesWithReview, error) {
	resp, err := c.raw.ListMyStoriesReviewStatus(ctx, params)
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
