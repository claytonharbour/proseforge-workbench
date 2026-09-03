// Package author provides author-facing read operations (bookshelf) shared by
// the CLI and the MCP server.
package author

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides author operations. It wraps the concrete *api.Client because
// the author methods live on the client rather than the ProseForgeAPI interface.
type Service struct {
	api    *api.Client
	logger *slog.Logger
}

// NewService creates an author Service.
func NewService(client *api.Client, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the author service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// Bookshelf returns an author's complete bookshelf — stories with titles,
// taglines, slugs, cover URLs, section images, and series membership.
func (s *Service) Bookshelf(ctx context.Context, handle, q, series, status, sort string, limit, offset int) (json.RawMessage, error) {
	s.logger.Info("author.Bookshelf", "handle", handle, "q", q, "series", series, "status", status, "sort", sort)
	return s.api.GetAuthorBookshelf(ctx, handle, q, series, status, sort, limit, offset)
}
