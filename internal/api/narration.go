package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// StartNarration triggers narration generation for a story.
func (c *Client) StartNarration(ctx context.Context, storyID string) error {
	body := gen.HandlersNarrationOptionsRequest{}
	resp, err := c.raw.PostStoryIdNarrate(ctx, storyID, body)
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

// GetNarration returns narration status and chapter details for a story.
func (c *Client) GetNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdNarration(ctx, storyID)
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

// RegenerateChapter re-narrates a specific chapter.
// If force is true, regenerates even if content hasn't changed.
// If voice is non-empty, overrides the narration-level voice for this chapter.
func (c *Client) RegenerateChapter(ctx context.Context, storyID, chapterID string, force bool, voice string) error {
	body := gen.HandlersRegenerateChapterRequest{}
	if force {
		body.Force = &force
	}
	if voice != "" {
		body.VoiceHint = &voice
	}
	resp, err := c.raw.PostStoryIdNarrationChaptersChapterIdRegenerate(ctx, storyID, chapterID, body)
	if err != nil {
		return fmt.Errorf("regenerate chapter %s for story %s: %w", chapterID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("regenerate chapter %s for story %s: %w", chapterID, storyID, err)
	}
	return nil
}

// ListVoices returns available TTS voices across all providers.
func (c *Client) ListVoices(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetTtsVoices(ctx)
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

// RetryChapter resets a failed/stuck chapter and re-queues it.
func (c *Client) RetryChapter(ctx context.Context, storyID, chapterID string) error {
	resp, err := c.raw.PostStoryIdNarrationChaptersChapterIdRetry(ctx, storyID, chapterID)
	if err != nil {
		return fmt.Errorf("retry chapter %s for story %s: %w", chapterID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("retry chapter %s for story %s: %w", chapterID, storyID, err)
	}
	return nil
}

// GetCredits returns the authenticated user's credit balance.
func (c *Client) GetCredits(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetCreditsMe(ctx)
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

// RebuildNarration rebuilds the audiobook from existing chapter audio.
// If chapterAnnouncements is true, TTS-generated chapter title announcements are inserted.
func (c *Client) RebuildNarration(ctx context.Context, storyID string, chapterAnnouncements bool) error {
	body := gen.HandlersNarrationOptionsRequest{}
	if chapterAnnouncements {
		body.Options = &gen.PostgresNarrationOptions{
			ChapterAnnouncements: &chapterAnnouncements,
		}
	}
	resp, err := c.raw.PostStoryIdNarrationRebuild(ctx, storyID, body)
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
	resp, err := c.raw.DeleteStoryIdNarration(ctx, storyID)
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
	resp, err := c.raw.PostStoryIdNarrationResume(ctx, storyID)
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

// CancelChapter cancels a specific chapter's narration.
func (c *Client) CancelChapter(ctx context.Context, storyID, chapterID string) error {
	resp, err := c.raw.PostStoryIdNarrationChaptersChapterIdCancel(ctx, storyID, chapterID)
	if err != nil {
		return fmt.Errorf("cancel chapter %s for story %s: %w", chapterID, storyID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("cancel chapter %s for story %s: %w", chapterID, storyID, err)
	}
	return nil
}

// ListSegments returns segment details for a chapter.
func (c *Client) ListSegments(ctx context.Context, storyID, chapterID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdNarrationChaptersChapterIdSegments(ctx, storyID, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list segments for chapter %s: %w", chapterID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list segments for chapter %s: %w", chapterID, err)
	}
	return json.RawMessage(body), nil
}

// RegenerateSegment re-narrates a single segment within a chapter.
// If voice is non-empty, overrides the voice for this segment only.
func (c *Client) RegenerateSegment(ctx context.Context, storyID, chapterID, segmentID, voice string) error {
	body := gen.HandlersRegenerateSegmentRequest{}
	if voice != "" {
		body.VoiceHint = &voice
	}
	resp, err := c.raw.PostStoryIdNarrationChaptersChapterIdSegmentsSegmentIdRegenerate(ctx, storyID, chapterID, segmentID, body)
	if err != nil {
		return fmt.Errorf("regenerate segment %s in chapter %s: %w", segmentID, chapterID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("regenerate segment %s in chapter %s: %w", segmentID, chapterID, err)
	}
	return nil
}

// PatchNarration patches multiple segments and/or chapters in one call, then rebuilds the audiobook.
func (c *Client) PatchNarration(ctx context.Context, storyID string, req gen.HandlersPatchNarrationRequest) (json.RawMessage, error) {
	resp, err := c.raw.PostStoryIdNarrationPatch(ctx, storyID, req)
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
	params := &gen.GetCreditsMeTransactionsParams{Limit: &limit}
	resp, err := c.raw.GetCreditsMeTransactions(ctx, params)
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
	resp, err := c.raw.GetStoryIdNarrationAudiobook(ctx, storyID)
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
