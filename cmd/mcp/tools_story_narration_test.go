package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestNarrationStatusAndDriftToolsRegisteredReadOnly(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerStoryTools(s, newClientResolver("http://example.test", "token", nil, nil))

	for _, name := range []string{"narration_status", "narration_drift"} {
		registered := s.GetTool(name)
		if registered == nil {
			t.Fatalf("missing tool %s", name)
		}
		if registered.Tool.Annotations.ReadOnlyHint == nil || !*registered.Tool.Annotations.ReadOnlyHint {
			t.Errorf("%s should be marked read-only", name)
		}
	}
}

func TestNarrationDescriptionsExplainPendingSections(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerStoryTools(s, newClientResolver("http://example.test", "token", nil, nil))

	for _, name := range []string{"narration_start", "narration_status", "narration_rebuild"} {
		registered := s.GetTool(name)
		if registered == nil {
			t.Fatalf("missing tool %s", name)
		}
		if !strings.Contains(registered.Tool.Description, "pending_sections") {
			t.Errorf("%s description should explain pending_sections", name)
		}
	}

	for _, name := range []string{"narration_status", "narration_rebuild"} {
		if !strings.Contains(s.GetTool(name).Tool.Description, "sections_not_narrated") {
			t.Errorf("%s description should explain sections_not_narrated", name)
		}
	}
}

func TestNarrationDriftReturnsOnlyDriftEnvelope(t *testing.T) {
	drift := map[string]any{
		"status":                 "stale",
		"contentChangedSections": []any{"section-1"},
		"nameChangedSections":    []any{},
		"reordered":              true,
		"insertedSections":       []any{"section-2"},
		"deletedSections":        []any{},
	}
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/story/story-1/narration" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "complete",
			"chapters": []any{map[string]any{"id": "section-1"}},
			"drift":    drift,
		})
	}))
	defer apiServer.Close()

	result := callNarrationTool(t, apiServer.URL, "narration_drift")
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload) != 1 {
		t.Fatalf("payload should contain only drift: %s", resultText(t, result))
	}
	var got map[string]any
	if err := json.Unmarshal(payload["drift"], &got); err != nil {
		t.Fatal(err)
	}
	if got["status"] != "stale" || got["reordered"] != true {
		t.Fatalf("drift = %#v", got)
	}
}

func TestNarrationDriftHandlesMissingEnvelope(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"processing"}`))
	}))
	defer apiServer.Close()

	result := callNarrationTool(t, apiServer.URL, "narration_drift")
	if got := resultText(t, result); got != `{"drift":null}` {
		t.Fatalf("result = %s", got)
	}
}

func TestNarrationDriftPreservesBackendErrors(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not_found","message":"Narration not found"}`))
	}))
	defer apiServer.Close()

	result := callNarrationTool(t, apiServer.URL, "narration_drift")
	if !result.IsError {
		t.Fatal("expected MCP error result")
	}
	var payload toolErrorResponse
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error != "not_found" || payload.StatusCode != http.StatusNotFound || payload.Retryable {
		t.Fatalf("payload = %#v", payload)
	}
}

func callNarrationTool(t *testing.T, apiURL, name string) *mcp.CallToolResult {
	t.Helper()
	s := server.NewMCPServer("test", "test")
	registerStoryTools(s, newClientResolver(apiURL, "token", slog.Default(), nil))
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = map[string]any{"story_id": "story-1"}
	result, err := s.GetTool(name).Handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
