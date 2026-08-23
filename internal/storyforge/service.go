// Package storyforge provides Story Forge chat and generation pipeline operations.
package storyforge

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides Story Forge operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a storyforge Service.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the storyforge service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// GetMeta returns the generated outline.
func (s *Service) GetMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("storyforge.GetMeta", "storyID", storyID)
	return s.api.GetStoryMeta(ctx, storyID)
}
