package room

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// The hint exists because a 403 on a room has two causes and names only one.
// A self-controlling table: every row states what it wants and why, so a row
// that stops holding is a failure rather than a silently-passing case.
func TestHintWrongEntityType(t *testing.T) {
	forbidden := &api.APIError{StatusCode: http.StatusForbidden}
	notFound := &api.APIError{StatusCode: http.StatusNotFound}
	plain := errors.New("dial tcp: connection refused")

	const marker = "--type conversation"

	tests := []struct {
		name       string
		err        error
		entityType string
		wantHint   bool
		why        string
	}{
		{"default type, 403", forbidden, "story", true,
			"the whole point: took the default and got told 'no access'"},
		{"empty type, 403", forbidden, "", true,
			"empty means the caller never chose, same as taking the default"},
		{"explicit series, 403", forbidden, "series", false,
			"they named a type, so they already considered it — nudging is noise"},
		{"explicit conversation, 403", forbidden, "conversation", false,
			"already on the type the hint would suggest; it would be circular"},
		{"default type, 404", notFound, "story", false,
			"404 already says the thing does not exist; no ambiguity to resolve"},
		{"default type, non-API error", plain, "story", false,
			"a transport failure is not a type problem and must not be dressed as one"},
		{"wrapped 403, default type", fmt.Errorf("read room messages: %w", forbidden), "story", true,
			"errors.As must see through service-layer wrapping, which is how it always arrives"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hintWrongEntityType(tt.err, tt.entityType)

			if gotHint := strings.Contains(got.Error(), marker); gotHint != tt.wantHint {
				t.Errorf("hint present = %v, want %v — %s", gotHint, tt.wantHint, tt.why)
			}

			// The original must survive identically either way: callers switch on
			// it. A hint that swallows the cause trades one bad message for another.
			if !errors.Is(got, tt.err) {
				t.Errorf("original error no longer unwrappable; callers matching on it would break")
			}
			if !tt.wantHint && got.Error() != tt.err.Error() {
				t.Errorf("unhinted error was modified:\n got %q\nwant %q", got, tt.err)
			}
		})
	}
}

// Guards the negative arm of the table above. If someone changes the hint text
// and forgets the marker, every wantHint:false row keeps passing while every
// wantHint:true row starts failing for the wrong reason — this says so plainly.
func TestHintMarkerIsPresentInHintText(t *testing.T) {
	got := hintWrongEntityType(&api.APIError{StatusCode: http.StatusForbidden}, "story")
	if !strings.Contains(got.Error(), "--type conversation") {
		t.Fatal("the marker TestHintWrongEntityType greps for is gone from the hint; " +
			"that test's negative rows would now pass vacuously")
	}
}

// === #355: room_send must say WHERE it posted ===================================

// Aldric's third acceptance criterion, and he called it right: "trivially true,
// and it is the thing that would have to break for this bug to return."
func TestSendResultEchoesDestination(t *testing.T) {
	sent := &api.SendRoomMessageResponse{
		ID: "1787376152177-0", Timestamp: "2026-08-22T05:22:32Z", Backend: "https://api.example.test",
	}
	res := SendResult{
		SendRoomMessageResponse: sent,
		Destination: &Destination{
			EntityType: "conversation",
			EntityID:   "df2c3383-91e9-4fd5-9ac4-d8c1a6ea12dc",
			Title:      "Leads Chat",
			Members:    8,
		},
	}

	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["id"] != sent.ID {
		t.Errorf("echoed id = %v, want %v — the confirmation must still identify the message", got["id"], sent.ID)
	}

	dest, ok := got["destination"].(map[string]any)
	if !ok {
		t.Fatal("no destination in the response: this is the whole of #355")
	}
	for field, want := range map[string]any{
		"entity_type": "conversation",
		"entity_id":   "df2c3383-91e9-4fd5-9ac4-d8c1a6ea12dc",
		"room":        "Leads Chat",
		"members":     float64(8),
	} {
		if dest[field] != want {
			t.Errorf("destination.%s = %v, want %v", field, dest[field], want)
		}
	}
}

// ⚠️ The title is the load-bearing field. A sender comparing two UUIDs confirms
// whichever they already believed — "Leads Chat" against an intent of "the fleet
// room" fails on sight. If the title ever silently stops being emitted, the echo
// still LOOKS present while losing the only part that catches a misroute.
func TestSendResultCarriesTheTitleSpecifically(t *testing.T) {
	raw, _ := json.Marshal(SendResult{
		SendRoomMessageResponse: &api.SendRoomMessageResponse{ID: "1-0"},
		Destination:             &Destination{EntityType: "story", EntityID: "abc", Title: "Whimsy of the Woods", Members: 13},
	})
	if !strings.Contains(string(raw), "Whimsy of the Woods") {
		t.Fatal("title absent from the echo; the id-only echo is what #246 already shipped and it caught nothing")
	}
}

func TestDestinationResolved(t *testing.T) {
	if (&Destination{EntityType: "story", EntityID: "abc"}).Resolved() {
		t.Error("a destination with no title must report unresolved, so callers can say 'not checked' rather than print a blank")
	}
	if !(&Destination{Title: "Leads Chat"}).Resolved() {
		t.Error("a destination with a title is resolved")
	}
}

