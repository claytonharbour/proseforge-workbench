// Package story provides story listing and reading operations.
package story

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

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
	s.logger.Info("story.List")
	result, err := s.api.ListStories(ctx, params)
	if err == nil && result != nil {
		count := 0
		if result.Stories != nil {
			count = len(*result.Stories)
		}
		s.logger.Debug("story.List", "count", count)
	}
	return result, err
}

// Get returns a single story by ID.
func (s *Service) Get(ctx context.Context, id string) (*api.Story, error) {
	s.logger.Info("story.Get", "id", id)
	return s.api.GetStory(ctx, id)
}

// GetWithContent returns a story with full section content included.
func (s *Service) GetWithContent(ctx context.Context, id string) (*api.Story, error) {
	s.logger.Info("story.GetWithContent", "id", id)
	return s.api.GetStoryWithContent(ctx, id)
}

// Export downloads a story in the given format (json, markdown, pdf).
func (s *Service) Export(ctx context.Context, id, format string) (string, error) {
	s.logger.Info("story.Export", "id", id, "format", format)
	return s.api.DownloadStory(ctx, id, format)
}

// GetSection returns a single section's content and metadata.
func (s *Service) GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
	s.logger.Info("story.GetSection", "storyID", storyID, "sectionID", sectionID)
	return s.api.GetSection(ctx, storyID, sectionID)
}

// ListSections returns a story's sections.
//
// Deliberately not derived from the story-level read: that read 404s for a
// contributor, which left a grantee able to fetch a section only if it already
// knew the id and with no tool that would tell it one.
func (s *Service) ListSections(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.ListSections", "storyID", storyID)
	return s.api.ListSections(ctx, storyID)
}

// GetQuality returns the code-based quality assessment for a story.
func (s *Service) GetQuality(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.GetQuality", "storyID", storyID)
	return s.api.GetQuality(ctx, storyID)
}

// AssessQuality triggers a code-based quality assessment for a story.
func (s *Service) AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error) {
	s.logger.Info("story.AssessQuality", "storyID", storyID, "force", force)
	return s.api.AssessQuality(ctx, storyID, force)
}

// AssessQualityAtVersion runs a synchronous quality assessment against a specific version SHA.
func (s *Service) AssessQualityAtVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	s.logger.Info("story.AssessQualityAtVersion", "storyID", storyID, "sha", sha)
	return s.api.AssessQualityAtVersion(ctx, storyID, sha)
}

