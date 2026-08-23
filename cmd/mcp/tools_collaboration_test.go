package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestCollaborationToolsRegistered(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerCollaborationTools(s, newClientResolver("http://example.test", "token", nil, nil))
	want := []string{
		"contributor_list", "contributor_grant", "contributor_revoke", "contributor_leave",
		"series_contributor_list", "series_contributor_grant", "series_contributor_revoke",
		"shared_stories", "contribution_list", "contribution_get", "story_contribution_accounting", "series_contribution_accounting", "contribution_diff",
		"contribution_ready", "contribution_ready_for_owner", "contribution_discard", "contribution_request_changes", "contribution_merge",
		"contribution_sync", "contribution_suggest",
	}
	for _, name := range want {
		tool := s.GetTool(name)
		if tool == nil {
			t.Fatalf("missing tool %s", name)
		}
		if name != "shared_stories" && len(tool.Tool.InputSchema.Required) == 0 {
			t.Errorf("%s should require story/contribution arguments", name)
		}
	}
	if tool := s.GetTool("contribution_discard"); tool == nil || tool.Tool.Annotations.DestructiveHint == nil || !*tool.Tool.Annotations.DestructiveHint {
		t.Error("contribution_discard should be marked destructive")
	}
	if tool := s.GetTool("contributor_revoke"); tool == nil || tool.Tool.Annotations.DestructiveHint == nil || !*tool.Tool.Annotations.DestructiveHint {
		t.Error("contributor_revoke should be marked destructive")
	}
	if tool := s.GetTool("contributor_leave"); tool == nil || tool.Tool.Annotations.DestructiveHint == nil || !*tool.Tool.Annotations.DestructiveHint {
		t.Error("contributor_leave should be marked destructive")
	}
}

func TestCollaborationMutationErrorsPreserveBackendContract(t *testing.T) {
	tests := []struct {
		name      string
		tool      string
		args      map[string]any
		status    int
		body      string
		wantCode  string
		wantMsg   string
		retryable bool
	}{
		{
			name: "mark ready conflict", tool: "contribution_ready",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "contrib-1"},
			status: http.StatusConflict, body: `{"error":"invalid_transition","message":"Not allowed from the current status"}`,
			wantCode: "conflict", wantMsg: "Not allowed from the current status",
		},
		{
			name: "ready for owner conflict", tool: "contribution_ready_for_owner",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "contrib-1"},
			status: http.StatusConflict, body: `{"error":"invalid_transition","message":"Contribution already merged"}`,
			wantCode: "conflict", wantMsg: "Contribution already merged",
		},
		{
			name: "discard forbidden", tool: "contribution_discard",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "contrib-1"},
			status: http.StatusForbidden, body: `{"error":"forbidden","message":"not your contribution"}`,
			wantCode: "forbidden", wantMsg: "Access denied. You don't have permission for this operation.",
		},
		{
			name: "request changes conflict", tool: "contribution_request_changes",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "contrib-1"},
			status: http.StatusConflict, body: `{"error":"invalid_transition","message":"Contribution already merged"}`,
			wantCode: "conflict", wantMsg: "Contribution already merged",
		},
		{
			name: "merge missing", tool: "contribution_merge",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "missing"},
			status: http.StatusNotFound, body: `{"error":"not_found","message":"Contribution not found"}`,
			wantCode: "not_found", wantMsg: "Contribution not found",
		},
		{
			name: "sync conflict", tool: "contribution_sync",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "contrib-1"},
			status: http.StatusConflict, body: `{"error":"merge_conflict","message":"Conflict requires resolutions"}`,
			wantCode: "conflict", wantMsg: "Conflict requires resolutions",
		},
		{
			name: "suggest validation", tool: "contribution_suggest",
			args:   map[string]any{"story_id": "story-1", "contribution_id": "contrib-1", "section_id": "section-1", "content": "suggested"},
			status: http.StatusUnprocessableEntity, body: `{"error":"validation","message":"Suggested content is required"}`,
			wantCode: "unprocessable", wantMsg: "Suggested content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer apiServer.Close()

			mcpServer := server.NewMCPServer("test", "test")
			registerCollaborationTools(mcpServer, newClientResolver(apiServer.URL, "token", slog.Default(), nil))
			registered := mcpServer.GetTool(tt.tool)
			if registered == nil {
				t.Fatalf("missing tool %s", tt.tool)
			}
			req := mcp.CallToolRequest{}
			req.Params.Name = tt.tool
			req.Params.Arguments = tt.args
			result, err := registered.Handler(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsError {
				t.Fatal("expected MCP error result")
			}
			var payload toolErrorResponse
			if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error != tt.wantCode || payload.StatusCode != tt.status || payload.Retryable != tt.retryable {
				t.Errorf("payload = %#v", payload)
			}
			wantMessage := tt.wantMsg + " (" + apiServer.URL + ")"
			if payload.Message != wantMessage {
				t.Errorf("message = %q, want %q", payload.Message, wantMessage)
			}
		})
	}
}
