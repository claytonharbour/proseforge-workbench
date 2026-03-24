package story

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

func TestList(t *testing.T) {
	title := "Test Story"
	stories := []gen.HandlersStoryResponse{{Title: &title}}
	total := 1
	mock := &api.MockClient{
		ListStoriesFn: func(ctx context.Context, params *gen.GetStoriesParams) (*api.StoryList, error) {
			return &api.StoryList{Stories: &stories, Total: &total}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.List(context.Background(), &gen.GetStoriesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total == nil || *result.Total != 1 {
		t.Errorf("expected total 1, got %v", result.Total)
	}
	if result.Stories == nil || len(*result.Stories) != 1 {
		t.Fatalf("expected 1 story, got %d", len(*result.Stories))
	}
	if got := *(*result.Stories)[0].Title; got != "Test Story" {
		t.Errorf("expected title 'Test Story', got '%s'", got)
	}
}

func TestGet(t *testing.T) {
	id := "abc-123"
	title := "My Story"
	mock := &api.MockClient{
		GetStoryFn: func(ctx context.Context, storyID string) (*api.Story, error) {
			if storyID != id {
				t.Errorf("expected story ID %s, got %s", id, storyID)
			}
			return &api.Story{Id: &id, Title: &title}, nil
		},
	}

	svc := NewService(mock)
	result, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Id == nil || *result.Id != id {
		t.Errorf("expected ID %s, got %v", id, result.Id)
	}
	if result.Title == nil || *result.Title != title {
		t.Errorf("expected title %s, got %v", title, result.Title)
	}
}

func TestExport(t *testing.T) {
	mock := &api.MockClient{
		DownloadStoryFn: func(ctx context.Context, id string, format string) (string, error) {
			if format != "markdown" {
				t.Errorf("expected format markdown, got %s", format)
			}
			return "# My Story\n\nContent here", nil
		},
	}

	svc := NewService(mock)
	content, err := svc.Export(context.Background(), "story-1", "markdown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "# My Story\n\nContent here" {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestGetSection(t *testing.T) {
	expected := json.RawMessage(`{"name":"Chapter 1","content":"Once upon a time..."}`)
	mock := &api.MockClient{
		GetSectionFn: func(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
			if storyID != "s1" || sectionID != "sec1" {
				t.Errorf("unexpected IDs: story=%s, section=%s", storyID, sectionID)
			}
			return expected, nil
		},
	}

	svc := NewService(mock)
	data, err := svc.GetSection(context.Background(), "s1", "sec1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}
}

func TestGetQuality(t *testing.T) {
	expected := json.RawMessage(`{"score":85}`)
	mock := &api.MockClient{
		GetQualityFn: func(ctx context.Context, storyID string) (json.RawMessage, error) {
			return expected, nil
		},
	}

	svc := NewService(mock)
	data, err := svc.GetQuality(context.Background(), "story-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}
}

func TestAssessQuality(t *testing.T) {
	expected := json.RawMessage(`{"assessed":true}`)
	mock := &api.MockClient{
		AssessQualityFn: func(ctx context.Context, storyID string, force bool) (json.RawMessage, error) {
			if !force {
				t.Error("expected force=true")
			}
			return expected, nil
		},
	}

	svc := NewService(mock)
	data, err := svc.AssessQuality(context.Background(), "story-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}
}

func TestGetInsights(t *testing.T) {
	expected := json.RawMessage(`{"insights":"good"}`)
	mock := &api.MockClient{
		GetInsightsFn: func(ctx context.Context, storyID string) (json.RawMessage, error) {
			return expected, nil
		},
	}

	svc := NewService(mock)
	data, err := svc.GetInsights(context.Background(), "story-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}
}