// GetInsights returns combined quality and AI analysis information for a story.
func (s *Service) GetInsights(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.GetInsights", "storyID", storyID)
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

// Publish publishes a story with optional visibility ("public" or "members").
// Pass "" to use the server default.
func (s *Service) Publish(ctx context.Context, id, visibility string) error {
	s.logger.Info("story.Publish", "id", id, "visibility", visibility)
	return s.api.PublishStory(ctx, id, visibility)
}

// UpdateVisibility changes the visibility of a published story.
func (s *Service) UpdateVisibility(ctx context.Context, id, visibility string) error {
	s.logger.Info("story.UpdateVisibility", "id", id, "visibility", visibility)
	return s.api.UpdateVisibility(ctx, id, visibility)
}

// Unpublish unpublishes a story.
func (s *Service) Unpublish(ctx context.Context, id string) error {
	s.logger.Info("story.Unpublish", "id", id)
	return s.api.UnpublishStory(ctx, id)
}

// CreatePitch creates a new story in pitch status (pre-writing idea).
func (s *Service) CreatePitch(ctx context.Context, req api.CreateStoryRequest) (*api.Story, error) {
	s.logger.Info("story.CreatePitch", "title", req.Title, "genreId", req.GenreId)
	return s.api.CreatePitch(ctx, req)
}

// Promote promotes a pitch to draft status.
func (s *Service) Promote(ctx context.Context, id string) error {
	s.logger.Info("story.Promote", "id", id)
	return s.api.PromoteStory(ctx, id)
}

// UpsertMeta writes story planning data (creates if missing, updates if present).
func (s *Service) UpsertMeta(ctx context.Context, storyID, metaType, content string) (json.RawMessage, error) {
	s.logger.Info("story.UpsertMeta", "storyID", storyID, "metaType", metaType)
	return s.api.UpsertStoryMeta(ctx, storyID, metaType, content)
}

// CreateSection creates a new section in a story.
func (s *Service) CreateSection(ctx context.Context, storyID string, req api.CreateSectionRequest) (json.RawMessage, error) {
	s.logger.Info("story.CreateSection", "storyID", storyID, "name", req.Name)
	return s.api.CreateSection(ctx, storyID, req)
}

// WriteSection updates a section's content, name, and/or position. Returns the
// raw API response body containing the updated section, the full sections list
// (with renumbered sortOrders), commitSha (empty string when no git commit
// happened — e.g. order-only edits), and current story wordCount.
func (s *Service) WriteSection(ctx context.Context, storyID, sectionID string, req api.UpdateSectionRequest) (json.RawMessage, error) {
	s.logger.Info("story.WriteSection", "storyID", storyID, "sectionID", sectionID)
	return s.api.WriteSection(ctx, storyID, sectionID, req)
}

// ReorderSection moves a section to a 0-indexed position within its story,
// renumbering siblings. Out-of-range values clamp to the nearest end.
func (s *Service) ReorderSection(ctx context.Context, storyID, sectionID string, order int) (json.RawMessage, error) {
	s.logger.Info("story.ReorderSection", "storyID", storyID, "sectionID", sectionID, "order", order)
	return s.api.ReorderSection(ctx, storyID, sectionID, order)
}

// Delete permanently deletes a story.
func (s *Service) Delete(ctx context.Context, id string) error {
	s.logger.Info("story.Delete", "id", id)
	return s.api.DeleteStory(ctx, id)
}

// DeleteSection deletes a section from a story.
func (s *Service) DeleteSection(ctx context.Context, storyID, sectionID string) error {
	s.logger.Info("story.DeleteSection", "storyID", storyID, "sectionID", sectionID)
	return s.api.DeleteSection(ctx, storyID, sectionID)
}

// RestoreVersion restores a story to a previous version.
func (s *Service) RestoreVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	s.logger.Info("story.RestoreVersion", "storyID", storyID, "sha", sha)
	return s.api.RestoreVersion(ctx, storyID, sha)
}

// GetMetaStale returns sections affected by meta changes.
func (s *Service) GetMetaStale(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.GetMetaStale", "storyID", storyID)
	return s.api.GetMetaStale(ctx, storyID)
}

// AcknowledgeMetaStale dismisses all meta staleness warnings.
func (s *Service) AcknowledgeMetaStale(ctx context.Context, storyID string) error {
	s.logger.Info("story.AcknowledgeMetaStale", "storyID", storyID)
	return s.api.AcknowledgeMetaStale(ctx, storyID)
}

// RegenerateTagline queues AI tagline regeneration.
func (s *Service) RegenerateTagline(ctx context.Context, storyID string) error {
	s.logger.Info("story.RegenerateTagline", "storyID", storyID)
	return s.api.RegenerateTagline(ctx, storyID)
}

// RegenerateTitle queues AI title regeneration.
func (s *Service) RegenerateTitle(ctx context.Context, storyID string) error {
	s.logger.Info("story.RegenerateTitle", "storyID", storyID)
	return s.api.RegenerateTitle(ctx, storyID)
}

// RegenerateStaleNarration auto-detects and regenerates stale narration sections.
func (s *Service) RegenerateStaleNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.RegenerateStaleNarration", "storyID", storyID)
	return s.api.RegenerateStaleNarration(ctx, storyID)
}

// AcknowledgeNarrationStale dismisses narration staleness in bulk.
func (s *Service) AcknowledgeNarrationStale(ctx context.Context, storyID string) error {
	s.logger.Info("story.AcknowledgeNarrationStale", "storyID", storyID)
	return s.api.AcknowledgeNarrationStale(ctx, storyID)
}

// ResolveVanityURL resolves a vanity URL (@handle/slug) to story metadata.
func (s *Service) ResolveVanityURL(ctx context.Context, handle, slug string) (json.RawMessage, error) {
	s.logger.Info("story.ResolveVanityURL", "handle", handle, "slug", slug)
	return s.api.ResolveVanityURL(ctx, handle, slug)
}

// ListGenres returns all available genres.
func (s *Service) ListGenres(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("story.ListGenres")
	return s.api.ListGenres(ctx)
}

