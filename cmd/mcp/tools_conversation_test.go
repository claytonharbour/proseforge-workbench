package main

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

// #821 — the conversation tool surface. Two things are pinned here and neither is cosmetic.
func TestConversationToolSurface(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerConversationTools(s, newClientResolver("http://example.test", "token", nil, nil))

	want := map[string][]string{
		"conversation_list":   {},
		"conversation_create": {},
		"conversation_add":    {"conversation_id", "principal_id"},
		"conversation_leave":  {"conversation_id"},
		// #1173 — removing SOMEONE ELSE. Distinct from leave, which revokes your own.
		"conversation_remove": {"conversation_id", "principal_id"},
	}

	for name, required := range want {
		tool := s.GetTool(name)
		if tool == nil {
			t.Fatalf("missing tool %s — registered in main.go?", name)
		}
		got := map[string]bool{}
		for _, r := range tool.Tool.InputSchema.Required {
			got[r] = true
		}
		for _, r := range required {
			if !got[r] {
				t.Errorf("%s does not require %q — a caller can omit it and get a confusing "+
					"server-side error instead of a clear one", name, r)
			}
		}
	}

	// 🛑 THE ONE THAT MATTERS. conversation_add names a person by PRINCIPAL ID, never by
	// email, because forge/proseforge#911 stopped the API from disclosing friends' addresses
	// — a caller holds an id and has no address to send. That is easy to "helpfully" undo
	// later by adding an `email` argument for symmetry with the story/series grant tools,
	// which would reintroduce a parameter no picker can populate.
	//
	// ⚠️ Asserting the ABSENCE of a field, so this fails the moment someone adds one back.
	// 🛑 remove REQUIRES a principal id; leave does not take one. If remove ever loses that
	// requirement it has quietly become "leave", and the creator-only removal of another
	// member would be gone with no test going red.
	rm := s.GetTool("conversation_remove")
	if rm == nil {
		t.Fatal("conversation_remove missing")
	}
	rmRequired := map[string]bool{}
	for _, r := range rm.Tool.InputSchema.Required {
		rmRequired[r] = true
	}
	if !rmRequired["principal_id"] {
		t.Error("conversation_remove does not require principal_id — it cannot name who to remove")
	}
	if lv := s.GetTool("conversation_leave"); lv != nil {
		for _, r := range lv.Tool.InputSchema.Required {
			if r == "principal_id" {
				t.Error("conversation_leave requires principal_id — leave is about the CALLER, not another person")
			}
		}
	}

	add := s.GetTool("conversation_add")
	if _, present := add.Tool.InputSchema.Properties["email"]; present {
		t.Error("conversation_add grew an `email` argument. #911 means a client does not know " +
			"its friends' addresses; principal_id is the only name it can supply.")
	}
	if _, present := add.Tool.InputSchema.Properties["principal_id"]; !present {
		t.Fatal("conversation_add lost principal_id — there is now no way to name a member")
	}

	// ⚑ Vacuity guard: if the loop above ever iterates nothing, it passes while asserting
	// nothing at all.
	if len(want) != 5 {
		t.Fatalf("checked %d tools, expected 5", len(want))
	}

	// Messaging is deliberately NOT here — room_send/room_read with entity_type
	// "conversation" already do it. A conversation_send would be a second name for one
	// primitive, and would have to re-implement cursors, filtering and archive semantics.
	for _, absent := range []string{"conversation_send", "conversation_read"} {
		if s.GetTool(absent) != nil {
			t.Errorf("%s exists — messaging belongs to room_send/room_read with "+
				"entity_type 'conversation', not to a parallel tool", absent)
		}
	}

	// The description has to say how to talk in one, or a model has the id and no idea
	// what to do with it.
	if d := s.GetTool("conversation_list").Tool.Description; !strings.Contains(d, "room_send") {
		t.Error("conversation_list description does not point at room_send — a caller is left " +
			"holding an id with no route to using it")
	}
}
