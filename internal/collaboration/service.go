// Package collaboration provides grant-based story collaboration operations.
package collaboration

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Service wraps typed API operations for shared stories and contributions.
type Service struct {
	api    *api.Client
	logger *slog.Logger
}

// NewService creates a collaboration service.
func NewService(client *api.Client, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the service logger.
func WithLogger(logger *slog.Logger) Option { return func(s *Service) { s.logger = logger } }

func (s *Service) SharedStories(ctx context.Context) (*gen.HandlersSharedStoriesResponse, error) {
	s.logger.Info("collaboration.SharedStories")
	return s.api.GetSharedStories(ctx)
}

func (s *Service) Grants(ctx context.Context, storyID string) (*gen.HandlersGrantListResponse, error) {
	s.logger.Info("collaboration.Grants", "storyID", storyID)
	return s.api.ListStoryGrants(ctx, storyID)
}

func (s *Service) Grant(ctx context.Context, storyID, email, capability string) (*gen.HandlersGrantResponse, error) {
	s.logger.Info("collaboration.Grant", "storyID", storyID, "email", email, "capability", capability)
	return s.api.GrantStoryCapability(ctx, storyID, email, capability)
}

func (s *Service) Revoke(ctx context.Context, storyID, email, capability string) error {
	s.logger.Info("collaboration.Revoke", "storyID", storyID, "email", email, "capability", capability)
	return s.api.RevokeStoryCapability(ctx, storyID, email, capability)
}

// Leave removes every story grant held by the authenticated caller.
func (s *Service) Leave(ctx context.Context, storyID string) error {
	s.logger.Info("collaboration.Leave", "storyID", storyID)
	return s.api.LeaveStory(ctx, storyID)
}

func (s *Service) SeriesGrants(ctx context.Context, seriesID string) (*gen.HandlersGrantListResponse, error) {
	s.logger.Info("collaboration.SeriesGrants", "seriesID", seriesID)
	return s.api.ListSeriesGrants(ctx, seriesID)
}

func (s *Service) SeriesGrant(ctx context.Context, seriesID, email, capability string) (*gen.HandlersGrantResponse, error) {
	s.logger.Info("collaboration.SeriesGrant", "seriesID", seriesID, "email", email, "capability", capability)
	return s.api.GrantSeriesCapability(ctx, seriesID, email, capability)
}

func (s *Service) SeriesRevoke(ctx context.Context, seriesID, email, capability string) error {
	s.logger.Info("collaboration.SeriesRevoke", "seriesID", seriesID, "email", email, "capability", capability)
	return s.api.RevokeSeriesCapability(ctx, seriesID, email, capability)
}

func (s *Service) Contributions(ctx context.Context, storyID string) (*gen.HandlersContributionsResponse, error) {
	s.logger.Info("collaboration.Contributions", "storyID", storyID)
	return s.api.ListContributions(ctx, storyID)
}

func (s *Service) StoryAccounting(ctx context.Context, storyID string) (*gen.ContributionsAccounting, error) {
	s.logger.Info("collaboration.StoryAccounting", "storyID", storyID)
	return s.api.StoryContributionAccounting(ctx, storyID)
}

func (s *Service) SeriesAccounting(ctx context.Context, seriesID string) (*gen.ContributionsAccounting, error) {
	s.logger.Info("collaboration.SeriesAccounting", "seriesID", seriesID)
	return s.api.SeriesContributionAccounting(ctx, seriesID)
}

func (s *Service) Mine(ctx context.Context, storyID string) (*gen.HandlersContributionEnvelope, error) {
	s.logger.Info("collaboration.Mine", "storyID", storyID)
	return s.api.GetMyContribution(ctx, storyID)
}

func (s *Service) Diff(ctx context.Context, storyID, contributionID string) (*gen.HandlersContributionDiffResponse, error) {
	s.logger.Info("collaboration.Diff", "storyID", storyID, "contributionID", contributionID)
	return s.api.GetContributionDiff(ctx, storyID, contributionID)
}

func (s *Service) Ready(ctx context.Context, storyID, contributionID string) (*gen.HandlersContributionEnvelope, error) {
	s.logger.Info("collaboration.Ready", "storyID", storyID, "contributionID", contributionID)
	return s.api.MarkContributionReady(ctx, storyID, contributionID)
}

func (s *Service) ReadyForOwner(ctx context.Context, storyID, contributionID string) error {
	s.logger.Info("collaboration.ReadyForOwner", "storyID", storyID, "contributionID", contributionID)
	return s.api.NotifyOwnerContributionReady(ctx, storyID, contributionID)
}

func (s *Service) Discard(ctx context.Context, storyID, contributionID string) error {
	s.logger.Info("collaboration.Discard", "storyID", storyID, "contributionID", contributionID)
	return s.api.DiscardContribution(ctx, storyID, contributionID)
}

func (s *Service) RequestChanges(ctx context.Context, storyID, contributionID string) (*gen.HandlersContributionEnvelope, error) {
	s.logger.Info("collaboration.RequestChanges", "storyID", storyID, "contributionID", contributionID)
	return s.api.RequestContributionChanges(ctx, storyID, contributionID)
}

// Merge accepts a contribution into trunk. selections is optional per-file
// partial accept (proseforge#812); nil takes everything.
//
// ⚠️ Log the selection COUNT, not just the ids. A partial merge that took fewer
// files than the operator meant is invisible in a log line that says only
// "Merge storyID=… contributionID=…", and the contributor cannot see their own
// diff to notice (proseforge#765).
func (s *Service) Merge(ctx context.Context, storyID, contributionID string,
	selections map[string]bool) (*gen.HandlersContributionEnvelope, error) {
	if selections == nil {
		s.logger.Info("collaboration.Merge", "storyID", storyID, "contributionID", contributionID, "selections", "all")
	} else {
		accepted := 0
		for _, take := range selections {
			if take {
				accepted++
			}
		}
		s.logger.Info("collaboration.Merge", "storyID", storyID, "contributionID", contributionID,
			"paths", len(selections), "accepted", accepted, "rejected", len(selections)-accepted)
	}
	return s.api.MergeContribution(ctx, storyID, contributionID, selections)
}

func (s *Service) Sync(ctx context.Context, storyID, contributionID string, resolutions map[string]string) (*gen.HandlersPullContributionResponse, error) {
	s.logger.Info("collaboration.Sync", "storyID", storyID, "contributionID", contributionID)
	return s.api.SyncContribution(ctx, storyID, contributionID, resolutions)
}

func (s *Service) SuggestSection(ctx context.Context, storyID, contributionID, sectionID, content string) (*gen.HandlersContributionEnvelope, error) {
	s.logger.Info("collaboration.SuggestSection", "storyID", storyID, "contributionID", contributionID, "sectionID", sectionID)
	return s.api.SuggestContributionSection(ctx, storyID, contributionID, sectionID, content)
}

// Suggestions lists the suggestions raised against a contribution.
func (s *Service) Suggestions(ctx context.Context, storyID, contributionID string) (json.RawMessage, error) {
	s.logger.Info("collaboration.Suggestions", "storyID", storyID, "contributionID", contributionID)
	return s.api.ListContributionSuggestions(ctx, storyID, contributionID)
}

// ResolveSuggestion accepts or rejects one suggestion. Resolve each one before
// merging: only accepted text reaches trunk.
func (s *Service) ResolveSuggestion(ctx context.Context, storyID, contributionID, suggestionID, status string) (json.RawMessage, error) {
	s.logger.Info("collaboration.ResolveSuggestion",
		"storyID", storyID, "contributionID", contributionID, "suggestionID", suggestionID, "status", status)
	return s.api.UpdateContributionSuggestionStatus(ctx, storyID, contributionID, suggestionID, status)
}
