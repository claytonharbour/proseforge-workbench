package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/feedback"
)

func registerFeedbackTools(s *server.MCPServer, r *clientResolver) {
	// feedback_list
	s.AddTool(
		tool("feedback_list",
			mcp.WithDescription("List feedback reviews for a story. Returns review IDs for use with feedback_get and feedback_item_add."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.List(ctx, id)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// feedback_get
	s.AddTool(
		tool("feedback_get",
			mcp.WithDescription("Get feedback review details with all feedback items inline. Requires story_id and review_id from feedback_list."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewID, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.GetFull(ctx, storyID, reviewID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// feedback_suggestions
	s.AddTool(
		tool("feedback_suggestions",
			mcp.WithDescription("List suggestions for a feedback review. Requires story_id and review_id from feedback_list."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewID, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.GetSuggestions(ctx, storyID, reviewID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// feedback_diff
	s.AddTool(
		tool("feedback_diff",
			mcp.WithDescription("Get diff of suggested changes for a feedback review. Requires story_id and review_id."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewID, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			result, err := svc.GetDiff(ctx, storyID, reviewID)
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(result)
		},
	)

	// feedback_create
	s.AddTool(
		tool("feedback_create",
			mcp.WithDescription("Create a new AI feedback review for a story (owner only)"),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			id, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			review, err := svc.Create(ctx, id, api.StartAIReviewRequest{})
			if err != nil {
				return toolError(err, client), nil
			}
			return jsonResult(review)
		},
	)

	// feedback_item_add
	s.AddTool(
		tool("feedback_item_add",
			mcp.WithDescription("Add a feedback item to a review. Requires story_id and review_id. Use to submit replacement suggestions, strengths, opportunities, or context notes."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Item type: replacement, strength, opportunity, suggestion, context")),
			mcp.WithString("text", mcp.Required(), mcp.Description("For replacement: original text. For others: feedback text")),
			mcp.WithString("section_id", mcp.Description("Section ID to attach feedback to (required for replacement; auto-assigned to first section for other types if omitted)")),
			mcp.WithString("suggested", mcp.Description("Replacement text (for replacement type)")),
			mcp.WithString("rationale", mcp.Description("Why this improves the writing")),
			mcp.WithString("context_type", mcp.Description("For type=context: characters, plot, tone, threads (default: general)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewID, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			itemType, err := requireArg(req, "type")
			if err != nil {
				return toolError(err), nil
			}
			text, err := requireArg(req, "text")
			if err != nil {
				return toolError(err), nil
			}

			item := api.AddFeedbackItemRequest{
				Type: &itemType,
				Text: &text,
			}
			if s := optionalArg(req, "section_id"); s != "" {
				item.SectionId = &s
			}
			if s := optionalArg(req, "suggested"); s != "" {
				item.Suggested = &s
			}
			if s := optionalArg(req, "rationale"); s != "" {
				item.Rationale = &s
			}
			if itemType == "context" {
				ct := optionalArg(req, "context_type")
				if ct == "" {
					return mcp.NewToolResultError("context_type is required when type=context (valid: characters, plot, tone, threads)"), nil
				}
				item.ContextType = &ct
			}

			if err := svc.AddItem(ctx, storyID, reviewID, item); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Feedback item added."), nil
		},
	)

	// feedback_section_update
	s.AddTool(
		tool("feedback_section_update",
			mcp.WithDescription("Rewrite a section's content in the feedback branch. Requires story_id, review_id, and section_id. Use for structural fixes: cutting bloat, fixing POV, eliminating duplicates."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
			mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID to rewrite")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Full rewritten section content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewID, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			sectionID, err := requireArg(req, "section_id")
			if err != nil {
				return toolError(err), nil
			}
			content, err := requireArg(req, "content")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.UpdateSection(ctx, storyID, reviewID, sectionID, content); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Section %s updated (%d characters).", sectionID, len(content))), nil
		},
	)

	// feedback_submit
	s.AddTool(
		tool("feedback_submit",
			mcp.WithDescription("Submit a review, marking it as ready for the author. Call after adding all feedback items."),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			id, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}
			if err := svc.Submit(ctx, id); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Review %s submitted.", id)), nil
		},
	)

	// feedback_incorporate — uses feedback.Service.IncorporateAll instead of inline diff parsing
	s.AddTool(
		tool("feedback_incorporate",
			mcp.WithDescription("Incorporate feedback changes into the story. Author operation — applies accepted feedback to the story."),
			mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
			mcp.WithString("review_id", mcp.Required(), mcp.Description("Review ID")),
			mcp.WithBoolean("accept_all", mcp.Description("Accept all changes (default: true)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			svc := feedback.NewService(client, feedback.WithLogger(r.logger))
			storyID, err := requireArg(req, "story_id")
			if err != nil {
				return toolError(err), nil
			}
			reviewID, err := requireArg(req, "review_id")
			if err != nil {
				return toolError(err), nil
			}

			if err := svc.IncorporateAll(ctx, storyID, reviewID); err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText("Feedback incorporated."), nil
		},
	)
}
