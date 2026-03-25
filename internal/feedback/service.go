// Package feedback provides feedback review operations including the
// incorporate-all workflow that was previously duplicated in CLI and MCP.
package feedback

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides feedback-related operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a FeedbackService.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the feedback service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// List returns all feedback reviews for a story.
func (s *Service) List(ctx context.Context, storyID string) (*api.FeedbackReviewList, error) {
	return s.api.GetFeedbackReviews(ctx, storyID)
}

// Get returns a single feedback review by ID.
// Pass include values (e.g. "items") to embed related data in the response.
func (s *Service) Get(ctx context.Context, storyID, reviewID string, include ...string) (*api.FeedbackReview, error) {
	return s.api.GetFeedbackReview(ctx, storyID, reviewID, include...)
}

// GetFull returns a feedback review with items included.
func (s *Service) GetFull(ctx context.Context, storyID, reviewID string) (*api.FeedbackReviewWithItems, error) {
	return s.api.GetFeedbackReviewFull(ctx, storyID, reviewID)
}

// GetDiff returns the diff of suggested changes for a feedback review.
func (s *Service) GetDiff(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
	return s.api.GetFeedbackDiff(ctx, storyID, reviewID)
}

// GetSuggestions returns the full feedback response including sections and suggestions.
func (s *Service) GetSuggestions(ctx context.Context, storyID, reviewID string) (*api.FullFeedback, error) {
	return s.api.GetFeedbackSuggestions(ctx, storyID, reviewID)
}

// Create creates a new feedback review for a story.
func (s *Service) Create(ctx context.Context, storyID string, req api.StartAIReviewRequest) (*api.FeedbackReview, error) {
	return s.api.CreateFeedbackReview(ctx, storyID, req)
}

// AddItem adds a feedback item to a review. If sectionId is missing for
// non-replacement types, it auto-assigns to the first section of the story.
func (s *Service) AddItem(ctx context.Context, storyID, reviewID string, req api.AddFeedbackItemRequest) error {
	itemType := ""
	if req.Type != nil {
		itemType = *req.Type
	}

	if (req.SectionId == nil || *req.SectionId == "") && itemType != "replacement" {
		story, err := s.api.GetStory(ctx, storyID)
		if err == nil && story.Sections != nil && len(*story.Sections) > 0 {
			firstID := (*story.Sections)[0].Id
			if firstID != nil {
				req.SectionId = firstID
				s.logger.Debug("feedback.AddItem", "auto_section", *firstID, "reason", "story-level observation")
			}
		}
	}

	s.logger.Info("feedback.AddItem", "storyID", storyID, "reviewID", reviewID, "type", itemType)
	return s.api.AddFeedbackItem(ctx, storyID, reviewID, req)
}

// UpdateSection rewrites a section's content in the feedback branch.
func (s *Service) UpdateSection(ctx context.Context, storyID, reviewID, sectionID, content string) error {
	s.logger.Info("feedback.UpdateSection", "storyID", storyID, "reviewID", reviewID, "sectionID", sectionID)
	return s.api.UpdateSectionContent(ctx, storyID, reviewID, sectionID, content)
}

// Submit submits a review, marking it as ready for the author.
func (s *Service) Submit(ctx context.Context, reviewID string) error {
	s.logger.Info("feedback.Submit", "reviewID", reviewID)
	return s.api.SubmitReview(ctx, reviewID)
}

// IncorporateAll fetches the diff, builds a selections map that accepts every
// changed file, and calls incorporate. Polls for the diff to be ready first,
// since buffer sync to git is asynchronous after feedback_section_update.
func (s *Service) IncorporateAll(ctx context.Context, storyID, reviewID string) error {
	diff, err := s.waitForDiff(ctx, storyID, reviewID)
	if err != nil {
		return err
	}

	selections := make(map[string]bool)
	if diff.Files != nil {
		for _, f := range *diff.Files {
			if f.Path != nil {
				selections[*f.Path] = true
			}
		}
	}

	s.logger.Info("feedback.IncorporateAll", "storyID", storyID, "reviewID", reviewID, "files", len(selections))

	req := api.IncorporateRequest{Selections: &selections}
	if err := s.api.IncorporateFeedback(ctx, storyID, reviewID, req); err != nil {
		return err
	}

	return s.waitForCompletion(ctx, storyID, reviewID)
}

// IncorporateSelective incorporates only the paths specified in selections.
// Polls for the diff to be ready first to ensure buffer sync is complete.
func (s *Service) IncorporateSelective(ctx context.Context, storyID, reviewID string, selections map[string]bool) error {
	if _, err := s.waitForDiff(ctx, storyID, reviewID); err != nil {
		return err
	}

	req := api.IncorporateRequest{Selections: &selections}
	if err := s.api.IncorporateFeedback(ctx, storyID, reviewID, req); err != nil {
		return err
	}

	return s.waitForCompletion(ctx, storyID, reviewID)
}

// waitForDiff polls the feedback diff until files appear or timeout.
// Buffer sync after feedback_section_update is asynchronous — the diff may be
// empty if we check immediately after submit. Returns the diff once ready.
func (s *Service) waitForDiff(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
	timeout := 15 * time.Second
	interval := 1 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		diff, err := s.api.GetFeedbackDiff(ctx, storyID, reviewID)
		if err != nil {
			return nil, fmt.Errorf("polling diff: %w", err)
		}

		fileCount := 0
		if diff.Files != nil {
			fileCount = len(*diff.Files)
		}

		if fileCount > 0 {
			s.logger.Debug("feedback.waitForDiff", "reviewID", reviewID, "files", fileCount)
			return diff, nil
		}

		if time.Now().After(deadline) {
			s.logger.Warn("feedback.waitForDiff", "reviewID", reviewID, "status", "timeout", "files", 0)
			return diff, nil // Return empty diff rather than failing
		}

		s.logger.Debug("feedback.waitForDiff", "reviewID", reviewID, "status", "waiting", "files", 0)

		select {
		case <-ctx.Done():
			s.logger.Warn("feedback.waitForDiff", "reviewID", reviewID, "status", "context_done", "files", 0)
			return diff, nil // Return empty diff rather than failing
		case <-time.After(interval):
		}
	}
}

// waitForCompletion polls the review status until it reaches api.ReviewStatusCompleted or times out.
// Incorporate is async (River job) — the assessment must wait for it to finish.
func (s *Service) waitForCompletion(ctx context.Context, storyID, reviewID string) error {
	timeout := 30 * time.Second
	interval := 2 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		review, err := s.api.GetFeedbackReview(ctx, storyID, reviewID)
		if err != nil {
			return fmt.Errorf("polling review status: %w", err)
		}

		status := ""
		if review.Status != nil {
			status = *review.Status
		}

		s.logger.Debug("feedback.waitForCompletion", "reviewID", reviewID, "status", status)

		if status == api.ReviewStatusCompleted {
			s.logger.Info("feedback.IncorporateAll", "reviewID", reviewID, "status", api.ReviewStatusCompleted)
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}

	s.logger.Warn("feedback.waitForCompletion", "reviewID", reviewID, "status", "timeout")
	return nil // Don't fail on timeout — incorporate may still complete
}
