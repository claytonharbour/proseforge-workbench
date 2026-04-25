package api

import (
	"context"
	"encoding/json"
	"io"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// MockClient implements ProseForgeAPI with function fields for testing.
// Each function field can be set to control the mock's behavior. If a function
// field is nil, calling the method will panic (indicating a missing test setup).
type MockClient struct {
	// Stories
	ListStoriesFn                  func(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error)
	GetStoryFn                     func(ctx context.Context, id string) (*Story, error)
	GetStoryWithContentFn          func(ctx context.Context, id string) (*Story, error)
	CreateStoryFn                  func(ctx context.Context, req CreateStoryRequest) (*Story, error)
	UpdateStoryFn                  func(ctx context.Context, id string, req UpdateStoryRequest) error
	PublishStoryFn                 func(ctx context.Context, id string, visibility string) error
	UnpublishStoryFn               func(ctx context.Context, id string) error
	CreatePitchFn                  func(ctx context.Context, req CreateStoryRequest) (*Story, error)
	PromoteStoryFn                 func(ctx context.Context, id string) error
	UpsertStoryMetaFn              func(ctx context.Context, storyID, metaType, content string) (json.RawMessage, error)
	DeleteStoryFn                  func(ctx context.Context, id string) error
	DeleteSectionFn                func(ctx context.Context, storyID, sectionID string) error
	RestoreVersionFn               func(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	GetMetaStaleFn                 func(ctx context.Context, storyID string) (json.RawMessage, error)
	AcknowledgeMetaStaleFn         func(ctx context.Context, storyID string) error
	RegenerateTaglineFn            func(ctx context.Context, storyID string) error
	RegenerateTitleFn              func(ctx context.Context, storyID string) error
	RegenerateStaleNarrationFn     func(ctx context.Context, storyID string) (json.RawMessage, error)
	AcknowledgeNarrationStaleFn    func(ctx context.Context, storyID string) error
	UpdateVisibilityFn             func(ctx context.Context, id string, visibility string) error
	DownloadStoryFn                func(ctx context.Context, id string, format string) (string, error)
	ResolveVanityURLFn             func(ctx context.Context, handle, slug string) (json.RawMessage, error)
	ListStoriesWithReviewStatusFn  func(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error)
	ListVersionsFn                func(ctx context.Context, storyID string, params *gen.GetStoryIdVersionsParams) (json.RawMessage, error)
	GetVersionFn                  func(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	DiffVersionsFn                func(ctx context.Context, storyID, fromSha, toSha string) (json.RawMessage, error)

	// Sections
	GetSectionFn     func(ctx context.Context, storyID, sectionID string) (json.RawMessage, error)
	CreateSectionFn  func(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error)
	WriteSectionFn   func(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) error

	// Genres
	ListGenresFn func(ctx context.Context) (json.RawMessage, error)

	// Quality & Insights
	GetQualityFn              func(ctx context.Context, storyID string) (json.RawMessage, error)
	AssessQualityFn           func(ctx context.Context, storyID string, force bool) (json.RawMessage, error)
	AssessQualityAtVersionFn  func(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	GetInsightsFn             func(ctx context.Context, storyID string) (json.RawMessage, error)

	// Reviews
	AddReviewerFn        func(ctx context.Context, storyID string, req AddReviewerRequest) (*Reviewer, error)
	ListReviewersFn      func(ctx context.Context, storyID string) (*ReviewersList, error)
	AcceptReviewFn       func(ctx context.Context, reviewID string) error
	DeclineReviewFn      func(ctx context.Context, reviewID string) error
	ApproveStoryFn       func(ctx context.Context, reviewID string) error
	RejectStoryFn        func(ctx context.Context, reviewID string, req ReviewFeedbackRequest) error
	ListPendingReviewsFn func(ctx context.Context, params *gen.GetReviewsPendingParams) (*PendingReviews, error)

	// Feedback
	GetFeedbackReviewsFn      func(ctx context.Context, storyID string) (*FeedbackReviewList, error)
	GetFeedbackReviewFn       func(ctx context.Context, storyID, reviewID string, include ...string) (*FeedbackReview, error)
	GetFeedbackReviewFullFn   func(ctx context.Context, storyID, reviewID string) (*FeedbackReviewWithItems, error)
	GetFeedbackDiffFn         func(ctx context.Context, storyID, reviewID string) (*ReviewDiffResponse, error)
	GetFeedbackSuggestionsFn  func(ctx context.Context, storyID, reviewID string) (*FullFeedback, error)
	CreateFeedbackReviewFn    func(ctx context.Context, storyID string, req StartAIReviewRequest) (*FeedbackReview, error)
	AddFeedbackItemFn         func(ctx context.Context, storyID, reviewID string, req AddFeedbackItemRequest) error
	SubmitReviewFn            func(ctx context.Context, reviewID string) error
	UpdateSectionContentFn    func(ctx context.Context, storyID, reviewID, sectionID, content string) error
	IncorporateFeedbackFn     func(ctx context.Context, storyID, reviewID string, req IncorporateRequest) error

	// Reviewer Pool
	RequestReviewerFn            func(ctx context.Context, req CreateReviewerRequestReq) error
	RespondToReviewerRequestFn   func(ctx context.Context, requestID string, req RespondToReviewerReq) error
	ListAvailableReviewersFn     func(ctx context.Context) (*AvailableReviewerList, error)
	ListMyReviewersFn            func(ctx context.Context) ([]Reviewer, error)

	// Narration
	StartNarrationFn     func(ctx context.Context, storyID string, voice string) error
	GetNarrationFn       func(ctx context.Context, storyID string) (json.RawMessage, error)
	GetAudiobookFn       func(ctx context.Context, storyID string) (json.RawMessage, error)
	RegenerateChapterFn  func(ctx context.Context, storyID, chapterID string, force bool, voice string) error
	RetryChapterFn       func(ctx context.Context, storyID, chapterID string) error
	RebuildNarrationFn   func(ctx context.Context, storyID string, chapterAnnouncements bool) error
	DeleteNarrationFn    func(ctx context.Context, storyID string) error
	ResumeNarrationFn    func(ctx context.Context, storyID string) error
	CancelChapterFn      func(ctx context.Context, storyID, chapterID string) error
	ListVoicesFn            func(ctx context.Context) (json.RawMessage, error)
	ListSegmentsFn          func(ctx context.Context, storyID, chapterID string) (json.RawMessage, error)
	RegenerateSegmentFn     func(ctx context.Context, storyID, chapterID, segmentID, voice string) error
	PatchNarrationFn        func(ctx context.Context, storyID string, req gen.HandlersPatchNarrationRequest) (json.RawMessage, error)

	// Credits
	GetCreditsFn       func(ctx context.Context) (json.RawMessage, error)
	EstimateCreditsFn  func(ctx context.Context, params *gen.GetCreditsEstimateParams) (json.RawMessage, error)
	GetCreditHistoryFn func(ctx context.Context, limit int) (json.RawMessage, error)

	// Series
	ListSeriesFn    func(ctx context.Context) (*SeriesList, error)
	CreateSeriesFn  func(ctx context.Context, req CreateSeriesReq) (*Series, error)
	GetSeriesByIDFn func(ctx context.Context, id string) (*Series, error)
	UpdateSeriesFn  func(ctx context.Context, id string, req UpdateSeriesReq) error
	ArchiveSeriesFn func(ctx context.Context, id string) error
	GetWorldFn      func(ctx context.Context, seriesID string) (json.RawMessage, error)
	UpdateWorldFn   func(ctx context.Context, seriesID string, content string) error
	GetTimelineFn              func(ctx context.Context, seriesID string) (json.RawMessage, error)
	ListTimelineSectionsFn     func(ctx context.Context, seriesID string) (json.RawMessage, error)
	GetTimelineSectionFn       func(ctx context.Context, seriesID, slug string) (json.RawMessage, error)
	UpdateTimelineSectionFn    func(ctx context.Context, seriesID, slug, title, content string) (json.RawMessage, error)
	DeleteTimelineSectionFn    func(ctx context.Context, seriesID, slug string) error
	ReorderSeriesStoriesFn     func(ctx context.Context, seriesID string, storyIDs []string) error
	ReorderTimelineSectionsFn  func(ctx context.Context, seriesID string, slugs []string) error

	// Series Characters
	CreateCharacterFn func(ctx context.Context, seriesID string, req CreateCharacterReq) (*Character, error)
	ListCharactersFn  func(ctx context.Context, seriesID string) (*CharacterList, error)
	GetCharacterFn    func(ctx context.Context, seriesID, slug string) (*Character, error)
	UpdateCharacterFn func(ctx context.Context, seriesID, slug string, req UpdateCharacterReq) error
	DeleteCharacterFn func(ctx context.Context, seriesID, slug string) error

	// Series Stories
	ListSeriesStoriesFn      func(ctx context.Context, seriesID string) (json.RawMessage, error)
	AddStoryToSeriesFn       func(ctx context.Context, seriesID, storyID string) error
	RemoveStoryFromSeriesFn  func(ctx context.Context, seriesID, storyID string) error

	// Story Forge Pipeline
	GetStoryMetaFn         func(ctx context.Context, storyID string) (json.RawMessage, error)

	// Series Plan
	PlanStoryFn func(ctx context.Context, seriesID string, req PlanStoryReq) (*PlanStoryResp, error)

	// Images
	GenerateImageFn       func(ctx context.Context, req gen.HandlersGenerateImageRequest) (json.RawMessage, error)
	UploadImageFn         func(ctx context.Context, contentType string, body io.Reader) (json.RawMessage, error)
	GetImageFn            func(ctx context.Context, id string) (json.RawMessage, error)
	ListImagesFn          func(ctx context.Context, params *gen.GetImagesParams) (json.RawMessage, error)
	RegenerateImageFn     func(ctx context.Context, id string, req gen.HandlersRegenerateRequest) (json.RawMessage, error)
	ListStoryImagesFn     func(ctx context.Context, storyID string) (json.RawMessage, error)
	AttachImageToStoryFn  func(ctx context.Context, storyID, imageID string, req gen.HandlersAddToStoryRequest) error
	SetStoryImageCoverFn  func(ctx context.Context, storyID, imageID string) error
}

// Compile-time assertion: *MockClient implements ProseForgeAPI.
var _ ProseForgeAPI = (*MockClient)(nil)

// Stories

func (m *MockClient) ListStories(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error) {
	return m.ListStoriesFn(ctx, params)
}

func (m *MockClient) GetStory(ctx context.Context, id string) (*Story, error) {
	return m.GetStoryFn(ctx, id)
}

func (m *MockClient) GetStoryWithContent(ctx context.Context, id string) (*Story, error) {
	return m.GetStoryWithContentFn(ctx, id)
}

func (m *MockClient) DownloadStory(ctx context.Context, id string, format string) (string, error) {
	return m.DownloadStoryFn(ctx, id, format)
}

func (m *MockClient) ResolveVanityURL(ctx context.Context, handle, slug string) (json.RawMessage, error) {
	return m.ResolveVanityURLFn(ctx, handle, slug)
}

func (m *MockClient) ListStoriesWithReviewStatus(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error) {
	return m.ListStoriesWithReviewStatusFn(ctx, params)
}

func (m *MockClient) CreateStory(ctx context.Context, req CreateStoryRequest) (*Story, error) {
	return m.CreateStoryFn(ctx, req)
}

func (m *MockClient) UpdateStory(ctx context.Context, id string, req UpdateStoryRequest) error {
	return m.UpdateStoryFn(ctx, id, req)
}

func (m *MockClient) PublishStory(ctx context.Context, id string, visibility string) error {
	return m.PublishStoryFn(ctx, id, visibility)
}

func (m *MockClient) UnpublishStory(ctx context.Context, id string) error {
	return m.UnpublishStoryFn(ctx, id)
}

func (m *MockClient) CreatePitch(ctx context.Context, req CreateStoryRequest) (*Story, error) {
	return m.CreatePitchFn(ctx, req)
}

func (m *MockClient) PromoteStory(ctx context.Context, id string) error {
	return m.PromoteStoryFn(ctx, id)
}

func (m *MockClient) UpsertStoryMeta(ctx context.Context, storyID, metaType, content string) (json.RawMessage, error) {
	return m.UpsertStoryMetaFn(ctx, storyID, metaType, content)
}

func (m *MockClient) UpdateVisibility(ctx context.Context, id string, visibility string) error {
	return m.UpdateVisibilityFn(ctx, id, visibility)
}

func (m *MockClient) DeleteStory(ctx context.Context, id string) error {
	return m.DeleteStoryFn(ctx, id)
}

func (m *MockClient) DeleteSection(ctx context.Context, storyID, sectionID string) error {
	return m.DeleteSectionFn(ctx, storyID, sectionID)
}

func (m *MockClient) RestoreVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	return m.RestoreVersionFn(ctx, storyID, sha)
}

func (m *MockClient) GetMetaStale(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.GetMetaStaleFn(ctx, storyID)
}

func (m *MockClient) AcknowledgeMetaStale(ctx context.Context, storyID string) error {
	return m.AcknowledgeMetaStaleFn(ctx, storyID)
}

func (m *MockClient) RegenerateTagline(ctx context.Context, storyID string) error {
	return m.RegenerateTaglineFn(ctx, storyID)
}

func (m *MockClient) RegenerateTitle(ctx context.Context, storyID string) error {
	return m.RegenerateTitleFn(ctx, storyID)
}

func (m *MockClient) RegenerateStaleNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.RegenerateStaleNarrationFn(ctx, storyID)
}

func (m *MockClient) AcknowledgeNarrationStale(ctx context.Context, storyID string) error {
	return m.AcknowledgeNarrationStaleFn(ctx, storyID)
}

func (m *MockClient) ListVersions(ctx context.Context, storyID string, params *gen.GetStoryIdVersionsParams) (json.RawMessage, error) {
	return m.ListVersionsFn(ctx, storyID, params)
}

func (m *MockClient) GetVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	return m.GetVersionFn(ctx, storyID, sha)
}

func (m *MockClient) DiffVersions(ctx context.Context, storyID, fromSha, toSha string) (json.RawMessage, error) {
	return m.DiffVersionsFn(ctx, storyID, fromSha, toSha)
}

// Sections

func (m *MockClient) GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error) {
	return m.GetSectionFn(ctx, storyID, sectionID)
}

