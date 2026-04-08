package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetStoryMeta returns the generated outline (story.md, characters.md, plot.md).
func (c *Client) GetStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdMeta(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get story meta %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get story meta %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}
