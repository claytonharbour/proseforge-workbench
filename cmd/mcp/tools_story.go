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
			mcp.WithDescription("Browse stories with pagination, filtering, and search. Filter by status (pitch, draft, published, generating, failed), search by title, sort by date or rating. Pitch stories are excluded from default lists — use status=pitch explicitly."),
			mcp.WithString("status", mcp.Description("Filter: pitch, draft, published, unpublished, generating, failed, all (default excludes pitches)")),
			mcp.WithString("q", mcp.Description("Search query (matches title and other fields)")),
			mcp.WithString("sort", mcp.Description("Sort: date_desc (default), date_asc, updated_desc, updated_asc, rating_desc, rating_asc")),
			mcp.WithBoolean("narration", mcp.Description("Filter to stories with narration")),
			mcp.WithBoolean("audiobook", mcp.Description("Filter to stories with completed audiobook")),
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
			if q := optionalArg(req, "q"); q != "" {
				params.Q = &q
			}
			if s := optionalArg(req, "sort"); s != "" {
				params.Sort = &s
			}
			if optionalBoolArg(req, "narration") {
				t := true
				params.Narration = &t
			}
			if optionalBoolArg(req, "audiobook") {
				t := true
				params.Audiobook = &t
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
			mcp.WithDescription("Resolve a vanity URL (@handle/slug) to a story ID and metadata. Use when given a public URL like app.proseforge.ai/@handle/slug/read"),
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
			mcp.WithDescription("Get story metadata and section IDs. Returns thin response by default (no content). Set include_content=true for full text inline, or use story_export for formatted output."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithBoolean("include_content", mcp.Description("Include full section content (default false — use story_export for reading)")),
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
			if optionalBoolArg(req, "include_content") {
				result, err := svc.GetWithContent(ctx, id)
				if err != nil {
					return toolError(err, client), nil
				}
				return jsonResult(result)
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
			mcp.WithDescription("Download a story in the specified format. Synchronously generates and streams content. Use this to read a story, not story_get."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("format", mcp.Description("Export format: json (default), markdown, pdf, epub")),
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

	// story_assess_version
	s.AddTool(
		tool("story_assess_version",
			mcp.WithDescription("Assess quality at a specific version SHA. Synchronous — returns scores inline, no polling needed. Use story_versions to find SHAs. Great for before/after comparisons across rewrites."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("sha", mcp.Required(), mcp.Description("Version SHA from story_versions")),
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
			sha, err := requireArg(req, "sha")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.AssessQualityAtVersion(ctx, id, sha)
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
			mcp.WithDescription("Create a new story in draft status. Genre is resolved by name. To create a pre-writing idea instead, use story_pitch_create. For the full story lifecycle (pitch → draft → published), read the docs://series-workflow resource."),
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
			mcp.WithDescription("Publish a story. Only draft stories can be published. Set visibility to 'public' (default) or 'members' (login required)."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("visibility", mcp.Description("'public' or 'members' (default: public)")),
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
			visibility := optionalArg(req, "visibility")
			if err := svc.Publish(ctx, id, visibility); err != nil {
				return toolError(err, client), nil
			}
			msg := "Story published"
			if visibility != "" {
				msg += " with visibility: " + visibility
			}
			return mcp.NewToolResultText(msg + "."), nil
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

	// story_pitch_create
	s.AddTool(
		tool("story_pitch_create",
			mcp.WithDescription("Create a pitch — a pre-writing story idea. Pitches have planning data (meta) but no sections. Next steps: use story_meta_upsert to add premise, characters, and plot outline, then story_promote to transition to draft when ready to write sections. See docs://series-workflow for the full lifecycle."),
			mcp.WithString("genre", mcp.Required(), mcp.Description("Genre name (e.g., \"Historical Fiction\")")),
			mcp.WithString("title", mcp.Description("Story title (optional)")),
			mcp.WithString("tagline", mcp.Description("Story tagline (optional, set via update)")),
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

			result, err := svc.CreatePitch(ctx, createReq)
			if err != nil {
				return toolError(err, client), nil
			}

			// If tagline was provided, set it via update
			if tagline := optionalArg(req, "tagline"); tagline != "" && result.Id != nil {
				updateReq := api.UpdateStoryRequest{Tagline: &tagline}
				if err := svc.Update(ctx, *result.Id, updateReq); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("pitch created (%s) but failed to set tagline: %v", *result.Id, err)), nil
				}
			}

			return jsonResult(result)
		},
	)

	// story_promote
	s.AddTool(
		tool("story_promote",
			mcp.WithDescription("Promote a pitch to draft status. Enables section creation and publishing. Only works on stories with status 'pitch'."),
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
			if err := svc.Promote(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Pitch promoted to draft."), nil
		},
	)

	// story_meta_upsert
	s.AddTool(
		tool("story_meta_upsert",
			mcp.WithDescription("Write story planning data as markdown. Creates the file if missing, updates if present. Works on any story regardless of status.\n\nFormats by type:\n- 'story': Use ## headers for Genre & Tone, Central Theme, Setting, Core Conflict. The pipeline injects the entire file as context for section generation.\n- 'characters': Use ## CharacterName headers with Role, Background, Motivation underneath. Include whatever details matter — the AI reads the full text.\n- 'plot': Use ## Section 1, ## Section 2, etc. as headers with a brief summary of each section's plot beat. The pipeline parses these headers to inject per-section context during generation. Without ## Section N headers, the pipeline falls back to the whole outline."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("meta_type", mcp.Required(), mcp.Description("Type: 'story' (premise/overview), 'characters' (character profiles with ## Name headers), or 'plot' (plot outline with ## Section N headers)")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Planning data content (markdown)")),
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
			metaType, err := requireArg(req, "meta_type")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.UpsertMeta(ctx, storyID, metaType, content)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// story_update_visibility
	s.AddTool(
		tool("story_update_visibility",
			mcp.WithDescription("Change visibility of a published story. 'public' = anyone can read, 'members' = login required."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("visibility", mcp.Required(), mcp.Description("'public' or 'members'")),
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
			visibility, err := requireArg(req, "visibility")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.UpdateVisibility(ctx, id, visibility); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Visibility updated to " + visibility + "."), nil
		},
	)

	// section_create
	s.AddTool(
		tool("section_create",
			mcp.WithDescription("Create a new section in a story. For story content — not for planning data. For planning data (premise, characters, plot outline), use story_meta_upsert instead."),
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
			mcp.WithDescription("Primary write endpoint for BYOAI agents. Content stored in git with version history. Triggers quality assessment automatically."),
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
			mcp.WithDescription("Queue TTS narration for a story. Generates audio per-chapter and assembles M4B audiobook. Costs credits — use credits_estimate first. Poll with narration_status. Use narration_voices to list available voices."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("voice", mcp.Description("TTS voice name (e.g., 'Kore'). Use narration_voices to list options. Omit for server default.")),
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
			voice := optionalArg(req, "voice")
			if err := svc.StartNarration(ctx, storyID, voice); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration started."), nil
		},
	)

	// narration_status
	s.AddTool(
		tool("narration_status",
			mcp.WithDescription("Poll narration progress. Returns overall status, per-chapter status, stale chapter detection, and audiobook URL when complete. Provides chapter IDs for narration_regenerate, narration_segments, etc."),
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
			mcp.WithDescription("Re-narrate a chapter. By default only works if content changed (stale). Set force=true for voice changes or unconditional re-TTS."),
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
			mcp.WithDescription("Re-assemble M4B audiobook from existing chapter audio with updated metadata. No TTS, no credits."),
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
			mcp.WithDescription("Remove all narration data and audio for a story. Not reversible. Re-narrate with narration_start."),
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
			mcp.WithDescription("Re-TTS a single segment, optionally with a different voice. Rejects if chapter content changed — use narration_regenerate for stale chapters."),
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
			mcp.WithDescription("Batch update: re-TTS segments/chapters with new voices, restitch, rebuild audiobook — one call. Batch alternative to narration_segment_regenerate."),
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

	// credits_estimate
	s.AddTool(
		tool("credits_estimate",
			mcp.WithDescription("Returns estimated cost AND whether user can afford it. Call before narration, generation, or AI analysis."),
			mcp.WithString("operation", mcp.Required(), mcp.Description("Operation: narrate, generate, rewrite, image, avatar, patch, insights")),
			mcp.WithNumber("sections", mcp.Description("Number of sections (for narrate, generate, rewrite)")),
			mcp.WithNumber("segments", mcp.Description("Number of segments (for patch)")),
			mcp.WithBoolean("images", mcp.Description("Include image generation (for generate)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			op, err := requireArg(req, "operation")
			if err != nil {
				return toolError(err), nil
			}
			params := &gen.GetCreditsEstimateParams{Operation: op}
			if v := optionalIntArg(req, "sections", 0); v > 0 {
				params.Sections = &v
			}
			if v := optionalIntArg(req, "segments", 0); v > 0 {
				params.Segments = &v
			}
			if optionalBoolArg(req, "images") {
				t := true
				params.Images = &t
			}
			result, err := svc.EstimateCredits(ctx, params)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// credits_history
	s.AddTool(
		tool("credits_history",
			mcp.WithDescription("Paginated credit transaction history. Shows grants, reservations, and settlements. Filter by provider."),
			mcp.WithNumber("limit", mcp.Description("Max results (default 20, max 100)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := story.NewService(client, story.WithLogger(r.logger))
			limit := optionalIntArg(req, "limit", 20)
			result, err := svc.GetCreditHistory(ctx, limit)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// credits_balance
	s.AddTool(
		tool("credits_balance",
			mcp.WithDescription("Returns current credit balance and per-provider costs. Check before expensive operations."),
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

	// story_versions
	s.AddTool(
		tool("story_versions",
			mcp.WithDescription("List version history (git commits) for a story. Returns SHAs for use with story_version_get and story_version_diff."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithNumber("limit", mcp.Description("Max results (default 50, max 100)")),
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
			params := &gen.GetStoryIdVersionsParams{}
			if l := optionalIntArg(req, "limit", 0); l > 0 {
				params.Limit = &l
			}
			data, err := svc.ListVersions(ctx, id, params)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_version_get
	s.AddTool(
		tool("story_version_get",
			mcp.WithDescription("Get story content at a specific version (git SHA). Use story_versions to list available SHAs."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("sha", mcp.Required(), mcp.Description("Git commit SHA from story_versions")),
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
			sha, err := requireArg(req, "sha")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.GetVersion(ctx, id, sha)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_version_diff
	s.AddTool(
		tool("story_version_diff",
			mcp.WithDescription("Show diff between two story versions. Requires two SHAs from story_versions."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("from_sha", mcp.Required(), mcp.Description("Starting version SHA")),
			mcp.WithString("to_sha", mcp.Required(), mcp.Description("Ending version SHA")),
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
			fromSha, err := requireArg(req, "from_sha")
			if err != nil {
				return toolError(err), nil
			}
			toSha, err := requireArg(req, "to_sha")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.DiffVersions(ctx, id, fromSha, toSha)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_delete
	s.AddTool(
		tool("story_delete",
			mcp.WithDescription("Permanently delete a story, all sections, and git repo. Cancels running generation/narration. Not reversible."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
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
			if err := svc.Delete(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story deleted."), nil
		},
	)

	// section_delete
	s.AddTool(
		tool("section_delete",
			mcp.WithDescription("Delete a section from a story. Not reversible."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
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
			if err := svc.DeleteSection(ctx, storyID, sectionID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Section deleted."), nil
		},
	)

	// story_version_restore
	s.AddTool(
		tool("story_version_restore",
			mcp.WithDescription("Restore story content to a previous version. Use story_versions to find SHAs."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("sha", mcp.Required(), mcp.Description("Git SHA from story_versions")),
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
			sha, err := requireArg(req, "sha")
			if err != nil {
				return toolError(err), nil
			}
			data, err := svc.RestoreVersion(ctx, id, sha)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_meta_stale
	s.AddTool(
		tool("story_meta_stale",
			mcp.WithDescription("Check which sections are stale after editing planning data (meta). Returns affected section IDs."),
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
			data, err := svc.GetMetaStale(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// story_meta_acknowledge
	s.AddTool(
		tool("story_meta_acknowledge",
			mcp.WithDescription("Dismiss all staleness warnings after cosmetic meta edits that don't require section regeneration."),
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
			if err := svc.AcknowledgeMetaStale(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Staleness acknowledged."), nil
		},
	)

	// story_regenerate_tagline
	s.AddTool(
		tool("story_regenerate_tagline",
			mcp.WithDescription("Regenerate story tagline using AI. Queues a background job."),
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
			if err := svc.RegenerateTagline(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Tagline regeneration queued."), nil
		},
	)

	// story_regenerate_title
	s.AddTool(
		tool("story_regenerate_title",
			mcp.WithDescription("Regenerate story title using AI. Queues a background job."),
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
			if err := svc.RegenerateTitle(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Title regeneration queued."), nil
		},
	)

	// narration_regenerate_stale
	s.AddTool(
		tool("narration_regenerate_stale",
			mcp.WithDescription("Auto-detect chapters with changed content and regenerate their narration. Only re-narrates stale chapters."),
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
			data, err := svc.RegenerateStaleNarration(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// narration_acknowledge
	s.AddTool(
		tool("narration_acknowledge",
			mcp.WithDescription("Dismiss all narration staleness warnings in bulk. Use when content changes don't require re-narration."),
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
			if err := svc.AcknowledgeNarrationStale(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration staleness acknowledged."), nil
		},
	)
}