func (m *MockClient) CreateSection(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error) {
	return m.CreateSectionFn(ctx, storyID, req)
}

func (m *MockClient) WriteSection(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) error {
	return m.WriteSectionFn(ctx, storyID, sectionID, req)
}

// Genres

func (m *MockClient) ListGenres(ctx context.Context) (json.RawMessage, error) {
	return m.ListGenresFn(ctx)
}

// Quality & Insights

func (m *MockClient) GetQuality(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.GetQualityFn(ctx, storyID)
}

func (m *MockClient) AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error) {
	return m.AssessQualityFn(ctx, storyID, force)
}

func (m *MockClient) AssessQualityAtVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error) {
	return m.AssessQualityAtVersionFn(ctx, storyID, sha)
}

func (m *MockClient) GetInsights(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.GetInsightsFn(ctx, storyID)
}

// Reviews

func (m *MockClient) AddReviewer(ctx context.Context, storyID string, req AddReviewerRequest) (*Reviewer, error) {
	return m.AddReviewerFn(ctx, storyID, req)
}

func (m *MockClient) ListReviewers(ctx context.Context, storyID string) (*ReviewersList, error) {
	return m.ListReviewersFn(ctx, storyID)
}

func (m *MockClient) AcceptReview(ctx context.Context, reviewID string) error {
	return m.AcceptReviewFn(ctx, reviewID)
}

