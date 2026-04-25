package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateListingRequest is the request body for creating a listing.
type CreateListingRequest struct {
	Store  string `json:"store"`
	Format string `json:"format"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

// UpdateListingRequest is the request body for updating a listing.
type UpdateListingRequest struct {
	Store  string `json:"store,omitempty"`
	Format string `json:"format,omitempty"`
	Status string `json:"status,omitempty"`
	URL    string `json:"url,omitempty"`
}

// ListListings returns all listings for a story.
func (c *Client) ListListings(ctx context.Context, storyID string) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/v1/story/%s/listings", c.baseURL, storyID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create list listings request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	defer resp.Body.Close()

	var result json.RawMessage
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	return result, nil
}

// CreateListing adds a listing to a story.
func (c *Client) CreateListing(ctx context.Context, storyID string, listing CreateListingRequest) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/v1/story/%s/listings", c.baseURL, storyID)
	body, err := json.Marshal(listing)
	if err != nil {
		return nil, fmt.Errorf("marshal listing: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create listing request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	defer resp.Body.Close()

	var result json.RawMessage
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	return result, nil
}

// UpdateListing updates an existing listing.
func (c *Client) UpdateListing(ctx context.Context, listingID string, listing UpdateListingRequest) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/v1/listings/%s", c.baseURL, listingID)
	body, err := json.Marshal(listing)
	if err != nil {
		return nil, fmt.Errorf("marshal listing update: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create update listing request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}
	defer resp.Body.Close()

	var result json.RawMessage
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}
	return result, nil
}

// DeleteListing removes a listing.
func (c *Client) DeleteListing(ctx context.Context, listingID string) error {
	url := fmt.Sprintf("%s/api/v1/listings/%s", c.baseURL, listingID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create delete listing request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result json.RawMessage
		if err := decode(resp, &result); err != nil {
			return fmt.Errorf("delete listing: %w", err)
		}
	}
	return nil
}
