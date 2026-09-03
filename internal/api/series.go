package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ListSeries returns the authenticated user's series.
func (c *Client) ListSeries(ctx context.Context) (*SeriesList, error) {
	resp, err := c.raw.ListSeries(ctx)
	if err != nil {
		return nil, fmt.Errorf("list series: %w", err)
	}
	defer resp.Body.Close()

	var result SeriesList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list series: %w", err)
	}
	return &result, nil
}

// CreateSeries creates a new series.
func (c *Client) CreateSeries(ctx context.Context, req CreateSeriesReq) (*Series, error) {
	resp, err := c.raw.CreateSeries(ctx, gen.CreateSeriesJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("create series: %w", err)
	}
	defer resp.Body.Close()

	var result Series
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create series: %w", err)
	}
	return &result, nil
}

// GetSeriesByID returns a single series by ID.
func (c *Client) GetSeriesByID(ctx context.Context, id string) (*Series, error) {
	resp, err := c.raw.GetSeries(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get series %s: %w", id, err)
	}
	defer resp.Body.Close()

	var result Series
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get series %s: %w", id, err)
	}
	return &result, nil
}

// UpdateSeries updates a series' metadata.
func (c *Client) UpdateSeries(ctx context.Context, id string, req UpdateSeriesReq) error {
	resp, err := c.raw.UpdateSeries(ctx, id, gen.UpdateSeriesJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("update series %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update series %s: %w", id, err)
	}
	return nil
}

// ArchiveSeries archives (deletes) a series.
func (c *Client) ArchiveSeries(ctx context.Context, id string) error {
	resp, err := c.raw.DeleteSeries(ctx, id)
	if err != nil {
		return fmt.Errorf("archive series %s: %w", id, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("archive series %s: %w", id, err)
	}
	return nil
}

// GetWorld returns the world overview document for a series.
func (c *Client) GetWorld(ctx context.Context, seriesID string) (json.RawMessage, error) {
	resp, err := c.raw.GetSeriesWorld(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("get world %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get world %s: %w", seriesID, err)
	}
	return json.RawMessage(body), nil
}

// UpdateWorld updates the world overview document for a series.
func (c *Client) UpdateWorld(ctx context.Context, seriesID string, content string) error {
	req := gen.UpdateSeriesWorldJSONRequestBody{Content: &content}
	resp, err := c.raw.UpdateSeriesWorld(ctx, seriesID, req)
	if err != nil {
		return fmt.Errorf("update world %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update world %s: %w", seriesID, err)
	}
	return nil
}

// GetTimeline returns the canon timeline for a series.
func (c *Client) GetTimeline(ctx context.Context, seriesID string) (json.RawMessage, error) {
	resp, err := c.raw.ListSeriesTimelines(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("get timeline %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get timeline %s: %w", seriesID, err)
	}
	return json.RawMessage(body), nil
}

// ListTimelineSections returns the list of timeline sections (slugs, titles, sort order).
func (c *Client) ListTimelineSections(ctx context.Context, seriesID string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455): a hand-built path is
	// invisible to `make generate-stale`, so a rename upstream becomes a runtime
	// 404 instead of a compile error.
	resp, err := c.raw.ListSeriesTimelineSections(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("list timeline sections %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list timeline sections %s: %w", seriesID, err)
	}
	return json.RawMessage(body), nil
}

