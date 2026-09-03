package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/notification"
)

func registerNotificationTools(s *server.MCPServer, r *clientResolver) {
	s.AddTool(tool("notification_list",
		mcp.WithDescription("List a bounded page of notifications for the authenticated account. Use this to verify collaboration handoffs, actor attribution, story links, and unread state."),
		mcp.WithNumber("limit", mcp.Description("Maximum notifications to return (default 25, range 1-100)")),
		mcp.WithNumber("offset", mcp.Description("Pagination offset (default 0)")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		limit := optionalIntArg(req, "limit", 25)
		if limit < 1 || limit > 100 {
			return toolError(invalidInputErrorf("limit must be between 1 and 100")), nil
		}
		offset := optionalIntArg(req, "offset", 0)
		if offset < 0 {
			return toolError(invalidInputErrorf("offset must be zero or greater")), nil
		}
		result, err := notification.NewService(client, notification.WithLogger(r.logger)).List(ctx, limit, offset)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("notification_unread_count",
		mcp.WithDescription("Get the unread notification count for the authenticated account without fetching notification bodies."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		result, err := notification.NewService(client, notification.WithLogger(r.logger)).UnreadCount(ctx)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})
}
