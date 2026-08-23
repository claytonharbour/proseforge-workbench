package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Listing operations wrap the generated client (c.raw.*) + gen request types.
// The hand-written raw-HTTP implementation was retired; the wire format matched
// the swagger exactly (store / format / status / url), so behavior is unchanged
// — only the transport moved onto codegen. Listings expose external store URLs
// (Amazon, Apple Books, etc.) for a story.

// ListListings returns all listings for a story.
func (c *Client) ListListings(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdListings(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	defer resp.Body.Close()
	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	return json.RawMessage(body), nil
}

// CreateListing adds a listing to a story.
func (c *Client) CreateListing(ctx context.Context, storyID, store, format, status, url string) (json.RawMessage, error) {
	body := gen.PostStoryIdListingsJSONRequestBody{
		Store:  &store,
		Format: &format,
		Status: &status,
		Url:    &url,
	}
	resp, err := c.raw.PostStoryIdListings(ctx, storyID, body)
	if err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	defer resp.Body.Close()
	out, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	return json.RawMessage(out), nil
}

// UpdateListing updates an existing listing. Empty fields are left unchanged.
func (c *Client) UpdateListing(ctx context.Context, listingID, store, format, status, url string) (json.RawMessage, error) {
	body := gen.PutListingsIdJSONRequestBody{}
	if store != "" {
		body.Store = &store
	}
	if format != "" {
		body.Format = &format
	}
	if status != "" {
		body.Status = &status
	}
	if url != "" {
		body.Url = &url
	}
	resp, err := c.raw.PutListingsId(ctx, listingID, body)
	if err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}
	defer resp.Body.Close()
	out, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}
	return json.RawMessage(out), nil
}

// DeleteListing removes a listing.
func (c *Client) DeleteListing(ctx context.Context, listingID string) error {
	resp, err := c.raw.DeleteListingsId(ctx, listingID)
	if err != nil {
		return fmt.Errorf("delete listing: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete listing: %w", err)
	}
	return nil
}
