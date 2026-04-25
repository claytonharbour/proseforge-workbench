package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

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
func (c *Client) WriteSection(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) error {
	resp, err := c.raw.PutStoryIdSectionsSectionId(ctx, storyID, sectionID, gen.PutStoryIdSectionsSectionIdJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("write section %s in story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("write section %s in story %s: %w", sectionID, storyID, err)
	}
	return nil
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
