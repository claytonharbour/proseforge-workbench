package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/friends"
)

// Friendship is separate from access. Use the friends directory to establish
// trust, then use contributor_grant for immediate story or room access.

func registerFriendsTools(s *server.MCPServer, r *clientResolver) {
	s.AddTool(tool("friends_list",
		mcp.WithDescription("List your accepted friends. Friendship is a trust relationship, not story access; use contributor_grant for immediate story or room access. Returns people only by default; AI reviewers are opt-in."),
		mcp.WithString("include", mcp.Description("Set to \"system\" to include the AI reviewers alongside people. Omit for people only — they are separate populations and mixing them silently is rarely what a caller wants.")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		data, err := friends.NewService(client, friends.WithLogger(r.logger)).
			List(ctx, optionalArg(req, "include"))
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(data, client)
	})

	s.AddTool(tool("friends_search",
		mcp.WithDescription("Find people you could send a friend request to. Matches name and email. Returns each candidate's user id for friends_add."),
		mcp.WithString("search", mcp.Required(), mcp.Description("Full email address, or part of a display name. Required — an empty term used to return the entire user directory.")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		search, err := requireArg(req, "search")
		if err != nil {
			return toolError(err), nil
		}
		data, err := friends.NewService(client, friends.WithLogger(r.logger)).Candidates(ctx, search)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(data, client)
	})

	s.AddTool(tool("friends_add",
		mcp.WithDescription("Send a friend request. This is a request, not a friendship — the other person accepts or declines it, and nothing exists between you until they do."),
		mcp.WithString("user_id", mcp.Required(), mcp.Description("User id of the person to ask, from friends_search")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		userID, err := requireArg(req, "user_id")
		if err != nil {
			return toolError(err), nil
		}
		if err := friends.NewService(client, friends.WithLogger(r.logger)).Request(ctx, userID); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Friend request sent to " + userID + ". Nothing is shared until they accept."), nil
	})

	s.AddTool(tool("friends_requests",
		mcp.WithDescription("Friend requests awaiting your answer, or ones you have sent. The id on each row is the REQUEST id — that is what friends_respond takes, not the person's user id."),
		mcp.WithBoolean("outgoing", mcp.Description("Show requests you sent instead of ones awaiting you")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		svc := friends.NewService(client, friends.WithLogger(r.logger))

		var data any
		var callErr error
		if optionalBoolArg(req, "outgoing") {
			data, callErr = svc.Outgoing(ctx)
		} else {
			data, callErr = svc.Incoming(ctx)
		}
		if callErr != nil {
			return toolError(callErr, client), nil
		}
		return jsonResultWithBackend(data, client)
	})

	s.AddTool(tool("friends_respond",
		mcp.WithDescription("Accept or decline an incoming friend request. Takes the REQUEST id from friends_requests, not the requester's user id — the two are different and the wrong one fails."),
		mcp.WithString("request_id", mcp.Required(), mcp.Description("Request id, from friends_requests")),
		mcp.WithBoolean("accept", mcp.Required(), mcp.Description("true to accept, false to decline")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		requestID, err := requireArg(req, "request_id")
		if err != nil {
			return toolError(err), nil
		}
		accept := optionalBoolArg(req, "accept")
		if err := friends.NewService(client, friends.WithLogger(r.logger)).
			Respond(ctx, requestID, accept); err != nil {
			return toolError(err, client), nil
		}
		verb := "declined"
		if accept {
			verb = "accepted"
		}
		return mcp.NewToolResultText("Friend request " + requestID + " " + verb + "."), nil
	})

	s.AddTool(tool("friends_remove",
		mcp.WithDescription("End a friendship. Does not revoke story access already granted — use contributor_revoke for that; the two are separate and ending a friendship leaves existing grants standing."),
		mcp.WithString("friend_id", mcp.Required(), mcp.Description("User id of the friend to remove, from friends_list")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		friendID, err := requireArg(req, "friend_id")
		if err != nil {
			return toolError(err), nil
		}
		if err := friends.NewService(client, friends.WithLogger(r.logger)).Remove(ctx, friendID); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Friendship with " + friendID + " ended. Any story access already granted is unaffected."), nil
	})
}
