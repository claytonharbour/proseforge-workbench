package api

import (
	"context"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ListNotifications returns a bounded page of notifications for the authenticated user.
func (c *Client) ListNotifications(ctx context.Context, limit, offset int) (*gen.HandlersNotificationListResponse, error) {
	params := &gen.ListNotificationsParams{Limit: &limit, Offset: &offset}
	resp, err := c.raw.ListNotifications(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersNotificationListResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return &result, nil
}

// GetNotificationUnreadCount returns the current user's unread notification count.
func (c *Client) GetNotificationUnreadCount(ctx context.Context) (*gen.HandlersUnreadCountResponse, error) {
	resp, err := c.raw.GetNotificationsUnreadCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get notification unread count: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersUnreadCountResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get notification unread count: %w", err)
	}
	return &result, nil
}
