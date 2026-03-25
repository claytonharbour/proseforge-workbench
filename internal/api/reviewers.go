package api

import (
	"context"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// RequestReviewer sends a reviewer request to another user.
func (c *Client) RequestReviewer(ctx context.Context, req CreateReviewerRequestReq) error {
	resp, err := c.raw.PostReviewersRequest(ctx, gen.PostReviewersRequestJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("request reviewer: %w", err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("request reviewer: %w", err)
	}
	return nil
}

// RespondToReviewerRequest accepts or declines a reviewer request.
func (c *Client) RespondToReviewerRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error {
	resp, err := c.raw.PostReviewersRespondId(ctx, requestID, gen.PostReviewersRespondIdJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("respond to reviewer request %s: %w", requestID, err)
	}
	defer resp.Body.Close()

	_, err = checkResponse(resp)
	if err != nil {
		return fmt.Errorf("respond to reviewer request %s: %w", requestID, err)
	}
	return nil
}

// ListAvailableReviewers returns users who have opted in as reviewers.
func (c *Client) ListAvailableReviewers(ctx context.Context) (*AvailableReviewerList, error) {
	resp, err := c.raw.GetReviewersAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("list available reviewers: %w", err)
	}
	defer resp.Body.Close()

	var result AvailableReviewerList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list available reviewers: %w", err)
	}
	return &result, nil
}

// ListMyReviewers returns the authenticated user's accepted reviewers.
func (c *Client) ListMyReviewers(ctx context.Context) ([]Reviewer, error) {
	resp, err := c.raw.GetReviewersMy(ctx)
	if err != nil {
		return nil, fmt.Errorf("list my reviewers: %w", err)
	}
	defer resp.Body.Close()

	var wrapper struct {
		Reviewers []Reviewer `json:"reviewers"`
	}
	if err := decode(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("list my reviewers: %w", err)
	}
	return wrapper.Reviewers, nil
}