// StartNarration triggers narration generation for a story.
// Pass voice="" to use the server default.
//
// narration_start is destructive on a story that already has narration: the
// backend deletes the existing track and re-narrates from scratch. To prevent
// an accidental delete + credit burn (proseforge-workbench#238), this refuses
// to run when the story already has a non-in-flight narration unless
// confirmReplace is true. The guard is a convenience safety net, not an
// authorization control — it fails open if the narration state can't be read.
//
// confirmReplace also forwards to the backend as the #477 force flag: the
// backend now independently 409s (narration_exists) on a terminal-state
// narration unless force is set, so a deliberate replace must send it or the
// call fails server-side (proseforge-workbench#245). An in-flight narration is
// never force-deletable — the backend returns 409 conflict regardless.
// emphasis is the tri-state #642 prosody toggle: nil omits the field so the
// TTS worker applies its own default; true/false pin it for this narration.
func (s *Service) StartNarration(ctx context.Context, storyID, voice string, confirmReplace bool, emphasis *bool) error {
	s.logger.Info("story.StartNarration", "storyID", storyID, "voice", voice, "confirmReplace", confirmReplace)
	if !confirmReplace {
		if state, guarded := s.guardableNarration(ctx, storyID); guarded {
			return fmt.Errorf("story already has narration (status: %s) — narration_start deletes it and re-narrates from scratch, double-burning credits. "+
				"To change voice without losing the existing track, regenerate sections with the new voice (narration_regenerate with voice=…) then narration_rebuild. "+
				"To restart a failed narration, use narration_resume or narration_retry. "+
				"To intentionally delete and re-narrate anyway, pass confirm_replace=true.", state)
		}
	}
	return s.api.StartNarration(ctx, storyID, voice, confirmReplace, emphasis)
}

// guardableNarration reports whether the story already has a non-in-flight
// narration that narration_start would destructively replace. Returns the
// narration status and true when the guard should fire. Fails open (false)
// when there is no narration (404) or the state can't be read.
func (s *Service) guardableNarration(ctx context.Context, storyID string) (string, bool) {
	raw, err := s.api.GetNarration(ctx, storyID)
	if err != nil {
		return "", false // no narration (404) or transient read failure — don't block
	}
	var st struct {
		Status      string `json:"status"`
		NarrationID string `json:"narration_id"`
	}
	if json.Unmarshal(raw, &st) != nil {
		return "", false
	}
	if st.NarrationID == "" && st.Status == "" {
		return "", false // no existing narration record
	}
	switch st.Status {
	case "pending", "processing", "in_progress", "in-progress", "generating", "queued", "running", "started":
		return "", false // in-flight — excluded from the guard per #238
	default:
		return st.Status, true // complete / error / partial / unknown-terminal → guard
	}
}

// GetNarration returns narration status and per-section details.
func (s *Service) GetNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.GetNarration", "storyID", storyID)
	return s.api.GetNarration(ctx, storyID)
}

// RegenerateSection re-narrates a specific section.
// If force is true, regenerates even if content hasn't changed.
// If voice is non-empty, overrides the narration-level voice for this section.
func (s *Service) RegenerateSection(ctx context.Context, storyID, sectionID string, force bool, voice string) error {
	s.logger.Info("story.RegenerateSection", "storyID", storyID, "sectionID", sectionID, "force", force, "voice", voice)
	return s.api.RegenerateSection(ctx, storyID, sectionID, force, voice)
}

// ListVoices returns available TTS voices across all providers.
func (s *Service) ListVoices(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("story.ListVoices")
	return s.api.ListVoices(ctx)
}

// RetrySection resets a failed/stuck section and re-queues it.
func (s *Service) RetrySection(ctx context.Context, storyID, sectionID string) error {
	s.logger.Info("story.RetrySection", "storyID", storyID, "sectionID", sectionID)
	return s.api.RetrySection(ctx, storyID, sectionID)
}

// RebuildNarration rebuilds the audiobook from existing per-section audio.
// If sectionAnnouncements is true, TTS-generated section title announcements are inserted.
// emphasis follows the #642 tri-state: nil leaves the stored narration option
// untouched (partial-merge on the backend); true/false updates it — note the
// new value only affects audio on a subsequent re-TTS, not the free rebuild.
func (s *Service) RebuildNarration(ctx context.Context, storyID string, sectionAnnouncements bool, emphasis *bool) error {
	s.logger.Info("story.RebuildNarration", "storyID", storyID, "sectionAnnouncements", sectionAnnouncements)
	return s.api.RebuildNarration(ctx, storyID, sectionAnnouncements, emphasis)
}

