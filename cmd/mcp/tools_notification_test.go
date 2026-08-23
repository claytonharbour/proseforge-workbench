package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestNotificationToolsRegistered(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerNotificationTools(s, newClientResolver("http://example.test", "token", nil, nil))
	for _, name := range []string{"notification_list", "notification_unread_count"} {
		tool := s.GetTool(name)
		if tool == nil {
			t.Fatalf("missing tool %s", name)
		}
		if tool.Tool.Annotations.ReadOnlyHint == nil || !*tool.Tool.Annotations.ReadOnlyHint {
			t.Errorf("%s should be marked read-only", name)
		}
	}
}

func TestNotificationListRejectsUnboundedLimit(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerNotificationTools(s, newClientResolver("http://example.test", "token", nil, nil))
	req := mcp.CallToolRequest{}
	req.Params.Name = "notification_list"
	req.Params.Arguments = map[string]any{"limit": float64(101)}
	result, err := s.GetTool("notification_list").Handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var payload toolErrorResponse
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error != "invalid_input" || payload.Message != "limit must be between 1 and 100" {
		t.Fatalf("payload = %#v", payload)
	}
}
