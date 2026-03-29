package api

import (
	"context"
	"encoding/json"
	"fmt"

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

// UpdateTimeline updates the canon timeline for a series.
func (c *Client) UpdateTimeline(ctx context.Context, seriesID string, content string) error {
	req := gen.PutSeriesIdTimelineJSONRequestBody{Content: &content}
	resp, err := c.raw.PutSeriesIdTimeline(ctx, seriesID, req)
	if err != nil {
		return fmt.Errorf("update timeline %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update timeline %s: %w", seriesID, err)
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

// CreateSeriesChat starts a new world-building chat session for a series.
func (c *Client) CreateSeriesChat(ctx context.Context, seriesID string) (*SeriesChatSession, error) {
	resp, err := c.raw.PostSeriesIdChatSessions(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("create series chat for %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	var result SeriesChatSession
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create series chat for %s: %w", seriesID, err)
	}
	return &result, nil
}

// ListSeriesChats returns chat sessions for a series.
func (c *Client) ListSeriesChats(ctx context.Context, seriesID string) (json.RawMessage, error) {
	resp, err := c.raw.GetSeriesIdChatSessions(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("list series chats for %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list series chats for %s: %w", seriesID, err)
	}
	return json.RawMessage(body), nil
}

// GetSeriesChat returns a chat session with its messages.
func (c *Client) GetSeriesChat(ctx context.Context, seriesID, sessionID string) (*SeriesChatSession, error) {
	resp, err := c.raw.GetSeriesIdChatSessionsSessionId(ctx, seriesID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get series chat %s: %w", sessionID, err)
	}
	defer resp.Body.Close()

	var result SeriesChatSession
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get series chat %s: %w", sessionID, err)
	}
	return &result, nil
}

// SendSeriesChatMessage sends a message and gets the AI response.
func (c *Client) SendSeriesChatMessage(ctx context.Context, seriesID, sessionID string, req SeriesChatSendReq) (*SeriesChatSendResp, error) {
	resp, err := c.raw.PostSeriesIdChatSessionsSessionIdMessages(ctx, seriesID, sessionID, gen.PostSeriesIdChatSessionsSessionIdMessagesJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("send series chat message in %s: %w", sessionID, err)
	}
	defer resp.Body.Close()

	var result SeriesChatSendResp
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("send series chat message in %s: %w", sessionID, err)
	}
	return &result, nil
}

// FinalizeSeriesChat finalizes a chat session.
func (c *Client) FinalizeSeriesChat(ctx context.Context, seriesID, sessionID string) (*SeriesChatFinalizeResp, error) {
	resp, err := c.raw.PostSeriesIdChatSessionsSessionIdFinalize(ctx, seriesID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("finalize series chat %s: %w", sessionID, err)
	}
	defer resp.Body.Close()

	var result SeriesChatFinalizeResp
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("finalize series chat %s: %w", sessionID, err)
	}
	return &result, nil
}

// HarvestSeriesChat extracts metadata from a chat session to git.
func (c *Client) HarvestSeriesChat(ctx context.Context, seriesID, sessionID string) (json.RawMessage, error) {
	resp, err := c.raw.PostSeriesIdChatSessionsSessionIdHarvest(ctx, seriesID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("harvest series chat %s: %w", sessionID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("harvest series chat %s: %w", sessionID, err)
	}
	return json.RawMessage(body), nil
}

// HarvestAllSeriesChats harvests all chat sessions for a series.
func (c *Client) HarvestAllSeriesChats(ctx context.Context, seriesID string) (json.RawMessage, error) {
	resp, err := c.raw.PostSeriesIdChatHarvest(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("harvest all series chats for %s: %w", seriesID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("harvest all series chats for %s: %w", seriesID, err)
	}
	return json.RawMessage(body), nil
}
