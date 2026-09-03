package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// GetAuthorBookshelf returns an author's complete bookshelf.
// Works unauthenticated (published only) or authenticated (includes unpublished).
//
// 🛑 TRIES THE CANONICAL SINGULAR ROUTE FIRST, FALLS BACK TO THE PLURAL ON 404.
//
// The two spellings are being migrated (#459, forge/proseforge#1288) and the
// cutover is NOT simultaneous across environments: dev drops the plural before
// demo and prod gain the singular, so at the time of writing there is no single
// spelling that works everywhere —
//
//	demo   /authors/{h}/books 200   ·   /author/{h}/books 404
//	dev    the reverse, once the deletion deploys
//
// ⚠️ A client that picks one breaks the other environment. This tries the
// destination first so the fallback disappears on its own once the singular is
// everywhere, and costs nothing but a 404 until then.
//
// ⛔ FALLS BACK ON 404 ONLY. A 401, 403 or 500 is a real failure and must
// surface — retrying those against a second URL would turn one clear error into
// two confusing ones.
//
// ⚑ Not silent: both attempts go through c.httpClient, so `--debug` shows the
// 404 and the retry. Remove this fallback when /author/{handle}/books answers on
// demo and prod — that is the last box on #459.
func (c *Client) GetAuthorBookshelf(ctx context.Context, handle string, q, series, status, sort string, limit, offset int) (json.RawMessage, error) {
	body, err := c.authorBookshelfCanonical(ctx, handle, q, series, status, sort, limit, offset)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return c.authorBookshelfPlural(ctx, handle, q, series, status, sort, limit, offset)
	}
	return body, err
}

// authorBookshelfCanonical issues the request against the singular route via the
// GENERATED client (#455), so a rename or removal upstream is a compile error
// here rather than a 404 in front of a user. This is the arm that survives:
// /author/{handle}/books is in the spec, the plural is not.
//
// ⚠️ Pointers are taken ONLY for values the hand-built code would have sent.
// That is the OPPOSITE of the body-carrying migrations in this package, where
// nil silently drops a field the server always received. Here omitempty is the
// intended behaviour and matches the previous `if q != ""` conditionals exactly
// — so do not "fix" this by addressing every field.
func (c *Client) authorBookshelfCanonical(ctx context.Context, handle string, q, series, status, sort string, limit, offset int) (json.RawMessage, error) {
	params := &gen.ListAuthorBooksParams{}
	if q != "" {
		params.Q = &q
	}
	if series != "" {
		params.Series = &series
	}
	if status != "" {
		params.Status = &status
	}
	if sort != "" {
		params.Sort = &sort
	}
	if limit > 0 {
		params.Limit = &limit
	}
	if offset > 0 {
		params.Offset = &offset
	}

	resp, err := c.raw.ListAuthorBooks(ctx, handle, params)
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

// authorBookshelfPlural issues the request against the DEPRECATED plural route.
//
// ⛔ Stays hand-built deliberately: /authors/{handle}/books is absent from the
// spec (it is being deleted), so there is no generated method to call. It is the
// one hand-built URL left in this file, pinned by TestHandBuiltURLsDoNotGrow.
// Delete this function — and the fallback above — when the singular answers on
// demo and prod. That is the last box on #459.
func (c *Client) authorBookshelfPlural(ctx context.Context, handle string, q, series, status, sort string, limit, offset int) (json.RawMessage, error) {
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
