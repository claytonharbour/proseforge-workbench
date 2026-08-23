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
	resp, err := c.raw.GetStoryIdQuality(ctx, storyID)
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
	params := &gen.PostStoryIdQualityAssessParams{Force: &force}
	resp, err := c.raw.PostStoryIdQualityAssess(ctx, storyID, params)
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

// AssessQualityAtVersion runs a synchronous quality assessment against a specific version SHA.
// Unlike AssessQuality, this returns scores inline (no polling needed).
func (c *Client) AssessQualityAtVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/v1/story/%s/quality/assess/%s?force=true", c.baseURL, storyID, sha)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("assess quality at version %s for story %s: %w", sha, storyID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
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
	resp, err := c.raw.GetStoryIdInsights(ctx, storyID)
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
