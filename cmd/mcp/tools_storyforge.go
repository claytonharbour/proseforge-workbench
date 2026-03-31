package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/storyforge"
)

func registerStoryForgeTools(s *server.MCPServer, r *clientResolver) {
	// storyforge_chat_create — Start Story Forge Chat interview
	s.AddTool(
		tool("storyforge_chat_create",
			mcp.WithDescription("Start a new Story Forge Chat interview session"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.CreateSession(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// storyforge_chat_list — List chat sessions
	s.AddTool(
		tool("storyforge_chat_list",
			mcp.WithDescription("List Story Forge Chat sessions"),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.ListSessions(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// storyforge_chat_get — Get session with messages
	s.AddTool(
		tool("storyforge_chat_get",
			mcp.WithDescription("Get a Story Forge Chat session with its full message history"),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.GetSession(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// storyforge_chat_send — Send message
	s.AddTool(
		tool("storyforge_chat_send",
			mcp.WithDescription("Send a message in a Story Forge Chat interview and get AI response"),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}
			sendReq := api.ChatSendReq{Content: &content}
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.SendMessage(ctx, id, sendReq)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// storyforge_chat_finalize — Finalize interview
	s.AddTool(
		tool("storyforge_chat_finalize",
			mcp.WithDescription("Finalize a Story Forge Chat interview and trigger story generation"),
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Chat session ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "session_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.Finalize(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// storyforge_status — Poll generation progress
	s.AddTool(
		tool("storyforge_status",
			mcp.WithDescription("Get the current status of story generation pipeline"),
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
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.GetStatus(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// storyforge_meta_get — Review generated outline
	s.AddTool(
		tool("storyforge_meta_get",
			mcp.WithDescription("Get the generated story outline (story.md, characters.md, plot.md)"),
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
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.GetMeta(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// storyforge_meta_approve — Approve outline
	s.AddTool(
		tool("storyforge_meta_approve",
			mcp.WithDescription("Approve the generated outline and start section generation"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.ApproveMeta(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// storyforge_meta_regenerate — Retry outline generation
	s.AddTool(
		tool("storyforge_meta_regenerate",
			mcp.WithDescription("Regenerate the story outline (free retry)"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.RegenerateMeta(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// storyforge_resume — Resume failed generation
	s.AddTool(
		tool("storyforge_resume",
			mcp.WithDescription("Resume a failed or paused story generation"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
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
			svc := storyforge.NewService(client, storyforge.WithLogger(r.logger))
			result, err := svc.ResumeGeneration(ctx, storyID)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)
}
