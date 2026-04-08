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

func TestCreateStory(t *testing.T) {
	title := "New Story"
	genreID := "genre-1"
	resultID := "story-new"

	mock := &api.MockClient{
		CreateStoryFn: func(ctx context.Context, req api.CreateStoryRequest) (*api.Story, error) {
			if req.Title == nil || *req.Title != title {
				t.Errorf("expected title %s, got %v", title, req.Title)
			}
			if req.GenreId == nil || *req.GenreId != genreID {
				t.Errorf("expected genre ID %s, got %v", genreID, req.GenreId)
			}
			return &api.Story{Id: &resultID, Title: &title}, nil
		},
	}

	svc := NewService(mock)
	req := api.CreateStoryRequest{Title: &title, GenreId: &genreID}
	result, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Id == nil || *result.Id != resultID {
		t.Errorf("expected ID %s, got %v", resultID, result.Id)
	}
	if result.Title == nil || *result.Title != title {
		t.Errorf("expected title %s, got %v", title, result.Title)
	}
}

func TestUpdateStory(t *testing.T) {
	var called bool
	newTitle := "Updated Title"

	mock := &api.MockClient{
		UpdateStoryFn: func(ctx context.Context, id string, req api.UpdateStoryRequest) error {
			if id != "story-1" {
				t.Errorf("expected story ID story-1, got %s", id)
			}
			if req.Title == nil || *req.Title != newTitle {
				t.Errorf("expected title %s, got %v", newTitle, req.Title)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.UpdateStoryRequest{Title: &newTitle}
	err := svc.Update(context.Background(), "story-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected UpdateStory to be called")
	}
}

func TestPublishStory(t *testing.T) {
	var called bool
	mock := &api.MockClient{
		PublishStoryFn: func(ctx context.Context, id string, visibility string) error {
			if id != "story-1" {
				t.Errorf("expected story ID story-1, got %s", id)
			}
			if visibility != "members" {
				t.Errorf("expected visibility members, got %s", visibility)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.Publish(context.Background(), "story-1", "members")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected PublishStory to be called")
	}
}

func TestCreateSection(t *testing.T) {
	name := "Chapter 1"
	expected := json.RawMessage(`{"id":"sec-new","name":"Chapter 1"}`)

	mock := &api.MockClient{
		CreateSectionFn: func(ctx context.Context, storyID string, req api.CreateSectionRequest) (json.RawMessage, error) {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			if req.Name == nil || *req.Name != name {
				t.Errorf("expected name %s, got %v", name, req.Name)
			}
			return expected, nil
		},
	}

	svc := NewService(mock)
	req := api.CreateSectionRequest{Name: &name}
	data, err := svc.CreateSection(context.Background(), "story-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}
}

func TestWriteSection(t *testing.T) {
	var called bool
	content := "Once upon a time in a land far away..."

	mock := &api.MockClient{
		WriteSectionFn: func(ctx context.Context, storyID, sectionID string, req api.UpdateSectionRequest) error {
			if storyID != "story-1" {
				t.Errorf("expected story ID story-1, got %s", storyID)
			}
			if sectionID != "sec-1" {
				t.Errorf("expected section ID sec-1, got %s", sectionID)
			}
			if req.Content == nil || *req.Content != content {
				t.Errorf("expected content %q, got %v", content, req.Content)
			}
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	req := api.UpdateSectionRequest{Content: &content}
	err := svc.WriteSection(context.Background(), "story-1", "sec-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected WriteSection to be called")
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
