package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/series"
)

func registerSeriesTools(s *server.MCPServer, r *clientResolver) {
	// series_list — List user's series
	s.AddTool(
		tool("series_list",
			mcp.WithDescription("List the authenticated user's series"),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.List(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_create — Create a new series
	s.AddTool(
		tool("series_create",
			mcp.WithDescription("Create a new series"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Series name")),
			mcp.WithString("description", mcp.Description("Series description")),
			mcp.WithString("genre_id", mcp.Description("Genre ID")),
			mcp.WithString("tone_id", mcp.Description("Tone ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			name, err := requireArg(req, "name")
			if err != nil {
				return toolError(err), nil
			}
			createReq := api.CreateSeriesReq{Name: &name}
			if desc := optionalArg(req, "description"); desc != "" {
				createReq.Description = &desc
			}
			if gid := optionalArg(req, "genre_id"); gid != "" {
				createReq.GenreId = &gid
			}
			if tid := optionalArg(req, "tone_id"); tid != "" {
				createReq.ToneId = &tid
			}

			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.Create(ctx, createReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_get — Get series details
	s.AddTool(
		tool("series_get",
			mcp.WithDescription("Get series details by ID"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.Get(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_update — Update series metadata
	s.AddTool(
		tool("series_update",
			mcp.WithDescription("Update a series' metadata (name, description, genre, tone)"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("name", mcp.Description("New series name")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("genre_id", mcp.Description("New genre ID")),
			mcp.WithString("tone_id", mcp.Description("New tone ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			updateReq := api.UpdateSeriesReq{}
			if name := optionalArg(req, "name"); name != "" {
				updateReq.Name = &name
			}
			if desc := optionalArg(req, "description"); desc != "" {
				updateReq.Description = &desc
			}
			if gid := optionalArg(req, "genre_id"); gid != "" {
				updateReq.GenreId = &gid
			}
			if tid := optionalArg(req, "tone_id"); tid != "" {
				updateReq.ToneId = &tid
			}

			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.Update(ctx, id, updateReq); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Series updated."), nil
		},
	)

	// series_archive — Archive a series
	s.AddTool(
		tool("series_archive",
			mcp.WithDescription("Archive (delete) a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.Archive(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Series archived."), nil
		},
	)

	// series_world_get — Get world overview
	s.AddTool(
		tool("series_world_get",
			mcp.WithDescription("Get the world overview document for a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.GetWorld(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// series_world_update — Update world overview
	s.AddTool(
		tool("series_world_update",
			mcp.WithDescription("Update the world overview document for a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("World overview content (markdown)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.UpdateWorld(ctx, id, content); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("World overview updated."), nil
		},
	)

	// series_timeline_get — Get canon timeline
	s.AddTool(
		tool("series_timeline_get",
			mcp.WithDescription("Get the full canon timeline (all sections assembled). For updating individual book timelines, use series_timeline_sections to list slugs, then series_timeline_section_update to write a specific book's events."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.GetTimeline(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// series_timeline_sections — List timeline sections
	s.AddTool(
		tool("series_timeline_sections",
			mcp.WithDescription("List timeline sections with slugs and titles. Use slugs with series_timeline_section_get and series_timeline_section_update to read/write individual sections."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.ListTimelineSections(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_timeline_section_get — Get a single timeline section by slug
	s.AddTool(
		tool("series_timeline_section_get",
			mcp.WithDescription("Get a single timeline section by slug. Use series_timeline_sections to list available slugs."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Timeline section slug (from series_timeline_sections)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.GetTimelineSection(ctx, id, slug)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_timeline_section_update — Update a single timeline section by slug
	s.AddTool(
		tool("series_timeline_section_update",
			mcp.WithDescription("Update a single timeline section by slug. Write events for one book without rewriting the full timeline. Use series_timeline_sections to find the slug for your book."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Timeline section slug (from series_timeline_sections)")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Section title (e.g. 'Book 3: Dead Reckoning')")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Section content (markdown — list of events)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			title, err := requireArg(req, "title")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.UpdateTimelineSection(ctx, id, slug, title, content)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_timeline_section_delete — Delete a timeline section
	s.AddTool(
		tool("series_timeline_section_delete",
			mcp.WithDescription("Delete a timeline section. Not reversible."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Timeline section slug")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.DeleteTimelineSection(ctx, id, slug); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Timeline section deleted."), nil
		},
	)

	// series_stories_reorder — Set book number order
	s.AddTool(
		tool("series_stories_reorder",
			mcp.WithDescription("Set book number order for stories in a series. Pass all story IDs in desired order."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("story_ids", mcp.Required(), mcp.Description("Ordered list of story IDs — first becomes book 1, second becomes book 2, etc.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			idsStr, err := requireArg(req, "story_ids")
			if err != nil {
				return toolError(err), nil
			}
			var ids []string
			if err := json.Unmarshal([]byte(idsStr), &ids); err != nil {
				return mcp.NewToolResultError("story_ids must be a JSON array of strings"), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.ReorderSeriesStories(ctx, id, ids); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story order updated."), nil
		},
	)

	// series_timeline_reorder — Set timeline section order
	s.AddTool(
		tool("series_timeline_reorder",
			mcp.WithDescription("Set display order for timeline sections. Pass all slugs in desired order."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slugs", mcp.Required(), mcp.Description("Ordered list of timeline section slugs")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slugsStr, err := requireArg(req, "slugs")
			if err != nil {
				return toolError(err), nil
			}
			var slugs []string
			if err := json.Unmarshal([]byte(slugsStr), &slugs); err != nil {
				return mcp.NewToolResultError("slugs must be a JSON array of strings"), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.ReorderTimelineSections(ctx, id, slugs); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Timeline section order updated."), nil
		},
	)

	// series_character_create — Create character in series
	s.AddTool(
		tool("series_character_create",
			mcp.WithDescription("Create a character in a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
			mcp.WithString("role", mcp.Description("Character role (e.g. protagonist, recurring, minor)")),
			mcp.WithString("profile", mcp.Description("Character profile (markdown — description, voice, relationships)")),
			mcp.WithString("status", mcp.Description("Character status (e.g. active, deceased, archived)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			name, err := requireArg(req, "name")
			if err != nil {
				return toolError(err), nil
			}
			createReq := api.CreateCharacterReq{Name: &name}
			if role := optionalArg(req, "role"); role != "" {
				createReq.Role = &role
			}
			if profile := optionalArg(req, "profile"); profile != "" {
				createReq.Profile = &profile
			}
			if status := optionalArg(req, "status"); status != "" {
				createReq.Status = &status
			}

			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.CreateCharacter(ctx, seriesID, createReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_character_list — List characters
	s.AddTool(
		tool("series_character_list",
			mcp.WithDescription("List all characters in a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.ListCharacters(ctx, seriesID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_character_get — Get character profile
	s.AddTool(
		tool("series_character_get",
			mcp.WithDescription("Get a character's profile by slug"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Character slug")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.GetCharacter(ctx, seriesID, slug)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_character_update — Update character profile
	s.AddTool(
		tool("series_character_update",
			mcp.WithDescription("Update a character's profile"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Character slug")),
			mcp.WithString("name", mcp.Description("New character name")),
			mcp.WithString("role", mcp.Description("New role")),
			mcp.WithString("profile", mcp.Description("New profile (markdown)")),
			mcp.WithString("status", mcp.Description("New status")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			updateReq := api.UpdateCharacterReq{}
			if name := optionalArg(req, "name"); name != "" {
				updateReq.Name = &name
			}
			if role := optionalArg(req, "role"); role != "" {
				updateReq.Role = &role
			}
			if profile := optionalArg(req, "profile"); profile != "" {
				updateReq.Profile = &profile
			}
			if status := optionalArg(req, "status"); status != "" {
				updateReq.Status = &status
			}

			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.UpdateCharacter(ctx, seriesID, slug, updateReq); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Character updated."), nil
		},
	)

	// series_character_delete — Delete character
	s.AddTool(
		tool("series_character_delete",
			mcp.WithDescription("Delete a character from a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Character slug")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			slug, err := requireArg(req, "slug")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.DeleteCharacter(ctx, seriesID, slug); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Character deleted."), nil
		},
	)

	// series_stories_list — List stories in series
	s.AddTool(
		tool("series_stories_list",
			mcp.WithDescription("List stories linked to a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.ListStories(ctx, seriesID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// series_stories_add — Link story to series
	s.AddTool(
		tool("series_stories_add",
			mcp.WithDescription("Link an existing story to a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID to link")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.AddStory(ctx, seriesID, storyID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story linked to series."), nil
		},
	)

	// series_plan — Create StorySeed-seeded Story Forge Chat from series context
	s.AddTool(
		tool("series_plan",
			mcp.WithDescription("Create a Story Forge Chat session seeded with series context (world, characters, timeline). The AI interviews the author with full series awareness."),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithNumber("book_number", mcp.Description("Book number for this installment (0 = auto-detect next)")),
			mcp.WithString("notes", mcp.Description("Author notes injected into AI context")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			planReq := api.PlanStoryReq{}
			if bn := optionalIntArg(req, "book_number", 0); bn > 0 {
				planReq.BookNumber = &bn
			}
			if notes := optionalArg(req, "notes"); notes != "" {
				planReq.Notes = &notes
			}

			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.PlanStory(ctx, seriesID, planReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_stories_remove — Unlink story from series
	s.AddTool(
		tool("series_stories_remove",
			mcp.WithDescription("Remove a story from a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID to unlink")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			seriesID, err := requireArg(req, "series_id")
			if err != nil {
				return toolError(err), nil
			}
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			if err := svc.RemoveStory(ctx, seriesID, storyID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Story removed from series."), nil
		},
	)
}
