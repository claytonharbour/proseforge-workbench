// Package review orchestrates the story review workflow.
package review

import (
	"context"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Service provides review-related operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a ReviewService.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the review service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// ListPending returns reviews pending for the authenticated reviewer.
func (s *Service) ListPending(ctx context.Context, params *gen.GetReviewsPendingParams) (*api.PendingReviews, error) {
	return s.api.ListPendingReviews(ctx, params)
}

// AddReviewer adds a reviewer to a story (called by the author).
func (s *Service) AddReviewer(ctx context.Context, storyID string, req api.AddReviewerRequest) (*api.Reviewer, error) {
	return s.api.AddReviewer(ctx, storyID, req)
}

// Accept accepts a review assignment (called by the reviewer).
func (s *Service) Accept(ctx context.Context, reviewID string) error {
	s.logger.Info("review.Accept", "reviewID", reviewID)
	return s.api.AcceptReview(ctx, reviewID)
}

// Decline declines a review assignment (called by the reviewer).
func (s *Service) Decline(ctx context.Context, reviewID string) error {
	s.logger.Info("review.Decline", "reviewID", reviewID)
	return s.api.DeclineReview(ctx, reviewID)
}

// Approve approves a story after review (called by the reviewer).
func (s *Service) Approve(ctx context.Context, reviewID string) error {
	s.logger.Info("review.Approve", "reviewID", reviewID)
	return s.api.ApproveStory(ctx, reviewID)
}

// Reject rejects a story after review with optional feedback (called by the reviewer).
func (s *Service) Reject(ctx context.Context, reviewID string, req api.ReviewFeedbackRequest) error {
	s.logger.Info("review.Reject", "reviewID", reviewID)
	return s.api.RejectStory(ctx, reviewID, req)
}
