// Package conversation provides standalone-room operations (forge/proseforge#821) shared by
// the CLI and the MCP server.
//
// ⚑ There is no Send or Read here on purpose. Messaging a conversation is the EXISTING room
// API with entity_type "conversation" — room_send / room_read already work against it and
// need no new code. What had no path was creating one and managing who is in it, and that is
// all this package does.
package conversation

import (
	"context"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service wraps the concrete *api.Client, mirroring internal/room: the conversation methods
// live on the client rather than on the ProseForgeAPI interface.
type Service struct {
	api    *api.Client
	logger *slog.Logger
}

func NewService(client *api.Client, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Option func(*Service)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) { s.logger = logger }
}

func (s *Service) List(ctx context.Context) (*api.ConversationList, error) {
	s.logger.Info("conversation.List")
	return s.api.ListConversations(ctx)
}

func (s *Service) Create(ctx context.Context, title string) (*api.Conversation, error) {
	s.logger.Info("conversation.Create", "title", title)
	return s.api.CreateConversation(ctx, title)
}

func (s *Service) AddMember(ctx context.Context, id, principalID, capability string) error {
	s.logger.Info("conversation.AddMember", "conversation", id, "principal", principalID, "capability", capability)
	return s.api.AddConversationMember(ctx, id, principalID, capability)
}

// RemoveMember revokes another member's access (#1173). Creator-only, enforced server-side.
func (s *Service) RemoveMember(ctx context.Context, id, principalID string) error {
	return s.api.RemoveConversationMember(ctx, id, principalID)
}

func (s *Service) Leave(ctx context.Context, id string) error {
	s.logger.Info("conversation.Leave", "conversation", id)
	return s.api.LeaveConversation(ctx, id)
}
