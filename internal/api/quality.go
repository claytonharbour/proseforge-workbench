package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// GetQuality returns the code-based quality assessment for a story.
func (c *Client) GetQuality(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdQuality(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("getting quality for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("getting quality for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// AssessQuality triggers a code-based quality assessment for a story.
func (c *Client) AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error) {
	params := &gen.PostStoryIdQualityAssessParams{Force: &force}
	resp, err := c.raw.PostStoryIdQualityAssess(ctx, storyID, params)
	if err != nil {
		return nil, fmt.Errorf("assessing quality for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("assessing quality for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// GetInsights returns combined quality and AI analysis information for a story.
func (c *Client) GetInsights(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdInsights(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("getting insights for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("getting insights for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}
