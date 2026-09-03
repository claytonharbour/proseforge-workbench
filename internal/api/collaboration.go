package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ListStoryGrants returns the owner's story grant roster.
func (c *Client) ListStoryGrants(ctx context.Context, storyID string) (*gen.HandlersGrantListResponse, error) {
	resp, err := c.raw.ListStoryGrants(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("list story grants: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersGrantListResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list story grants: %w", err)
	}
	return &result, nil
}

// GrantStoryCapability grants one capability to an existing user by email.
func (c *Client) GrantStoryCapability(ctx context.Context, storyID, email, capability string) (*gen.HandlersGrantResponse, error) {
	body := gen.CreateStoryGrantJSONRequestBody{Email: &email, Capability: &capability}
	resp, err := c.raw.CreateStoryGrant(ctx, storyID, body)
	if err != nil {
		return nil, fmt.Errorf("grant story capability: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersGrantResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("grant story capability: %w", err)
	}
	return &result, nil
}

// RevokeStoryCapability revokes one capability from an existing user by email.
func (c *Client) RevokeStoryCapability(ctx context.Context, storyID, email, capability string) error {
	body := gen.DeleteStoryGrantJSONRequestBody{Email: &email, Capability: &capability}
	resp, err := c.raw.DeleteStoryGrant(ctx, storyID, body)
	if err != nil {
		return fmt.Errorf("revoke story capability: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("revoke story capability: %w", err)
	}
	return nil
}

// LeaveStory removes every grant held by the authenticated caller. Owners
// cannot leave their own story; ownership is not represented as a grant.
func (c *Client) LeaveStory(ctx context.Context, storyID string) error {
	resp, err := c.raw.DeleteMyStoryGrant(ctx, storyID)
	if err != nil {
		return fmt.Errorf("leave story: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("leave story: %w", err)
	}
	return nil
}

// ListSeriesGrants returns the series owner's grant roster.
func (c *Client) ListSeriesGrants(ctx context.Context, seriesID string) (*gen.HandlersGrantListResponse, error) {
	resp, err := c.raw.ListSeriesGrants(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("list series grants: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersGrantListResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list series grants: %w", err)
	}
	return &result, nil
}

// GrantSeriesCapability grants one capability to an existing user by email.
func (c *Client) GrantSeriesCapability(ctx context.Context, seriesID, email, capability string) (*gen.HandlersGrantResponse, error) {
	body := gen.CreateSeriesGrantJSONRequestBody{Email: &email, Capability: &capability}
	resp, err := c.raw.CreateSeriesGrant(ctx, seriesID, body)
	if err != nil {
		return nil, fmt.Errorf("grant series capability: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersGrantResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("grant series capability: %w", err)
	}
	return &result, nil
}

// RevokeSeriesCapability revokes one capability from an existing user by email.
func (c *Client) RevokeSeriesCapability(ctx context.Context, seriesID, email, capability string) error {
	body := gen.DeleteSeriesGrantJSONRequestBody{Email: &email, Capability: &capability}
	resp, err := c.raw.DeleteSeriesGrant(ctx, seriesID, body)
	if err != nil {
		return fmt.Errorf("revoke series capability: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("revoke series capability: %w", err)
	}
	return nil
}

// GetSharedStories returns stories shared with the authenticated user.
func (c *Client) GetSharedStories(ctx context.Context) (*gen.HandlersSharedStoriesResponse, error) {
	resp, err := c.raw.ListMyUsersSharedStories(ctx)
	if err != nil {
		return nil, fmt.Errorf("get shared stories: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersSharedStoriesResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get shared stories: %w", err)
	}
	return &result, nil
}

// ListContributions returns contributions for a story. This is owner-facing.
func (c *Client) ListContributions(ctx context.Context, storyID string) (*gen.HandlersContributionsResponse, error) {
	resp, err := c.raw.ListStoryContributions(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersContributionsResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}
	return &result, nil
}

// StoryContributionAccounting returns contribution totals and snapshots for a story.
func (c *Client) StoryContributionAccounting(ctx context.Context, storyID string) (*gen.ContributionsAccounting, error) {
	resp, err := c.raw.GetStoryContributionsAccounting(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get story contribution accounting: %w", err)
	}
	defer resp.Body.Close()
	var result gen.ContributionsAccounting
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get story contribution accounting: %w", err)
	}
	return &result, nil
}

// SeriesContributionAccounting returns contribution totals and snapshots across a series.
func (c *Client) SeriesContributionAccounting(ctx context.Context, seriesID string) (*gen.ContributionsAccounting, error) {
	resp, err := c.raw.GetSeriesContributionsAccounting(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("get series contribution accounting: %w", err)
	}
	defer resp.Body.Close()
	var result gen.ContributionsAccounting
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get series contribution accounting: %w", err)
	}
	return &result, nil
}

// GetMyContribution returns the authenticated user's active contribution.
func (c *Client) GetMyContribution(ctx context.Context, storyID string) (*gen.HandlersContributionEnvelope, error) {
	resp, err := c.raw.ListMyStoryContributions(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get contribution: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersContributionEnvelope
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get contribution: %w", err)
	}
	return &result, nil
}

// GetContributionDiff returns a contribution's branch-vs-trunk diff.
func (c *Client) GetContributionDiff(ctx context.Context, storyID, contributionID string) (*gen.HandlersContributionDiffResponse, error) {
	resp, err := c.raw.GetStoryContributionsDiff(ctx, storyID, contributionID)
	if err != nil {
		return nil, fmt.Errorf("get contribution diff: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersContributionDiffResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get contribution diff: %w", err)
	}
	return &result, nil
}

func (c *Client) contributionAction(ctx context.Context, action string, storyID, contributionID string, call func() (*http.Response, error)) (*gen.HandlersContributionEnvelope, error) {
	resp, err := call()
	if err != nil {
		return nil, fmt.Errorf("%s contribution: %w", action, err)
	}
	defer resp.Body.Close()
	var result gen.HandlersContributionEnvelope
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("%s contribution: %w", action, err)
	}
	return &result, nil
}

// MarkContributionReady submits a contribution for owner review.
func (c *Client) MarkContributionReady(ctx context.Context, storyID, contributionID string) (*gen.HandlersContributionEnvelope, error) {
	return c.contributionAction(ctx, "mark ready", storyID, contributionID, func() (*http.Response, error) {
		return c.raw.ReadyStoryContribution(ctx, storyID, contributionID)
	})
}

// NotifyOwnerContributionReady tells the owner that a delegated reviewer has
// finished reviewing a ready contribution.
func (c *Client) NotifyOwnerContributionReady(ctx context.Context, storyID, contributionID string) error {
	resp, err := c.raw.CreateStoryContributionsReadyForOwner(ctx, storyID, contributionID)
	if err != nil {
		return fmt.Errorf("notify owner contribution ready: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("notify owner contribution ready: %w", err)
	}
	return nil
}

// DiscardContribution removes the authenticated contributor's unmerged branch
// and buffered edits without changing the story owner's trunk.
func (c *Client) DiscardContribution(ctx context.Context, storyID, contributionID string) error {
	resp, err := c.raw.DeleteStoryContribution(ctx, storyID, contributionID)
	if err != nil {
		return fmt.Errorf("discard contribution: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("discard contribution: %w", err)
	}
	return nil
}

// RequestContributionChanges asks a contributor to revise their work.
func (c *Client) RequestContributionChanges(ctx context.Context, storyID, contributionID string) (*gen.HandlersContributionEnvelope, error) {
	return c.contributionAction(ctx, "request changes", storyID, contributionID, func() (*http.Response, error) {
		return c.raw.CreateStoryContributionsRequestChange(ctx, storyID, contributionID)
	})
}

// MergeContribution merges a contribution into the owner's story.
//
// selections is optional per-file partial accept (proseforge#812): path -> accept.
// nil merges everything, which is the long-standing behaviour.
//
// 🛑 When supplied it is FAIL-CLOSED — it must name EVERY changed path. An absent
// path is a 400 `invalid_selections`, NOT an implicit accept, and that is the
// opposite of the feedback-merge path. So a caller translating section ids to
// paths must expand to the complete set before calling; a partial map produces a
// rejected request, and a silently-wrong one would produce "a partial merge that
// looks complete" (forge/proseforge-workbench#275).
func (c *Client) MergeContribution(ctx context.Context, storyID, contributionID string,
	selections map[string]bool) (*gen.HandlersContributionEnvelope, error) {
	body := gen.MergeStoryContributionJSONRequestBody{}
	if selections != nil {
		body.Selections = &selections
	}
	return c.contributionAction(ctx, "merge", storyID, contributionID, func() (*http.Response, error) {
		return c.raw.MergeStoryContribution(ctx, storyID, contributionID, body)
	})
}

// SyncContribution pulls trunk changes into a contribution branch. The backend
// authorizes either the contributor or the story owner and resolves the target
// branch from the contribution ID.
func (c *Client) SyncContribution(ctx context.Context, storyID, contributionID string, resolutions map[string]string) (*gen.HandlersPullContributionResponse, error) {
	return c.PullContribution(ctx, storyID, contributionID, "master", resolutions)
}

// PullContribution pulls trunk changes into a named contribution branch.
func (c *Client) PullContribution(ctx context.Context, storyID, contributionID, branch string, resolutions map[string]string) (*gen.HandlersPullContributionResponse, error) {
	body := gen.CreateStoryContributionsPullJSONRequestBody{}
	if resolutions != nil {
		body.Resolutions = &resolutions
	}
	resp, err := c.raw.CreateStoryContributionsPull(ctx, storyID, contributionID, branch, body)
	if err != nil {
		return nil, fmt.Errorf("pull contribution: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersPullContributionResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("pull contribution: %w", err)
	}
	return &result, nil
}

// ListContributionSuggestions returns the suggestions raised against a
// contribution — the read half of the write path SuggestContributionSection
// already covers (forge/proseforge#813 P4).
//
// Returned raw. The payload is the same four-array bucketing the review surface
// serves, and that surface is mid-rework upstream, so decoding it into a shape
// here would bake in a structure that is still moving.
func (c *Client) ListContributionSuggestions(ctx context.Context, storyID, contributionID string) (json.RawMessage, error) {
	resp, err := c.raw.ListStoryContributionsSuggestions(ctx, storyID, contributionID)
	if err != nil {
		return nil, fmt.Errorf("list contribution suggestions: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list contribution suggestions: %w", err)
	}
	return json.RawMessage(body), nil
}

// UpdateContributionSuggestionStatus moves a single suggestion's status.
//
// The vocabulary has two halves, and they are not interchangeable: replacements
// (which carry original/suggested text) take accepted or rejected, while
// advisory items take acknowledged or dismissed. pending undoes either.
//
// Validated here rather than sent through, because status is the only thing
// distinguishing accepting from rejecting — a typo would otherwise arrive as a
// plausible-looking no-op.
//
// The set is deliberately NOT validated here. It is being simplified upstream,
// and a hard-coded list in a shipped client turns every future change into a
// false rejection of a value the server accepts — a worse failure than the one
// it guards against, and one only a workbench release can clear. The generated
// model comment already disagrees with the route documentation (three values
// against five), which is the same drift arriving early.
//
// The server is the authority. An unknown status comes back as an API error
// naming it, so a typo still fails loudly rather than becoming a silent no-op —
// which was the only reason to validate locally.
//
// This is per-suggestion selection: reject one, accept another, then merge, and
// only the accepted text reaches trunk (#812).
func (c *Client) UpdateContributionSuggestionStatus(ctx context.Context, storyID, contributionID, suggestionID, status string) (json.RawMessage, error) {
	if status == "" {
		return nil, fmt.Errorf("suggestion status is required")
	}

	body := gen.UpdateStoryContributionsSuggestionJSONRequestBody{Status: &status}
	resp, err := c.raw.UpdateStoryContributionsSuggestion(
		ctx, storyID, contributionID, suggestionID, body)
	if err != nil {
		return nil, fmt.Errorf("update suggestion status: %w", err)
	}
	defer resp.Body.Close()

	raw, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("update suggestion status: %w", err)
	}
	return json.RawMessage(raw), nil
}

// SuggestContributionSection writes an owner suggestion to a contribution section.
func (c *Client) SuggestContributionSection(ctx context.Context, storyID, contributionID, sectionID, content string) (*gen.HandlersContributionEnvelope, error) {
	body := gen.UpdateStoryContributionsSectionsSuggestJSONRequestBody{Content: &content}
	resp, err := c.raw.UpdateStoryContributionsSectionsSuggest(ctx, storyID, contributionID, sectionID, body)
	if err != nil {
		return nil, fmt.Errorf("suggest contribution section: %w", err)
	}
	defer resp.Body.Close()
	var result gen.HandlersContributionEnvelope
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("suggest contribution section: %w", err)
	}
	return &result, nil
}