func (m *MockClient) DeclineReview(ctx context.Context, reviewID string) error {
	return m.DeclineReviewFn(ctx, reviewID)
}

func (m *MockClient) ApproveStory(ctx context.Context, reviewID string) error {
	return m.ApproveStoryFn(ctx, reviewID)
}

func (m *MockClient) RejectStory(ctx context.Context, reviewID string, req ReviewFeedbackRequest) error {
	return m.RejectStoryFn(ctx, reviewID, req)
}

func (m *MockClient) ListPendingReviews(ctx context.Context, params *gen.GetReviewsPendingParams) (*PendingReviews, error) {
	return m.ListPendingReviewsFn(ctx, params)
}

// Feedback

func (m *MockClient) GetFeedbackReviews(ctx context.Context, storyID string) (*FeedbackReviewList, error) {
	return m.GetFeedbackReviewsFn(ctx, storyID)
}

func (m *MockClient) GetFeedbackReview(ctx context.Context, storyID, reviewID string, include ...string) (*FeedbackReview, error) {
	return m.GetFeedbackReviewFn(ctx, storyID, reviewID, include...)
}

func (m *MockClient) GetFeedbackReviewFull(ctx context.Context, storyID, reviewID string) (*FeedbackReviewWithItems, error) {
	return m.GetFeedbackReviewFullFn(ctx, storyID, reviewID)
}

