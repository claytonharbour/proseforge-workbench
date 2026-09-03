package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

const demoBannerMaxSuppression = 24 * time.Hour

// ValidateDemoBannerUpdate mirrors the backend's bounded suppression rule so
// callers fail before making an admin request that cannot succeed.
func ValidateDemoBannerUpdate(enabled bool, expiresAt *time.Time) error {
	if enabled {
		return nil
	}
	now := time.Now()
	if expiresAt == nil || !expiresAt.After(now) || expiresAt.After(now.Add(demoBannerMaxSuppression)) {
		return fmt.Errorf("a suppression expiry within the next 24 hours is required")
	}
	return nil
}

// GetDemoBanner returns the authenticated admin view of the runtime banner.
func (c *Client) GetDemoBanner(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetAdminDemoBanner(ctx)
	if err != nil {
		return nil, fmt.Errorf("get demo banner: %w", err)
	}
	defer resp.Body.Close()
	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get demo banner: %w", err)
	}
	return json.RawMessage(body), nil
}

// UpdateDemoBanner changes the runtime banner and returns the server's audit
// fields (environment, effective visibility, expiry, and update time).
func (c *Client) UpdateDemoBanner(ctx context.Context, enabled bool, expiresAt *time.Time) (json.RawMessage, error) {
	if err := ValidateDemoBannerUpdate(enabled, expiresAt); err != nil {
		return nil, err
	}
	body := gen.HandlersUpdateDemoBannerRequest{Enabled: &enabled}
	if !enabled && expiresAt != nil {
		expires := expiresAt.UTC().Format(time.RFC3339)
		body.ExpiresAt = &expires
	}
	resp, err := c.raw.UpdateAdminDemoBanner(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("update demo banner: %w", err)
	}
	defer resp.Body.Close()
	result, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("update demo banner: %w", err)
	}
	return json.RawMessage(result), nil
}
