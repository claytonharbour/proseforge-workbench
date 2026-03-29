package main

import (
	"context"

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
			mcp.WithDescription("Get the canon timeline for a series"),
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

	// series_timeline_update — Update canon timeline
	s.AddTool(
		tool("series_timeline_update",
			mcp.WithDescription("Update the canon timeline for a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Timeline content (markdown)")),
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
			if err := svc.UpdateTimeline(ctx, id, content); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Timeline updated."), nil
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

	// series_chat_create — Start world-building chat
	s.AddTool(
		tool("series_chat_create",
			mcp.WithDescription("Start a new world-building chat session for a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
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
			result, err := svc.CreateChat(ctx, seriesID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_chat_list — List chat sessions
	s.AddTool(
		tool("series_chat_list",
			mcp.WithDescription("List world-building chat sessions for a series"),
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
			result, err := svc.ListChats(ctx, seriesID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// series_chat_get — Get chat session with messages
	s.AddTool(
		tool("series_chat_get",
			mcp.WithDescription("Get a world-building chat session with its messages"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
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
			sessionID, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.GetChat(ctx, seriesID, sessionID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_chat_send — Send message in chat
	s.AddTool(
		tool("series_chat_send",
			mcp.WithDescription("Send a message in a world-building chat session and get AI response"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message content")),
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
			sessionID, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}
			sendReq := api.SeriesChatSendReq{Content: &content}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.SendChatMessage(ctx, seriesID, sessionID, sendReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_chat_finalize — Finalize chat session
	s.AddTool(
		tool("series_chat_finalize",
			mcp.WithDescription("Finalize a world-building chat session"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
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
			sessionID, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.FinalizeChat(ctx, seriesID, sessionID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// series_chat_harvest — Extract metadata to git
	s.AddTool(
		tool("series_chat_harvest",
			mcp.WithDescription("Extract world/character/timeline metadata from a chat session to git"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
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
			sessionID, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := series.NewService(client, series.WithLogger(r.logger))
			result, err := svc.HarvestChat(ctx, seriesID, sessionID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// series_harvest_all — Harvest all sessions
	s.AddTool(
		tool("series_harvest_all",
			mcp.WithDescription("Harvest metadata from all chat sessions in a series"),
			mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
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
			result, err := svc.HarvestAllChats(ctx, seriesID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
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
