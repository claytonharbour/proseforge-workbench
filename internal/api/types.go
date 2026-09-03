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

// Reviewer pool types.
//
// These were aliases for generated types until the availability concept was
// removed upstream (forge/proseforge#807/#808) and the generated types went with
// it. They are declared locally now so the deprecated signatures in reviewers.go
// keep compiling and their callers keep getting the explanatory error rather
// than a build break. Nothing populates them — ListAvailableReviewers never
// reaches a route. They retire with the rest of the reviewer surface (#277).
//
// wirePartner: none — their generated counterparts were DELETED upstream with the
// availability concept (#807/#808). There is nothing left to drift against; these
// exist only so the deprecated signatures compile.
type (
	AvailableReviewer struct {
		Email *string `json:"email,omitempty"`
		Id    *string `json:"id,omitempty"`
		Name  *string `json:"name,omitempty"`
	}
	AvailableReviewerList struct {
		Reviewers *[]AvailableReviewer `json:"reviewers,omitempty"`
		Total     *int                 `json:"total,omitempty"`
	}
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
// wirePartner: HandlersReviewDetailResponse
type FeedbackReviewWithItems struct {
	Review *FeedbackReview          `json:"review,omitempty"`
	Items  *FeedbackReviewItemsData `json:"items,omitempty"`
}

// FeedbackReviewItemsData contains the structured feedback items.
// wirePartner: HandlersSuggestionsResponse
type FeedbackReviewItemsData struct {
	SessionID        string                `json:"sessionId"`
	Sections         []FeedbackSectionData `json:"sections"`
	TotalSuggestions int                   `json:"totalSuggestions"`
	HasConflicts     bool                  `json:"hasConflicts"`
}

// FeedbackSectionContext is the section-level analysis metadata (characters
// mentioned, plot points, unresolved threads, tone summary, general notes).
// Aliased from the generated type so the wire shape stays in sync with the API.
type FeedbackSectionContext = gen.FeedbackSectionContext

// FeedbackSectionData contains feedback items for a single section.
// wirePartner: HandlersSectionWithSuggestions
type FeedbackSectionData struct {
	SectionID     string                  `json:"sectionId"`
	SectionTitle  string                  `json:"sectionTitle"`
	Rating        float64                 `json:"rating"`
	Suggestions   []FeedbackItemDetail    `json:"suggestions"`
	Strengths     []FeedbackItemDetail    `json:"strengths"`
	Opportunities []FeedbackItemDetail    `json:"opportunities"`
	Comments      []FeedbackItemDetail    `json:"comments"`
	Context       *FeedbackSectionContext `json:"context,omitempty"`
}

// FeedbackItemDetail is a single feedback item within a section.
// wirePartner: HandlersSuggestionItem
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

	// Series Plan types
	PlanStoryReq  = gen.HandlersPlanStoryRequest
	PlanStoryResp = gen.HandlersPlanStoryResponse

	// Series Chat types
	SeriesChatSession      = gen.HandlersSeriesChatSessionResponse
	SeriesChatMessage      = gen.HandlersSeriesChatMsgResponse
	SeriesChatSendReq      = gen.HandlersSeriesChatSendMessageRequest
	SeriesChatSendResp     = gen.HandlersSeriesChatSendMessageResponse
	SeriesChatFinalizeResp = gen.HandlersSeriesChatFinalizeResponse
)

// Request types
type (
	AddReviewerRequest          = gen.HandlersAddReviewerRequest
	ReviewFeedbackRequest       = gen.HandlersReviewFeedbackRequest
	RequestReviewFromPoolReq    = gen.HandlersRequestReviewFromPoolRequest
	RequestReviewFromPoolResp   = gen.HandlersRequestReviewFromPoolResponse
	AddFeedbackItemRequest      = gen.HandlersAddFeedbackItemRequest
	StartAIReviewRequest        = gen.HandlersStartAIReviewRequest
	IncorporateRequest          = gen.HandlersIncorporateRequest
	UpdateSectionContentRequest = gen.HandlersUpdateSectionContentRequest
	UpdateSuggestionStatusReq   = gen.HandlersUpdateSuggestionStatusRequest
	RespondToReviewerReq        = gen.HandlersRespondToFriendRequestRequest
	CreateReviewerRequestReq    = gen.HandlersCreateFriendRequestRequest
	CreateStoryRequest          = gen.HandlersCreateStoryRequest
	UpdateStoryRequest          = gen.HandlersUpdateStoryRequest
	CreateSectionRequest        = gen.HandlersCreateSectionRequest
	UpdateSectionRequest        = gen.HandlersUpdateSectionRequest
)

// Review status values returned by the API.
const (
	ReviewStatusPending   = "pending"
	ReviewStatusRunning   = "running"
	ReviewStatusCompleted = "completed"
)

// Story Forge Chat types
type (
	ChatSession         = gen.HandlersChatSessionResponse
	ChatMessage         = gen.HandlersChatMsgResponse
	ChatSendReq         = gen.HandlersSendMessageRequest
	ChatSendResp        = gen.HandlersSendMessageResponse
	ChatFinalizeResp    = gen.HandlersFinalizeResponse
	GenerateCompleteReq = gen.HandlersGenerateCompleteStoryRequest
)

// Image types
type (
	GenerateImageRequest   = gen.HandlersGenerateImageRequest
	RegenerateImageRequest = gen.HandlersRegenerateRequest
	AddImageToStoryRequest = gen.HandlersAddToStoryRequest
)

// Common types
type (
	ErrorResponse   = gen.HandlersErrorResponse
	MessageResponse = gen.HandlersMessageResponse
)
