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

func TestRoomIdentityArgumentsAreOptional(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerRoomTools(s, newClientResolver("http://example.test", "token", nil, nil))

	for _, name := range []string{"room_list", "room_send", "room_cursor_get", "room_cursor_set"} {
		tool := s.GetTool(name)
		if tool == nil {
			t.Fatalf("missing %s", name)
		}
		for _, required := range tool.Tool.InputSchema.Required {
			if required == "agent" || required == "agent_handle" {
				t.Errorf("%s still requires %s", name, required)
			}
		}
	}
}

func TestRoomReadSurfacesEffectiveBackend(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messages":[],"lastId":"message-1"}`))
	}))
	defer apiServer.Close()

	s := server.NewMCPServer("test", "test")
	registerRoomTools(s, newClientResolver(apiServer.URL, "token", slog.Default(), nil))
	req := mcp.CallToolRequest{}
	req.Params.Name = "room_read"
	req.Params.Arguments = map[string]any{"entity_id": "series-1", "entity_type": "series"}
	result, err := s.GetTool("room_read").Handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["backend"] != apiServer.URL {
		t.Fatalf("backend = %v, want %q", payload["backend"], apiServer.URL)
	}
}

func TestRoomListPreservesSilentMemberRoster(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"entityType":"story","entityId":"story-1","title":"A Story","canPost":true,"members":[{"id":"owner-1","name":"Owner"},{"id":"silent-1","name":"Silent Collaborator","vanityHandle":"silent"}]}]`))
	}))
	defer apiServer.Close()

	s := server.NewMCPServer("test", "test")
	registerRoomTools(s, newClientResolver(apiServer.URL, "token", slog.Default(), nil))
	req := mcp.CallToolRequest{}
	req.Params.Name = "room_list"
	result, err := s.GetTool("room_list").Handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var payload []map[string]any
	if err := json.Unmarshal([]byte(resultText(t, result)), &payload); err != nil {
		t.Fatal(err)
	}
	members, ok := payload[0]["members"].([]any)
	if !ok || len(members) != 2 {
		t.Fatalf("members = %#v", payload[0]["members"])
	}
}

// TestEveryEntityTypeDescriptionNamesEverySupportedType is the guard for a failure that is
// invisible to the compiler, to `go vet`, and to every other test in this file.
//
// 🛑 THE CAPABILITY AND ITS DISCOVERABILITY ARE SEPARATE THINGS. entityTypeArg is a
// free-form string and the server owns the allowlist, so `entity_type: "conversation"`
// worked from the moment #821 registered the type — through code nobody changed here. But
// all eight entity_type descriptions named story, series and bundle and stopped, so no
// model would ever try the fourth. A tool argument a model is never told about is, from the
// model's side, indistinguishable from one that does not exist.
//
// ⚠️ EIGHT COPIES OF ONE SENTENCE IS THE DRIFT SHAPE. Nothing made them agree; they agreed
// because someone pasted carefully. This asserts both halves — every type named, and every
// copy identical — because a fix applied to seven of eight sites is the more likely mistake
// than a fix applied to none.
func TestEveryEntityTypeDescriptionNamesEverySupportedType(t *testing.T) {
	// The types the SERVER accepts (validEntityTypes in the API's room handler). This list
	// lives in another repo, so it cannot be imported — it is restated here deliberately,
	// and this test is the reminder to update it when that one changes.
	supported := []string{"story", "series", "bundle", "conversation"}

	s := server.NewMCPServer("test", "test")
	registerRoomTools(s, newClientResolver("http://example.test", "token", nil, nil))

	seen := map[string]string{} // description -> first tool that used it
	checked := 0

	for _, name := range []string{
		"room_send", "room_message_delete", "room_read", "room_cursor_get",
		"room_cursor_set", "room_status", "room_presence", "room_react",
		"room_archive", "room_unarchive",
	} {
		tool := s.GetTool(name)
		if tool == nil {
			t.Fatalf("missing tool %s — the list in this test has drifted from the registrations", name)
		}
		prop, ok := tool.Tool.InputSchema.Properties["entity_type"].(map[string]any)
		if !ok {
			t.Fatalf("%s has no entity_type property; this test claims to cover it", name)
		}
		desc, _ := prop["description"].(string)
		if desc == "" {
			t.Fatalf("%s entity_type has no description at all", name)
		}
		checked++

		for _, want := range supported {
			if !strings.Contains(desc, want) {
				t.Errorf("%s entity_type description does not mention %q — the server accepts "+
					"it and no model will ever try it:\n  %s", name, want, desc)
			}
		}
		if first, dup := seen[desc]; dup {
			_ = first
		} else {
			seen[desc] = name
		}
	}

	// ⚑ THE VACUITY GUARD. If the tool-name list above ever stops matching reality, the
	// loop silently checks nothing and this test passes while asserting nothing at all.
	if checked != 10 {
		t.Fatalf("checked %d entity_type descriptions, expected 10", checked)
	}
	if len(seen) != 1 {
		t.Errorf("the 10 entity_type descriptions are not identical (%d distinct variants) — "+
			"they drift by copy, so a fix applied to some sites reads as applied to all: %v",
			len(seen), seen)
	}
}