// Degrading to type+id still satisfies the minimum #355 asks for. What must NOT
// happen is a blank title rendering as though it were checked.
func TestUnresolvedDestinationKeepsTypeAndIDButOmitsTitle(t *testing.T) {
	raw, _ := json.Marshal(&Destination{EntityType: "conversation", EntityID: "df2c3383"})
	s := string(raw)
	for _, want := range []string{"conversation", "df2c3383"} {
		if !strings.Contains(s, want) {
			t.Errorf("unresolved destination dropped %q; type+id is the floor, not optional", want)
		}
	}
	for _, unwanted := range []string{`"room"`, `"members"`} {
		if strings.Contains(s, unwanted) {
			t.Errorf("unresolved destination emitted %s; an empty title reads as 'no title', not 'not checked'", unwanted)
		}
	}
}

// 🛑 This runs AFTER a successful send. A lookup failure must never turn a
// delivered message into a command that reports failure — that would be a worse
// bug than the one being fixed, because the caller would retry and double-post.
func TestResolveDestinationNeverFailsWhenLookupDoes(t *testing.T) {
	client, err := api.New("http://127.0.0.1:9", "no-such-token") // nothing listens on port 9
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	got := NewService(client).ResolveDestination(context.Background(), "conversation", "df2c3383")

	if got == nil {
		t.Fatal("returned nil on lookup failure; callers dereference this unconditionally")
	}
	if got.EntityType != "conversation" || got.EntityID != "df2c3383" {
		t.Errorf("lost type/id on degrade: %+v — these come from the caller and cannot fail", got)
	}
	if got.Resolved() {
		t.Error("claimed resolved with no reachable backend")
	}
}

// === #346: a limited read must say what it may have missed ======================

// Three benches drew confident false negatives from `room read` in one night, all
// from documented default behaviour. Each row records the incident shape it
// guards, so a row that stops holding names what comes back.
func TestWindowWarnings(t *testing.T) {
	const (
		oldest    = "OLDEST"
		window    = "the --limit you set"
		servercap = "server's default cap"
		scan      = "scanned at most"
	)
	tests := []struct {
		name             string
		returned, limit  int
		order            string
		filtered, paging bool
		want             []string
		why              string
	}{
		{"no limit, room smaller than the cap", 50, 0, "", false, false, nil,
			"under the server cap unfiltered = provably the whole room; order hides nothing"},
		{"NO LIMIT, cap reached", 1000, 0, "", false, false, []string{oldest, servercap},
			"the worst case and the one my first cut missed: a plain `room read` returns 1000 " +
				"ending FOUR DAYS ago, silently. @Tuner measured it, I reproduced it"},
		{"desc, under limit, unfiltered", 3, 10, "desc", false, false, nil,
			"provably the whole tail: right order, room exhausted before the limit"},
		{"asc but under the limit", 3, 10, "", false, false, nil,
			"asc only hides things when something lies beyond the edge; here nothing does. " +
				"My first cut warned anyway — over-warning is what makes warnings ignorable"},
		{"asc while paging with --since", 60, 60, "asc", false, true, []string{window},
			"paging WANTS oldest-first so no ordering warning, but a full window still means more"},
		{"window exactly full", 3, 3, "desc", false, false, []string{window},
			"@Gordon: limit=3, his reply was message four — one outside the window"},
		{"asc AND full", 3, 3, "", false, false, []string{oldest, window},
			"both hazards apply; reporting only one leaves the other looking checked"},
		{"filtered and short", 2, 10, "desc", true, false, []string{scan},
			"a filtered read draws from the scanned window, so a short result proves nothing"},
		{"filtered cursor poll keeping nothing", 0, 60, "asc", true, true, nil,
			"normal for a watcher: the cursor still advances past the window, nothing is skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WindowWarnings(tt.returned, tt.limit, tt.order, tt.filtered, tt.paging)

			if len(got) != len(tt.want) {
				t.Fatalf("got %d warning(s), want %d — %s\n got: %v", len(got), len(tt.want), tt.why, got)
			}
			for i, marker := range tt.want {
				if !strings.Contains(got[i], marker) {
					t.Errorf("warning %d = %q, want one containing %q — %s", i, got[i], marker, tt.why)
				}
			}
		})
	}
}

// 🛑 The silent rows above are the load-bearing ones: a warning on every read
// gets tuned out, and then it protects nobody. This asserts the quiet cases stay
// quiet, so "added a warning" can't quietly become "warns constantly".
func TestWindowWarningsStayQuietWhenTheResultIsTrustworthy(t *testing.T) {
	for _, q := range []struct {
		name             string
		returned, limit  int
		order            string
		filtered, paging bool
	}{
		{"unlimited read", 500, 0, "", false, false},
		{"newest-first, exhausted", 7, 50, "desc", false, false},
		{"cursor poll", 12, 60, "asc", false, true},
		{"filtered cursor poll", 0, 60, "asc", true, true},
	} {
		if got := WindowWarnings(q.returned, q.limit, q.order, q.filtered, q.paging); got != nil {
			t.Errorf("%s: warned %v on a trustworthy result; noise here is what makes the real warnings ignorable", q.name, got)
		}
	}
}
