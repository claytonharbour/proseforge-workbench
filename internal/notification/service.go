// Package notification provides read-only notification operations.
package notification

import (
	"context"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

type Service struct {
	api    *api.Client
	logger *slog.Logger
}

type Option func(*Service)

func WithLogger(logger *slog.Logger) Option { return func(s *Service) { s.logger = logger } }

func NewService(client *api.Client, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) List(ctx context.Context, limit, offset int) (*gen.HandlersNotificationListResponse, error) {
	s.logger.Info("notification.List", "limit", limit, "offset", offset)
	return s.api.ListNotifications(ctx, limit, offset)
}

func (s *Service) UnreadCount(ctx context.Context) (*gen.HandlersUnreadCountResponse, error) {
	s.logger.Info("notification.UnreadCount")
	return s.api.GetNotificationUnreadCount(ctx)
}
