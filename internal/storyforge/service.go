// Package storyforge provides Story Forge chat and generation pipeline operations.
package storyforge

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides Story Forge operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a storyforge Service.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the storyforge service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// CreateSession starts a new Story Forge Chat interview.
func (s *Service) CreateSession(ctx context.Context) (*api.ChatSession, error) {
	s.logger.Info("storyforge.CreateSession")
	return s.api.CreateChatSession(ctx)
}

// ListSessions returns the user's chat sessions.
func (s *Service) ListSessions(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("storyforge.ListSessions")
	return s.api.ListChatSessions(ctx)
}

// GetSession returns a chat session with its messages.
func (s *Service) GetSession(ctx context.Context, id string) (*api.ChatSession, error) {
	s.logger.Info("storyforge.GetSession", "id", id)
	return s.api.GetChatSession(ctx, id)
}

// SendMessage sends a message and gets the AI response.
func (s *Service) SendMessage(ctx context.Context, id string, req api.ChatSendReq) (*api.ChatSendResp, error) {
	s.logger.Info("storyforge.SendMessage", "sessionID", id)
	return s.api.SendChatMessage(ctx, id, req)
}

// Finalize finalizes a chat interview and triggers story generation.
func (s *Service) Finalize(ctx context.Context, id string) (*api.ChatFinalizeResp, error) {
	s.logger.Info("storyforge.Finalize", "sessionID", id)
	return s.api.FinalizeChatSession(ctx, id)
}

// GetStatus polls the generation pipeline status.
func (s *Service) GetStatus(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("storyforge.GetStatus", "storyID", storyID)
	return s.api.GetGenerationStatus(ctx, storyID)
}

// GetMeta returns the generated outline.
func (s *Service) GetMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("storyforge.GetMeta", "storyID", storyID)
	return s.api.GetStoryMeta(ctx, storyID)
}

// ApproveMeta approves the outline and starts section generation.
func (s *Service) ApproveMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("storyforge.ApproveMeta", "storyID", storyID)
	return s.api.ApproveStoryMeta(ctx, storyID)
}

// RegenerateMeta triggers a free retry on outline generation.
func (s *Service) RegenerateMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("storyforge.RegenerateMeta", "storyID", storyID)
	return s.api.RegenerateStoryMeta(ctx, storyID)
}

// ResumeGeneration resumes a failed or paused generation.
func (s *Service) ResumeGeneration(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("storyforge.ResumeGeneration", "storyID", storyID)
	return s.api.ResumeGeneration(ctx, storyID)
}
