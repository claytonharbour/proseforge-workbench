package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/storyforge"
)

func registerStoryForgeTools(s *server.MCPServer, r *clientResolver) {
	// story_meta_get — Read story planning data
	s.AddTool(
		tool("story_meta_get",
			mcp.WithDescription("Read all story planning data in one call. Returns three markdown documents: story (premise/genre/theme/setting/conflict), characters (## Name headers with profiles), plot (## Section N headers with plot beats). Use with story_meta_upsert to write individual documents."),
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
}
