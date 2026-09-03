package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdminTools(s *server.MCPServer, r *clientResolver) {
	s.AddTool(tool("demo_banner_get",
		mcp.WithDescription("Read the runtime demo banner admin state, including enabled/show status, environment, expiry, and update timestamp. Requires an admin credential."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		data, err := client.GetDemoBanner(ctx)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(data, client)
	})

	s.AddTool(tool("demo_banner_update",
		mcp.WithDescription("Update the runtime demo banner. Set enabled=true to restore the banner and clear expiry. Set enabled=false only with expires_at in RFC3339 within the next 24 hours; the server response includes effective state and audit fields. Requires an admin credential."),
		mcp.WithBoolean("enabled", mcp.Required(), mcp.Description("true to show the banner; false to suppress it temporarily")),
		mcp.WithString("expires_at", mcp.Description("Required when enabled=false. RFC3339 timestamp strictly in the future and no more than 24 hours away.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		enabled := optionalBoolArg(req, "enabled")
		var expiresAt *time.Time
		if raw := optionalArg(req, "expires_at"); raw != "" {
			parsed, parseErr := time.Parse(time.RFC3339, raw)
			if parseErr != nil {
				return toolError(fmt.Errorf("expires_at must be RFC3339: %w", parseErr)), nil
			}
			expiresAt = &parsed
		}
		data, err := client.UpdateDemoBanner(ctx, enabled, expiresAt)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(data, client)
	})
}
