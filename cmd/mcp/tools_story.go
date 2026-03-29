package main

import (
	"context"
	"encoding/json"
	"fmt"

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
			mcp.WithDescription("List stories for the authenticated user. Use to browse stories when you don't have an ID."),
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

	// story_resolve
	s.AddTool(
		tool("story_resolve",
			mcp.WithDescription("Resolve a vanity URL (@handle/slug) to a story ID and metadata. Use when given a public URL like app.proseforge.com/@handle/slug/read"),
			mcp.WithString("handle", mcp.Required(), mcp.Description("Author handle (without @)")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Story slug from the URL")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			handle, err := requireArg(req, "handle")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.ResolveVanityURL(ctx, handle, slug)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// story_get
	s.AddTool(
		tool("story_get",
			mcp.WithDescription("Get story details including section IDs. Requires story_id from story_list or story_resolve."),
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
			mcp.WithDescription("Export/download a story in the specified format. Requires story_id. Returns full content — use this to read a story, not story_get."),
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
			mcp.WithDescription("Get a single section's content and metadata (context-efficient). Requires story_id and section_id — get section IDs from story_get."),
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
			mcp.WithDescription("Get code-based quality assessment scores for a story. Requires story_id."),
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
			mcp.WithDescription("Trigger a code-based quality assessment for a story. Poll with story_quality for results."),
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
			mcp.WithDescription("Create a new story. Genre is specified by name and resolved to an ID. Returns story_id for use with section_create and other tools."),
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

			genreID, err := svc.ResolveGenreID(ctx, genreName)
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
			mcp.WithDescription("Create a new section in a story. Requires story_id from story_create or story_get."),
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
			mcp.WithDescription("Write/update content in a section. Requires story_id and section_id."),
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

	// narration_start
	s.AddTool(
		tool("narration_start",
			mcp.WithDescription("Start narration/audiobook generation for a story. Requires story_id. Story should be published first. Poll with narration_status."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			if err := svc.StartNarration(ctx, storyID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration started."), nil
		},
	)

	// narration_status
	s.AddTool(
		tool("narration_status",
			mcp.WithDescription("Get narration status and chapter details for a story. Returns chapter IDs needed by narration_regenerate, narration_segments, etc."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			result, err := svc.GetNarration(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// narration_audiobook
	s.AddTool(
		tool("narration_audiobook",
			mcp.WithDescription("Get audiobook download info for a story. Returns download URLs after narration completes."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			result, err := svc.GetAudiobook(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// narration_voices
	s.AddTool(
		tool("narration_voices",
			mcp.WithDescription("List available TTS voices across all providers. Use voice names with narration_regenerate or narration_patch."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			result, err := svc.ListVoices(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// narration_regenerate
	s.AddTool(
		tool("narration_regenerate",
			mcp.WithDescription("Regenerate narration for a specific chapter. Supports force regeneration and voice override."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("chapter_id", mcp.Required(), mcp.Description("Chapter ID (from narration_status)")),
			mcp.WithBoolean("force", mcp.Description("Force regeneration even if content hasn't changed (default false)")),
			mcp.WithString("voice", mcp.Description("Voice override for this chapter (e.g., Puck, af_sarah). Use narration_voices to list options.")),
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
			chapterID, err := requireArg(req, "chapter_id")
			if err != nil {
				return toolError(err), nil
			}
			force := optionalBoolArg(req, "force")
			voice := optionalArg(req, "voice")
			if err := svc.RegenerateChapter(ctx, storyID, chapterID, force, voice); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Chapter %s regeneration started.", chapterID)), nil
		},
	)

	// narration_retry
	s.AddTool(
		tool("narration_retry",
			mcp.WithDescription("Retry a failed or stuck chapter narration. Resets error state and re-queues."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("chapter_id", mcp.Required(), mcp.Description("Chapter ID (from narration_status)")),
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
			chapterID, err := requireArg(req, "chapter_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.RetryChapter(ctx, storyID, chapterID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Chapter %s retry started.", chapterID)), nil
		},
	)

	// narration_rebuild
	s.AddTool(
		tool("narration_rebuild",
			mcp.WithDescription("Reassemble audiobook from existing chapter audio. No TTS, no credits. Use after fixing individual chapters."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithBoolean("chapter_announcements", mcp.Description("Insert TTS-generated chapter title announcements (default false)")),
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
			announcements := optionalBoolArg(req, "chapter_announcements")
			if err := svc.RebuildNarration(ctx, storyID, announcements); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration rebuild started."), nil
		},
	)

	// narration_delete
	s.AddTool(
		tool("narration_delete",
			mcp.WithDescription("Delete all narration data for a story. Start fresh with narration_start after."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			if err := svc.DeleteNarration(ctx, storyID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration deleted."), nil
		},
	)

	// narration_resume
	s.AddTool(
		tool("narration_resume",
			mcp.WithDescription("Resume a stuck narration that stopped mid-processing"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			if err := svc.ResumeNarration(ctx, storyID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration resumed."), nil
		},
	)

	// narration_chapter_cancel
	s.AddTool(
		tool("narration_chapter_cancel",
			mcp.WithDescription("Cancel a specific chapter's narration without deleting the whole narration"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("chapter_id", mcp.Required(), mcp.Description("Chapter ID")),
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
			chapterID, err := requireArg(req, "chapter_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.CancelChapter(ctx, storyID, chapterID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Chapter %s cancelled.", chapterID)), nil
		},
	)

	// narration_segments
	s.AddTool(
		tool("narration_segments",
			mcp.WithDescription("List segments for a chapter with text content, voice, and provider info. Requires story_id and chapter_id from narration_status. Returns segment IDs for narration_segment_regenerate."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("chapter_id", mcp.Required(), mcp.Description("Chapter ID")),
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
			chapterID, err := requireArg(req, "chapter_id")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.ListSegments(ctx, storyID, chapterID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// narration_segment_regenerate
	s.AddTool(
		tool("narration_segment_regenerate",
			mcp.WithDescription("Regenerate a single segment's audio within a chapter. Requires segment_id from narration_segments. Use narration_voices to list available voices."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("chapter_id", mcp.Required(), mcp.Description("Chapter ID")),
			mcp.WithString("segment_id", mcp.Required(), mcp.Description("Segment ID (from narration_segments)")),
			mcp.WithString("voice", mcp.Description("Voice override for this segment (e.g., Kore, af_sarah). Use narration_voices to list options.")),
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
			chapterID, err := requireArg(req, "chapter_id")
			if err != nil {
				return toolError(err), nil
			}
			segmentID, err := requireArg(req, "segment_id")
			if err != nil {
				return toolError(err), nil
			}
			voice := optionalArg(req, "voice")
			if err := svc.RegenerateSegment(ctx, storyID, chapterID, segmentID, voice); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Segment %s regeneration started.", segmentID)), nil
		},
	)

	// narration_patch
	s.AddTool(
		tool("narration_patch",
			mcp.WithDescription("Batch patch multiple segments and/or chapters with voice changes. Rebuilds audiobook once when done. Batch alternative to narration_segment_regenerate — one call, one rebuild."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("segments_json", mcp.Description("JSON array of {chapter_id, segment_id, voice} objects")),
			mcp.WithString("chapters_json", mcp.Description("JSON array of {chapter_id, voice} objects")),
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

			patchReq := gen.HandlersPatchNarrationRequest{}

			if segsJSON := optionalArg(req, "segments_json"); segsJSON != "" {
				var segs []gen.NarrationPatchSegmentEntry
				if err := json.Unmarshal([]byte(segsJSON), &segs); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("invalid segments_json: %v", err)), nil
				}
				patchReq.Segments = &segs
			}

			if chsJSON := optionalArg(req, "chapters_json"); chsJSON != "" {
				var chs []gen.NarrationPatchChapterEntry
				if err := json.Unmarshal([]byte(chsJSON), &chs); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("invalid chapters_json: %v", err)), nil
				}
				patchReq.Chapters = &chs
			}

			if patchReq.Segments == nil && patchReq.Chapters == nil {
				return mcp.NewToolResultError("specify segments_json and/or chapters_json"), nil
			}

			result, err := svc.PatchNarration(ctx, storyID, patchReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// credits_balance
	s.AddTool(
		tool("credits_balance",
			mcp.WithDescription("Get the authenticated user's credit balance. Check before expensive operations like narration."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			result, err := svc.GetCredits(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)
}
