package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

func registerListingTools(s *server.MCPServer, r *clientResolver) {
	// listing_list — List listings for a story
	s.AddTool(
		tool("listing_list",
			mcp.WithDescription(
				"List external store listings for a story (Amazon, Apple Books, etc.). "+
					"Returns store, format, status, and URL for each listing."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}

			result, err := client.ListListings(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// listing_create — Add a listing to a story
	s.AddTool(
		tool("listing_create",
			mcp.WithDescription(
				"Add an external store listing to a story. Stores: amazon, google_play, "+
					"apple_books, audible, kobo, other. Formats: ebook, audiobook, paperback. "+
					"Status: live, preorder, draft, pulled."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("store", mcp.Required(), mcp.Description("Store: amazon, google_play, apple_books, audible, kobo, other")),
			mcp.WithString("format", mcp.Required(), mcp.Description("Format: ebook, audiobook, paperback")),
			mcp.WithString("status", mcp.Required(), mcp.Description("Status: live, preorder, draft, pulled")),
			mcp.WithString("listing_url", mcp.Required(), mcp.Description("Store listing URL")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			store, err := requireArg(req, "store")
			if err != nil {
				return toolError(err), nil
			}
			format, err := requireArg(req, "format")
			if err != nil {
				return toolError(err), nil
			}
			status, err := requireArg(req, "status")
			if err != nil {
				return toolError(err), nil
			}
			listingURL, err := requireArg(req, "listing_url")
			if err != nil {
				return toolError(err), nil
			}

			result, err := client.CreateListing(ctx, storyID, api.CreateListingRequest{
				Store:  store,
				Format: format,
				Status: status,
				URL:    listingURL,
			})
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// listing_update — Update a listing
	s.AddTool(
		tool("listing_update",
			mcp.WithDescription("Update an existing store listing. Only provided fields are changed."),
			mcp.WithString("listing_id", mcp.Required(), mcp.Description("Listing ID")),
			mcp.WithString("store", mcp.Description("Store: amazon, google_play, apple_books, audible, kobo, other")),
			mcp.WithString("format", mcp.Description("Format: ebook, audiobook, paperback")),
			mcp.WithString("status", mcp.Description("Status: live, preorder, draft, pulled")),
			mcp.WithString("listing_url", mcp.Description("Store listing URL")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			listingID, err := requireArg(req, "listing_id")
			if err != nil {
				return toolError(err), nil
			}

			result, err := client.UpdateListing(ctx, listingID, api.UpdateListingRequest{
				Store:  optionalArg(req, "store"),
				Format: optionalArg(req, "format"),
				Status: optionalArg(req, "status"),
				URL:    optionalArg(req, "listing_url"),
			})
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// listing_delete — Delete a listing
	s.AddTool(
		tool("listing_delete",
			mcp.WithDescription("Delete a store listing."),
			mcp.WithString("listing_id", mcp.Required(), mcp.Description("Listing ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			listingID, err := requireArg(req, "listing_id")
			if err != nil {
				return toolError(err), nil
			}

			if err := client.DeleteListing(ctx, listingID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Listing deleted."), nil
		},
	)
}
