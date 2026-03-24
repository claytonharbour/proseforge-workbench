package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/claytonharbour/proseforge-workbench/internal/review"
)

func registerReviewTools(s *server.MCPServer, r *clientResolver) {
	// review_list
	s.AddTool(
		tool("review_list",
			mcp.WithDescription("List pending reviews assigned to the authenticated user"),
			mcp.WithNumber("limit", mcp.Description("Max results (1-100, default 25)")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := review.NewService(client, review.WithLogger(r.logger))
			limit := optionalIntArg(req, "limit", 25)
			params := &gen.GetReviewsPendingParams{Limit: &limit}
			result, err := svc.ListPending(ctx, params)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// review_accept
	s.AddTool(
		tool("review_accept",
			mcp.WithDescription("Accept a review assignment"),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := review.NewService(client, review.WithLogger(r.logger))
			id, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.Accept(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Review %s accepted.", id)), nil
		},
	)

	// review_decline
	s.AddTool(
		tool("review_decline",
			mcp.WithDescription("Decline a review assignment"),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := review.NewService(client, review.WithLogger(r.logger))
			id, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.Decline(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Review %s declined.", id)), nil
		},
	)
}
