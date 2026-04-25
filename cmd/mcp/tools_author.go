package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAuthorTools(s *server.MCPServer, r *clientResolver) {
	s.AddTool(
		tool("author_bookshelf",
			mcp.WithDescription(
				"Get an author's complete bookshelf — stories with titles, taglines, slugs, "+
					"cover URLs, section images, and series membership in a single call. "+
					"Replaces the N+1 pattern of story_list + per-story story_images.\n\n"+
					"Works unauthenticated for published stories. Authenticated calls also "+
					"include drafts and pitches."),
			mcp.WithString("handle", mcp.Required(), mcp.Description("Author vanity handle (e.g. 'claytonharbour')")),
			mcp.WithString("q", mcp.Description("Search title, tagline, or series name")),
			mcp.WithString("series", mcp.Description("Filter by series slug")),
			mcp.WithString("status", mcp.Description("Filter by status: draft, published, pitch")),
			mcp.WithString("sort", mcp.Description("Sort order: 'series' (default), 'title', 'newest', 'oldest'")),
			mcp.WithNumber("limit", mcp.Description("Max results (default 50, max 100)")),
			mcp.WithNumber("offset", mcp.Description("Pagination offset")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			handle, err := requireArg(req, "handle")
			if err != nil {
				return toolError(err), nil
			}

			q := optionalArg(req, "q")
			series := optionalArg(req, "series")
			status := optionalArg(req, "status")
			sort := optionalArg(req, "sort")
			limit := optionalIntArg(req, "limit", 0)
			offset := optionalIntArg(req, "offset", 0)

			result, err := client.GetAuthorBookshelf(ctx, handle, q, series, status, sort, limit, offset)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)
}
