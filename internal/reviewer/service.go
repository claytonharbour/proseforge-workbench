// Package reviewer provides reviewer pool operations.
package reviewer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides reviewer pool operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a ReviewerService.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the reviewer service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// ListMy returns the authenticated user's accepted reviewers.
func (s *Service) ListMy(ctx context.Context) ([]api.Reviewer, error) {
	return s.api.ListMyReviewers(ctx)
}

// ListAvailable returns users who have opted in as reviewers.
func (s *Service) ListAvailable(ctx context.Context) (json.RawMessage, error) {
	return s.api.ListAvailableReviewers(ctx)
}

// Request sends a reviewer request to another user.
func (s *Service) Request(ctx context.Context, req api.CreateReviewerRequestReq) error {
	return s.api.RequestReviewer(ctx, req)
}

// Respond accepts or declines a reviewer request.
func (s *Service) Respond(ctx context.Context, requestID string, req api.RespondToReviewerReq) error {
	return s.api.RespondToReviewerRequest(ctx, requestID, req)
}