func (m *MockClient) GetFeedbackDiff(ctx context.Context, storyID, reviewID string) (*ReviewDiffResponse, error) {
	return m.GetFeedbackDiffFn(ctx, storyID, reviewID)
}

func (m *MockClient) GetFeedbackSuggestions(ctx context.Context, storyID, reviewID string) (*FullFeedback, error) {
	return m.GetFeedbackSuggestionsFn(ctx, storyID, reviewID)
}

func (m *MockClient) CreateFeedbackReview(ctx context.Context, storyID string, req StartAIReviewRequest) (*FeedbackReview, error) {
	return m.CreateFeedbackReviewFn(ctx, storyID, req)
}

func (m *MockClient) AddFeedbackItem(ctx context.Context, storyID, reviewID string, req AddFeedbackItemRequest) error {
	return m.AddFeedbackItemFn(ctx, storyID, reviewID, req)
}

func (m *MockClient) SubmitReview(ctx context.Context, reviewID string) error {
	return m.SubmitReviewFn(ctx, reviewID)
}

func (m *MockClient) UpdateSectionContent(ctx context.Context, storyID, reviewID, sectionID, content string) error {
	return m.UpdateSectionContentFn(ctx, storyID, reviewID, sectionID, content)
}

func (m *MockClient) IncorporateFeedback(ctx context.Context, storyID, reviewID string, req IncorporateRequest) error {
	return m.IncorporateFeedbackFn(ctx, storyID, reviewID, req)
}

