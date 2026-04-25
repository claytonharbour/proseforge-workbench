package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// entityTypeArg returns the entity type, defaulting to "story".
func entityTypeArg(req mcp.CallToolRequest) string {
	if v := optionalArg(req, "entity_type"); v != "" {
		return v
	}
	return "story"
}

func registerRoomTools(s *server.MCPServer, r *clientResolver) {
	// room_send — Send a message to a room
	s.AddTool(
		tool("room_send",
			mcp.WithDescription(
				"Post a message to a conversation room. Rooms are broadcast streams — all "+
					"participants see all messages. Available on stories and series.\n\n"+
					"Sending the first message creates the room automatically — no setup needed.\n\n"+
					"Include your identity in 'agent' and optionally 'perspective' (your craft "+
					"lens) and 'target' (which topic this relates to)."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story or series ID")),
			mcp.WithString("agent", mcp.Required(), mcp.Description("Your identity — title-of-the-moment")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message body (markdown)")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default) or 'series'")),
			mcp.WithString("perspective", mcp.Description("Craft lens (artificer, keeper, quality, security, etc.)")),
			mcp.WithString("target", mcp.Description("Topic this relates to (plot/Section 3, characters/Mara, process/status)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			entityID, err := requireArg(req, "entity_id")
			if err != nil {
				return toolError(err), nil
			}
			agent, err := requireArg(req, "agent")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}

			msg := api.SendRoomMessageRequest{
				Agent:       agent,
				Perspective: optionalArg(req, "perspective"),
				Target:      optionalArg(req, "target"),
				Content:     content,
			}

			result, err := client.SendRoomMessage(ctx, entityTypeArg(req), entityID, msg)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// room_read — Read messages from a room
	s.AddTool(
		tool("room_read",
			mcp.WithDescription(
				"Read messages from a conversation room. Returns the full history, or a "+
					"delta from a cursor position.\n\n"+
					"Use 'since' (the lastId from a previous read) to get only new messages. "+
					"Omit 'since' for full history.\n\n"+
					"This is a broadcast read — every caller sees all messages."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story or series ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default) or 'series'")),
			mcp.WithString("since", mcp.Description("Cursor: read messages after this ID (from previous lastId)")),
			mcp.WithString("order", mcp.Description("Sort order: 'asc' (oldest-first, default — use for reading conversations) or 'desc' (newest-first — use for peeking at the latest messages without loading full history)")),
			mcp.WithNumber("limit", mcp.Description("Max messages to return (default 1000)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			entityID, err := requireArg(req, "entity_id")
			if err != nil {
				return toolError(err), nil
			}

			since := optionalArg(req, "since")
			order := optionalArg(req, "order")
			limit := optionalIntArg(req, "limit", 1000)

			result, err := client.ReadRoomMessages(ctx, entityTypeArg(req), entityID, since, limit, order)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// room_status — Get room status
	s.AddTool(
		tool("room_status",
			mcp.WithDescription(
				"Check whether a room exists, is active or archived, and how many "+
					"messages it contains."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story or series ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default) or 'series'")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			entityID, err := requireArg(req, "entity_id")
			if err != nil {
				return toolError(err), nil
			}

			result, err := client.GetRoomStatus(ctx, entityTypeArg(req), entityID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// room_archive — Archive a room
	s.AddTool(
		tool("room_archive",
			mcp.WithDescription("Archive a room. Reads still work, writes are rejected. Use room_unarchive to re-enable."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story or series ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default) or 'series'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			entityID, err := requireArg(req, "entity_id")
			if err != nil {
				return toolError(err), nil
			}

			if err := client.ArchiveRoom(ctx, entityTypeArg(req), entityID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Room archived."), nil
		},
	)

	// room_unarchive — Unarchive a room
	s.AddTool(
		tool("room_unarchive",
			mcp.WithDescription("Unarchive a room, re-enabling writes."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story or series ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default) or 'series'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}

			entityID, err := requireArg(req, "entity_id")
			if err != nil {
				return toolError(err), nil
			}

			if err := client.UnarchiveRoom(ctx, entityTypeArg(req), entityID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Room unarchived."), nil
		},
	)
}
