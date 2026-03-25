// Package story provides story listing and reading operations.
package story

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Service provides story-related operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a StoryService.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the story service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// List returns the authenticated user's stories.
func (s *Service) List(ctx context.Context, params *gen.GetStoriesParams) (*api.StoryList, error) {
	result, err := s.api.ListStories(ctx, params)
	if err == nil && result != nil {
		count := 0
		if result.Stories != nil {
			count = len(*result.Stories)
		}
		s.logger.Info("story.List", "count", count)
	}
	return result, err
}

// Get returns a single story by ID.
func (s *Service) Get(ctx context.Context, id string) (*api.Story, error) {
	s.logger.Info("story.Get", "id", id)
	return s.api.GetStory(ctx, id)
}

// Export downloads a story in the given format (json, markdown, pdf).
func (s *Service) Export(ctx context.Context, id, format string) (string, error) {
	s.logger.Info("story.Export", "id", id, "format", format)
	return s.api.DownloadStory(ctx, id, format)
}

// GetSection returns a single section's content and metadata.
func (s *Service) GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
	return s.api.GetSection(ctx, storyID, sectionID)
}

// GetQuality returns the code-based quality assessment for a story.
func (s *Service) GetQuality(ctx context.Context, storyID string) (json.RawMessage, error) {
	return s.api.GetQuality(ctx, storyID)
}

// AssessQuality triggers a code-based quality assessment for a story.
func (s *Service) AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error) {
	return s.api.AssessQuality(ctx, storyID, force)
}

// GetInsights returns combined quality and AI analysis information for a story.
func (s *Service) GetInsights(ctx context.Context, storyID string) (json.RawMessage, error) {
	return s.api.GetInsights(ctx, storyID)
}

// Create creates a new story.
func (s *Service) Create(ctx context.Context, req api.CreateStoryRequest) (*api.Story, error) {
	s.logger.Info("story.Create", "title", req.Title, "genreId", req.GenreId)
	return s.api.CreateStory(ctx, req)
}

// Update updates a story's metadata.
func (s *Service) Update(ctx context.Context, id string, req api.UpdateStoryRequest) error {
	s.logger.Info("story.Update", "id", id)
	return s.api.UpdateStory(ctx, id, req)
}

// Publish publishes a story.
func (s *Service) Publish(ctx context.Context, id string) error {
	s.logger.Info("story.Publish", "id", id)
	return s.api.PublishStory(ctx, id)
}

// Unpublish unpublishes a story.
func (s *Service) Unpublish(ctx context.Context, id string) error {
	s.logger.Info("story.Unpublish", "id", id)
	return s.api.UnpublishStory(ctx, id)
}

// CreateSection creates a new section in a story.
func (s *Service) CreateSection(ctx context.Context, storyID string, req api.CreateSectionRequest) (json.RawMessage, error) {
	s.logger.Info("story.CreateSection", "storyID", storyID, "name", req.Name)
	return s.api.CreateSection(ctx, storyID, req)
}

// WriteSection updates a section's content.
func (s *Service) WriteSection(ctx context.Context, storyID, sectionID string, req api.UpdateSectionRequest) error {
	s.logger.Info("story.WriteSection", "storyID", storyID, "sectionID", sectionID)
	return s.api.WriteSection(ctx, storyID, sectionID, req)
}

// ListGenres returns all available genres.
func (s *Service) ListGenres(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("story.ListGenres")
	return s.api.ListGenres(ctx)
}
