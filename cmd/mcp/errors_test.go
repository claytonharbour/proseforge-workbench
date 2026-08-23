package main

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) != 1 {
		t.Fatalf("content length = %d, want 1", len(result.Content))
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want mcp.TextContent", result.Content[0])
	}
	return content.Text
}

func TestToolErrorClassifiesInvalidInput(t *testing.T) {
	result := toolErrorWithURL(invalidInputErrorf("parse resolutions JSON: unexpected end of JSON input"), "http://example.test")
	if !result.IsError {
		t.Fatal("expected MCP error result")
	}
	var payload toolErrorResponse
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error != "invalid_input" {
		t.Errorf("error = %q, want invalid_input", payload.Error)
	}
	if payload.Message != "parse resolutions JSON: unexpected end of JSON input" {
		t.Errorf("message = %q", payload.Message)
	}
	if payload.Retryable {
		t.Error("invalid input must not be retryable")
	}
	if payload.StatusCode != 0 {
		t.Errorf("status code = %d, want 0", payload.StatusCode)
	}
}

func TestRequireArgReturnsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		message string
	}{
		{name: "missing", args: map[string]any{}, message: "missing required argument: story_id"},
		{name: "wrong type", args: map[string]any{"story_id": 42}, message: "argument story_id must be a string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.args
			_, err := requireArg(req, "story_id")
			if err == nil {
				t.Fatal("expected argument error")
			}
			result := toolError(err)
			var payload toolErrorResponse
			if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error != "invalid_input" || payload.Message != tt.message {
				t.Fatalf("payload = %#v", payload)
			}
		})
	}
}
