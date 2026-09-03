// Package listing provides external-store listing operations (Amazon, Apple
// Books, etc.) shared by the CLI and the MCP server.
package listing

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides listing operations. It wraps the concrete *api.Client because
// the listing methods live on the client rather than the ProseForgeAPI interface.
type Service struct {
	api    *api.Client
	logger *slog.Logger
}

// NewService creates a listing Service.
func NewService(client *api.Client, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the listing service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// List returns all listings for a story.
func (s *Service) List(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("listing.List", "storyID", storyID)
	return s.api.ListListings(ctx, storyID)
}

// Create adds a listing to a story.
func (s *Service) Create(ctx context.Context, storyID, store, format, status, url string) (json.RawMessage, error) {
	s.logger.Info("listing.Create", "storyID", storyID, "store", store)
	return s.api.CreateListing(ctx, storyID, store, format, status, url)
}

// Update updates an existing listing (empty fields are left unchanged).
func (s *Service) Update(ctx context.Context, listingID, store, format, status, url string) (json.RawMessage, error) {
	s.logger.Info("listing.Update", "listingID", listingID)
	return s.api.UpdateListing(ctx, listingID, store, format, status, url)
}

// Delete removes a listing.
func (s *Service) Delete(ctx context.Context, listingID string) error {
	s.logger.Info("listing.Delete", "listingID", listingID)
	return s.api.DeleteListing(ctx, listingID)
}
