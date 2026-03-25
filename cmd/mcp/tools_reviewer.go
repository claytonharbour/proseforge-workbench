package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/review"
	"github.com/claytonharbour/proseforge-workbench/internal/reviewer"
)

func registerReviewerTools(s *server.MCPServer, r *clientResolver) {
	// reviewer_add (author adds reviewer to story)
	s.AddTool(
		tool("reviewer_add",
			mcp.WithDescription("Add a reviewer to a story (author operation)"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("reviewer_id", mcp.Description("Reviewer's user ID")),
			mcp.WithString("email", mcp.Description("Reviewer's email (alternative to reviewer_id)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := review.NewService(client, review.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewerID := optionalArg(req, "reviewer_id")
			email := optionalArg(req, "email")
			if reviewerID == "" && email == "" {
				return mcp.NewToolResultError("provide reviewer_id or email"), nil
			}
			reqBody := api.AddReviewerRequest{}
			if reviewerID != "" {
				reqBody.ReviewerId = &reviewerID
			}
			if email != "" {
				reqBody.Email = &email
			}
			result, err := svc.AddReviewer(ctx, storyID, reqBody)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// reviewer_available (list users who opted in as reviewers)
	s.AddTool(
		tool("reviewer_available",
			mcp.WithDescription("List users available as reviewers"),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := reviewer.NewService(client, reviewer.WithLogger(r.logger))
			reviewers, err := svc.ListAvailable(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(reviewers)
		},
	)
}
