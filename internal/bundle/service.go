// Package bundle provides bundle (multi-story packaging) operations.
//
// Bundles are packaging recipes: a named ordering of stories with interstitial
// transition text, exportable as EPUB/PDF/markdown/JSON. This service is the
// single home for bundle business logic — both the CLI (cmd/cli) and the MCP
// server (cmd/mcp) call it, so request-building lives here once.
package bundle

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides bundle operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a bundle Service.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the bundle service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// List returns the authenticated user's bundles.
func (s *Service) List(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("bundle.List")
	return s.api.ListBundles(ctx)
}

// Get returns a single bundle with its entries.
func (s *Service) Get(ctx context.Context, id string) (json.RawMessage, error) {
	s.logger.Info("bundle.Get", "id", id)
	return s.api.GetBundle(ctx, id)
}

// Create creates a new bundle.
func (s *Service) Create(ctx context.Context, name, intro string) (json.RawMessage, error) {
	s.logger.Info("bundle.Create", "name", name)
	return s.api.CreateBundle(ctx, name, intro)
}

// Update updates a bundle's name and/or intro.
func (s *Service) Update(ctx context.Context, id, name, intro string) error {
	s.logger.Info("bundle.Update", "id", id)
	return s.api.UpdateBundle(ctx, id, name, intro)
}

// Delete deletes a bundle (the referenced stories are not affected).
func (s *Service) Delete(ctx context.Context, id string) error {
	s.logger.Info("bundle.Delete", "id", id)
	return s.api.DeleteBundle(ctx, id)
}

// AddEntry adds a story to a bundle with optional interstitial transition text
// (the "comment" / "Transition" field).
func (s *Service) AddEntry(ctx context.Context, bundleID, storyID, transition string) (json.RawMessage, error) {
	s.logger.Info("bundle.AddEntry", "bundleID", bundleID, "storyID", storyID)
	return s.api.AddBundleEntry(ctx, bundleID, storyID, transition)
}

// UpdateEntry updates an entry's interstitial transition text.
func (s *Service) UpdateEntry(ctx context.Context, bundleID, entryID, transition string) error {
	s.logger.Info("bundle.UpdateEntry", "bundleID", bundleID, "entryID", entryID)
	return s.api.UpdateBundleEntry(ctx, bundleID, entryID, transition)
}

// RemoveEntry removes an entry from a bundle (the story is not affected).
func (s *Service) RemoveEntry(ctx context.Context, bundleID, entryID string) error {
	s.logger.Info("bundle.RemoveEntry", "bundleID", bundleID, "entryID", entryID)
	return s.api.RemoveBundleEntry(ctx, bundleID, entryID)
}

// ReorderEntries sets the display order of entries by entry ID.
func (s *Service) ReorderEntries(ctx context.Context, bundleID string, entryIDs []string) error {
	s.logger.Info("bundle.ReorderEntries", "bundleID", bundleID, "count", len(entryIDs))
	return s.api.ReorderBundleEntries(ctx, bundleID, entryIDs)
}

// Export renders a bundle as epub, pdf, markdown, or json.
func (s *Service) Export(ctx context.Context, id, format string) ([]byte, error) {
	s.logger.Info("bundle.Export", "id", id, "format", format)
	return s.api.ExportBundle(ctx, id, format)
}

// SetCover attaches an image as the bundle cover.
func (s *Service) SetCover(ctx context.Context, bundleID, imageID string) error {
	s.logger.Info("bundle.SetCover", "bundleID", bundleID, "imageID", imageID)
	return s.api.SetBundleCover(ctx, bundleID, imageID)
}
