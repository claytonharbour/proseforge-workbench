package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/collaboration"
)

func registerCollaborationTools(s *server.MCPServer, r *clientResolver) {
	s.AddTool(tool("contributor_list",
		mcp.WithDescription("Owner/admin action: list a story's contributor grants, including grantee email, principal ID, capability, and grant timestamp."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Grants(ctx, storyID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("contributor_grant",
		mcp.WithDescription("Owner/admin action: immediately grant one capability to an existing user by email on a story. Supported capabilities, in ladder order, are story:view, story:feedback, story:edit, story:review, story:merge, plus room:enter and room:post. Repeating a grant is idempotent."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Existing grantee email")),
		mcp.WithString("capability", mcp.Required(), mcp.Description("Capability to grant")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		email, err := requireArg(req, "email")
		if err != nil {
			return toolError(err), nil
		}
		capability, err := requireArg(req, "capability")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Grant(ctx, storyID, email, capability)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("contributor_revoke",
		mcp.WithDescription("Owner/admin action: revoke one capability from an existing user by email. Revocation is idempotent and remains available after subscription lapse."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Existing grantee email")),
		mcp.WithString("capability", mcp.Required(), mcp.Description("Capability to revoke")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		email, err := requireArg(req, "email")
		if err != nil {
			return toolError(err), nil
		}
		capability, err := requireArg(req, "capability")
		if err != nil {
			return toolError(err), nil
		}
		if err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Revoke(ctx, storyID, email, capability); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Story capability revoked."), nil
	})

	s.AddTool(tool("contributor_leave",
		mcp.WithDescription("Contributor action: leave a story by removing every story grant held by the authenticated account. The owner cannot leave their own story because ownership is not a grant."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		if err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Leave(ctx, storyID); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Story access left."), nil
	})

	s.AddTool(tool("series_contributor_list",
		mcp.WithDescription("Owner/admin action: list a series' contributor grants, including room capabilities. Grants are immediately effective."),
		mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		seriesID, err := requireArg(req, "series_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).SeriesGrants(ctx, seriesID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("series_contributor_grant",
		mcp.WithDescription("Owner/admin action: immediately grant one capability on a series to an existing user by email. This applies to the series room; use story grants for individual story access."),
		mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Existing grantee email")),
		mcp.WithString("capability", mcp.Required(), mcp.Description("Capability to grant, such as room:enter or room:post")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		seriesID, err := requireArg(req, "series_id")
		if err != nil {
			return toolError(err), nil
		}
		email, err := requireArg(req, "email")
		if err != nil {
			return toolError(err), nil
		}
		capability, err := requireArg(req, "capability")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).SeriesGrant(ctx, seriesID, email, capability)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("series_contributor_revoke",
		mcp.WithDescription("Owner/admin action: revoke one capability from a series user by email."),
		mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Existing grantee email")),
		mcp.WithString("capability", mcp.Required(), mcp.Description("Capability to revoke")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		seriesID, err := requireArg(req, "series_id")
		if err != nil {
			return toolError(err), nil
		}
		email, err := requireArg(req, "email")
		if err != nil {
			return toolError(err), nil
		}
		capability, err := requireArg(req, "capability")
		if err != nil {
			return toolError(err), nil
		}
		if err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).SeriesRevoke(ctx, seriesID, email, capability); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Series capability revoked."), nil
	})

	s.AddTool(tool("shared_stories",
		mcp.WithDescription("List stories shared with the authenticated account. Each result includes the owner, grant status, capabilities, and story identity."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).SharedStories(ctx)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("contribution_list",
		mcp.WithDescription("List contributions for a story. Owner-facing overview of contributor branches and lifecycle status."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Contributions(ctx, storyID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("contribution_get",
		// ⚠️ This description is the ONLY place an MCP consumer learns the legal
		// status values — the served spec carries no enum on the status property.
		// It listed four of six, and the two it omitted are the ones an agent
		// most needs: `discarded` is TERMINAL, and an unrecognised status reads
		// as "some in-progress thing", so the natural response is to resume work
		// the contributor deliberately threw away (#348).
		//
		// 🛑 `running` and `failed` are deliberately NOT here. They are upstream
		// #816 and do not exist yet; listing them early is the same error facing
		// the other way.
		mcp.WithDescription("Get the authenticated account's active contribution for a story. Use this before editing to learn what state the work is in and which branch/base it uses. "+
			"Statuses: draft (being written) · ready (submitted for owner review) · changes_requested (owner sent it back — the contributor revises) · "+
			"changes_suggested (owner proposed specific edits — resolve them) · merged (TERMINAL, accepted into trunk) · discarded (TERMINAL, thrown away). "+
			"changes_requested and changes_suggested are DIFFERENT states, not aliases. Do not resume or re-submit a merged or discarded contribution."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Mine(ctx, storyID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("story_contribution_accounting",
		mcp.WithDescription("Owner/reviewer action: return contributor totals and dated submitted/merged snapshots for a story. Requires story review access; use this for accounting, not branch content."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).StoryAccounting(ctx, storyID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("series_contribution_accounting",
		mcp.WithDescription("Owner/reviewer action: return contributor totals and dated submitted/merged snapshots across every story in a series. Requires series access; use this for accounting, not branch content."),
		mcp.WithString("series_id", mcp.Required(), mcp.Description("Series ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		seriesID, err := requireArg(req, "series_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).SeriesAccounting(ctx, seriesID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("contribution_diff",
		mcp.WithDescription("Owner action: show a contribution's branch-versus-trunk diff, including per-section content and conflict indicators."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Diff(ctx, storyID, contributionID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResult(result)
	})

	s.AddTool(tool("contribution_ready",
		mcp.WithDescription("Mark the authenticated contributor's work ready for owner review. This changes lifecycle state; it does not merge the branch."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return contributionAction(ctx, req, r, func(s *collaboration.Service, storyID, contributionID string) (any, error) {
			return s.Ready(ctx, storyID, contributionID)
		})
	})

	s.AddTool(tool("contribution_ready_for_owner",
		mcp.WithDescription("Delegated reviewer action: notify the story owner that review of a ready contribution is complete. This sends a handoff notification without changing contribution state or merging."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		if err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).ReadyForOwner(ctx, storyID, contributionID); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Story owner notified that the contribution is ready."), nil
	})

	s.AddTool(tool("contribution_discard",
		mcp.WithDescription("Discard the authenticated contributor's unmerged contribution. Permanently removes its branch and buffered edits without changing the owner's story."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{DestructiveHint: mcp.ToBoolPtr(true)}),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		if err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Discard(ctx, storyID, contributionID); err != nil {
			return toolError(err, client), nil
		}
		return mcp.NewToolResultText("Contribution discarded."), nil
	})

	s.AddTool(tool("contribution_request_changes",
		mcp.WithDescription("Owner action: return a contribution to draft status so the contributor can revise it."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return contributionAction(ctx, req, r, func(s *collaboration.Service, storyID, contributionID string) (any, error) {
			return s.RequestChanges(ctx, storyID, contributionID)
		})
	})

	s.AddTool(tool("contribution_merge",
		mcp.WithDescription("Owner action: merge a contribution into the owner's story. The backend enforces ownership and conflict rules. "+
			"Optionally accept only SOME of the changed files with `selections` — a JSON object of file path -> true/false. "+
			"FAIL-CLOSED: if you supply selections it must name EVERY changed path; an absent path is a 400 invalid_selections, "+
			"NOT an implicit accept. Call contribution_diff first to get the full changed-path list. "+
			"Keys are file paths (content/<sectionId>.md), not bare section ids. Omit selections to merge everything."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
		mcp.WithString("selections", mcp.Description(
			`Optional JSON object mapping EVERY changed file path to true (accept) or false (reject), e.g. {"content/abc.md": true, "content/def.md": false}. Omit to accept all.`)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		// nil (omitted) means "take everything". An explicitly empty object is
		// NOT the same and is passed through so the fail-closed backend decides,
		// rather than this layer inventing an "accept nothing" semantic.
		var selections map[string]bool
		if raw := optionalArg(req, "selections"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &selections); err != nil {
				return toolError(invalidInputErrorf("parse selections JSON: %v", err)), nil
			}
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).
			Merge(ctx, storyID, contributionID, selections)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(result, client)
	})

	s.AddTool(tool("contribution_sync",
		mcp.WithDescription("Sync trunk into a contribution branch using a three-way merge. Allowed for the contributor or story owner; provide a JSON object of resolutions when the backend reports conflicts."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
		mcp.WithString("resolutions", mcp.Description("Optional JSON object mapping conflicted paths to resolved content")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		var resolutions map[string]string
		if raw := optionalArg(req, "resolutions"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &resolutions); err != nil {
				return toolError(invalidInputErrorf("parse resolutions JSON: %v", err)), nil
			}
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Sync(ctx, storyID, contributionID, resolutions)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(result, client)
	})

	s.AddTool(tool("contribution_suggest",
		mcp.WithDescription("Owner action: attach a suggested section revision to a contribution without editing the contributor's branch directly."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
		mcp.WithString("section_id", mcp.Required(), mcp.Description("Section ID")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Suggested section content")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
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
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).SuggestSection(ctx, storyID, contributionID, sectionID, content)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(result, client)
	})

	s.AddTool(tool("contribution_suggestions",
		mcp.WithDescription("List the suggestions raised against a contribution, with their accepted/rejected/pending status. This is the read half of contribution_suggest — use it to see what is outstanding before merging. Visible to the story's reviewers and to the contribution's own author."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).Suggestions(ctx, storyID, contributionID)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(result, client)
	})

	s.AddTool(tool("contribution_suggestion_resolve",
		mcp.WithDescription("Accept or reject a single suggestion on a contribution. Resolve each suggestion before contribution_merge: only accepted text reaches the owner's trunk, and a rejected one stays rejected through the merge."),
		mcp.WithString("story_id", mcp.Required(), mcp.Description("Story ID")),
		mcp.WithString("contribution_id", mcp.Required(), mcp.Description("Contribution ID")),
		mcp.WithString("suggestion_id", mcp.Required(), mcp.Description("Suggestion ID, from contribution_suggestions")),
		mcp.WithString("status", mcp.Required(), mcp.Description("For a replacement (one carrying original/suggested text): accepted or rejected. For an advisory item: acknowledged or dismissed. pending undoes either. The two halves are not interchangeable — check the suggestion's shape in contribution_suggestions first.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := r.resolve(req)
		if err != nil {
			return toolError(err), nil
		}
		storyID, err := requireArg(req, "story_id")
		if err != nil {
			return toolError(err), nil
		}
		contributionID, err := requireArg(req, "contribution_id")
		if err != nil {
			return toolError(err), nil
		}
		suggestionID, err := requireArg(req, "suggestion_id")
		if err != nil {
			return toolError(err), nil
		}
		status, err := requireArg(req, "status")
		if err != nil {
			return toolError(err), nil
		}
		result, err := collaboration.NewService(client, collaboration.WithLogger(r.logger)).
			ResolveSuggestion(ctx, storyID, contributionID, suggestionID, status)
		if err != nil {
			return toolError(err, client), nil
		}
		return jsonResultWithBackend(result, client)
	})
}

func contributionAction(ctx context.Context, req mcp.CallToolRequest, r *clientResolver, action func(*collaboration.Service, string, string) (any, error)) (*mcp.CallToolResult, error) {
	client, err := r.resolve(req)
	if err != nil {
		return toolError(err), nil
	}
	storyID, err := requireArg(req, "story_id")
	if err != nil {
		return toolError(err), nil
	}
	contributionID, err := requireArg(req, "contribution_id")
	if err != nil {
		return toolError(err), nil
	}
	result, err := action(collaboration.NewService(client, collaboration.WithLogger(r.logger)), storyID, contributionID)
	if err != nil {
		return toolError(err, client), nil
	}
	return jsonResultWithBackend(result, client)
}
