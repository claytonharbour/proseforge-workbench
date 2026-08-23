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

func TestServerVersionToolRegistered(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerServerTools(s, newClientResolver("http://example.test", "token", nil, nil), nil)

	tool := s.GetTool("server_version")
	if tool == nil {
		t.Fatal("missing server_version tool")
	}
	if tool.Tool.Annotations.ReadOnlyHint == nil || !*tool.Tool.Annotations.ReadOnlyHint {
		t.Fatal("server_version should be marked read-only")
	}
}

func TestServerVersionToolReturnsEmbeddedVersion(t *testing.T) {
	const want = "test-build-123"
	previous := Version
	Version = want
	t.Cleanup(func() { Version = previous })

	s := server.NewMCPServer("test", "test")
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/version" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"buildTime":"2026-08-19T00:00:00Z","gitCommit":"abc123","component":"api","allComponents":"/api/v1/admin/version"}`))
	}))
	defer apiServer.Close()
	registerServerTools(s, newClientResolver(apiServer.URL, "token", slog.Default(), nil), map[string]string{"PROSEFORGE_WORKER_VERSION": "worker-7"})
	req := mcp.CallToolRequest{}
	req.Params.Name = "server_version"
	result, err := s.GetTool("server_version").Handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["mcp"].(map[string]any)["version"] != want || payload["server"].(map[string]any)["gitCommit"] != "abc123" || payload["worker"].(map[string]any)["version"] != "worker-7" {
		t.Fatalf("payload = %#v", payload)
	}
}