// GetTimelineSection returns a single timeline section by slug.
func (c *Client) GetTimelineSection(ctx context.Context, seriesID, slug string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.GetSeriesTimeline(ctx, seriesID, slug)
	if err != nil {
		return nil, fmt.Errorf("get timeline section %s in series %s: %w", slug, seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get timeline section %s in series %s: %w", slug, seriesID, err)
	}
	return json.RawMessage(body), nil
}

// UpdateTimelineSection updates a single timeline section by slug.
func (c *Client) UpdateTimelineSection(ctx context.Context, seriesID, slug, title, content string) (json.RawMessage, error) {
	// Generated client, not a hand-built URL (#455).
	//
	// ⚠️ The generated body uses *string with omitempty, where the hand-built
	// payload used plain string. ALWAYS take the address — passing nil would
	// OMIT the field, which the previous code never did: it sent "title":""
	// for an empty title. Addressing the value preserves the wire bytes exactly.
	resp, err := c.raw.UpdateSeriesTimelineSection(ctx, seriesID, slug,
		gen.UpdateSeriesTimelineSectionJSONRequestBody{Title: &title, Content: &content})
	if err != nil {
		return nil, fmt.Errorf("update timeline section %s in series %s: %w", slug, seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("update timeline section %s in series %s: %w", slug, seriesID, err)
	}
	return json.RawMessage(body), nil
}

// DeleteTimelineSection removes a timeline section.
func (c *Client) DeleteTimelineSection(ctx context.Context, seriesID, slug string) error {
	// Generated client, not a hand-built URL (#455).
	resp, err := c.raw.DeleteSeriesTimeline(ctx, seriesID, slug)
	if err != nil {
		return fmt.Errorf("delete timeline section %s in series %s: %w", slug, seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete timeline section %s in series %s: %w", slug, seriesID, err)
	}
	return nil
}

// ReorderSeriesStories sets story order in a series.
func (c *Client) ReorderSeriesStories(ctx context.Context, seriesID string, storyIDs []string) error {
	// Generated client, not a hand-built URL (#455).
	//
	// ⚠️ &storyIDs, never nil: the generated field is *[]string with omitempty,
	// so nil would OMIT "storyIds" entirely. A pointer to a nil slice still
	// marshals as "storyIds":null, which is what the hand-built payload sent.
	resp, err := c.raw.UpdateSeriesStoriesOrder(ctx, seriesID,
		gen.UpdateSeriesStoriesOrderJSONRequestBody{StoryIds: &storyIDs})
	if err != nil {
		return fmt.Errorf("reorder stories in series %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("reorder stories in series %s: %w", seriesID, err)
	}
	return nil
}

// ReorderTimelineSections sets timeline section order.
func (c *Client) ReorderTimelineSections(ctx context.Context, seriesID string, slugs []string) error {
	// Generated client, not a hand-built URL (#455). &slugs, never nil — see
	// ReorderSeriesStories above for why.
	resp, err := c.raw.UpdateSeriesTimelineOrder(ctx, seriesID,
		gen.UpdateSeriesTimelineOrderJSONRequestBody{Slugs: &slugs})
	if err != nil {
		return fmt.Errorf("reorder timeline sections in series %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("reorder timeline sections in series %s: %w", seriesID, err)
	}
	return nil
}

// CreateCharacter creates a character in a series.
func (c *Client) CreateCharacter(ctx context.Context, seriesID string, req CreateCharacterReq) (*Character, error) {
	resp, err := c.raw.CreateSeriesCharacter(ctx, seriesID, gen.CreateSeriesCharacterJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("create character in series %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	var result Character
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create character in series %s: %w", seriesID, err)
	}
	return &result, nil
}

// ListCharacters returns all characters in a series.
func (c *Client) ListCharacters(ctx context.Context, seriesID string) (*CharacterList, error) {
	resp, err := c.raw.ListSeriesCharacters(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("list characters in series %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	var result CharacterList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list characters in series %s: %w", seriesID, err)
	}
	return &result, nil
}

// GetCharacter returns a character by slug.
func (c *Client) GetCharacter(ctx context.Context, seriesID, slug string) (*Character, error) {
	resp, err := c.raw.GetSeriesCharacter(ctx, seriesID, slug)
	if err != nil {
		return nil, fmt.Errorf("get character %s in series %s: %w", slug, seriesID, err)
	}
	defer resp.Body.Close()

	var result Character
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get character %s in series %s: %w", slug, seriesID, err)
	}
	return &result, nil
}

// UpdateCharacter updates a character's profile.
func (c *Client) UpdateCharacter(ctx context.Context, seriesID, slug string, req UpdateCharacterReq) error {
	resp, err := c.raw.UpdateSeriesCharacter(ctx, seriesID, slug, gen.UpdateSeriesCharacterJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("update character %s in series %s: %w", slug, seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update character %s in series %s: %w", slug, seriesID, err)
	}
	return nil
}

// DeleteCharacter removes a character from a series.
func (c *Client) DeleteCharacter(ctx context.Context, seriesID, slug string) error {
	resp, err := c.raw.DeleteSeriesCharacter(ctx, seriesID, slug)
	if err != nil {
		return fmt.Errorf("delete character %s in series %s: %w", slug, seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete character %s in series %s: %w", slug, seriesID, err)
	}
	return nil
}

// ListSeriesStories returns stories linked to a series.
func (c *Client) ListSeriesStories(ctx context.Context, seriesID string) (json.RawMessage, error) {
	resp, err := c.raw.ListSeriesStories(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("list stories in series %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list stories in series %s: %w", seriesID, err)
	}
	return json.RawMessage(body), nil
}

// AddStoryToSeries links an existing story to a series.
func (c *Client) AddStoryToSeries(ctx context.Context, seriesID, storyID string) error {
	req := gen.CreateSeriesStoryJSONRequestBody{StoryId: &storyID}
	resp, err := c.raw.CreateSeriesStory(ctx, seriesID, req)
	if err != nil {
		return fmt.Errorf("add story %s to series %s: %w", storyID, seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("add story %s to series %s: %w", storyID, seriesID, err)
	}
	return nil
}

// RemoveStoryFromSeries unlinks a story from a series.
func (c *Client) RemoveStoryFromSeries(ctx context.Context, seriesID, storyID string) error {
	resp, err := c.raw.DeleteSeriesStory(ctx, seriesID, storyID)
	if err != nil {
		return fmt.Errorf("remove story %s from series %s: %w", storyID, seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("remove story %s from series %s: %w", storyID, seriesID, err)
	}
	return nil
}

// PlanStory creates a StorySeed-seeded Story Forge Chat session from series context.
func (c *Client) PlanStory(ctx context.Context, seriesID string, req PlanStoryReq) (*PlanStoryResp, error) {
	resp, err := c.raw.CreateSeriesStoriesPlan(ctx, seriesID, gen.CreateSeriesStoriesPlanJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("plan story in series %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	var result PlanStoryResp
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("plan story in series %s: %w", seriesID, err)
	}
	return &result, nil
}