// DeleteNarration deletes all narration data for a story.
func (s *Service) DeleteNarration(ctx context.Context, storyID string) error {
	s.logger.Info("story.DeleteNarration", "storyID", storyID)
	return s.api.DeleteNarration(ctx, storyID)
}

// ResumeNarration resumes a stuck narration.
func (s *Service) ResumeNarration(ctx context.Context, storyID string) error {
	s.logger.Info("story.ResumeNarration", "storyID", storyID)
	return s.api.ResumeNarration(ctx, storyID)
}

// CancelSection cancels a specific section's narration.
func (s *Service) CancelSection(ctx context.Context, storyID, sectionID string) error {
	s.logger.Info("story.CancelSection", "storyID", storyID, "sectionID", sectionID)
	return s.api.CancelSection(ctx, storyID, sectionID)
}

// ListSegments returns segment details for a section.
func (s *Service) ListSegments(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
	s.logger.Info("story.ListSegments", "storyID", storyID, "sectionID", sectionID)
	return s.api.ListSegments(ctx, storyID, sectionID)
}

// RegenerateSegment re-narrates a single segment within a section.
// If voice is non-empty, overrides the voice for this segment only.
func (s *Service) RegenerateSegment(ctx context.Context, storyID, sectionID, segmentID, voice string) error {
	s.logger.Info("story.RegenerateSegment", "storyID", storyID, "sectionID", sectionID, "segmentID", segmentID, "voice", voice)
	return s.api.RegenerateSegment(ctx, storyID, sectionID, segmentID, voice)
}

// PatchNarration patches multiple segments and/or sections in one call, then rebuilds the audiobook.
func (s *Service) PatchNarration(ctx context.Context, storyID string, req gen.HandlersPatchNarrationRequest) (json.RawMessage, error) {
	s.logger.Info("story.PatchNarration", "storyID", storyID)
	return s.api.PatchNarration(ctx, storyID, req)
}

// EstimateCredits returns the estimated cost for an operation before executing it.
func (s *Service) EstimateCredits(ctx context.Context, params *gen.GetCreditsEstimateParams) (json.RawMessage, error) {
	s.logger.Info("story.EstimateCredits", "operation", params.Operation)
	return s.api.EstimateCredits(ctx, params)
}

// GetCreditHistory returns recent credit transactions.
func (s *Service) GetCreditHistory(ctx context.Context, limit int) (json.RawMessage, error) {
	s.logger.Info("story.GetCreditHistory", "limit", limit)
	return s.api.GetCreditHistory(ctx, limit)
}

// GetCredits returns the authenticated user's credit balance.
func (s *Service) GetCredits(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("story.GetCredits")
	return s.api.GetCredits(ctx)
}

// GetAudiobook returns audiobook download info for a story.
func (s *Service) GetAudiobook(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.GetAudiobook", "storyID", storyID)
	return s.api.GetAudiobook(ctx, storyID)
}

// GenerateImage triggers async AI image generation. Returns 202 — poll with GetImage.
func (s *Service) GenerateImage(ctx context.Context, req gen.HandlersGenerateImageRequest) (json.RawMessage, error) {
	s.logger.Info("story.GenerateImage", "storyId", req.StoryId, "userPrompt", req.UserPrompt)
	return s.api.GenerateImage(ctx, req)
}

// UploadImage uploads a pre-made image via multipart form.
func (s *Service) UploadImage(ctx context.Context, contentType string, body io.Reader) (json.RawMessage, error) {
	s.logger.Info("story.UploadImage")
	return s.api.UploadImage(ctx, contentType, body)
}

