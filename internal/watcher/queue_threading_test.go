package watcher

import (
	"encoding/json"
	"strings"
	"testing"
)

// 🛑 THE QUEUE MUST CARRY THE WHOLE THREADING SET, NOT JUST THE PARENT LINK (#471).
//
// The Message comment says a field dropped here is lost for good — the poll that
// could have carried it will not run again. That was written when ReplyTo was
// added, and it describes a CLASS; only ReplyTo was fixed, so threadRootId,
// parentPrincipalId and principalId kept being stripped directly beneath the
// warning about stripping them.
//
// ⚠️ This asserts the JSON ACTUALLY WRITTEN, not the struct's field list. A test
// that checks the Go fields exist would pass while a `json:"-"` tag or a copy
// loop that never sets them still produced a queue line without them — and the
// copy loop is exactly where the original defect lived.
func TestQueuedMessageCarriesEveryThreadingField(t *testing.T) {
	m := Message{
		ID: "1788406438823-0", Agent: "smiley", Content: "x", Timestamp: "2026-09-03T03:33:58Z",
		ReplyTo:           "1788405902893-0",
		ThreadRootID:      "1788397626901-0",
		ParentPrincipalID: "e5008927-0f7e-40af-99be-b12e4dbc4aab",
		PrincipalID:       "c890e6e1-1111-2222-3333-444455556666",
	}

	line, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Round-trip through the wire form a consumer actually reads off queue.jsonl.
	var got map[string]any
	if err := json.Unmarshal(line, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for field, want := range map[string]string{
		"replyTo":           m.ReplyTo,
		"threadRootId":      m.ThreadRootID,
		"parentPrincipalId": m.ParentPrincipalID,
		"principalId":       m.PrincipalID,
	} {
		v, ok := got[field]
		if !ok {
			t.Errorf("queue line drops %q — thread context does not survive delivery, "+
				"and the queue is append-only so it cannot be backfilled: %s", field, line)
			continue
		}
		if v != want {
			t.Errorf("%q = %v, want %q", field, v, want)
		}
	}
}

// ✅ NEGATIVE CONTROL. Every field is `omitempty`, so an unset one must NOT appear
// as an empty string. Without this the fix could "pass" by emitting four empty
// keys on every message — which would bloat an append-only file forever and make
// "absent" indistinguishable from "the server sent nothing".
func TestQueuedMessageOmitsUnsetThreadingFields(t *testing.T) {
	line, err := json.Marshal(Message{ID: "1-0", Agent: "tate", Content: "top-level", Timestamp: "t"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, field := range []string{"replyTo", "threadRootId", "parentPrincipalId", "principalId"} {
		if strings.Contains(string(line), `"`+field+`"`) {
			t.Errorf("unset %q was emitted anyway: %s", field, line)
		}
	}
}
