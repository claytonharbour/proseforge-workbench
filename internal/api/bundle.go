package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Bundle operations wrap the generated client (c.raw.*) and gen request types.
//
// The previous hand-written implementation (raw http.NewRequest + local structs)
// has been retired: it predated complete swagger coverage and sent the stale
// `comment` field for entry transitions, which the current server silently drops.
// The generated request body uses `transition`, which is what the server honors —
// so this migration also fixes interstitial transition text not persisting.

// ListBundles returns all bundles owned by the authenticated user.
func (c *Client) ListBundles(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.ListBundles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list bundles: %w", err)
	}
	defer resp.Body.Close()
	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list bundles: %w", err)
	}
	return json.RawMessage(body), nil
}

// GetBundle returns a bundle with all its entries.
func (c *Client) GetBundle(ctx context.Context, id string) (json.RawMessage, error) {
	resp, err := c.raw.GetBundle(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get bundle %s: %w", id, err)
	}
	defer resp.Body.Close()
	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get bundle %s: %w", id, err)
	}
	return json.RawMessage(body), nil
}

// CreateBundle creates a new bundle.
func (c *Client) CreateBundle(ctx context.Context, name, intro string) (json.RawMessage, error) {
	body := gen.CreateBundleJSONRequestBody{Name: &name}
	if intro != "" {
		body.Intro = &intro
	}
	resp, err := c.raw.CreateBundle(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("create bundle: %w", err)
	}
	defer resp.Body.Close()
	out, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("create bundle: %w", err)
	}
	return json.RawMessage(out), nil
}

// UpdateBundle updates a bundle's name and/or intro. Empty fields are left unchanged.
func (c *Client) UpdateBundle(ctx context.Context, id, name, intro string) error {
	body := gen.UpdateBundleJSONRequestBody{}
	if name != "" {
		body.Name = &name
	}
	if intro != "" {
		body.Intro = &intro
	}
	resp, err := c.raw.UpdateBundle(ctx, id, body)
	if err != nil {
		return fmt.Errorf("update bundle %s: %w", id, err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update bundle %s: %w", id, err)
	}
	return nil
}

// DeleteBundle deletes a bundle. The referenced stories are not affected.
func (c *Client) DeleteBundle(ctx context.Context, id string) error {
	resp, err := c.raw.DeleteBundle(ctx, id)
	if err != nil {
		return fmt.Errorf("delete bundle %s: %w", id, err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete bundle %s: %w", id, err)
	}
	return nil
}

// AddBundleEntry adds a story as an entry. transition is the interstitial text
// shown with the entry in the export (the server field is `transition`).
func (c *Client) AddBundleEntry(ctx context.Context, bundleID, storyID, transition string) (json.RawMessage, error) {
	body := gen.CreateBundleEntryJSONRequestBody{StoryId: &storyID}
	if transition != "" {
		body.Transition = &transition
	}
	resp, err := c.raw.CreateBundleEntry(ctx, bundleID, body)
	if err != nil {
		return nil, fmt.Errorf("add entry to bundle %s: %w", bundleID, err)
	}
	defer resp.Body.Close()
	out, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("add entry to bundle %s: %w", bundleID, err)
	}
	return json.RawMessage(out), nil
}

// UpdateBundleEntry updates an entry's transition text.
func (c *Client) UpdateBundleEntry(ctx context.Context, bundleID, entryID, transition string) error {
	resp, err := c.raw.UpdateBundleEntry(ctx, bundleID, entryID,
		gen.UpdateBundleEntryJSONRequestBody{Transition: &transition})
	if err != nil {
		return fmt.Errorf("update entry %s in bundle %s: %w", entryID, bundleID, err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("update entry %s in bundle %s: %w", entryID, bundleID, err)
	}
	return nil
}

// RemoveBundleEntry removes an entry from a bundle. The story is not affected.
func (c *Client) RemoveBundleEntry(ctx context.Context, bundleID, entryID string) error {
	resp, err := c.raw.DeleteBundleEntry(ctx, bundleID, entryID)
	if err != nil {
		return fmt.Errorf("remove entry %s from bundle %s: %w", entryID, bundleID, err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("remove entry %s from bundle %s: %w", entryID, bundleID, err)
	}
	return nil
}

// ReorderBundleEntries sets the display order of entries by entry ID.
func (c *Client) ReorderBundleEntries(ctx context.Context, bundleID string, entryIDs []string) error {
	resp, err := c.raw.UpdateBundleEntriesOrder(ctx, bundleID,
		gen.UpdateBundleEntriesOrderJSONRequestBody{EntryIds: &entryIDs})
	if err != nil {
		return fmt.Errorf("reorder entries in bundle %s: %w", bundleID, err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("reorder entries in bundle %s: %w", bundleID, err)
	}
	return nil
}

// ExportBundle exports a bundle in the given format (epub, json, markdown, pdf).
func (c *Client) ExportBundle(ctx context.Context, id, format string) ([]byte, error) {
	var (
		resp *http.Response
		err  error
	)
	switch format {
	case "epub":
		resp, err = c.raw.GetBundleExportEpub(ctx, id)
	case "json":
		resp, err = c.raw.GetBundleExportJson(ctx, id)
	case "markdown":
		resp, err = c.raw.GetBundleExportMarkdown(ctx, id)
	case "pdf":
		resp, err = c.raw.GetBundleExportPdf(ctx, id)
	default:
		return nil, fmt.Errorf("unsupported export format %q (use epub, json, markdown, pdf)", format)
	}
	if err != nil {
		return nil, fmt.Errorf("export bundle %s as %s: %w", id, format, err)
	}
	defer resp.Body.Close()
	data, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("export bundle %s as %s: %w", id, format, err)
	}
	return data, nil
}

// SetBundleCover associates an image as the bundle's cover image.
func (c *Client) SetBundleCover(ctx context.Context, bundleID, imageID string) error {
	resp, err := c.raw.UpdateBundleCover(ctx, bundleID, imageID)
	if err != nil {
		return fmt.Errorf("set cover for bundle %s: %w", bundleID, err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("set cover for bundle %s: %w", bundleID, err)
	}
	return nil
}
