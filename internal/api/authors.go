package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAuthorBookshelf returns an author's complete bookshelf.
// Works unauthenticated (published only) or authenticated (includes unpublished).
func (c *Client) GetAuthorBookshelf(ctx context.Context, handle string, q, series, status, sort string, limit, offset int) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/v1/authors/%s/books", c.baseURL, handle)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create author bookshelf request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	qp := req.URL.Query()
	if q != "" {
		qp.Set("q", q)
	}
	if series != "" {
		qp.Set("series", series)
	}
	if status != "" {
		qp.Set("status", status)
	}
	if sort != "" {
		qp.Set("sort", sort)
	}
	if limit > 0 {
		qp.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		qp.Set("offset", fmt.Sprintf("%d", offset))
	}
	req.URL.RawQuery = qp.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get author bookshelf: %w", err)
	}
	defer resp.Body.Close()

	var result json.RawMessage
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get author bookshelf: %w", err)
	}
	return result, nil
}
