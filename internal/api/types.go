package api

import "github.com/claytonharbour/proseforge-workbench/internal/api/gen"

// Type aliases for the generated types the workbench uses.
// These provide shorter names and insulate callers from the generated package.

// Story types
type (
	Story             = gen.HandlersStoryResponse
	StoryList         = gen.HandlersStoryListResponse
	StorySection      = gen.HandlersStorySectionResponse
	StoryBrief        = gen.HandlersStoryBriefResponse
	StoryWithReview   = gen.HandlersStoryWithReviewSummaryResponse
	StoriesWithReview = gen.HandlersStoriesWithReviewStatusResponse
)

// Review types
type (
	Reviewer       = gen.HandlersReviewerResponse
	ReviewersList  = gen.HandlersReviewersListResponse
	ReviewSummary  = gen.HandlersReviewSummaryResponse
	PendingReviews = gen.HandlersPendingReviewsResponse
)

// Reviewer pool types
type (
	AvailableReviewer     = gen.HandlersAvailableReviewerResponse
	AvailableReviewerList = gen.HandlersAvailableReviewersListResponse
)

// Feedback types
type (
	FeedbackReview     = gen.HandlersFeedbackReviewResponse
	FeedbackReviewList = gen.HandlersFeedbackReviewListResponse
	FeedbackSuggestion = gen.HandlersFeedbackSuggestionResponse
	FullFeedback       = gen.HandlersFullFeedbackResponse
	DiffResponse       = gen.HandlersDiffResponse
	ReviewDiffResponse = gen.HandlersReviewDiffResponse
	ReviewDiffFile     = gen.HandlersReviewDiffFile
	FileDiff           = gen.HandlersFileDiffResponse
)

// FeedbackReviewWithItems is the response shape when ?include=items is passed.
// The API wraps the review in {"review": {...}, "items": {...}}.
type FeedbackReviewWithItems struct {
	Review *FeedbackReview          `json:"review,omitempty"`
	Items  *FeedbackReviewItemsData `json:"items,omitempty"`
}

// FeedbackReviewItemsData contains the structured feedback items.
type FeedbackReviewItemsData struct {
	SessionID        string                    `json:"sessionId"`
	Sections         []FeedbackSectionData     `json:"sections"`
	TotalSuggestions int                       `json:"totalSuggestions"`
	HasConflicts     bool                      `json:"hasConflicts"`
}

// FeedbackSectionData contains feedback items for a single section.
type FeedbackSectionData struct {
	SectionID     string                       `json:"sectionId"`
	SectionTitle  string                       `json:"sectionTitle"`
	Rating        float64                      `json:"rating"`
	Suggestions   []FeedbackItemDetail         `json:"suggestions"`
	Strengths     []FeedbackItemDetail         `json:"strengths"`
	Opportunities []FeedbackItemDetail         `json:"opportunities"`
	Comments      []FeedbackItemDetail         `json:"comments"`
	Context       map[string][]string          `json:"context"`
}

// FeedbackItemDetail is a single feedback item within a section.
type FeedbackItemDetail struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Original    string `json:"original,omitempty"`
	Suggested   string `json:"suggested,omitempty"`
	Text        string `json:"text,omitempty"`
	Rationale   string `json:"rationale,omitempty"`
	Status      string `json:"status"`
	HasConflict bool   `json:"hasConflict"`
	CanApply    bool   `json:"canApply"`
	Source      string `json:"source"`
}

// Genre types
type (
	Genre = gen.HandlersGenreResponse
)

// Series types
type (
	Series             = gen.HandlersSeriesResponse
	SeriesList         = gen.HandlersSeriesListResponse
	CreateSeriesReq    = gen.HandlersCreateSeriesRequest
	UpdateSeriesReq    = gen.HandlersUpdateSeriesRequest
	UpdateContentReq   = gen.HandlersUpdateContentRequest
	Character          = gen.HandlersCharacterResponse
	CharacterList      = gen.HandlersCharacterListResponse
	CreateCharacterReq = gen.HandlersCreateCharacterRequest
	UpdateCharacterReq = gen.HandlersUpdateCharacterRequest
	AddStoryReq        = gen.HandlersAddStoryRequest

	// Series Chat types
	SeriesChatSession     = gen.HandlersSeriesChatSessionResponse
	SeriesChatMessage     = gen.HandlersSeriesChatMsgResponse
	SeriesChatSendReq     = gen.HandlersSeriesChatSendMessageRequest
	SeriesChatSendResp    = gen.HandlersSeriesChatSendMessageResponse
	SeriesChatFinalizeResp = gen.HandlersSeriesChatFinalizeResponse
)

// Request types
type (
	AddReviewerRequest        = gen.HandlersAddReviewerRequest
	ReviewFeedbackRequest     = gen.HandlersReviewFeedbackRequest
	RequestReviewFromPoolReq  = gen.HandlersRequestReviewFromPoolRequest
	RequestReviewFromPoolResp = gen.HandlersRequestReviewFromPoolResponse
	AddFeedbackItemRequest    = gen.HandlersAddFeedbackItemRequest
	StartAIReviewRequest      = gen.HandlersStartAIReviewRequest
	IncorporateRequest           = gen.HandlersIncorporateRequest
	UpdateSectionContentRequest  = gen.HandlersUpdateSectionContentRequest
	UpdateSuggestionStatusReq = gen.HandlersUpdateSuggestionStatusRequest
	RespondToReviewerReq      = gen.HandlersRespondToReviewerRequestRequest
	CreateReviewerRequestReq  = gen.HandlersCreateReviewerRequestRequest
	CreateStoryRequest        = gen.HandlersCreateStoryRequest
	UpdateStoryRequest        = gen.HandlersUpdateStoryRequest
	CreateSectionRequest      = gen.HandlersCreateSectionRequest
	UpdateSectionRequest      = gen.HandlersUpdateSectionRequest
)

// Review status values returned by the API.
const (
	ReviewStatusPending   = "pending"
	ReviewStatusRunning   = "running"
	ReviewStatusCompleted = "completed"
)

// Story Forge Chat types
type (
	ChatSession       = gen.HandlersChatSessionResponse
	ChatMessage       = gen.HandlersChatMsgResponse
	ChatSendReq       = gen.HandlersSendMessageRequest
	ChatSendResp      = gen.HandlersSendMessageResponse
	ChatFinalizeResp  = gen.HandlersFinalizeResponse
	GenerateCompleteReq = gen.HandlersGenerateCompleteStoryRequest
)

// Image types
type (
	GenerateImageRequest = gen.HandlersGenerateImageRequest
	RegenerateImageRequest = gen.HandlersRegenerateRequest
	AddImageToStoryRequest = gen.HandlersAddToStoryRequest
)

// Common types
type (
	ErrorResponse   = gen.HandlersErrorResponse
	MessageResponse = gen.HandlersMessageResponse
)
