package reviewer

import (
	"context"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

func boolPtr(b bool) *bool { return &b }
func strPtr(s string) *string { return &s }

func TestListAvailable(t *testing.T) {
	reviewers := []api.AvailableReviewer{
		{Id: strPtr("user-1")},
		{Id: strPtr("user-2")},
	}

	mock := &api.MockClient{
		ListAvailableReviewersFn: func(ctx context.Context) (*api.AvailableReviewerList, error) {
			return &api.AvailableReviewerList{Reviewers: &reviewers}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.ListAvailable(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Reviewers == nil {
		t.Fatal("expected reviewers list, got nil")
	}
	if len(*result.Reviewers) != 2 {
		t.Errorf("expected 2 reviewers, got %d", len(*result.Reviewers))
	}
}

func TestRequestReviewer(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		RequestReviewerFn: func(ctx context.Context, req api.CreateReviewerRequestReq) error {
			if req.ReviewerId == nil || *req.ReviewerId != "user-42" {
				t.Errorf("expected reviewer ID user-42, got %v", req.ReviewerId)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.CreateReviewerRequestReq{ReviewerId: strPtr("user-42")}
	err := svc.Request(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected RequestReviewer to be called")
	}
}

func TestRespondToRequestAccept(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		RespondToReviewerRequestFn: func(ctx context.Context, requestID string, req api.RespondToReviewerReq) error {
			if requestID != "req-1" {
				t.Errorf("expected request ID req-1, got %s", requestID)
			}
			if req.Accept == nil || !*req.Accept {
				t.Error("expected accept=true")
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.RespondToReviewerReq{Accept: boolPtr(true)}
	err := svc.Respond(context.Background(), "req-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected RespondToReviewerRequest to be called")
	}
}

func TestRespondToRequestDecline(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		RespondToReviewerRequestFn: func(ctx context.Context, requestID string, req api.RespondToReviewerReq) error {
			if requestID != "req-2" {
				t.Errorf("expected request ID req-2, got %s", requestID)
			}
			if req.Accept == nil || *req.Accept {
				t.Error("expected accept=false")
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.RespondToReviewerReq{Accept: boolPtr(false)}
	err := svc.Respond(context.Background(), "req-2", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected RespondToReviewerRequest to be called")
	}
}
