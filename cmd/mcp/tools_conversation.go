package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/conversation"
)

// ⚑ FOUR TOOLS, NOT SIX. Messaging a conversation is `room_send` / `room_read` with
// entity_type "conversation" — already working, now documented (1724804). Adding
// conversation_send/read would be a second name for one primitive, and the room tools
// already carry the cursor, filtering and archive semantics that would have to be
// duplicated alongside them.
func registerConversationTools(s *server.MCPServer, r *clientResolver) {
	s.AddTool(
		tool("conversation_list",
			mcp.WithDescription("List conversations the authenticated account can reach — ones you started and ones you were added to. A conversation is a standalone room with no story or series behind it: use it to talk to specific people instead of broadcasting into a shared entity's room. Each entry gives the id, title (null for a 1:1 — identify it by its members instead), whether you are the owner, and timestamps. Message it with room_send/room_read using entity_type 'conversation'."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			result, err := conversation.NewService(client, conversation.WithLogger(r.logger)).List(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	s.AddTool(
		tool("conversation_create",
			mcp.WithDescription("Start a new conversation — a standalone room with no story or series attached. Returns its id, which you then use with conversation_add to bring people in and with room_send (entity_type 'conversation') to talk.\n\nThe conversation is created EMPTY: membership is granted per person afterwards, not listed here. Title is optional and best left blank for a 1:1, where the members identify it better than a name would."),
			mcp.WithString("title", mcp.Description("Optional name. Omit for a 1:1 — it will be identified by who is in it.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			result, err := conversation.NewService(client, conversation.WithLogger(r.logger)).
				Create(ctx, optionalArg(req, "title"))
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	s.AddTool(
		tool("conversation_add",
			mcp.WithDescription("Add someone to a conversation, naming them by PRINCIPAL ID — the id from friends_list, not an email address. The API stopped disclosing friends' email addresses, so an id is what a caller actually holds.\n\nOnly the creator can add members, and only people you are already friends with: a refusal here is usually about the friendship, not the capability. Capability is 'room:post' (read and write, the default) or 'room:enter' (read only)."),
			mcp.WithString("conversation_id", mcp.Required(), mcp.Description("Conversation ID from conversation_create or conversation_list")),
			mcp.WithString("principal_id", mcp.Required(), mcp.Description("The person's principal ID — from friends_list. NOT an email address.")),
			mcp.WithString("capability", mcp.Description("'room:post' (read+write, default) or 'room:enter' (read only)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "conversation_id")
			if err != nil {
				return toolError(err), nil
			}
			principal, err := requireArg(req, "principal_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := conversation.NewService(client, conversation.WithLogger(r.logger))
			if err := svc.AddMember(ctx, id, principal, optionalArg(req, "capability")); err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(map[string]any{
				"conversationId": id, "principalId": principal, "added": true,
			})
		},
	)

	s.AddTool(
		tool("conversation_remove",
			mcp.WithDescription("Remove someone ELSE from a conversation, naming them by PRINCIPAL ID. Creator-only.\n\nThis is not conversation_leave with a different argument: leave revokes YOUR OWN access and any member can do it, this revokes another person's and only the creator can. Until this existed the only way out of a room was for the person being removed to remove themselves.\n\nRemoval revokes access; it does not delete anything they wrote."),
			mcp.WithString("conversation_id", mcp.Required(), mcp.Description("Conversation ID from conversation_list")),
			mcp.WithString("principal_id", mcp.Required(), mcp.Description("The person's principal ID — from conversation_list members or friends_list. NOT an email address.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "conversation_id")
			if err != nil {
				return toolError(err), nil
			}
			principal, err := requireArg(req, "principal_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := conversation.NewService(client, conversation.WithLogger(r.logger))
			if err := svc.RemoveMember(ctx, id, principal); err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(map[string]any{
				"conversationId": id, "principalId": principal, "removed": true,
			})
		},
	)

	s.AddTool(
		tool("conversation_leave",
			mcp.WithDescription("Leave a conversation — this is also how you REFUSE an invitation, since they are the same act at different moments. Your access is revoked immediately and you stop seeing its messages.\n\nThe creator cannot leave their own conversation: they hold it through ownership rather than a grant, so leaving would revoke nothing. Their equivalent is archiving the room, which closes it for everyone."),
			mcp.WithString("conversation_id", mcp.Required(), mcp.Description("Conversation ID to leave")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			id, err := requireArg(req, "conversation_id")
			if err != nil {
				return toolError(err), nil
			}
			svc := conversation.NewService(client, conversation.WithLogger(r.logger))
			if err := svc.Leave(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(map[string]any{"conversationId": id, "left": true})
		},
	)
}
