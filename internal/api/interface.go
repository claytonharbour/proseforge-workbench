package api

import (
	"context"
	"encoding/json"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// ProseForgeAPI defines the full set of operations the workbench uses against
// the ProseForge API. Services depend on this interface rather than on *Client
// directly, enabling testing with fakes and decoupling layers.
type ProseForgeAPI interface {
	// Stories
	ListStories(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error)
	GetStory(ctx context.Context, id string) (*Story, error)
	CreateStory(ctx context.Context, req CreateStoryRequest) (*Story, error)
	UpdateStory(ctx context.Context, id string, req UpdateStoryRequest) error
	PublishStory(ctx context.Context, id string) error
	UnpublishStory(ctx context.Context, id string) error
	DownloadStory(ctx context.Context, id string, format string) (string, error)
	ListStoriesWithReviewStatus(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error)

	// Sections
	GetSection(ctx context.Context, storyID, sectionID string) (json.RawMessage, error)
	CreateSection(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error)
	WriteSection(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) error

	// Genres
	ListGenres(ctx context.Context) (json.RawMessage, error)

	// Quality & Insights
	GetQuality(ctx context.Context, storyID string) (json.RawMessage, error)
	AssessQuality(ctx context.Context, storyID string, force bool) (json.RawMessage, error)
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
}

// Compile-time assertion: *Client implements ProseForgeAPI.
var _ ProseForgeAPI = (*Client)(nil)
