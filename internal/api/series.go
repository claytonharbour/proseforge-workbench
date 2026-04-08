package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ListSeries returns the authenticated user's series.
func (c *Client) ListSeries(ctx context.Context) (*SeriesList, error) {
	resp, err := c.raw.GetSeries(ctx)
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
	resp, err := c.raw.PostSeries(ctx, gen.PostSeriesJSONRequestBody(req))
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
	resp, err := c.raw.GetSeriesId(ctx, id)
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
	resp, err := c.raw.PutSeriesId(ctx, id, gen.PutSeriesIdJSONRequestBody(req))
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
	resp, err := c.raw.DeleteSeriesId(ctx, id)
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
	resp, err := c.raw.GetSeriesIdWorld(ctx, seriesID)
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
	req := gen.PutSeriesIdWorldJSONRequestBody{Content: &content}
	resp, err := c.raw.PutSeriesIdWorld(ctx, seriesID, req)
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
	resp, err := c.raw.GetSeriesIdTimeline(ctx, seriesID)
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
	url := fmt.Sprintf("%s/api/v1/series/%s/timeline/sections", c.baseURL, seriesID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("list timeline sections %s: %w", seriesID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
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
	url := fmt.Sprintf("%s/api/v1/series/%s/timeline/%s", c.baseURL, seriesID, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get timeline section %s in series %s: %w", slug, seriesID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
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
	payload := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{Title: title, Content: content}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal timeline section update: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/series/%s/timeline/%s", c.baseURL, seriesID, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("update timeline section %s in series %s: %w", slug, seriesID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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
	url := fmt.Sprintf("%s/api/v1/series/%s/timeline/%s", c.baseURL, seriesID, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("delete timeline section %s in series %s: %w", slug, seriesID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
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
	payload := struct {
		StoryIds []string `json:"storyIds"`
	}{StoryIds: storyIDs}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal reorder stories: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/series/%s/stories/order", c.baseURL, seriesID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("reorder stories in series %s: %w", seriesID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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
	payload := struct {
		Slugs []string `json:"slugs"`
	}{Slugs: slugs}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal reorder timeline sections: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/series/%s/timeline/order", c.baseURL, seriesID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("reorder timeline sections in series %s: %w", seriesID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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
	resp, err := c.raw.PostSeriesIdCharacters(ctx, seriesID, gen.PostSeriesIdCharactersJSONRequestBody(req))
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
	resp, err := c.raw.GetSeriesIdCharacters(ctx, seriesID)
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
	resp, err := c.raw.GetSeriesIdCharactersSlug(ctx, seriesID, slug)
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
	resp, err := c.raw.PutSeriesIdCharactersSlug(ctx, seriesID, slug, gen.PutSeriesIdCharactersSlugJSONRequestBody(req))
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
	resp, err := c.raw.DeleteSeriesIdCharactersSlug(ctx, seriesID, slug)
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
	resp, err := c.raw.GetSeriesIdStories(ctx, seriesID)
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
	req := gen.PostSeriesIdStoriesJSONRequestBody{StoryId: &storyID}
	resp, err := c.raw.PostSeriesIdStories(ctx, seriesID, req)
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
	resp, err := c.raw.DeleteSeriesIdStoriesStoryId(ctx, seriesID, storyID)
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
	resp, err := c.raw.PostSeriesIdStoriesPlan(ctx, seriesID, gen.PostSeriesIdStoriesPlanJSONRequestBody(req))
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