// UploadAndAttachImage uploads a pre-made image, attaches it to a story,
// and optionally sets it as the cover image. Caller provides the multipart
// content type and body (use buildMultipartUpload from the MCP tools package).
func (s *Service) UploadAndAttachImage(ctx context.Context, storyID string, contentType string, body io.Reader, cover bool) (json.RawMessage, error) {
	s.logger.Info("story.UploadAndAttachImage", "storyID", storyID, "cover", cover)

	result, err := s.api.UploadImage(ctx, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("upload image: %w", err)
	}

	// Parse image ID from upload response.
	// The response nests the ID under {"image": {"id": "..."}}.
	var uploaded struct {
		Image struct {
			ID string `json:"id"`
		} `json:"image"`
	}
	if err := json.Unmarshal(result, &uploaded); err != nil {
		return nil, fmt.Errorf("parse upload response: %w", err)
	}
	if uploaded.Image.ID == "" {
		return nil, fmt.Errorf("upload succeeded but response contained no image ID")
	}

	// Attach to story
	attachReq := gen.HandlersAddToStoryRequest{}
	if cover {
		t := true
		attachReq.IsPrimary = &t
	}
	if err := s.api.AttachImageToStory(ctx, storyID, uploaded.Image.ID, attachReq); err != nil {
		return nil, fmt.Errorf("attach image: %w", err)
	}

	return result, nil
}

// GetImage returns image details and generation status.
func (s *Service) GetImage(ctx context.Context, id string) (json.RawMessage, error) {
	s.logger.Info("story.GetImage", "id", id)
	return s.api.GetImage(ctx, id)
}

// ListImages returns the user's image library with pagination.
func (s *Service) ListImages(ctx context.Context, params *gen.GetImagesParams) (json.RawMessage, error) {
	s.logger.Info("story.ListImages")
	return s.api.ListImages(ctx, params)
}

// RegenerateImage re-rolls an existing image with optional new prompt. Returns 202.
func (s *Service) RegenerateImage(ctx context.Context, id string, req gen.HandlersRegenerateRequest) (json.RawMessage, error) {
	s.logger.Info("story.RegenerateImage", "id", id)
	return s.api.RegenerateImage(ctx, id, req)
}

// GenerateImageAndWait fires an image generation, polls until the job reaches
// a terminal state (completed/failed/cancelled), and returns the final image
// details. On timeout, returns the most recent image details with a non-nil
// error and the image_id so the caller can keep polling with image_get.
func (s *Service) GenerateImageAndWait(ctx context.Context, req gen.HandlersGenerateImageRequest, timeout time.Duration) (json.RawMessage, string, error) {
	s.logger.Info("story.GenerateImageAndWait", "storyId", req.StoryId, "timeout", timeout)

	initial, err := s.api.GenerateImage(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("generate: %w", err)
	}

	imageID := extractImageID(initial)
	if imageID == "" {
		return initial, "", fmt.Errorf("generate succeeded but response contained no image id; raw response returned for inspection")
	}

	deadline := time.Now().Add(timeout)
	interval := 2 * time.Second
	var last json.RawMessage = initial

	for {
		select {
		case <-ctx.Done():
			return last, imageID, ctx.Err()
		case <-time.After(interval):
		}

		current, err := s.api.GetImage(ctx, imageID)
		if err != nil {
			if time.Now().After(deadline) {
				return last, imageID, fmt.Errorf("polling failed and timeout reached: %w; image_id=%s", err, imageID)
			}
			continue
		}
		last = current

		status := extractImageStatus(current)
		if status == "completed" || status == "failed" || status == "cancelled" {
			return current, imageID, nil
		}

		if time.Now().After(deadline) {
			return last, imageID, fmt.Errorf("timeout after %s; image_id=%s still %q — keep polling with image_get", timeout, imageID, status)
		}
	}
}

// GenerateImageBatch fires N image generations in parallel and returns the
// initial 202 responses indexed in input order. Each entry is either a raw
// generate response (with image id for polling) or an error. Use
// GenerateImageAndWait for the wait-for-completion variant.
func (s *Service) GenerateImageBatch(ctx context.Context, requests []gen.HandlersGenerateImageRequest) ([]json.RawMessage, []error) {
	s.logger.Info("story.GenerateImageBatch", "count", len(requests))

	results := make([]json.RawMessage, len(requests))
	errs := make([]error, len(requests))

	var wg sync.WaitGroup
	for i, req := range requests {
		wg.Add(1)
		go func(idx int, r gen.HandlersGenerateImageRequest) {
			defer wg.Done()
			body, err := s.api.GenerateImage(ctx, r)
			results[idx] = body
			errs[idx] = err
		}(i, req)
	}
	wg.Wait()

	return results, errs
}

