package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ListSections returns a story's sections.
//
// This is the only content-read path a contributor can reach: the story-level
// read 404s for a grantee, so deriving the list from it left a bench able to
// read a section only if it already knew the id, with no way to learn one.
func (c *Client) ListSections(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdSections(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("list sections for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list sections for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// GetSection returns a single section's content and metadata.
func (c *Client) GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdSectionsSectionId(ctx, storyID, sectionID)
	if err != nil {
		return nil, fmt.Errorf("get section %s: %w", sectionID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get section %s: %w", sectionID, err)
	}
	return json.RawMessage(body), nil
}

// CreateSection creates a new section in a story.
func (c *Client) CreateSection(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error) {
	resp, err := c.raw.PostStoryIdSections(ctx, storyID, gen.PostStoryIdSectionsJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("create section in story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("create section in story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// WriteSection updates a section's content (and optionally other fields).
// Returns the API response body containing the updated section, the full
// sections list (with renumbered sortOrders), commit SHA, and word count.
func (c *Client) WriteSection(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) (json.RawMessage, error) {
	resp, err := c.raw.PutStoryIdSectionsSectionId(ctx, storyID, sectionID, gen.PutStoryIdSectionsSectionIdJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("write section %s in story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("write section %s in story %s: %w", sectionID, storyID, err)
	}
	return json.RawMessage(body), nil
}

// ReorderSection moves a section to a 0-indexed position within its story,
// renumbering siblings. Out-of-range values clamp to the nearest end.
// Backed by PATCH /story/{id}/sections/{sectionId}/order. Returns the API
// response body containing the moved section and the full sections list
// (with renumbered sortOrders).
func (c *Client) ReorderSection(ctx context.Context, storyID, sectionID string, order int) (json.RawMessage, error) {
	body := gen.HandlersMoveSectionRequest{Order: &order}
	resp, err := c.raw.PatchStoryIdSectionsSectionIdOrder(ctx, storyID, sectionID, body)
	if err != nil {
		return nil, fmt.Errorf("reorder section %s in story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	respBody, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("reorder section %s in story %s: %w", sectionID, storyID, err)
	}
	return json.RawMessage(respBody), nil
}

// ListGenres returns all available genres.
func (c *Client) ListGenres(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("list genres: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list genres: %w", err)
	}
	return json.RawMessage(body), nil
}
