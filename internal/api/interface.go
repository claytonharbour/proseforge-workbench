package api

import (
	"context"
	"encoding/json"
	"io"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ProseForgeAPI defines the full set of operations the workbench uses against
// the ProseForge API. Services depend on this interface rather than on *Client
// directly, enabling testing with fakes and decoupling layers.
type ProseForgeAPI interface {
	// Stories
	ListStories(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error)
	GetStory(ctx context.Context, id string) (*Story, error)
	GetStoryWithContent(ctx context.Context, id string) (*Story, error)
	CreateStory(ctx context.Context, req CreateStoryRequest) (*Story, error)
	UpdateStory(ctx context.Context, id string, req UpdateStoryRequest) error
	PublishStory(ctx context.Context, id string, visibility string) error
	UnpublishStory(ctx context.Context, id string) error
	UpdateVisibility(ctx context.Context, id string, visibility string) error
	DownloadStory(ctx context.Context, id string, format string) (string, error)
	ResolveVanityURL(ctx context.Context, handle, slug string) (json.RawMessage, error)
	ListStoriesWithReviewStatus(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error)
	ListVersions(ctx context.Context, storyID string, params *gen.GetStoryIdVersionsParams) (json.RawMessage, error)
	GetVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	DiffVersions(ctx context.Context, storyID, fromSha, toSha string) (json.RawMessage, error)

	// Sections
	GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error)
	CreateSection(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error)
	WriteSection(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) error

	// Genres
	ListGenres(ctx context.Context) (json.RawMessage, error)

	// Quality & Insights
	GetQuality(ctx context.Context, storyID string) (json.RawMessage, error)
	AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error)
	AssessQualityAtVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	GetInsights(ctx context.Context, storyID string) (json.RawMessage, error)

	// Reviews
	AddReviewer(ctx context.Context, storyID string, req AddReviewerRequest) (*Reviewer, error)
	ListReviewers(ctx context.Context, storyID string) (*ReviewersList, error)
	AcceptReview(ctx context.Context, reviewID string) error
	DeclineReview(ctx context.Context, reviewID string) error
	ApproveStory(ctx context.Context, reviewID string) error
	RejectStory(ctx context.Context, reviewID string, req ReviewFeedbackRequest) error
	ListPendingReviews(ctx context.Context, params *gen.GetReviewsPendingParams) (*PendingReviews, error)

	// Feedback
	GetFeedbackReviews(ctx context.Context, storyID string) (*FeedbackReviewList, error)
	GetFeedbackReview(ctx context.Context, storyID, reviewID string, include ...string) (*FeedbackReview, error)
	GetFeedbackReviewFull(ctx context.Context, storyID, reviewID string) (*FeedbackReviewWithItems, error)
	GetFeedbackDiff(ctx context.Context, storyID, reviewID string) (*ReviewDiffResponse, error)
	GetFeedbackSuggestions(ctx context.Context, storyID, reviewID string) (*FullFeedback, error)
	CreateFeedbackReview(ctx context.Context, storyID string, req StartAIReviewRequest) (*FeedbackReview, error)
	AddFeedbackItem(ctx context.Context, storyID, reviewID string, req AddFeedbackItemRequest) error
	SubmitReview(ctx context.Context, reviewID string) error
	UpdateSectionContent(ctx context.Context, storyID, reviewID, sectionID, content string) error
	IncorporateFeedback(ctx context.Context, storyID, reviewID string, req IncorporateRequest) error

	// Reviewer Pool
	RequestReviewer(ctx context.Context, req CreateReviewerRequestReq) error
	RespondToReviewerRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error
	ListAvailableReviewers(ctx context.Context) (*AvailableReviewerList, error)
	ListMyReviewers(ctx context.Context) ([]Reviewer, error)

	// Narration
	StartNarration(ctx context.Context, storyID string) error
	GetNarration(ctx context.Context, storyID string) (json.RawMessage, error)
	GetAudiobook(ctx context.Context, storyID string) (json.RawMessage, error)
	RegenerateChapter(ctx context.Context, storyID, chapterID string, force bool, voice string) error
	RetryChapter(ctx context.Context, storyID, chapterID string) error
	RebuildNarration(ctx context.Context, storyID string, chapterAnnouncements bool) error
	DeleteNarration(ctx context.Context, storyID string) error
	ResumeNarration(ctx context.Context, storyID string) error
	CancelChapter(ctx context.Context, storyID, chapterID string) error
	ListVoices(ctx context.Context) (json.RawMessage, error)
	ListSegments(ctx context.Context, storyID, chapterID string) (json.RawMessage, error)
	RegenerateSegment(ctx context.Context, storyID, chapterID, segmentID, voice string) error
	PatchNarration(ctx context.Context, storyID string, req gen.HandlersPatchNarrationRequest) (json.RawMessage, error)

	// Credits
	GetCredits(ctx context.Context) (json.RawMessage, error)
	EstimateCredits(ctx context.Context, params *gen.GetCreditsEstimateParams) (json.RawMessage, error)
	GetCreditHistory(ctx context.Context, limit int) (json.RawMessage, error)

	// Series
	ListSeries(ctx context.Context) (*SeriesList, error)
	CreateSeries(ctx context.Context, req CreateSeriesReq) (*Series, error)
	GetSeriesByID(ctx context.Context, id string) (*Series, error)
	UpdateSeries(ctx context.Context, id string, req UpdateSeriesReq) error
	ArchiveSeries(ctx context.Context, id string) error
	GetWorld(ctx context.Context, seriesID string) (json.RawMessage, error)
	UpdateWorld(ctx context.Context, seriesID string, content string) error
	GetTimeline(ctx context.Context, seriesID string) (json.RawMessage, error)
	UpdateTimeline(ctx context.Context, seriesID string, content string) error

	// Series Characters
	CreateCharacter(ctx context.Context, seriesID string, req CreateCharacterReq) (*Character, error)
	ListCharacters(ctx context.Context, seriesID string) (*CharacterList, error)
	GetCharacter(ctx context.Context, seriesID, slug string) (*Character, error)
	UpdateCharacter(ctx context.Context, seriesID, slug string, req UpdateCharacterReq) error
	DeleteCharacter(ctx context.Context, seriesID, slug string) error

	// Series Stories
	ListSeriesStories(ctx context.Context, seriesID string) (json.RawMessage, error)
	AddStoryToSeries(ctx context.Context, seriesID, storyID string) error
	RemoveStoryFromSeries(ctx context.Context, seriesID, storyID string) error

	// Series Chat
	CreateSeriesChat(ctx context.Context, seriesID string) (*SeriesChatSession, error)
	ListSeriesChats(ctx context.Context, seriesID string) (json.RawMessage, error)
	GetSeriesChat(ctx context.Context, seriesID, sessionID string) (*SeriesChatSession, error)
	SendSeriesChatMessage(ctx context.Context, seriesID, sessionID string, req SeriesChatSendReq) (*SeriesChatSendResp, error)
	FinalizeSeriesChat(ctx context.Context, seriesID, sessionID string) (*SeriesChatFinalizeResp, error)
	HarvestSeriesChat(ctx context.Context, seriesID, sessionID string) (json.RawMessage, error)
	HarvestAllSeriesChats(ctx context.Context, seriesID string) (json.RawMessage, error)

	// Series Plan
	PlanStory(ctx context.Context, seriesID string, req PlanStoryReq) (*PlanStoryResp, error)

	// Story Forge Chat
	CreateChatSession(ctx context.Context) (*ChatSession, error)
	ListChatSessions(ctx context.Context) (json.RawMessage, error)
	GetChatSession(ctx context.Context, id string) (*ChatSession, error)
	SendChatMessage(ctx context.Context, id string, req ChatSendReq) (*ChatSendResp, error)
	FinalizeChatSession(ctx context.Context, id string) (*ChatFinalizeResp, error)

	// Story Forge Pipeline
	GetGenerationStatus(ctx context.Context, storyID string) (json.RawMessage, error)
	GetStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error)
	ApproveStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error)
	RegenerateStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error)
	ResumeGeneration(ctx context.Context, storyID string) (json.RawMessage, error)

	// Images
	GenerateImage(ctx context.Context, req gen.HandlersGenerateImageRequest) (json.RawMessage, error)
	UploadImage(ctx context.Context, contentType string, body io.Reader) (json.RawMessage, error)
	GetImage(ctx context.Context, id string) (json.RawMessage, error)
	ListImages(ctx context.Context, params *gen.GetImagesParams) (json.RawMessage, error)
	RegenerateImage(ctx context.Context, id string, req gen.HandlersRegenerateRequest) (json.RawMessage, error)
	ListStoryImages(ctx context.Context, storyID string) (json.RawMessage, error)
	AttachImageToStory(ctx context.Context, storyID, imageID string, req gen.HandlersAddToStoryRequest) error
	SetStoryImageCover(ctx context.Context, storyID, imageID string) error
}

// Compile-time assertion: *Client implements ProseForgeAPI.
var _ ProseForgeAPI = (*Client)(nil)