// Reviewer Pool

func (m *MockClient) RequestReviewer(ctx context.Context, req CreateReviewerRequestReq) error {
	return m.RequestReviewerFn(ctx, req)
}

func (m *MockClient) RespondToReviewerRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error {
	return m.RespondToReviewerRequestFn(ctx, requestID, req)
}

func (m *MockClient) ListAvailableReviewers(ctx context.Context) (*AvailableReviewerList, error) {
	return m.ListAvailableReviewersFn(ctx)
}

func (m *MockClient) ListMyReviewers(ctx context.Context) ([]Reviewer, error) {
	return m.ListMyReviewersFn(ctx)
}

// Narration

func (m *MockClient) StartNarration(ctx context.Context, storyID string, voice string) error {
	return m.StartNarrationFn(ctx, storyID, voice)
}

func (m *MockClient) GetNarration(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.GetNarrationFn(ctx, storyID)
}

func (m *MockClient) GetAudiobook(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.GetAudiobookFn(ctx, storyID)
}

func (m *MockClient) RegenerateChapter(ctx context.Context, storyID, chapterID string, force bool, voice string) error {
	return m.RegenerateChapterFn(ctx, storyID, chapterID, force, voice)
}

func (m *MockClient) RetryChapter(ctx context.Context, storyID, chapterID string) error {
	return m.RetryChapterFn(ctx, storyID, chapterID)
}

