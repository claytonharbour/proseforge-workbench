package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PublicVersionInfo is the unauthenticated deployment metadata endpoint.
type PublicVersionInfo struct {
	GitCommit     string `json:"gitCommit"`
	BuildTime     string `json:"buildTime"`
	Component     string `json:"component"`
	AllComponents string `json:"allComponents"`
}

// GetPublicVersionInfo returns the backend API build metadata. This endpoint is
// intentionally public so deployment verification does not require admin access.
func (c *Client) GetPublicVersionInfo(ctx context.Context) (*PublicVersionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/version", nil)
	if err != nil {
		return nil, fmt.Errorf("get server version: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get server version: %w", err)
	}
	defer resp.Body.Close()
	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get server version: %w", err)
	}
	var result PublicVersionInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("get server version: %w", err)
	}
	return &result, nil
}
