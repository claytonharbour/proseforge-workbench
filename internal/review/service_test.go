package review

import (
	"context"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

func TestListPending(t *testing.T) {
	reviewID := "rev-1"
	storyID := "story-1"
	reviewStatus := "pending"
	reviews := []gen.HandlersReviewerResponse{
		{Id: &reviewID, StoryId: &storyID, Status: &reviewStatus},
	}
	total := 1

	mock := &api.MockClient{
		ListPendingReviewsFn: func(ctx context.Context, params *gen.GetReviewsPendingParams) (*api.PendingReviews, error) {
			return &api.PendingReviews{Reviews: &reviews, Total: &total}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.ListPending(context.Background(), &gen.GetReviewsPendingParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total == nil || *result.Total != 1 {
		t.Errorf("expected total 1, got %v", result.Total)
	}
	if result.Reviews == nil || len(*result.Reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(*result.Reviews))
	}
}

func TestAccept(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		AcceptReviewFn: func(ctx context.Context, reviewID string) error {
			if reviewID != "rev-1" {
				t.Errorf("expected review ID rev-1, got %s", reviewID)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.Accept(context.Background(), "rev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected AcceptReview to be called")
	}
}

func TestDecline(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		DeclineReviewFn: func(ctx context.Context, reviewID string) error {
			if reviewID != "rev-2" {
				t.Errorf("expected review ID rev-2, got %s", reviewID)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.Decline(context.Background(), "rev-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DeclineReview to be called")
	}
}

func TestApprove(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		ApproveStoryFn: func(ctx context.Context, reviewID string) error {
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.Approve(context.Background(), "rev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected ApproveStory to be called")
	}
}

func TestReject(t *testing.T) {
	var called bool
	feedback := "needs work"
	mock := &api.MockClient{
		RejectStoryFn: func(ctx context.Context, reviewID string, req api.ReviewFeedbackRequest) error {
			if req.Feedback == nil || *req.Feedback != feedback {
				t.Errorf("expected feedback '%s', got %v", feedback, req.Feedback)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.ReviewFeedbackRequest{Feedback: &feedback}
	err := svc.Reject(context.Background(), "rev-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected RejectStory to be called")
	}
}

func TestAddReviewer(t *testing.T) {
	reviewerID := "user-42"
	resultID := "reviewer-99"
	resultStatus := "pending"

	mock := &api.MockClient{
		AddReviewerFn: func(ctx context.Context, storyID string, req api.AddReviewerRequest) (*api.Reviewer, error) {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			if req.ReviewerId == nil || *req.ReviewerId != reviewerID {
				t.Errorf("expected reviewer ID %s, got %v", reviewerID, req.ReviewerId)
			}
			return &api.Reviewer{Id: &resultID, Status: &resultStatus}, nil
		},
	}

	svc := NewService(mock)
	req := api.AddReviewerRequest{ReviewerId: &reviewerID}
	result, err := svc.AddReviewer(context.Background(), "story-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Id == nil || *result.Id != resultID {
		t.Errorf("expected ID %s, got %v", resultID, result.Id)
	}
}
