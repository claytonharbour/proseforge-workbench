// Package friends provides the friends directory: the people a story owner can
// invite to collaborate. Renamed from the reviewer pool in forge/proseforge#808.
package friends

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Service provides friends operations backed by the ProseForge API.
type Service struct {
	api    api.ProseForgeAPI
	logger *slog.Logger
}

// NewService creates a friends Service.
func NewService(client api.ProseForgeAPI, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the friends service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// List returns the caller's accepted friendships. Pass include="system" to get
// the AI reviewers alongside people; empty lists people only.
func (s *Service) List(ctx context.Context, include string) (json.RawMessage, error) {
	s.logger.Info("friends.List", "include", include)
	return s.api.ListFriends(ctx, include)
}

// Count returns the number of accepted friendships.
func (s *Service) Count(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("friends.Count")
	return s.api.CountFriends(ctx)
}

// Candidates searches users who could be sent a friend request.
func (s *Service) Candidates(ctx context.Context, search string) (json.RawMessage, error) {
	s.logger.Info("friends.Candidates", "search", search)
	return s.api.ListFriendCandidates(ctx, search)
}

// Request sends a friend request to another user.
func (s *Service) Request(ctx context.Context, friendID string) error {
	s.logger.Info("friends.Request", "friendID", friendID)
	return s.api.RequestFriend(ctx, api.CreateReviewerRequestReq{FriendId: &friendID})
}

// Respond accepts or declines an incoming friend request.
func (s *Service) Respond(ctx context.Context, requestID string, accept bool) error {
	s.logger.Info("friends.Respond", "requestID", requestID, "accept", accept)
	return s.api.RespondToFriendRequest(ctx, requestID, api.RespondToReviewerReq{Accept: &accept})
}

// Incoming returns friend requests awaiting the caller's response.
func (s *Service) Incoming(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("friends.Incoming")
	return s.api.ListIncomingFriendRequests(ctx)
}

// Outgoing returns friend requests the caller has sent.
func (s *Service) Outgoing(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("friends.Outgoing")
	return s.api.ListOutgoingFriendRequests(ctx)
}

// Remove ends a friendship.
func (s *Service) Remove(ctx context.Context, friendID string) error {
	s.logger.Info("friends.Remove", "friendID", friendID)
	return s.api.RemoveFriend(ctx, friendID)
}
