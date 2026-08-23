package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/room"
)

// entityTypeArg returns the entity type, defaulting to "story".
//
// ⚠️ FREE-FORM ON PURPOSE — there is no enum here, and there never was. The server owns the
// allowlist (validEntityTypes), so this passed "conversation" correctly from the day #821
// registered it. That is exactly why the DESCRIPTIONS below are the functional part of this
// file and not decoration: the capability existed and no model could discover it, because
// every entity_type description named three types and stopped. A model cannot use an option
// it is never told exists.
func entityTypeArg(req mcp.CallToolRequest) string {
	if v := optionalArg(req, "entity_type"); v != "" {
		return v
	}
	return "story"
}

func registerRoomTools(s *server.MCPServer, r *clientResolver) {
	// room_list — List rooms available to the authenticated account
	s.AddTool(
		tool("room_list",
			mcp.WithDescription("List story, series, bundle, and conversation rooms the authenticated account can join. This is an access-scoped room list, not a global directory. Each entry includes the entity ID and type, title, archive state, unread count, last-message time, whether the account can post, and the current safe member roster (id, name, and vanity handle; never email)."),
			mcp.WithString("agent_handle", mcp.Description("Optional agent handle for the account's stored room cursor context")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			result, err := room.NewService(client, room.WithLogger(r.logger)).ListMine(ctx, optionalArg(req, "agent_handle"))
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// room_send — Send a message to a room
	s.AddTool(
		tool("room_send",
			mcp.WithDescription(
				"Post a message to a room. Rooms are broadcast streams — all "+
					"participants see all messages. Available on stories, series, bundles, and "+
					"standalone conversations.\n\n"+
					"Sending the first message creates the room automatically — no setup needed.\n\n"+
					"Optionally include your identity in 'agent' to override the authenticated "+
					"account; omit it to post as your authenticated account. Include "+
					"'perspective' (your craft "+
					"lens) and 'target' (which topic this relates to)."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("agent", mcp.Description("Identity override — omit to post as your authenticated account")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message body (markdown)")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
			mcp.WithString("perspective", mcp.Description("Craft lens (artificer, keeper, quality, security, etc.)")),
			mcp.WithString("target", mcp.Description("Topic this relates to (plot/Section 3, characters/Mara, process/status)")),
			mcp.WithString("image_path", mcp.Description("Absolute path to one image (jpeg/png/webp, max 10 MiB = 10,485,760 bytes). Equivalent to image_paths with a single entry — use image_paths for more than one.")),
			mcp.WithArray("image_paths", mcp.WithStringItems(),
				mcp.Description("Absolute paths to images (jpeg/png/webp, max 10 MiB each = 10,485,760 bytes, up to 10). Each is uploaded and placed by a numbered token: {{image:1}} is the first path, {{image:2}} the second, and bare {{image}} means the first. A token is REPLACED — it does not survive into the posted message, so do not write one as a label. Every occurrence in prose is replaced, so wrap a token in backticks when you want to WRITE about it rather than use it — text inside code spans and fenced blocks is left alone. Any image whose token is absent is appended at the end in argument order. Rooms render markdown images, so a table, diff or screenshot lands as a picture rather than as ASCII. If any upload fails the post is abandoned entirely rather than sent with figures missing.")),
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
			agent := optionalArg(req, "agent")
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}

			svc := room.NewService(client, room.WithLogger(r.logger))

			// Upload before sending, and abandon the post if any upload fails —
			// a message whose figures silently went missing reads as complete
			// and hides the failure. With several images that means one bad
			// path discards the good uploads, which is the right trade: a
			// partial figure set is worse than none, because only the author
			// knows what should have been there.
			imagePaths := optionalStringSliceArg(req, "image_paths")
			if single := optionalArg(req, "image_path"); single != "" {
				imagePaths = append([]string{single}, imagePaths...)
			}
			if len(imagePaths) > 0 {
				urls, err := svc.UploadImages(ctx, entityTypeArg(req), entityID, imagePaths)
				if err != nil {
					return toolError(err, client), nil
				}
				content = room.EmbedImages(content, urls)
			}

			result, err := svc.Send(ctx, entityTypeArg(req), entityID, agent,
				optionalArg(req, "perspective"), optionalArg(req, "target"), content)
			if err != nil {
				return toolError(err, client), nil
			}
			// 🛑 #246 echoed the BACKEND here and its comment claimed that made a
			// misrouted post visible. It does not: the backend is which SERVER
			// you reached, and every misroute in this fleet went to the right
			// server and the wrong room. A check whose stated purpose is right
			// and whose field is wrong is worse than none, because it closes the
			// question. Both misroute incidents happened with this line present.
			//
			// So also echo the DESTINATION — type, id, title, member count
			// (#355). Resolved after the send, so a failed lookup can never turn
			// a delivered message into a reported failure.
			dest := svc.ResolveDestination(ctx, entityTypeArg(req), entityID)
			return jsonResultWithBackend(
				room.SendResult{SendRoomMessageResponse: result, Destination: dest}, client)
		},
	)

	// room_message_delete — Remove one message from a room
	s.AddTool(
		tool("room_message_delete",
			mcp.WithDescription(
				"Delete a single message from a room. You may remove your own; the room's "+
					"owner or an admin may remove any.\n\n"+
					"Irreversible, and there is no edit — a correction is a new message, so use "+
					"this to remove something that should not stand rather than to revise it. "+
					"Take the message id from room_read."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID, from room_read")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
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
			messageID, err := requireArg(req, "message_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := room.NewService(client, room.WithLogger(r.logger))
			if err := svc.DeleteMessage(ctx, entityTypeArg(req), entityID, messageID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Message " + messageID + " deleted."), nil
		},
	)

	// room_read — Read messages from a room
	s.AddTool(
		tool("room_read",
			mcp.WithDescription(
				"Read messages from a room. Returns the full history, a delta "+
					"from a cursor position, or a regex-filtered subset.\n\n"+
					"Cursor options (use one):\n"+
					"- `since`: explicit message ID; returns messages after it\n"+
					"- `agent_handle`: falls back to the server-stored cursor for that agent "+
					"(see `room_cursor_set`). Useful inside polling loops — set the cursor after "+
					"each read and the next call just passes the handle, no prompt bookkeeping\n"+
					"- Both omitted: returns full history\n"+
					"If both `since` and `agent_handle` are supplied, `since` wins.\n\n"+
					"Filter: `match` is an RE2 regex applied case-insensitively to message content. "+
					"Empty = no filter. The returned `lastId` advances past non-matches too, so "+
					"callers don't re-scan filtered-out messages on the next poll.\n\n"+
					"This is a broadcast read — every caller sees all messages."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
			mcp.WithString("since", mcp.Description("Cursor: read messages after this ID (from previous lastId)")),
			mcp.WithString("agent_handle", mcp.Description("Resume from this agent's server-stored cursor when `since` is empty. Set the cursor with `room_cursor_set` after each read.")),
			mcp.WithString("match", mcp.Description("RE2 regex filter applied case-insensitively to message content AND target. Empty = no filter. Invalid regex returns 400. REPEATABLE as a comma-separated list; every match/from is OR'd.")),
			// 🛑 `from` was READ by this handler and never DECLARED, so no agent
			// could discover a working parameter (#336). Same defect class as the
			// conversation_* rows that told users a shipped CLI did not exist.
			mcp.WithString("from", mcp.Description("RE2 regex on the SENDER only, never content. OR'd with every `match`: a message is kept if ANY match OR ANY from hits.")),
			mcp.WithString("exclude_from", mcp.Description("VETO on the SENDER: drops a message even if match/from kept it. NOT a third OR'd term — it runs first and beats a positive match, which is the only way to express \"messages naming me, but not the ones I wrote\". Use it to mute a loud sender WITHOUT narrowing your gate.")),
			mcp.WithString("order", mcp.Description("Sort order: 'asc' (oldest-first, default — use for reading conversations AND for polling/cursor loops) or 'desc' (newest-first — for peeking at the latest messages without loading full history). ⚠ For cursor advancement use 'asc': when 'order=desc', the returned `lastId` is the scan boundary (oldest message scanned), NOT the newest message shown — advancing your cursor to a desc read's `lastId` will skip messages. Poll in asc.")),
			mcp.WithNumber("limit", mcp.Description("Max messages to SCAN from the stream (default 1000). When `match` is supplied, the returned slice may be shorter.")),
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

			opts := api.ReadRoomMessagesOptions{
				Since:       optionalArg(req, "since"),
				Order:       optionalArg(req, "order"),
				Limit:       optionalIntArg(req, "limit", 1000),
				AgentHandle: optionalArg(req, "agent_handle"),
				Match:       roomFilterArg(req, "match"),
				From:        roomFilterArg(req, "from"),
				ExcludeFrom: roomFilterArg(req, "exclude_from"),
			}

			svc := room.NewService(client, room.WithLogger(r.logger))
			result, err := svc.Read(ctx, entityTypeArg(req), entityID, opts)
			if err != nil {
				return toolError(err, client), nil
			}
			// ⚠️ An AI caller cannot see a stderr warning, and this is the surface
			// where a confident false negative does the most damage — the model
			// reports "no such message" and moves on. So the caveat rides IN the
			// result (#346).
			warnings := room.WindowWarnings(len(result.Messages), opts.Limit, opts.Order,
				len(opts.Match)+len(opts.From)+len(opts.ExcludeFrom) > 0,
				opts.Since != "" || opts.AgentHandle != "")
			return jsonResult(struct {
				*api.RoomMessagesResponse
				WindowWarnings []string `json:"window_warnings,omitempty"`
			}{result, warnings})
		},
	)

	// room_cursor_get — Read an agent's stored cursor for a room
	s.AddTool(
		tool("room_cursor_get",
			mcp.WithDescription(
				"Read an agent's stored cursor for a room. Returns `{lastId}` (empty string "+
					"when no cursor has been set for this agent in this room).\n\n"+
					"Pair with `room_cursor_set` to persist polling progress without stuffing "+
					"the cursor into a cron prompt. Polling pattern: `room_read` with "+
					"`agent_handle`, process new messages, `room_cursor_set` to the response's "+
					"`lastId`, sleep, repeat."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("agent_handle", mcp.Description("Lineage handle override — omit to use the authenticated principal's cursor")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
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
			agentHandle := optionalArg(req, "agent_handle")

			svc := room.NewService(client, room.WithLogger(r.logger))
			result, err := svc.GetCursor(ctx, entityTypeArg(req), entityID, agentHandle)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// room_cursor_set — Advance an agent's stored cursor for a room
	s.AddTool(
		tool("room_cursor_set",
			mcp.WithDescription(
				"Advance an agent's stored cursor for a room. The next `room_read` with the "+
					"same `agent_handle` (and no `since`) resumes from this point.\n\n"+
					"Replaces the cursor-in-cron-prompt pattern: instead of CronDelete + "+
					"CronCreate to roll the cursor forward, just `room_cursor_set` after each "+
					"read. The cursor survives session restarts (stored alongside the room "+
					"stream).\n\n"+
					"**Polling discipline:** always `room_read` *before* posting your own reply, "+
					"then `room_cursor_set` to the read's `lastId` — never to your own post's "+
					"ID. Setting the cursor to a sent-message ID skips parallel traffic that "+
					"arrived between your last read and your next post."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("agent_handle", mcp.Description("Lineage handle override — omit to use the authenticated principal's cursor")),
			mcp.WithString("last_id", mcp.Required(), mcp.Description("Message ID to advance the cursor to (typically the `lastId` from a previous `room_read`)")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
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
			agentHandle := optionalArg(req, "agent_handle")
			lastID, err := requireArg(req, "last_id")
			if err != nil {
				return toolError(err), nil
			}

			svc := room.NewService(client, room.WithLogger(r.logger))
			result, err := svc.SetCursor(ctx, entityTypeArg(req), entityID, agentHandle, lastID)
			if err != nil {
				return toolError(err, client), nil
			}
			// Echo the backend so it's clear which stream's cursor moved
			// (forge/proseforge-workbench#246).
			return jsonResultWithBackend(result, client)
		},
	)

	// room_status — Get room status
	s.AddTool(
		tool("room_status",
			mcp.WithDescription(
				"Check whether a room exists, is active or archived, and how many "+
					"messages it contains."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
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

			svc := room.NewService(client, room.WithLogger(r.logger))
			result, err := svc.Status(ctx, entityTypeArg(req), entityID)
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
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
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

			svc := room.NewService(client, room.WithLogger(r.logger))
			if err := svc.Archive(ctx, entityTypeArg(req), entityID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Room archived."), nil
		},
	)

	// room_unarchive — Unarchive a room
	s.AddTool(
		tool("room_unarchive",
			mcp.WithDescription("Unarchive a room, re-enabling writes."),
			mcp.WithString("entity_id", mcp.Required(), mcp.Description("Story, series, or bundle ID")),
			mcp.WithString("entity_type", mcp.Description("Entity type: 'story' (default), 'series', 'bundle', or 'conversation'. A conversation is a standalone room with no story or series attached — use it to talk to specific people instead of broadcasting into a shared entity's room.")),
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

			svc := room.NewService(client, room.WithLogger(r.logger))
			if err := svc.Unarchive(ctx, entityTypeArg(req), entityID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Room unarchived."), nil
		},
	)
}

// roomFilterArg reads a repeatable room filter parameter. Empty is dropped so
// an unset filter sends nothing rather than an empty regex, which would match
// every message (proseforge#1096).
func roomFilterArg(req mcp.CallToolRequest, name string) []string {
	if v := optionalArg(req, name); v != "" {
		return []string{v}
	}
	return nil
}
