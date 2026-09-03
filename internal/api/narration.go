package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// StartNarration triggers narration generation for a story.
// Pass voice="" to use the server default.
//
// force maps to the backend's #477 force flag: when true, narration_start
// DELETES an existing complete/error narration and restarts from scratch.
// Without it, starting on a story that already has a terminal-state narration
// returns 409 narration_exists; an in-flight narration returns 409 conflict
// (never force-deletable).
func (c *Client) StartNarration(ctx context.Context, storyID string, voice string, force bool, emphasis *bool) error {
	body := gen.HandlersNarrationOptionsRequest{}
	if voice != "" || emphasis != nil {
		opts := gen.PostgresNarrationOptions{Emphasis: emphasis}
		if voice != "" {
			opts.Voice = &voice
		}
		body.Options = &opts
	}
	if force {
		body.Force = &force
	}
	resp, err := c.raw.CreateStoryNarrate(ctx, storyID, body)
	if err != nil {
		return fmt.Errorf("start narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("start narration for story %s: %w", storyID, err)
	}
	return nil
}

// GetNarration returns narration status and per-section details for a story.
func (c *Client) GetNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryNarration(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get narration for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// RegenerateSection re-narrates a specific section.
// If force is true, regenerates even if content hasn't changed.
// If voice is non-empty, overrides the narration-level voice for this section.
func (c *Client) RegenerateSection(ctx context.Context, storyID, sectionID string, force bool, voice string) error {
	body := gen.HandlersRegenerateSectionRequest{}
	if force {
		body.Force = &force
	}
	if voice != "" {
		body.VoiceHint = &voice
	}
	resp, err := c.raw.RegenerateStoryNarrationSection(ctx, storyID, sectionID, body)
	if err != nil {
		return fmt.Errorf("regenerate section %s for story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("regenerate section %s for story %s: %w", sectionID, storyID, err)
	}
	return nil
}

// ListVoices returns available TTS voices across all providers.
func (c *Client) ListVoices(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.ListTtsVoices(ctx)
	if err != nil {
		return nil, fmt.Errorf("list TTS voices: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list TTS voices: %w", err)
	}
	return json.RawMessage(body), nil
}

// RetrySection resets a failed/stuck section and re-queues it.
func (c *Client) RetrySection(ctx context.Context, storyID, sectionID string) error {
	resp, err := c.raw.RetryStoryNarrationSection(ctx, storyID, sectionID)
	if err != nil {
		return fmt.Errorf("retry section %s for story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("retry section %s for story %s: %w", sectionID, storyID, err)
	}
	return nil
}

// GetCredits returns the authenticated user's credit balance.
func (c *Client) GetCredits(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.ListMyCredits(ctx)
	if err != nil {
		return nil, fmt.Errorf("get credits: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get credits: %w", err)
	}
	return json.RawMessage(body), nil
}

// RebuildNarration rebuilds the audiobook from existing per-section audio.
// Always sends the explicit section_announcements boolean so the API's
// partial-merge can toggle the flag off; a previous version only set
// Options when true, which meant rebuild-without-flag preserved the
// previously-stored true value (announcements appeared "sticky").
func (c *Client) RebuildNarration(ctx context.Context, storyID string, sectionAnnouncements bool, emphasis *bool) error {
	body := gen.HandlersNarrationOptionsRequest{
		Options: &gen.PostgresNarrationOptions{
			SectionAnnouncements: &sectionAnnouncements,
			Emphasis:             emphasis,
		},
	}
	resp, err := c.raw.RebuildStoryNarration(ctx, storyID, body)
	if err != nil {
		return fmt.Errorf("rebuild narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("rebuild narration for story %s: %w", storyID, err)
	}
	return nil
}

// DeleteNarration deletes all narration data for a story.
func (c *Client) DeleteNarration(ctx context.Context, storyID string) error {
	resp, err := c.raw.DeleteStoryNarration(ctx, storyID)
	if err != nil {
		return fmt.Errorf("delete narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("delete narration for story %s: %w", storyID, err)
	}
	return nil
}

// ResumeNarration resumes a stuck narration.
func (c *Client) ResumeNarration(ctx context.Context, storyID string) error {
	resp, err := c.raw.ResumeStoryNarration(ctx, storyID)
	if err != nil {
		return fmt.Errorf("resume narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("resume narration for story %s: %w", storyID, err)
	}
	return nil
}

// CancelSection cancels a specific section's narration.
func (c *Client) CancelSection(ctx context.Context, storyID, sectionID string) error {
	resp, err := c.raw.CancelStoryNarrationSection(ctx, storyID, sectionID)
	if err != nil {
		return fmt.Errorf("cancel section %s for story %s: %w", sectionID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("cancel section %s for story %s: %w", sectionID, storyID, err)
	}
	return nil
}

// ListSegments returns segment details for a section.
func (c *Client) ListSegments(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
	resp, err := c.raw.ListStoryNarrationSectionsSegments(ctx, storyID, sectionID)
	if err != nil {
		return nil, fmt.Errorf("list segments for section %s: %w", sectionID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list segments for section %s: %w", sectionID, err)
	}
	return json.RawMessage(body), nil
}

// RegenerateSegment re-narrates a single segment within a section.
// If voice is non-empty, overrides the voice for this segment only.
func (c *Client) RegenerateSegment(ctx context.Context, storyID, sectionID, segmentID, voice string) error {
	body := gen.HandlersRegenerateSegmentRequest{}
	if voice != "" {
		body.VoiceHint = &voice
	}
	resp, err := c.raw.RegenerateStoryNarrationSectionsSegment(ctx, storyID, sectionID, segmentID, body)
	if err != nil {
		return fmt.Errorf("regenerate segment %s in section %s: %w", segmentID, sectionID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("regenerate segment %s in section %s: %w", segmentID, sectionID, err)
	}
	return nil
}

// PatchNarration patches multiple segments and/or sections in one call, then rebuilds the audiobook.
func (c *Client) PatchNarration(ctx context.Context, storyID string, req gen.HandlersPatchNarrationRequest) (json.RawMessage, error) {
	resp, err := c.raw.CreateStoryNarrationPatch(ctx, storyID, req)
	if err != nil {
		return nil, fmt.Errorf("patch narration for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("patch narration for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// EstimateCredits returns the estimated credit cost for an operation.
func (c *Client) EstimateCredits(ctx context.Context, params *gen.GetCreditsEstimateParams) (json.RawMessage, error) {
	resp, err := c.raw.GetCreditsEstimate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("estimate credits: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("estimate credits: %w", err)
	}
	return json.RawMessage(body), nil
}

// GetCreditHistory returns recent credit transactions.
func (c *Client) GetCreditHistory(ctx context.Context, limit int) (json.RawMessage, error) {
	params := &gen.ListMyCreditsTransactionsParams{Limit: &limit}
	resp, err := c.raw.ListMyCreditsTransactions(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get credit history: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get credit history: %w", err)
	}
	return json.RawMessage(body), nil
}

// GetAudiobook returns the audiobook download info for a story.
func (c *Client) GetAudiobook(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryNarrationAudiobook(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get audiobook for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get audiobook for story %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}
