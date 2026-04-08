// Package series provides series management operations.
package series

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides series-related operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a series Service.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the series service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// List returns the authenticated user's series.
func (s *Service) List(ctx context.Context) (*api.SeriesList, error) {
	s.logger.Info("series.List")
	return s.api.ListSeries(ctx)
}

// Create creates a new series.
func (s *Service) Create(ctx context.Context, req api.CreateSeriesReq) (*api.Series, error) {
	s.logger.Info("series.Create", "name", req.Name)
	return s.api.CreateSeries(ctx, req)
}

// Get returns a single series by ID.
func (s *Service) Get(ctx context.Context, id string) (*api.Series, error) {
	s.logger.Info("series.Get", "id", id)
	return s.api.GetSeriesByID(ctx, id)
}

// Update updates a series' metadata.
func (s *Service) Update(ctx context.Context, id string, req api.UpdateSeriesReq) error {
	s.logger.Info("series.Update", "id", id)
	return s.api.UpdateSeries(ctx, id, req)
}

// Archive archives (deletes) a series.
func (s *Service) Archive(ctx context.Context, id string) error {
	s.logger.Info("series.Archive", "id", id)
	return s.api.ArchiveSeries(ctx, id)
}

// GetWorld returns the world overview document for a series.
func (s *Service) GetWorld(ctx context.Context, seriesID string) (json.RawMessage, error) {
	s.logger.Info("series.GetWorld", "seriesID", seriesID)
	return s.api.GetWorld(ctx, seriesID)
}

// UpdateWorld updates the world overview document for a series.
func (s *Service) UpdateWorld(ctx context.Context, seriesID, content string) error {
	s.logger.Info("series.UpdateWorld", "seriesID", seriesID)
	return s.api.UpdateWorld(ctx, seriesID, content)
}

// GetTimeline returns the canon timeline for a series.
func (s *Service) GetTimeline(ctx context.Context, seriesID string) (json.RawMessage, error) {
	s.logger.Info("series.GetTimeline", "seriesID", seriesID)
	return s.api.GetTimeline(ctx, seriesID)
}

// ListTimelineSections returns the list of timeline sections (slugs, titles, sort order).
func (s *Service) ListTimelineSections(ctx context.Context, seriesID string) (json.RawMessage, error) {
	s.logger.Info("series.ListTimelineSections", "seriesID", seriesID)
	return s.api.ListTimelineSections(ctx, seriesID)
}

// GetTimelineSection returns a single timeline section by slug.
func (s *Service) GetTimelineSection(ctx context.Context, seriesID, slug string) (json.RawMessage, error) {
	s.logger.Info("series.GetTimelineSection", "seriesID", seriesID, "slug", slug)
	return s.api.GetTimelineSection(ctx, seriesID, slug)
}

// UpdateTimelineSection updates a single timeline section by slug.
func (s *Service) UpdateTimelineSection(ctx context.Context, seriesID, slug, title, content string) (json.RawMessage, error) {
	s.logger.Info("series.UpdateTimelineSection", "seriesID", seriesID, "slug", slug)
	return s.api.UpdateTimelineSection(ctx, seriesID, slug, title, content)
}

// DeleteTimelineSection removes a timeline section.
func (s *Service) DeleteTimelineSection(ctx context.Context, seriesID, slug string) error {
	s.logger.Info("series.DeleteTimelineSection", "seriesID", seriesID, "slug", slug)
	return s.api.DeleteTimelineSection(ctx, seriesID, slug)
}

// ReorderSeriesStories sets story order in a series.
func (s *Service) ReorderSeriesStories(ctx context.Context, seriesID string, storyIDs []string) error {
	s.logger.Info("series.ReorderSeriesStories", "seriesID", seriesID)
	return s.api.ReorderSeriesStories(ctx, seriesID, storyIDs)
}

// ReorderTimelineSections sets timeline section order.
func (s *Service) ReorderTimelineSections(ctx context.Context, seriesID string, slugs []string) error {
	s.logger.Info("series.ReorderTimelineSections", "seriesID", seriesID)
	return s.api.ReorderTimelineSections(ctx, seriesID, slugs)
}

// CreateCharacter creates a character in a series.
func (s *Service) CreateCharacter(ctx context.Context, seriesID string, req api.CreateCharacterReq) (*api.Character, error) {
	s.logger.Info("series.CreateCharacter", "seriesID", seriesID, "name", req.Name)
	return s.api.CreateCharacter(ctx, seriesID, req)
}

// ListCharacters returns all characters in a series.
func (s *Service) ListCharacters(ctx context.Context, seriesID string) (*api.CharacterList, error) {
	s.logger.Info("series.ListCharacters", "seriesID", seriesID)
	return s.api.ListCharacters(ctx, seriesID)
}

// GetCharacter returns a character by slug.
func (s *Service) GetCharacter(ctx context.Context, seriesID, slug string) (*api.Character, error) {
	s.logger.Info("series.GetCharacter", "seriesID", seriesID, "slug", slug)
	return s.api.GetCharacter(ctx, seriesID, slug)
}

// UpdateCharacter updates a character's profile.
func (s *Service) UpdateCharacter(ctx context.Context, seriesID, slug string, req api.UpdateCharacterReq) error {
	s.logger.Info("series.UpdateCharacter", "seriesID", seriesID, "slug", slug)
	return s.api.UpdateCharacter(ctx, seriesID, slug, req)
}

// DeleteCharacter removes a character from a series.
func (s *Service) DeleteCharacter(ctx context.Context, seriesID, slug string) error {
	s.logger.Info("series.DeleteCharacter", "seriesID", seriesID, "slug", slug)
	return s.api.DeleteCharacter(ctx, seriesID, slug)
}

// ListStories returns stories linked to a series.
func (s *Service) ListStories(ctx context.Context, seriesID string) (json.RawMessage, error) {
	s.logger.Info("series.ListStories", "seriesID", seriesID)
	return s.api.ListSeriesStories(ctx, seriesID)
}

// AddStory links an existing story to a series.
func (s *Service) AddStory(ctx context.Context, seriesID, storyID string) error {
	s.logger.Info("series.AddStory", "seriesID", seriesID, "storyID", storyID)
	return s.api.AddStoryToSeries(ctx, seriesID, storyID)
}

// RemoveStory unlinks a story from a series.
func (s *Service) RemoveStory(ctx context.Context, seriesID, storyID string) error {
	s.logger.Info("series.RemoveStory", "seriesID", seriesID, "storyID", storyID)
	return s.api.RemoveStoryFromSeries(ctx, seriesID, storyID)
}

// PlanStory creates a StorySeed-seeded Story Forge Chat session from series context.
func (s *Service) PlanStory(ctx context.Context, seriesID string, req api.PlanStoryReq) (*api.PlanStoryResp, error) {
	s.logger.Info("series.PlanStory", "seriesID", seriesID)
	return s.api.PlanStory(ctx, seriesID, req)
}

