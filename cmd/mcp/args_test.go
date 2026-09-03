package main

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Registering through tool() must record every declared argument, including the
// auth params it appends, or the guard would reject legitimate calls.
func TestToolRecordsDeclaredArgs(t *testing.T) {
	tool("args_test_tool",
		mcp.WithString("entity_id", mcp.Required()),
		mcp.WithString("since"),
		mcp.WithString("agent_handle"),
	)

	for _, want := range []string{"entity_id", "since", "agent_handle", "url", "token", "credentials_file"} {
		if !declaredArgs["args_test_tool"][want] {
			t.Errorf("declared arg %q was not recorded", want)
		}
	}
}

// The defect this guards: an undeclared argument used to be dropped, and the
// call still succeeded while answering a different question. room_read with
// `since_id` instead of `since` returned the whole room — 488 messages, 1.1MB —
// with no error (forge/proseforge-workbench#279).
func TestUnknownArgsCatchesTheSinceIdMistake(t *testing.T) {
	tool("args_test_room_read",
		mcp.WithString("entity_id", mcp.Required()),
		mcp.WithString("since"),
		mcp.WithString("agent_handle"),
	)

	unknown := unknownArgs("args_test_room_read", map[string]any{
		"entity_id": "abc",
		"since_id":  "123-0",
	})
	if len(unknown) != 1 || unknown[0] != "since_id" {
		t.Fatalf("unknownArgs = %v, want [since_id]", unknown)
	}

	if got := suggestArg("args_test_room_read", "since_id"); got != "since" {
		t.Errorf("suggestArg(since_id) = %q, want \"since\"", got)
	}
	// camelCase for a snake_case argument is the other spelling people reach for.
	if got := suggestArg("args_test_room_read", "agentHandle"); got != "agent_handle" {
		t.Errorf("suggestArg(agentHandle) = %q, want \"agent_handle\"", got)
	}
}

// A correct call must pass cleanly — a guard that rejects valid arguments would
// be worse than the bug it replaces.
func TestUnknownArgsAllowsDeclaredArgs(t *testing.T) {
	tool("args_test_ok", mcp.WithString("entity_id"), mcp.WithString("since"))

	args := map[string]any{
		"entity_id":        "abc",
		"since":            "123-0",
		"credentials_file": "~/creds",
		"url":              "http://example.com",
	}
	if unknown := unknownArgs("args_test_ok", args); len(unknown) != 0 {
		t.Errorf("valid call rejected: %v", unknown)
	}
}

// A tool registered without the helper has no recorded schema; it must not be
// blocked, or an unregistered tool becomes uncallable.
func TestUnknownArgsIgnoresUnregisteredTool(t *testing.T) {
	if unknown := unknownArgs("never_registered", map[string]any{"whatever": 1}); unknown != nil {
		t.Errorf("unregistered tool should not be guarded, got %v", unknown)
	}
}

// Nonsense should be reported, not silently mapped to the nearest argument.
func TestSuggestArgDeclinesUnrelatedNames(t *testing.T) {
	tool("args_test_far", mcp.WithString("entity_id"), mcp.WithString("since"))
	if got := suggestArg("args_test_far", "xyzzy_frobnicate"); got != "" {
		t.Errorf("suggestArg on an unrelated name = %q, want empty", got)
	}
}
