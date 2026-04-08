// Package story provides story listing and reading operations.
package story

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

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

// WriteSection updates a section's content.
func (s *Service) WriteSection(ctx context.Context, storyID, sectionID string, req api.UpdateSectionRequest) error {
	s.logger.Info("story.WriteSection", "storyID", storyID, "sectionID", sectionID)
	return s.api.WriteSection(ctx, storyID, sectionID, req)
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

// RegenerateStaleNarration auto-detects and regenerates stale narration chapters.
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
func (s *Service) StartNarration(ctx context.Context, storyID, voice string) error {
	s.logger.Info("story.StartNarration", "storyID", storyID, "voice", voice)
	return s.api.StartNarration(ctx, storyID, voice)
}

// GetNarration returns narration status and chapter details.
func (s *Service) GetNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.GetNarration", "storyID", storyID)
	return s.api.GetNarration(ctx, storyID)
}

// RegenerateChapter re-narrates a specific chapter.
// If force is true, regenerates even if content hasn't changed.
// If voice is non-empty, overrides the narration-level voice for this chapter.
func (s *Service) RegenerateChapter(ctx context.Context, storyID, chapterID string, force bool, voice string) error {
	s.logger.Info("story.RegenerateChapter", "storyID", storyID, "chapterID", chapterID, "force", force, "voice", voice)
	return s.api.RegenerateChapter(ctx, storyID, chapterID, force, voice)
}

// ListVoices returns available TTS voices across all providers.
func (s *Service) ListVoices(ctx context.Context) (json.RawMessage, error) {
	s.logger.Info("story.ListVoices")
	return s.api.ListVoices(ctx)
}

// RetryChapter resets a failed/stuck chapter and re-queues it.
func (s *Service) RetryChapter(ctx context.Context, storyID, chapterID string) error {
	s.logger.Info("story.RetryChapter", "storyID", storyID, "chapterID", chapterID)
	return s.api.RetryChapter(ctx, storyID, chapterID)
}

// RebuildNarration rebuilds the audiobook from existing chapter audio.
// If chapterAnnouncements is true, TTS-generated chapter title announcements are inserted.
func (s *Service) RebuildNarration(ctx context.Context, storyID string, chapterAnnouncements bool) error {
	s.logger.Info("story.RebuildNarration", "storyID", storyID, "chapterAnnouncements", chapterAnnouncements)
	return s.api.RebuildNarration(ctx, storyID, chapterAnnouncements)
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

// CancelChapter cancels a specific chapter's narration.
func (s *Service) CancelChapter(ctx context.Context, storyID, chapterID string) error {
	s.logger.Info("story.CancelChapter", "storyID", storyID, "chapterID", chapterID)
	return s.api.CancelChapter(ctx, storyID, chapterID)
}

// ListSegments returns segment details for a chapter.
func (s *Service) ListSegments(ctx context.Context, storyID, chapterID string) (json.RawMessage, error) {
	s.logger.Info("story.ListSegments", "storyID", storyID, "chapterID", chapterID)
	return s.api.ListSegments(ctx, storyID, chapterID)
}

// RegenerateSegment re-narrates a single segment within a chapter.
// If voice is non-empty, overrides the voice for this segment only.
func (s *Service) RegenerateSegment(ctx context.Context, storyID, chapterID, segmentID, voice string) error {
	s.logger.Info("story.RegenerateSegment", "storyID", storyID, "chapterID", chapterID, "segmentID", segmentID, "voice", voice)
	return s.api.RegenerateSegment(ctx, storyID, chapterID, segmentID, voice)
}

// PatchNarration patches multiple segments and/or chapters in one call, then rebuilds the audiobook.
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

// ListStoryImages returns images attached to a story.
func (s *Service) ListStoryImages(ctx context.Context, storyID string) (json.RawMessage, error) {
	s.logger.Info("story.ListStoryImages", "storyID", storyID)
	return s.api.ListStoryImages(ctx, storyID)
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
