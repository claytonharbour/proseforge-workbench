package api

import (
	"context"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// GetFeedbackReviews returns all feedback reviews for a story.
func (c *Client) GetFeedbackReviews(ctx context.Context, storyID string) (*FeedbackReviewList, error) {
	resp, err := c.raw.GetStoryIdFeedbackReviews(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get feedback reviews for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	var result FeedbackReviewList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get feedback reviews for story %s: %w", storyID, err)
	}
	return &result, nil
}

// GetFeedbackReview returns a single feedback review by ID.
// Pass include values (e.g. "items") to embed related data in the response.
// When "items" is included, the API wraps the response in {"review": {...}, "items": {...}}.
// For the wrapped response, this returns the review portion only. Use
// GetFeedbackReviewWithItems for the full response including items.
func (c *Client) GetFeedbackReview(ctx context.Context, storyID, reviewID string, include ...string) (*FeedbackReview, error) {
	var editors []gen.RequestEditorFn
	hasItems := false
	if len(include) > 0 {
		editors = append(editors, withInclude(include...))
		for _, inc := range include {
			if inc == "items" {
				hasItems = true
			}
		}
	}

	resp, err := c.raw.GetStoryIdFeedbackReviewsReviewId(ctx, storyID, reviewID, editors...)
	if err != nil {
		return nil, fmt.Errorf("get feedback review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	if hasItems {
		var wrapper FeedbackReviewWithItems
		if err := decode(resp, &wrapper); err != nil {
			return nil, fmt.Errorf("get feedback review %s: %w", reviewID, err)
		}
		if wrapper.Review != nil {
			return wrapper.Review, nil
		}
		return &FeedbackReview{}, nil
	}

	var result FeedbackReview
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get feedback review %s: %w", reviewID, err)
	}
	return &result, nil
}

// GetFeedbackReviewFull returns a feedback review with items included.
// Returns the full wrapped response with both review metadata and feedback items.
func (c *Client) GetFeedbackReviewFull(ctx context.Context, storyID, reviewID string) (*FeedbackReviewWithItems, error) {
	editors := []gen.RequestEditorFn{withInclude("items")}

	resp, err := c.raw.GetStoryIdFeedbackReviewsReviewId(ctx, storyID, reviewID, editors...)
	if err != nil {
		return nil, fmt.Errorf("get feedback review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	var result FeedbackReviewWithItems
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get feedback review %s: %w", reviewID, err)
	}
	return &result, nil
}

// GetFeedbackDiff returns the diff of suggested changes for a feedback review.
func (c *Client) GetFeedbackDiff(ctx context.Context, storyID, reviewID string) (*ReviewDiffResponse, error) {
	resp, err := c.raw.GetStoryIdFeedbackReviewsReviewIdDiff(ctx, storyID, reviewID)
	if err != nil {
		return nil, fmt.Errorf("get feedback diff for review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	var result ReviewDiffResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get feedback diff for review %s: %w", reviewID, err)
	}
	return &result, nil
}

// GetFeedbackSuggestions returns the full feedback response including sections and suggestions.
func (c *Client) GetFeedbackSuggestions(ctx context.Context, storyID, reviewID string) (*FullFeedback, error) {
	resp, err := c.raw.GetStoryIdFeedbackReviewsReviewIdSuggestions(ctx, storyID, reviewID)
	if err != nil {
		return nil, fmt.Errorf("get suggestions for review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	var result FullFeedback
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get suggestions for review %s: %w", reviewID, err)
	}
	return &result, nil
}

// CreateFeedbackReview creates a new feedback review for a story.
func (c *Client) CreateFeedbackReview(ctx context.Context, storyID string, req StartAIReviewRequest) (*FeedbackReview, error) {
	resp, err := c.raw.PostStoryIdFeedbackReviews(ctx, storyID, gen.PostStoryIdFeedbackReviewsJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("create feedback review for story %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	var result FeedbackReview
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create feedback review for story %s: %w", storyID, err)
	}
	return &result, nil
}

// AddFeedbackItem adds a feedback item to a review.
func (c *Client) AddFeedbackItem(ctx context.Context, storyID, reviewID string, req AddFeedbackItemRequest) error {
	resp, err := c.raw.PostStoryIdFeedbackReviewsReviewIdItems(ctx, storyID, reviewID, gen.PostStoryIdFeedbackReviewsReviewIdItemsJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("add feedback item to review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("add feedback item to review %s: %w", reviewID, err)
	}
	return nil
}

// SubmitReview submits a review, marking it as ready for the author.
func (c *Client) SubmitReview(ctx context.Context, reviewID string) error {
	resp, err := c.raw.PostReviewIdSubmit(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("submit review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("submit review %s: %w", reviewID, err)
	}
	return nil
}

// UpdateSectionContent rewrites a section's content in the feedback branch.
func (c *Client) UpdateSectionContent(ctx context.Context, storyID, reviewID, sectionID, content string) error {
	req := gen.PatchStoryIdFeedbackReviewsReviewIdSectionsSectionIdContentJSONRequestBody{
		Content: &content,
	}
	resp, err := c.raw.PatchStoryIdFeedbackReviewsReviewIdSectionsSectionIdContent(ctx, storyID, reviewID, sectionID, req)
	if err != nil {
		return fmt.Errorf("update section %s content: %w", sectionID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("update section %s content: %w", sectionID, err)
	}
	return nil
}

// IncorporateFeedback incorporates accepted changes from a feedback review.
func (c *Client) IncorporateFeedback(ctx context.Context, storyID, reviewID string, req IncorporateRequest) error {
	resp, err := c.raw.PostStoryIdFeedbackReviewsReviewIdIncorporate(ctx, storyID, reviewID, gen.PostStoryIdFeedbackReviewsReviewIdIncorporateJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("incorporate feedback for review %s: %w", reviewID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("incorporate feedback for review %s: %w", reviewID, err)
	}
	return nil
}
