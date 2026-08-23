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

// sectionMutationResponse formats the PUT/PATCH-section response for an MCP
// tool result. When includeLayout is false, the heavy `sections` field (full
// renumbered layout) is dropped, leaving section/commitSha/wordCount as a
// lightweight confirmation. When true, the full upstream body is returned.
// On any parse error the raw body is returned unchanged — better to surface
// upstream output verbatim than to swallow it.
func sectionMutationResponse(body json.RawMessage, includeLayout bool) string {
	if includeLayout {
		return string(body)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return string(body)
	}
	delete(m, "sections")
	out, err := json.Marshal(m)
	if err != nil {
		return string(body)
	}
	return string(out)
}

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
			mcp.WithDescription("Download a story in the specified format. Synchronously generates and streams content. Use this to read a story, not story_get. json and markdown come back inline; pdf and epub are binary and are written to out_path instead, because binary cannot survive a text channel."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("format", mcp.Description("Export format: json (default), markdown, pdf, epub")),
			mcp.WithString("out_path", mcp.Description("Where to write the file. Required for pdf and epub; ignored for json and markdown. Supports ~/.")),
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
			// Refuse before the request when the destination is missing, so a
			// caller is not charged a multi-megabyte render to be told no.
			if isBinaryExportFormat(format) && optionalArg(req, "out_path") == "" {
				return binaryExportResult(format, "", nil)
			}
			content, err := svc.Export(ctx, id, format)
			if err != nil {
				return toolError(err, client), nil
			}
			if isBinaryExportFormat(format) {
				return binaryExportResult(format, optionalArg(req, "out_path"), []byte(content))
			}
			return mcp.NewToolResultText(content), nil
		},
	)

	// story_sections
	s.AddTool(
		tool("story_sections",
			mcp.WithDescription("List a story's sections — id, name, order, status — without their content. This is the section-id lookup that works when you are a contributor rather than the owner: story_get reads the story record, which a grantee cannot see, while this reads the sections directly. Pair it with story_section to fetch content."),
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
			data, err := svc.ListSections(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			// Project the fields this tool promises. The endpoint returns full
			// section bodies, and forwarding them made the discovery call cost
			// more context than reading the story — 92KB for ten sections, which
			// is the opposite of why this tool exists.
			return jsonResult(projectSectionList(data))
		},
	)

	// story_section
	s.AddTool(
		tool("story_section",
			mcp.WithDescription("Get a single section's content and metadata (context-efficient). Requires story_id and section_id — get section IDs from story_sections (works as a contributor) or story_get (owner only)."),
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
			mcp.WithDescription(
				"Update a section's content, name, and/or position. Only provided fields are "+
					"changed — omit `content` to rename without rewriting, omit `name` to update "+
					"content without renaming, omit `order` to leave position alone. At least "+
					"one of content, name, or order must be provided. Content writes are stored "+
					"in git with version history and trigger quality assessment automatically; "+
					"name and order changes do not produce a commit.\n\n"+
					"For move-only operations, prefer section_reorder — it's a dedicated tool "+
					"with no risk of accidental content overwrite."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
			mcp.WithString("content", mcp.Description("New section content. Optional — omit to rename or reorder without rewriting.")),
			mcp.WithString("name", mcp.Description("New section name (rename). Optional.")),
			mcp.WithNumber("order", mcp.Description("New 0-indexed position within the story. Renumbers siblings. Out-of-range values clamp to the nearest end. Optional.")),
			mcp.WithBoolean("include_layout", mcp.Description("Include the full sections list (with renumbered sortOrders) in the response. Default false — response confirms the updated section plus commitSha and wordCount. Set true when you need to verify the new layout in one call instead of following up with story_get.")),
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

			writeReq := api.UpdateSectionRequest{}
			if c := optionalArg(req, "content"); c != "" {
				writeReq.Content = &c
			}
			if n := optionalArg(req, "name"); n != "" {
				writeReq.Name = &n
			}
			if o := optionalIntArg(req, "order", -1); o >= 0 {
				writeReq.Order = &o
			}
			if writeReq.Content == nil && writeReq.Name == nil && writeReq.Order == nil {
				return mcp.NewToolResultError("at least one of content, name, or order is required"), nil
			}

			body, err := svc.WriteSection(ctx, storyID, sectionID, writeReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(sectionMutationResponse(body, optionalBoolArg(req, "include_layout"))), nil
		},
	)

	// section_reorder
	s.AddTool(
		tool("section_reorder",
			mcp.WithDescription(
				"Move a section to a 0-indexed position within its story, renumbering siblings. "+
					"Out-of-range values clamp to the nearest end. Dedicated move-only tool — "+
					"no risk of accidental content or name change. For combined edits (content + "+
					"move, rename + move), use section_write."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
			mcp.WithNumber("order", mcp.Required(), mcp.Description("0-indexed target position within the story")),
			mcp.WithBoolean("include_layout", mcp.Description("Include the full sections list (with renumbered sortOrders) in the response. Default false — response confirms the moved section with its new sortOrder. Set true when you need to verify the new layout in one call instead of following up with story_get.")),
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
			order := optionalIntArg(req, "order", -1)
			if order < 0 {
				return mcp.NewToolResultError("order is required and must be >= 0"), nil
			}

			body, err := svc.ReorderSection(ctx, storyID, sectionID, order)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(sectionMutationResponse(body, optionalBoolArg(req, "include_layout"))), nil
		},
	)

	// narration_start
	s.AddTool(
		tool("narration_start",
			mcp.WithDescription("Queue TTS narration for a story — generates audio per-section and assembles the M4B audiobook. Costs credits (use credits_estimate first); poll with narration_status; list voices with narration_voices.\n\n**Destructive — guarded.** On a story that ALREADY has narration, narration_start **deletes the existing track and re-renders EVERYTHING from scratch** — full credit burn, including sections you didn't touch. Pass `confirm_replace: true` to do it deliberately; that forwards the backend `force` flag (#477). The backend independently enforces this with two distinguishable 409s: `narration_exists` = a terminal-state (complete/error) track already exists → set `confirm_replace: true` to replace it; `conflict` = a narration is in-flight → **never force-deletable**, wait for it or cancel first (neither this tool nor `force` overrides an in-flight track). Before reaching for it:\n- **Changing the voice?** Don't use narration_start. Re-narrate each section with the new voice via `narration_regenerate` (voice=…, force=true), then `narration_rebuild` to reassemble — same credit cost as a full restart but **non-destructive** (the old track survives until rebuild), and cheaper still if only some sections change.\n- **A section was inserted after baseline?** Check `narration_status.pending_sections`; regenerate each listed section with `narration_regenerate` (or use `narration_regenerate_stale` for the batch), then call `narration_rebuild`.\n- **A narration failed or stuck?** Use `narration_resume` / `narration_retry`.\nReserve narration_start (or confirm_replace=true) for a fresh story with no narration, or a deliberate full re-narration."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("voice", mcp.Description("TTS voice name (e.g., 'Kore'). Use narration_voices to list options. Omit for server default.")),
			mcp.WithBoolean("emphasis", mcp.Description("Tri-state prosody toggle (#642): omit to use the TTS worker's default emphasis-pause behavior; true/false pins it for this narration. RunPod Kokoro only — other providers ignore it; pre-emphasis workers ignore the field entirely.")),
			mcp.WithBoolean("confirm_replace", mcp.Description("Required on a story that already has a terminal-state (complete/error) narration: forwards the backend force flag (#477) to DELETE the existing track and re-render everything from scratch. Default false — the call is rejected client-side with safe alternatives, and the backend 409s `narration_exists` without it. Has no effect on an in-flight narration (backend 409s `conflict` regardless). Set true only to intentionally discard and rebuild.")),
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
			confirmReplace := optionalBoolArg(req, "confirm_replace")
			emphasis := optionalBoolPtrArg(req, "emphasis")
			if err := svc.StartNarration(ctx, storyID, voice, confirmReplace, emphasis); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Narration started."), nil
		},
	)

	// narration_status
	s.AddTool(
		tool("narration_status",
			mcp.WithDescription("Poll narration progress. Returns overall status, per-section status, stale section detection, audiobook URL when complete, and a `drift` envelope `{status, contentChangedSections[], nameChangedSections[], reordered, insertedSections[], deletedSections[]}` describing how the assembled M4B has diverged from current section state. Also returns `pending_sections`: story-section IDs with no narration entry. When it is non-empty, regenerate every listed section (individually with `narration_regenerate` or as a batch with `narration_regenerate_stale`) before rebuilding; `narration_rebuild` will refuse with 409 `sections_not_narrated` until they are narrated. `drift.status`: `fresh` = synced; `stale` = call `narration_rebuild` (free, no TTS) for name/order changes, or re-narrate the listed `contentChangedSections` for content drift; `baseline_needed` = audiobook predates drift tracking, one free rebuild establishes the baseline. For drift-only polling, prefer `narration_drift` (~30× cheaper per call). Provides section IDs for narration_regenerate, narration_segments, etc."),
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

	// narration_drift
	s.AddTool(
		tool("narration_drift",
			mcp.WithDescription("Cheap drift check for monitoring loops. Returns just the `drift` envelope from `narration_status` (~100–500 bytes vs ~5–15 KB for the full status — roughly 30× less context per call). Use this when polling to check if an audiobook needs rebuild; use `narration_status` when you also need per-section progress, audio URLs, or voice info. Returns `{drift: {status, contentChangedSections[], nameChangedSections[], reordered, insertedSections[], deletedSections[]}}`. `status`: `fresh` = synced; `stale` = call `narration_rebuild` (free, no TTS) for name/order changes, or re-narrate listed `contentChangedSections` for content drift; `baseline_needed` = audiobook predates drift tracking, one free rebuild establishes baseline."),
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
			full, err := svc.GetNarration(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			var m map[string]json.RawMessage
			if err := json.Unmarshal(full, &m); err != nil {
				return toolError(fmt.Errorf("parse narration status: %w", err), client), nil
			}
			drift, ok := m["drift"]
			if !ok {
				return mcp.NewToolResultText(`{"drift":null}`), nil
			}
			out, err := json.Marshal(map[string]json.RawMessage{"drift": drift})
			if err != nil {
				return toolError(fmt.Errorf("marshal drift: %w", err), client), nil
			}
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	// narration_audiobook
	s.AddTool(
		tool("narration_audiobook",
			mcp.WithDescription("Get audiobook download metadata for a story: `{ready, status, audiobook_url, total_duration_ms, total_file_size_bytes, voice, narration_id, completed_sections, total_sections}`. Returns the M4B's download URL and stats once narration completes — it does NOT stream the audio bytes; fetch the file yourself from `audiobook_url`. `ready` is true only when narration status is `complete` and a URL exists; otherwise poll `narration_status`."),
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
			// Build a metadata envelope from the status payload rather than
			// the /narration/audiobook endpoint, which streams the raw M4B
			// binary — JSON-marshaling those bytes always fails on the first
			// null byte. The download URL + stats live in narration status.
			full, err := svc.GetNarration(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			out, err := audiobookMetadata(full)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(out)
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
			mcp.WithDescription("Re-narrate ONE section (targeted TTS — costs credits for that section only). By default only runs if the section's content changed (stale); set force=true for an unconditional re-TTS or a voice change. **This is the non-destructive voice-change building block:** to re-voice a whole book, call this per section with the new `voice` (force=true) then `narration_rebuild` to reassemble — same total cost as `narration_start` but it never deletes the existing track. Cf. `narration_start` (destructive full re-narration), `narration_regenerate_stale` (auto-targets un-narrated sections), `narration_rebuild` (free reassembly, no TTS)."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID (from narration_status)")),
			mcp.WithBoolean("force", mcp.Description("Force regeneration even if content hasn't changed (default false)")),
			mcp.WithString("voice", mcp.Description("Voice override for this section (e.g., Puck, af_sarah). Use narration_voices to list options.")),
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
			force := optionalBoolArg(req, "force")
			voice := optionalArg(req, "voice")
			if err := svc.RegenerateSection(ctx, storyID, sectionID, force, voice); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Section %s regeneration started.", sectionID)), nil
		},
	)

	// narration_retry
	s.AddTool(
		tool("narration_retry",
			mcp.WithDescription("Retry a failed or stuck section narration. Resets error state and re-queues."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID (from narration_status)")),
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
			if err := svc.RetrySection(ctx, storyID, sectionID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Section %s retry started.", sectionID)), nil
		},
	)

	// narration_rebuild
	s.AddTool(
		tool("narration_rebuild",
			mcp.WithDescription("Re-assemble the M4B audiobook from existing per-section audio with updated metadata. **Free — no TTS, no credits, non-destructive** (reuses the audio you already have). Call when `narration_status` reports `drift.status != \"fresh\"` (section name changes, reordering, inserts, deletes) or `drift.status == \"baseline_needed\"` (one-time baseline for pre-drift-tracking audiobooks). If `pending_sections` is non-empty, rebuild refuses with 409 `sections_not_narrated` and the missing section IDs; regenerate every listed section first with `narration_regenerate` or `narration_regenerate_stale`. For content drift on specific sections, regenerate those first (`narration_regenerate`) then rebuild. The trio: `narration_start` (destructive full TTS — deletes + re-narrates), `narration_regenerate[_stale]` (targeted TTS for changed/un-narrated sections), `narration_rebuild` (this — free reassembly)."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithBoolean("section_announcements", mcp.Description("Insert TTS-generated section title announcements (default false)")),
			mcp.WithBoolean("emphasis", mcp.Description("Tri-state prosody toggle (#642): omit to leave the stored narration option untouched; true/false updates it. The new value only affects audio on a subsequent re-TTS (narration_regenerate / narration_start) — rebuild itself reuses existing audio.")),
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
			announcements := optionalBoolArg(req, "section_announcements")
			emphasis := optionalBoolPtrArg(req, "emphasis")
			if err := svc.RebuildNarration(ctx, storyID, announcements, emphasis); err != nil {
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

	// narration_section_cancel
	s.AddTool(
		tool("narration_section_cancel",
			mcp.WithDescription("Cancel a specific section's narration without deleting the whole narration"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
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
			if err := svc.CancelSection(ctx, storyID, sectionID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Section %s cancelled.", sectionID)), nil
		},
	)

	// narration_segments
	s.AddTool(
		tool("narration_segments",
			mcp.WithDescription("List segments for a section with text content, voice, and provider info. Requires story_id and section_id from narration_status. Returns segment IDs for narration_segment_regenerate."),
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
			result, err := svc.ListSegments(ctx, storyID, sectionID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// narration_segment_regenerate
	s.AddTool(
		tool("narration_segment_regenerate",
			mcp.WithDescription("Re-TTS a single segment, optionally with a different voice. Rejects if section content changed — use narration_regenerate for stale sections."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
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
			sectionID, err := requireArg(req, "section_id")
			if err != nil {
				return toolError(err), nil
			}
			segmentID, err := requireArg(req, "segment_id")
			if err != nil {
				return toolError(err), nil
			}
			voice := optionalArg(req, "voice")
			if err := svc.RegenerateSegment(ctx, storyID, sectionID, segmentID, voice); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Segment %s regeneration started.", segmentID)), nil
		},
	)

	// narration_patch
	s.AddTool(
		tool("narration_patch",
			mcp.WithDescription("Batch update: re-TTS segments/sections with new voices, restitch, rebuild audiobook — one call. Batch alternative to narration_segment_regenerate."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("segments_json", mcp.Description("JSON array of {section_id, segment_id, voice} objects")),
			mcp.WithString("sections_json", mcp.Description("JSON array of {section_id, voice} objects")),
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

			if secsJSON := optionalArg(req, "sections_json"); secsJSON != "" {
				var secs []gen.NarrationPatchSectionEntry
				if err := json.Unmarshal([]byte(secsJSON), &secs); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("invalid sections_json: %v", err)), nil
				}
				patchReq.Sections = &secs
			}

			if patchReq.Segments == nil && patchReq.Sections == nil {
				return mcp.NewToolResultError("specify segments_json and/or sections_json"), nil
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

// sectionSummary is the section-discovery projection: enough to choose a
// section and fetch it with story_section, and nothing else.
type sectionSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Order  int    `json:"order"`
	Status string `json:"status"`
	Words  int    `json:"words,omitempty"`
}

// projectSectionList strips section content from a sections payload. Falls back
// to the raw payload if the shape is unexpected, so an upstream change degrades
// to verbose rather than to empty.
func projectSectionList(raw json.RawMessage) any {
	var payload struct {
		Sections []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Status  string `json:"status"`
			Content string `json:"content"`
		} `json:"sections"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || payload.Sections == nil {
		return json.RawMessage(raw)
	}

	out := make([]sectionSummary, 0, len(payload.Sections))
	for i, sec := range payload.Sections {
		out = append(out, sectionSummary{
			ID:     sec.ID,
			Name:   sec.Name,
			Order:  i,
			Status: sec.Status,
			Words:  len(strings.Fields(sec.Content)),
		})
	}
	return map[string]any{"sections": out, "count": len(out)}
}
