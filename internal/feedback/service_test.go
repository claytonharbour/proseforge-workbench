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
	status := api.ReviewStatusCompleted
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

func TestFeedbackList(t *testing.T) {
	reviewID := "rev-1"
	status := api.ReviewStatusCompleted
	reviews := []api.FeedbackReview{
		{Id: &reviewID, Status: &status},
	}
	total := 1

	mock := &api.MockClient{
		GetFeedbackReviewsFn: func(ctx context.Context, storyID string) (*api.FeedbackReviewList, error) {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			return &api.FeedbackReviewList{Reviews: &reviews, Total: &total}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.List(context.Background(), "story-1")
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

func TestFeedbackGet(t *testing.T) {
	reviewID := "rev-1"
	status := api.ReviewStatusCompleted
	review := &api.FeedbackReview{Id: &reviewID, Status: &status}
	items := &api.FeedbackReviewItemsData{SessionID: "sess-1", TotalSuggestions: 3}

	mock := &api.MockClient{
		GetFeedbackReviewFullFn: func(ctx context.Context, storyID, rID string) (*api.FeedbackReviewWithItems, error) {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			if rID != "rev-1" {
				t.Errorf("expected review ID rev-1, got %s", rID)
			}
			return &api.FeedbackReviewWithItems{Review: review, Items: items}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.GetFull(context.Background(), "story-1", "rev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Review == nil || result.Review.Id == nil || *result.Review.Id != "rev-1" {
		t.Errorf("expected review ID rev-1, got %v", result.Review)
	}
	if result.Items == nil || result.Items.TotalSuggestions != 3 {
		t.Errorf("expected 3 total suggestions, got %v", result.Items)
	}
}

func TestFeedbackAddItem(t *testing.T) {
	var called bool
	itemType := "suggestion"
	text := "Consider rephrasing this paragraph"
	sectionID := "sec-1"

	mock := &api.MockClient{
		AddFeedbackItemFn: func(ctx context.Context, storyID, reviewID string, req api.AddFeedbackItemRequest) error {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			if reviewID != "rev-1" {
				t.Errorf("expected review ID rev-1, got %s", reviewID)
			}
			if req.Type == nil || *req.Type != itemType {
				t.Errorf("expected type %s, got %v", itemType, req.Type)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.AddFeedbackItemRequest{
		Type:      &itemType,
		Text:      &text,
		SectionId: &sectionID,
	}
	err := svc.AddItem(context.Background(), "story-1", "rev-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected AddFeedbackItem to be called")
	}
}

func TestFeedbackSubmit(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		SubmitReviewFn: func(ctx context.Context, reviewID string) error {
			if reviewID != "rev-1" {
				t.Errorf("expected review ID rev-1, got %s", reviewID)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.Submit(context.Background(), "rev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected SubmitReview to be called")
	}
}

func TestFeedbackCreate(t *testing.T) {
	reviewID := "rev-new"
	status := api.ReviewStatusRunning

	mock := &api.MockClient{
		CreateFeedbackReviewFn: func(ctx context.Context, storyID string, req api.StartAIReviewRequest) (*api.FeedbackReview, error) {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			return &api.FeedbackReview{Id: &reviewID, Status: &status}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.Create(context.Background(), "story-1", api.StartAIReviewRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Id == nil || *result.Id != reviewID {
		t.Errorf("expected review ID %s, got %v", reviewID, result.Id)
	}
	if result.Status == nil || *result.Status != api.ReviewStatusRunning {
		t.Errorf("expected status running, got %v", result.Status)
	}
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
