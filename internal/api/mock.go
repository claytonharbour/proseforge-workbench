package api

import (
	"context"
	"encoding/json"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// MockClient implements ProseForgeAPI with function fields for testing.
// Each function field can be set to control the mock's behavior. If a function
// field is nil, calling the method will panic (indicating a missing test setup).
type MockClient struct {
	// Stories
	ListStoriesFn                  func(ctx context.Context, params *gen.GetStoriesParams) (*StoryList, error)
	GetStoryFn                     func(ctx context.Context, id string) (*Story, error)
	CreateStoryFn                  func(ctx context.Context, req CreateStoryRequest) (*Story, error)
	UpdateStoryFn                  func(ctx context.Context, id string, req UpdateStoryRequest) error
	PublishStoryFn                 func(ctx context.Context, id string) error
	UnpublishStoryFn               func(ctx context.Context, id string) error
	DownloadStoryFn                func(ctx context.Context, id string, format string) (string, error)
	ListStoriesWithReviewStatusFn  func(ctx context.Context, params *gen.GetStoriesMyReviewStatusParams) (*StoriesWithReview, error)

	// Sections
	GetSectionFn     func(ctx context.Context, storyID, sectionID string) (json.RawMessage, error)
	CreateSectionFn  func(ctx context.Context, storyID string, req CreateSectionRequest) (json.RawMessage, error)
	WriteSectionFn   func(ctx context.Context, storyID, sectionID string, req UpdateSectionRequest) error

	// Genres
	ListGenresFn func(ctx context.Context) (json.RawMessage, error)

	// Quality & Insights
	GetQualityFn    func(ctx context.Context, storyID string) (json.RawMessage, error)
	AssessQualityFn func(ctx context.Context, storyID string, force bool) (json.RawMessage, error)
	GetInsightsFn   func(ctx context.Context, storyID string) (json.RawMessage, error)

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
	ListAvailableReviewersFn     func(ctx context.Context) (json.RawMessage, error)
	ListMyReviewersFn            func(ctx context.Context) ([]Reviewer, error)
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

func (m *MockClient) DownloadStory(ctx context.Context, id string, format string) (string, error) {
	return m.DownloadStoryFn(ctx, id, format)
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

func (m *MockClient) PublishStory(ctx context.Context, id string) error {
	return m.PublishStoryFn(ctx, id)
}

func (m *MockClient) UnpublishStory(ctx context.Context, id string) error {
	return m.UnpublishStoryFn(ctx, id)
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

func (m *MockClient) ListAvailableReviewers(ctx context.Context) (json.RawMessage, error) {
	return m.ListAvailableReviewersFn(ctx)
}

func (m *MockClient) ListMyReviewers(ctx context.Context) ([]Reviewer, error) {
	return m.ListMyReviewersFn(ctx)
}