// extractImageID pulls the image id from a generate response, accepting the
// common shapes the API has used: top-level id, imageId, or nested image.id.
func extractImageID(body json.RawMessage) string {
	var top struct {
		Id      *string `json:"id"`
		ImageId *string `json:"imageId"`
		Image   *struct {
			Id *string `json:"id"`
		} `json:"image"`
	}
	if err := json.Unmarshal(body, &top); err != nil {
		return ""
	}
	if top.Id != nil && *top.Id != "" {
		return *top.Id
	}
	if top.ImageId != nil && *top.ImageId != "" {
		return *top.ImageId
	}
	if top.Image != nil && top.Image.Id != nil && *top.Image.Id != "" {
		return *top.Image.Id
	}
	return ""
}

// extractImageStatus pulls the status field from an image response.
func extractImageStatus(body json.RawMessage) string {
	var top struct {
		Status *string `json:"status"`
		Image  *struct {
			Status *string `json:"status"`
		} `json:"image"`
	}
	if err := json.Unmarshal(body, &top); err != nil {
		return ""
	}
	if top.Status != nil {
		return *top.Status
	}
	if top.Image != nil && top.Image.Status != nil {
		return *top.Image.Status
	}
	return ""
}

// ListStoryImages returns images attached to a story. Pass includeSections=true
// to also include images attached to the story's sections in one call.
func (s *Service) ListStoryImages(ctx context.Context, storyID string, includeSections bool) (json.RawMessage, error) {
	s.logger.Info("story.ListStoryImages", "storyID", storyID, "includeSections", includeSections)
	return s.api.ListStoryImages(ctx, storyID, includeSections)
}

// AttachImageToStory attaches an image to a story.
func (s *Service) AttachImageToStory(ctx context.Context, storyID, imageID string, req gen.HandlersAddToStoryRequest) error {
	s.logger.Info("story.AttachImageToStory", "storyID", storyID, "imageID", imageID)
	return s.api.AttachImageToStory(ctx, storyID, imageID, req)
}

// SetStoryImageCover sets an attached image as the story's primary/cover image.
func (s *Service) SetStoryImageCover(ctx context.Context, storyID, imageID string) error {
	s.logger.Info("story.SetStoryImageCover", "storyID", storyID, "imageID", imageID)
	return s.api.SetStoryImageCover(ctx, storyID, imageID)
}

// ListVersions returns version history (git commits) for a story.
func (s *Service) ListVersions(ctx context.Context, storyID string, params *gen.GetStoryIdVersionsParams) (json.RawMessage, error) {
	s.logger.Info("story.ListVersions", "storyID", storyID)
	return s.api.ListVersions(ctx, storyID, params)
}

// GetVersion returns story content at a specific version (git SHA).
func (s *Service) GetVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	s.logger.Info("story.GetVersion", "storyID", storyID, "sha", sha)
	return s.api.GetVersion(ctx, storyID, sha)
}

// DiffVersions returns the diff between two story versions.
func (s *Service) DiffVersions(ctx context.Context, storyID, fromSha, toSha string) (json.RawMessage, error) {
	s.logger.Info("story.DiffVersions", "storyID", storyID, "fromSha", fromSha, "toSha", toSha)
	return s.api.DiffVersions(ctx, storyID, fromSha, toSha)
}

// ResolveGenreID looks up a genre by name (case-insensitive) and returns its ID.
func (s *Service) ResolveGenreID(ctx context.Context, name string) (string, error) {
	s.logger.Info("story.ResolveGenreID", "name", name)
	data, err := s.ListGenres(ctx)
	if err != nil {
		return "", fmt.Errorf("resolve genre %q: %w", name, err)
	}

	var genres []api.Genre
	if err := json.Unmarshal(data, &genres); err != nil {
		return "", fmt.Errorf("resolve genre %q: parse response: %w", name, err)
	}

	target := strings.ToLower(strings.TrimSpace(name))
	for _, g := range genres {
		if g.Name != nil && strings.ToLower(*g.Name) == target {
			if g.Id == nil {
				return "", fmt.Errorf("genre %q has no ID", name)
			}
			return *g.Id, nil
		}
	}

	var available []string
	for _, g := range genres {
		if g.Name != nil {
			available = append(available, *g.Name)
		}
	}
	return "", fmt.Errorf("genre %q not found; available: %s", name, strings.Join(available, ", "))
}
