package api

import (
	"context"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// AddReviewer adds a reviewer to a story (called by the author).
func (c *Client) AddReviewer(ctx context.Context, storyID string, req AddReviewerRequest) (*Reviewer, error) {
	resp, err := c.raw.PostStoryIdReviewers(ctx, storyID, gen.PostStoryIdReviewersJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("adding reviewer to story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	var result Reviewer
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("adding reviewer to story %s: %w", storyID, err)
	}
	return &result, nil
}

// ListReviewers returns all reviewers for a story.
func (c *Client) ListReviewers(ctx context.Context, storyID string) (*ReviewersList, error) {
	resp, err := c.raw.GetStoryIdReviewers(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("listing reviewers for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	var result ReviewersList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("listing reviewers for story %s: %w", storyID, err)
	}
	return &result, nil
}

// AcceptReview accepts a review assignment (called by the reviewer).
func (c *Client) AcceptReview(ctx context.Context, reviewID string) error {
	resp, err := c.raw.PostReviewIdAccept(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("accepting review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("accepting review %s: %w", reviewID, err)
	}
	return nil
}

// DeclineReview declines a review assignment (called by the reviewer).
func (c *Client) DeclineReview(ctx context.Context, reviewID string) error {
	resp, err := c.raw.PostReviewIdDecline(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("declining review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("declining review %s: %w", reviewID, err)
	}
	return nil
}

// ApproveStory approves a story after review (called by the reviewer).
func (c *Client) ApproveStory(ctx context.Context, reviewID string) error {
	resp, err := c.raw.PostReviewIdApprove(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("approving review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("approving review %s: %w", reviewID, err)
	}
	return nil
}

// RejectStory rejects a story after review (called by the reviewer).
func (c *Client) RejectStory(ctx context.Context, reviewID string, req ReviewFeedbackRequest) error {
	resp, err := c.raw.PostReviewIdReject(ctx, reviewID, gen.PostReviewIdRejectJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("rejecting review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("rejecting review %s: %w", reviewID, err)
	}
	return nil
}

// ListPendingReviews returns reviews pending for the authenticated reviewer.
func (c *Client) ListPendingReviews(ctx context.Context, params *gen.GetReviewsPendingParams) (*PendingReviews, error) {
	resp, err := c.raw.GetReviewsPending(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("listing pending reviews: %w", err)
	}
	defer resp.Body.Close()

	var result PendingReviews
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("listing pending reviews: %w", err)
	}
	return &result, nil
}