func (m *MockClient) RebuildNarration(ctx context.Context, storyID string, chapterAnnouncements bool) error {
	return m.RebuildNarrationFn(ctx, storyID, chapterAnnouncements)
}

func (m *MockClient) DeleteNarration(ctx context.Context, storyID string) error {
	return m.DeleteNarrationFn(ctx, storyID)
}

func (m *MockClient) ResumeNarration(ctx context.Context, storyID string) error {
	return m.ResumeNarrationFn(ctx, storyID)
}

func (m *MockClient) CancelChapter(ctx context.Context, storyID, chapterID string) error {
	return m.CancelChapterFn(ctx, storyID, chapterID)
}

func (m *MockClient) ListVoices(ctx context.Context) (json.RawMessage, error) {
	return m.ListVoicesFn(ctx)
}

func (m *MockClient) ListSegments(ctx context.Context, storyID, chapterID string) (json.RawMessage, error) {
	return m.ListSegmentsFn(ctx, storyID, chapterID)
}

func (m *MockClient) RegenerateSegment(ctx context.Context, storyID, chapterID, segmentID, voice string) error {
	return m.RegenerateSegmentFn(ctx, storyID, chapterID, segmentID, voice)
}

func (m *MockClient) PatchNarration(ctx context.Context, storyID string, req gen.HandlersPatchNarrationRequest) (json.RawMessage, error) {
	return m.PatchNarrationFn(ctx, storyID, req)
}

// Credits

func (m *MockClient) GetCredits(ctx context.Context) (json.RawMessage, error) {
	return m.GetCreditsFn(ctx)
}

func (m *MockClient) EstimateCredits(ctx context.Context, params *gen.GetCreditsEstimateParams) (json.RawMessage, error) {
	return m.EstimateCreditsFn(ctx, params)
}

func (m *MockClient) GetCreditHistory(ctx context.Context, limit int) (json.RawMessage, error) {
	return m.GetCreditHistoryFn(ctx, limit)
}

// Series

func (m *MockClient) ListSeries(ctx context.Context) (*SeriesList, error) {
	return m.ListSeriesFn(ctx)
}

func (m *MockClient) CreateSeries(ctx context.Context, req CreateSeriesReq) (*Series, error) {
	return m.CreateSeriesFn(ctx, req)
}

func (m *MockClient) GetSeriesByID(ctx context.Context, id string) (*Series, error) {
	return m.GetSeriesByIDFn(ctx, id)
}

func (m *MockClient) UpdateSeries(ctx context.Context, id string, req UpdateSeriesReq) error {
	return m.UpdateSeriesFn(ctx, id, req)
}

func (m *MockClient) ArchiveSeries(ctx context.Context, id string) error {
	return m.ArchiveSeriesFn(ctx, id)
}

func (m *MockClient) GetWorld(ctx context.Context, seriesID string) (json.RawMessage, error) {
	return m.GetWorldFn(ctx, seriesID)
}

func (m *MockClient) UpdateWorld(ctx context.Context, seriesID string, content string) error {
	return m.UpdateWorldFn(ctx, seriesID, content)
}

func (m *MockClient) GetTimeline(ctx context.Context, seriesID string) (json.RawMessage, error) {
	return m.GetTimelineFn(ctx, seriesID)
}

func (m *MockClient) ListTimelineSections(ctx context.Context, seriesID string) (json.RawMessage, error) {
	return m.ListTimelineSectionsFn(ctx, seriesID)
}

func (m *MockClient) GetTimelineSection(ctx context.Context, seriesID, slug string) (json.RawMessage, error) {
	return m.GetTimelineSectionFn(ctx, seriesID, slug)
}

func (m *MockClient) UpdateTimelineSection(ctx context.Context, seriesID, slug, title, content string) (json.RawMessage, error) {
	return m.UpdateTimelineSectionFn(ctx, seriesID, slug, title, content)
}

