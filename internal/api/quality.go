package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// GetQuality returns the code-based quality assessment for a story.
func (c *Client) GetQuality(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryQuality(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get quality for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get quality for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// AssessQuality triggers a code-based quality assessment for a story.
func (c *Client) AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error) {
	params := &gen.AssessStoryQualityParams{Force: &force}
	resp, err := c.raw.AssessStoryQuality(ctx, storyID, params)
	if err != nil {
		return nil, fmt.Errorf("assess quality for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("assess quality for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// forceAssess appends ?force=true. The quality assess-at-version route accepts
// it but does not declare it in the spec, so no generated params struct carries
// it. See AssessQualityAtVersion and forge/proseforge#1310 (undocumented surface
// is invisible to every spec-derived guard).
func forceAssess(_ context.Context, req *http.Request) error {
	q := req.URL.Query()
	q.Set("force", "true")
	req.URL.RawQuery = q.Encode()
	return nil
}

// AssessQualityAtVersion runs a synchronous quality assessment against a specific version SHA.
// Unlike AssessQuality, this returns scores inline (no polling needed).
func (c *Client) AssessQualityAtVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	//
	// ⚠️ force=true is applied by a request editor rather than a params struct,
	// because the SPEC DOES NOT DOCUMENT IT ON THIS ROUTE. Its sibling
	// /story/{id}/quality/assess declares `force`; the /{sha} form declares only
	// id and sha, so the generated CreateStoryQualityAssess takes no params at
	// all. Migrating without this editor would silently drop force=true and
	// return CACHED scores instead of re-assessing — the same class of quiet
	// wire-format change as the omitempty trap, wearing different clothes.
	resp, err := c.raw.CreateStoryQualityAssess(ctx, storyID, sha, forceAssess)
	if err != nil {
		return nil, fmt.Errorf("assess quality at version %s for story %s: %w", sha, storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("assess quality at version %s for story %s: %w", sha, storyID, err)
	}
	return json.RawMessage(body), nil
}

// GetInsights returns combined quality and AI analysis information for a story.
func (c *Client) GetInsights(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.ListStoryInsights(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get insights for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get insights for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}
