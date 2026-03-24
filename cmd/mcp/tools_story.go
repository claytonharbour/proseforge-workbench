package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/claytonharbour/proseforge-workbench/internal/story"
)

func registerStoryTools(s *server.MCPServer, r *clientResolver) {
	// story_list
	s.AddTool(
		tool("story_list",
			mcp.WithDescription("List stories for the authenticated user"),
			mcp.WithString("status", mcp.Description("Filter by status: published, unpublished, all")),
			mcp.WithNumber("limit", mcp.Description("Max results (1-100, default 25)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			limit := optionalIntArg(req, "limit", 25)
			params := &gen.GetStoriesParams{Limit: &limit}
			if s := optionalArg(req, "status"); s != "" {
				params.Status = &s
			}
			result, err := svc.List(ctx, params)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// story_get
	s.AddTool(
		tool("story_get",
			mcp.WithDescription("Get story details including sections"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.Get(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// story_export
	s.AddTool(
		tool("story_export",
			mcp.WithDescription("Export/download a story in the specified format"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("format", mcp.Description("Export format: json (default), markdown, pdf")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			format := optionalArg(req, "format")
			if format == "" {
				format = "json"
			}
			content, err := svc.Export(ctx, id, format)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(content), nil
		},
	)

	// story_section
	s.AddTool(
		tool("story_section",
			mcp.WithDescription("Get a single section's content and metadata (context-efficient — doesn't load the full story)"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			sectionID, err := requireArg(req, "section_id")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.GetSection(ctx, storyID, sectionID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_quality
	s.AddTool(
		tool("story_quality",
			mcp.WithDescription("Get code-based quality assessment scores for a story"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.GetQuality(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_assess
	s.AddTool(
		tool("story_assess",
			mcp.WithDescription("Trigger a code-based quality assessment for a story"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithBoolean("force", mcp.Description("Force re-assessment even if content unchanged")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.AssessQuality(ctx, id, false)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_insights
	s.AddTool(
		tool("story_insights",
			mcp.WithDescription("Get combined quality and AI analysis insights for a story"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.GetInsights(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// genre_list
	s.AddTool(
		tool("genre_list",
			mcp.WithDescription("List available genres"),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			data, err := svc.ListGenres(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_create
	s.AddTool(
		tool("story_create",
			mcp.WithDescription("Create a new story. Genre is specified by name and resolved to an ID."),
			mcp.WithString("genre", mcp.Required(), mcp.Description("Genre name (e.g., \"Historical Fiction\")")),
			mcp.WithString("title", mcp.Description("Story title (optional)")),
			mcp.WithString("tagline", mcp.Description("Story tagline (optional, set via update after creation)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))

			genreName, err := requireArg(req, "genre")
			if err != nil {
				return toolError(err), nil
			}

			genreID, err := mcpResolveGenreID(ctx, svc, genreName)
			if err != nil {
				return toolError(err, client), nil
			}

			createReq := api.CreateStoryRequest{
				GenreId: &genreID,
			}
			if t := optionalArg(req, "title"); t != "" {
				createReq.Title = &t
			}

			result, err := svc.Create(ctx, createReq)
			if err != nil {
				return toolError(err, client), nil
			}

			// If tagline was provided, set it via update
			if tagline := optionalArg(req, "tagline"); tagline != "" && result.Id != nil {
				updateReq := api.UpdateStoryRequest{Tagline: &tagline}
				if err := svc.Update(ctx, *result.Id, updateReq); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("story created (%s) but failed to set tagline: %v", *result.Id, err)), nil
				}
			}

			return jsonResult(result)
		},
	)

	// story_update
	s.AddTool(
		tool("story_update",
			mcp.WithDescription("Update a story's metadata (title and/or tagline)"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("title", mcp.Description("New title")),
			mcp.WithString("tagline", mcp.Description("New tagline")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}

			updateReq := api.UpdateStoryRequest{}
			if t := optionalArg(req, "title"); t != "" {
				updateReq.Title = &t
			}
			if t := optionalArg(req, "tagline"); t != "" {
				updateReq.Tagline = &t
			}
			if updateReq.Title == nil && updateReq.Tagline == nil {
				return mcp.NewToolResultError("at least one of title or tagline is required"), nil
			}

			if err := svc.Update(ctx, id, updateReq); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story updated."), nil
		},
	)

	// story_publish
	s.AddTool(
		tool("story_publish",
			mcp.WithDescription("Publish a story"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.Publish(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story published."), nil
		},
	)

	// story_unpublish
	s.AddTool(
		tool("story_unpublish",
			mcp.WithDescription("Unpublish a story"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.Unpublish(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story unpublished."), nil
		},
	)

	// section_create
	s.AddTool(
		tool("section_create",
			mcp.WithDescription("Create a new section in a story"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Section name (e.g., \"Chapter 1\")")),
			mcp.WithNumber("order", mcp.Description("Position to insert at (0-indexed)")),
			mcp.WithString("content", mcp.Description("Initial content (optional)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			name, err := requireArg(req, "name")
			if err != nil {
				return toolError(err), nil
			}

			createReq := api.CreateSectionRequest{
				Name: &name,
			}
			if order := optionalIntArg(req, "order", -1); order >= 0 {
				createReq.Order = &order
			}
			if content := optionalArg(req, "content"); content != "" {
				createReq.Content = &content
			}

			data, err := svc.CreateSection(ctx, storyID, createReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// section_write
	s.AddTool(
		tool("section_write",
			mcp.WithDescription("Write/update content in a section"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Section content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			sectionID, err := requireArg(req, "section_id")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}

			writeReq := api.UpdateSectionRequest{
				Content: &content,
			}
			if err := svc.WriteSection(ctx, storyID, sectionID, writeReq); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Section content updated."), nil
		},
	)
}

// mcpResolveGenreID looks up a genre by name (case-insensitive) and returns its ID.
func mcpResolveGenreID(ctx context.Context, svc *story.Service, name string) (string, error) {
	data, err := svc.ListGenres(ctx)
	if err != nil {
		return "", fmt.Errorf("listing genres to resolve name: %w", err)
	}

	var genres []api.Genre
	if err := json.Unmarshal(data, &genres); err != nil {
		return "", fmt.Errorf("parsing genres: %w", err)
	}

	target := strings.ToLower(strings.TrimSpace(name))
	for _, g := range genres {
		if g.Name != nil && strings.ToLower(*g.Name) == target {
			if g.Id == nil {
				return "", fmt.Errorf("genre %q has no ID", name)
			}
			return *g.Id, nil
		}
	}

	var available []string
	for _, g := range genres {
		if g.Name != nil {
			available = append(available, *g.Name)
		}
	}
	return "", fmt.Errorf("genre %q not found; available: %s", name, strings.Join(available, ", "))
}