func (m *MockClient) DeleteTimelineSection(ctx context.Context, seriesID, slug string) error {
	return m.DeleteTimelineSectionFn(ctx, seriesID, slug)
}

func (m *MockClient) ReorderSeriesStories(ctx context.Context, seriesID string, storyIDs []string) error {
	return m.ReorderSeriesStoriesFn(ctx, seriesID, storyIDs)
}

func (m *MockClient) ReorderTimelineSections(ctx context.Context, seriesID string, slugs []string) error {
	return m.ReorderTimelineSectionsFn(ctx, seriesID, slugs)
}

// Series Characters

func (m *MockClient) CreateCharacter(ctx context.Context, seriesID string, req CreateCharacterReq) (*Character, error) {
	return m.CreateCharacterFn(ctx, seriesID, req)
}

func (m *MockClient) ListCharacters(ctx context.Context, seriesID string) (*CharacterList, error) {
	return m.ListCharactersFn(ctx, seriesID)
}

func (m *MockClient) GetCharacter(ctx context.Context, seriesID, slug string) (*Character, error) {
	return m.GetCharacterFn(ctx, seriesID, slug)
}

func (m *MockClient) UpdateCharacter(ctx context.Context, seriesID, slug string, req UpdateCharacterReq) error {
	return m.UpdateCharacterFn(ctx, seriesID, slug, req)
}

func (m *MockClient) DeleteCharacter(ctx context.Context, seriesID, slug string) error {
	return m.DeleteCharacterFn(ctx, seriesID, slug)
}

// Series Stories

func (m *MockClient) ListSeriesStories(ctx context.Context, seriesID string) (json.RawMessage, error) {
	return m.ListSeriesStoriesFn(ctx, seriesID)
}

func (m *MockClient) AddStoryToSeries(ctx context.Context, seriesID, storyID string) error {
	return m.AddStoryToSeriesFn(ctx, seriesID, storyID)
}

func (m *MockClient) RemoveStoryFromSeries(ctx context.Context, seriesID, storyID string) error {
	return m.RemoveStoryFromSeriesFn(ctx, seriesID, storyID)
}

// Series Plan

func (m *MockClient) PlanStory(ctx context.Context, seriesID string, req PlanStoryReq) (*PlanStoryResp, error) {
	return m.PlanStoryFn(ctx, seriesID, req)
}

// Story Forge Pipeline

func (m *MockClient) GetStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.GetStoryMetaFn(ctx, storyID)
}

// Images

func (m *MockClient) GenerateImage(ctx context.Context, req gen.HandlersGenerateImageRequest) (json.RawMessage, error) {
	return m.GenerateImageFn(ctx, req)
}

func (m *MockClient) UploadImage(ctx context.Context, contentType string, body io.Reader) (json.RawMessage, error) {
	return m.UploadImageFn(ctx, contentType, body)
}

func (m *MockClient) GetImage(ctx context.Context, id string) (json.RawMessage, error) {
	return m.GetImageFn(ctx, id)
}

func (m *MockClient) ListImages(ctx context.Context, params *gen.GetImagesParams) (json.RawMessage, error) {
	return m.ListImagesFn(ctx, params)
}

func (m *MockClient) RegenerateImage(ctx context.Context, id string, req gen.HandlersRegenerateRequest) (json.RawMessage, error) {
	return m.RegenerateImageFn(ctx, id, req)
}

func (m *MockClient) ListStoryImages(ctx context.Context, storyID string) (json.RawMessage, error) {
	return m.ListStoryImagesFn(ctx, storyID)
}

func (m *MockClient) AttachImageToStory(ctx context.Context, storyID, imageID string, req gen.HandlersAddToStoryRequest) error {
	return m.AttachImageToStoryFn(ctx, storyID, imageID, req)
}

func (m *MockClient) SetStoryImageCover(ctx context.Context, storyID, imageID string) error {
	return m.SetStoryImageCoverFn(ctx, storyID, imageID)
}
