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

// UpdateTimeline updates the canon timeline for a series.
func (s *Service) UpdateTimeline(ctx context.Context, seriesID, content string) error {
	s.logger.Info("series.UpdateTimeline", "seriesID", seriesID)
	return s.api.UpdateTimeline(ctx, seriesID, content)
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

// CreateChat starts a new world-building chat session for a series.
func (s *Service) CreateChat(ctx context.Context, seriesID string) (*api.SeriesChatSession, error) {
	s.logger.Info("series.CreateChat", "seriesID", seriesID)
	return s.api.CreateSeriesChat(ctx, seriesID)
}

// ListChats returns chat sessions for a series.
func (s *Service) ListChats(ctx context.Context, seriesID string) (json.RawMessage, error) {
	s.logger.Info("series.ListChats", "seriesID", seriesID)
	return s.api.ListSeriesChats(ctx, seriesID)
}

// GetChat returns a chat session with its messages.
func (s *Service) GetChat(ctx context.Context, seriesID, sessionID string) (*api.SeriesChatSession, error) {
	s.logger.Info("series.GetChat", "seriesID", seriesID, "sessionID", sessionID)
	return s.api.GetSeriesChat(ctx, seriesID, sessionID)
}

// SendChatMessage sends a message and gets the AI response.
func (s *Service) SendChatMessage(ctx context.Context, seriesID, sessionID string, req api.SeriesChatSendReq) (*api.SeriesChatSendResp, error) {
	s.logger.Info("series.SendChatMessage", "seriesID", seriesID, "sessionID", sessionID)
	return s.api.SendSeriesChatMessage(ctx, seriesID, sessionID, req)
}

// FinalizeChat finalizes a chat session.
func (s *Service) FinalizeChat(ctx context.Context, seriesID, sessionID string) (*api.SeriesChatFinalizeResp, error) {
	s.logger.Info("series.FinalizeChat", "seriesID", seriesID, "sessionID", sessionID)
	return s.api.FinalizeSeriesChat(ctx, seriesID, sessionID)
}

// HarvestChat extracts metadata from a chat session to git.
func (s *Service) HarvestChat(ctx context.Context, seriesID, sessionID string) (json.RawMessage, error) {
	s.logger.Info("series.HarvestChat", "seriesID", seriesID, "sessionID", sessionID)
	return s.api.HarvestSeriesChat(ctx, seriesID, sessionID)
}

// HarvestAllChats harvests all chat sessions for a series.
func (s *Service) HarvestAllChats(ctx context.Context, seriesID string) (json.RawMessage, error) {
	s.logger.Info("series.HarvestAllChats", "seriesID", seriesID)
	return s.api.HarvestAllSeriesChats(ctx, seriesID)
}
