package feedback

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

func strPtr(s string) *string { return &s }

func completedReviewFn() func(ctx context.Context, storyID, reviewID string, include ...string) (*api.FeedbackReview, error) {
	status := "completed"
	return func(ctx context.Context, storyID, reviewID string, include ...string) (*api.FeedbackReview, error) {
		return &api.FeedbackReview{Status: &status}, nil
	}
}

func makeDiffResponse(paths ...string) *api.ReviewDiffResponse {
	files := make([]api.ReviewDiffFile, len(paths))
	for i, p := range paths {
		files[i] = api.ReviewDiffFile{Path: strPtr(p)}
	}
	return &api.ReviewDiffResponse{Files: &files}
}

func TestIncorporateAll_HappyPath(t *testing.T) {
	var capturedReq api.IncorporateRequest
	mock := &api.MockClient{
		GetFeedbackDiffFn: func(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
			return makeDiffResponse("section-1.md", "section-2.md"), nil
		},
		IncorporateFeedbackFn: func(ctx context.Context, storyID, reviewID string, req api.IncorporateRequest) error {
			capturedReq = req
			return nil
		},
		GetFeedbackReviewFn: completedReviewFn(),
	}

	svc := NewService(mock)
	err := svc.IncorporateAll(context.Background(), "story-1", "review-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Selections == nil {
		t.Fatal("expected selections to be set")
	}
	sel := *capturedReq.Selections
	if len(sel) != 2 {
		t.Fatalf("expected 2 selections, got %d", len(sel))
	}
	if !sel["section-1.md"] {
		t.Error("expected section-1.md to be true")
	}
	if !sel["section-2.md"] {
		t.Error("expected section-2.md to be true")
	}
}

func TestIncorporateAll_EmptyDiff(t *testing.T) {
	// waitForDiff polls until timeout when diff is always empty.
	// Use a context with short deadline to avoid waiting 15s.
	var incorporateCalled bool
	mock := &api.MockClient{
		GetFeedbackDiffFn: func(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
			return makeDiffResponse(), nil
		},
		GetFeedbackReviewFn: completedReviewFn(),
		IncorporateFeedbackFn: func(ctx context.Context, storyID, reviewID string, req api.IncorporateRequest) error {
			incorporateCalled = true
			if req.Selections == nil {
				t.Error("expected selections to be non-nil")
			} else if len(*req.Selections) != 0 {
				t.Errorf("expected 0 selections, got %d", len(*req.Selections))
			}
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	svc := NewService(mock)
	err := svc.IncorporateAll(ctx, "story-1", "review-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !incorporateCalled {
		t.Error("expected IncorporateFeedback to be called even with empty diff")
	}
}

func TestIncorporateAll_DiffError(t *testing.T) {
	mock := &api.MockClient{
		GetFeedbackDiffFn: func(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
			return nil, fmt.Errorf("network error")
		},
	}

	svc := NewService(mock)
	err := svc.IncorporateAll(context.Background(), "story-1", "review-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "polling diff: network error" {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestIncorporateAll_IncorporateError(t *testing.T) {
	mock := &api.MockClient{
		GetFeedbackDiffFn: func(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
			return makeDiffResponse("a.md"), nil
		},
		GetFeedbackReviewFn: completedReviewFn(),
		IncorporateFeedbackFn: func(ctx context.Context, storyID, reviewID string, req api.IncorporateRequest) error {
			return fmt.Errorf("server error")
		},
	}

	svc := NewService(mock)
	err := svc.IncorporateAll(context.Background(), "story-1", "review-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "server error" {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestIncorporateSelective(t *testing.T) {
	var capturedReq api.IncorporateRequest
	mock := &api.MockClient{
		GetFeedbackDiffFn: func(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
			return makeDiffResponse("a.md", "b.md"), nil
		},
		IncorporateFeedbackFn: func(ctx context.Context, storyID, reviewID string, req api.IncorporateRequest) error {
			capturedReq = req
			return nil
		},
		GetFeedbackReviewFn: completedReviewFn(),
	}

	svc := NewService(mock)
	selections := map[string]bool{"a.md": true, "b.md": false}
	err := svc.IncorporateSelective(context.Background(), "story-1", "review-1", selections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Selections == nil {
		t.Fatal("expected selections to be set")
	}
	sel := *capturedReq.Selections
	if sel["a.md"] != true {
		t.Error("expected a.md to be true")
	}
	if sel["b.md"] != false {
		t.Error("expected b.md to be false")
	}
}

func TestIncorporateAll_DiffPolling(t *testing.T) {
	// Simulate buffer sync delay: diff empty on first call, ready on second.
	calls := 0
	mock := &api.MockClient{
		GetFeedbackDiffFn: func(ctx context.Context, storyID, reviewID string) (*api.ReviewDiffResponse, error) {
			calls++
			if calls < 2 {
				return makeDiffResponse(), nil // empty
			}
			return makeDiffResponse("section-1.md"), nil // ready
		},
		IncorporateFeedbackFn: func(ctx context.Context, storyID, reviewID string, req api.IncorporateRequest) error {
			return nil
		},
		GetFeedbackReviewFn: completedReviewFn(),
	}

	svc := NewService(mock)
	err := svc.IncorporateAll(context.Background(), "story-1", "review-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 diff polls, got %d", calls)
	}
}
