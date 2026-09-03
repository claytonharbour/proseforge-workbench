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
	ListStories(ctx context.Context, params *gen.ListStoriesParams) (*StoryList, error)
	GetStory(ctx context.Context, id string) (*Story, error)
	GetStoryWithContent(ctx context.Context, id string) (*Story, error)
	CreateStory(ctx context.Context, req CreateStoryRequest) (*Story, error)
	UpdateStory(ctx context.Context, id string, req UpdateStoryRequest) error
	PublishStory(ctx context.Context, id string, visibility string) error
	UnpublishStory(ctx context.Context, id string) error
	CreatePitch(ctx context.Context, req CreateStoryRequest) (*Story, error)
	PromoteStory(ctx context.Context, id string) error
	UpsertStoryMeta(ctx context.Context, storyID, metaType, content string) (json.RawMessage, error)
	DeleteStory(ctx context.Context, id string) error
	DeleteSection(ctx context.Context, storyID, sectionID string) error
	RestoreVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	GetMetaStale(ctx context.Context, storyID string) (json.RawMessage, error)
	AcknowledgeMetaStale(ctx context.Context, storyID string) error
	RegenerateTagline(ctx context.Context, storyID string) error
	RegenerateTitle(ctx context.Context, storyID string) error
	RegenerateStaleNarration(ctx context.Context, storyID string) (json.RawMessage, error)
	AcknowledgeNarrationStale(ctx context.Context, storyID string) error
	UpdateVisibility(ctx context.Context, id string, visibility string) error
	DownloadStory(ctx context.Context, id string, format string) (string, error)
	ResolveVanityURL(ctx context.Context, handle, slug string) (json.RawMessage, error)
	ListStoriesWithReviewStatus(ctx context.Context, params *gen.ListMyStoriesReviewStatusParams) (*StoriesWithReview, error)
	ListVersions(ctx context.Context, storyID string, params *gen.ListStoryVersionsParams) (json.RawMessage, error)
	GetVersion(ctx context.Context, storyID, sha string) (json.RawMessage, error)
	DiffVersions(ctx context.Context, storyID, fromSha, toSha string) (json.RawMessage, error)

	// Sections
	ListSections(ctx context.Context, storyID string) (json.RawMessage, error)
	GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error)
	CreateSection(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error)
	WriteSection(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) (json.RawMessage, error)
	ReorderSection(ctx context.Context, storyID, sectionID string, order int) (json.RawMessage, error)

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

	// Friends (the reviewer pool, renamed upstream in forge/proseforge#808)
	ListFriends(ctx context.Context, include string) (json.RawMessage, error)
	CountFriends(ctx context.Context) (json.RawMessage, error)
	ListFriendCandidates(ctx context.Context, search string) (json.RawMessage, error)
	RequestFriend(ctx context.Context, req CreateReviewerRequestReq) error
	RespondToFriendRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error
	ListIncomingFriendRequests(ctx context.Context) (json.RawMessage, error)
	ListOutgoingFriendRequests(ctx context.Context) (json.RawMessage, error)
	RemoveFriend(ctx context.Context, friendID string) error

	// Reviewer Pool
	RequestReviewer(ctx context.Context, req CreateReviewerRequestReq) error
	RespondToReviewerRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error
	ListAvailableReviewers(ctx context.Context) (*AvailableReviewerList, error)
	ListMyReviewers(ctx context.Context) ([]Reviewer, error)

	// Narration
	StartNarration(ctx context.Context, storyID string, voice string, force bool, emphasis *bool) error
	GetNarration(ctx context.Context, storyID string) (json.RawMessage, error)
	GetAudiobook(ctx context.Context, storyID string) (json.RawMessage, error)
	RegenerateSection(ctx context.Context, storyID, sectionID string, force bool, voice string) error
	RetrySection(ctx context.Context, storyID, sectionID string) error
	RebuildNarration(ctx context.Context, storyID string, sectionAnnouncements bool, emphasis *bool) error
	DeleteNarration(ctx context.Context, storyID string) error
	ResumeNarration(ctx context.Context, storyID string) error
	CancelSection(ctx context.Context, storyID, sectionID string) error
	ListVoices(ctx context.Context) (json.RawMessage, error)
	ListSegments(ctx context.Context, storyID, sectionID string) (json.RawMessage, error)
	RegenerateSegment(ctx context.Context, storyID, sectionID, segmentID, voice string) error
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
	ListTimelineSections(ctx context.Context, seriesID string) (json.RawMessage, error)
	GetTimelineSection(ctx context.Context, seriesID, slug string) (json.RawMessage, error)
	UpdateTimelineSection(ctx context.Context, seriesID, slug, title, content string) (json.RawMessage, error)
	DeleteTimelineSection(ctx context.Context, seriesID, slug string) error
	ReorderSeriesStories(ctx context.Context, seriesID string, storyIDs []string) error
	ReorderTimelineSections(ctx context.Context, seriesID string, slugs []string) error

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

	// Series Plan
	PlanStory(ctx context.Context, seriesID string, req PlanStoryReq) (*PlanStoryResp, error)

	// Story Forge Pipeline
	GetStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error)

	// Images
	GenerateImage(ctx context.Context, req gen.HandlersGenerateImageRequest) (json.RawMessage, error)
	UploadImage(ctx context.Context, contentType string, body io.Reader) (json.RawMessage, error)
	GetImage(ctx context.Context, id string) (json.RawMessage, error)
	ListImages(ctx context.Context, params *gen.ListImagesParams) (json.RawMessage, error)
	RegenerateImage(ctx context.Context, id string, req gen.HandlersRegenerateRequest) (json.RawMessage, error)
	ListStoryImages(ctx context.Context, storyID string, includeSections bool) (json.RawMessage, error)
	AttachImageToStory(ctx context.Context, storyID, imageID string, req gen.HandlersAddToStoryRequest) error
	SetStoryImageCover(ctx context.Context, storyID, imageID string) error

	// Bundles
	ListBundles(ctx context.Context) (json.RawMessage, error)
	GetBundle(ctx context.Context, id string) (json.RawMessage, error)
	CreateBundle(ctx context.Context, name, intro string) (json.RawMessage, error)
	UpdateBundle(ctx context.Context, id, name, intro string) error
	DeleteBundle(ctx context.Context, id string) error
	AddBundleEntry(ctx context.Context, bundleID, storyID, transition string) (json.RawMessage, error)
	UpdateBundleEntry(ctx context.Context, bundleID, entryID, transition string) error
	RemoveBundleEntry(ctx context.Context, bundleID, entryID string) error
	ReorderBundleEntries(ctx context.Context, bundleID string, entryIDs []string) error
	ExportBundle(ctx context.Context, id, format string) ([]byte, error)
	SetBundleCover(ctx context.Context, bundleID, imageID string) error

	WhoAmI(ctx context.Context) (json.RawMessage, error)
}

// Compile-time assertion: *Client implements ProseForgeAPI.
var _ ProseForgeAPI = (*Client)(nil)
