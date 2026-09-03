package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRoomIdentityMayBeOmitted(t *testing.T) {
	tests := []struct {
		name       string
		handle     string
		wantQuery  string
		wantAgent  bool
		wantHandle bool
	}{
		{name: "authenticated principal", wantQuery: "", wantAgent: false, wantHandle: false},
		{name: "explicit override", handle: "Rowan", wantQuery: "agentHandle=Rowan", wantAgent: true, wantHandle: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotQuery url.Values
			var sendBody, cursorBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.Query()
				if r.Method == http.MethodGet {
					_, _ = w.Write([]byte(`{"lastId":"cursor-1"}`))
					return
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				if r.Method == http.MethodPut {
					cursorBody = body
				} else {
					sendBody = body
				}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost {
					_, _ = w.Write([]byte(`{"id":"message-1","timestamp":"now"}`))
				}
			}))
			defer srv.Close()

			client, err := New(srv.URL, "test-token", WithRetry(0))
			if err != nil {
				t.Fatal(err)
			}
			res, err := client.SendRoomMessage(context.Background(), "series", "series-1", SendRoomMessageRequest{
				Agent: tt.handle, Content: "hello",
			})
			if err != nil {
				t.Fatalf("send: %v", err)
			}
			if res.ID != "message-1" {
				t.Fatalf("id = %q, want message-1", res.ID)
			}
			if res.Backend != srv.URL {
				t.Errorf("backend = %q, want %q", res.Backend, srv.URL)
			}
			if got, ok := sendBody["agent"]; ok != tt.wantAgent || (ok && got != tt.handle) {
				t.Errorf("agent body = %#v, present = %v, want present %v value %q", got, ok, tt.wantAgent, tt.handle)
			}

			_, err = client.GetRoomCursor(context.Background(), "series", "series-1", tt.handle)
			if err != nil {
				t.Fatalf("get cursor: %v", err)
			}
			if got := gotQuery.Get("agentHandle"); (got != "") != tt.wantHandle || got != tt.handle {
				t.Errorf("agentHandle query = %q, want %q", got, tt.handle)
			}

			cursor, err := client.SetRoomCursor(context.Background(), "series", "series-1", tt.handle, "cursor-1")
			if err != nil {
				t.Fatalf("set cursor: %v", err)
			}
			if cursor.Backend != srv.URL || cursor.Status != "cursor set" {
				t.Errorf("cursor response = %#v, want backend %q and status cursor set", cursor, srv.URL)
			}
			if got, ok := cursorBody["agentHandle"]; ok != tt.wantHandle || (ok && got != tt.handle) {
				t.Errorf("cursor agentHandle body = %#v, present = %v, want present %v value %q", got, ok, tt.wantHandle, tt.handle)
			}
		})
	}
}

func TestRoomReadsExposeBackend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/room/series/series-1/messages":
			_, _ = w.Write([]byte(`{"messages":[],"lastId":"message-1"}`))
		case r.URL.Path == "/api/v1/room/series/series-1/cursor":
			_, _ = w.Write([]byte(`{"lastId":"message-1"}`))
		case r.URL.Path == "/api/v1/room/series/series-1/status":
			_, _ = w.Write([]byte(`{"exists":true,"archived":false,"messageCount":1}`))
		default:
			t.Fatalf("unexpected request: %s", r.URL.RequestURI())
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	messages, err := client.ReadRoomMessages(context.Background(), "series", "series-1", ReadRoomMessagesOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := client.GetRoomCursor(context.Background(), "series", "series-1", "")
	if err != nil {
		t.Fatal(err)
	}
	status, err := client.GetRoomStatus(context.Background(), "series", "series-1")
	if err != nil {
		t.Fatal(err)
	}
	if messages.Backend != srv.URL || cursor.Backend != srv.URL || status.Backend != srv.URL {
		t.Fatalf("backends = messages:%q cursor:%q status:%q, want %q", messages.Backend, cursor.Backend, status.Backend, srv.URL)
	}
}

func TestListMyRoomsUsesGeneratedContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/rooms/mine" || r.URL.Query().Get("agentHandle") != "Rowan" {
			t.Fatalf("request = %s, want /api/v1/rooms/mine?agentHandle=Rowan", r.URL.RequestURI())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"entityType":"series","entityId":"series-1","title":"Test Series","archived":false,"canPost":true,"unreadCount":2,"members":[{"id":"owner-1","name":"Owner"},{"id":"silent-1","name":"Silent Collaborator","vanityHandle":"silent"}]}]`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	rooms, err := client.ListMyRooms(context.Background(), "Rowan")
	if err != nil {
		t.Fatal(err)
	}
	if len(*rooms) != 1 || (*rooms)[0].EntityId == nil || *(*rooms)[0].EntityId != "series-1" {
		t.Fatalf("rooms = %#v", rooms)
	}
	if (*rooms)[0].Members == nil || len(*(*rooms)[0].Members) != 2 || *(*(*rooms)[0].Members)[1].Name != "Silent Collaborator" {
		t.Fatalf("members = %#v", (*rooms)[0].Members)
	}
}

// 🛑 A HAND-WRITTEN STRUCT SEES ONLY WHAT IT LISTS (#463).
//
// RoomMessage is not generated, so a field the server sends and this struct omits
// vanishes silently: `pfw room read -o json` returned 0 of 60 messages carrying
// principalId while the API returned 60. There is no error, no warning, and the
// reader concludes the server does not send it — the same shape as #444.
//
// Measured on demo 2026-08-29, one 60-message window:
//
//	                    API   pfw room read -o json
//	replyTo              41   41   ✅ (already present)
//	threadRootId         38    0   ⛔
//	parentPrincipalId    38    0   ⛔
//	principalId          60    0   ⛔
//
// ⚠️ @Gordon's original report also named replyTo; that part was a window artefact
// (an ascending window predating threaded replies). The three below are real.
func TestRoomMessageKeepsTheThreadingAndIdentityFields(t *testing.T) {
	// A realistic reply: threadRootId DIFFERS from replyTo, which is the case that
	// makes grouping-by-parent wrong.
	const payload = `{"messages":[{
		"id":"1788040855729-0","agent":"Wayland","content":"body","timestamp":"2026-08-29T22:00:55Z",
		"replyTo":"1788040779445-0","threadRootId":"1788040466388-0",
		"parentPrincipalId":"c9463d02-e0c3-451a-a0c4-9be957fc941e",
		"principalId":"cd657cdc-6c3e-453d-be0e-92210518be7c"}]}`

	var got RoomMessagesResponse
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Messages) != 1 {
		t.Fatalf("decoded %d messages", len(got.Messages))
	}
	m := got.Messages[0]

	for _, f := range []struct{ name, got, want string }{
		{"ReplyTo", m.ReplyTo, "1788040779445-0"},
		{"ThreadRootID", m.ThreadRootID, "1788040466388-0"},
		{"ParentPrincipalID", m.ParentPrincipalID, "c9463d02-e0c3-451a-a0c4-9be957fc941e"},
		{"PrincipalID", m.PrincipalID, "cd657cdc-6c3e-453d-be0e-92210518be7c"},
	} {
		if f.got != f.want {
			t.Errorf("%s = %q, want %q — the server sent it and the struct dropped it, "+
				"which is invisible to every caller", f.name, f.got, f.want)
		}
	}

	// ⛔ THE ARM THAT STOPS THE FIX BEING COSMETIC. ThreadRootID and ReplyTo must stay
	// DISTINCT fields: a deep thread has a root that is not the immediate parent, and
	// collapsing them fragments one conversation into several.
	if m.ThreadRootID == m.ReplyTo {
		t.Error("ThreadRootID and ReplyTo decoded to the same value; this fixture " +
			"deliberately differs, so they are being read from one key")
	}

	// A non-reply must leave the threading fields empty rather than inventing them —
	// omitempty means absent, and absent must not become a zero-value pointer.
	var plain RoomMessagesResponse
	if err := json.Unmarshal([]byte(`{"messages":[{"id":"1-0","agent":"Tate","content":"x","timestamp":"t","principalId":"p"}]}`), &plain); err != nil {
		t.Fatal(err)
	}
	if p := plain.Messages[0]; p.ReplyTo != "" || p.ThreadRootID != "" || p.ParentPrincipalID != "" {
		t.Errorf("a top-level message decoded with threading fields set: %+v", p)
	} else if p.PrincipalID != "p" {
		t.Errorf("PrincipalID lost on a non-reply: %q", p.PrincipalID)
	}
}

// 🛑 room status ANSWERED THE WRONG QUESTION (#463 class, found by the same method).
//
// The server sends canPost/canModerate and RoomStatusResponse did not list them, so
// `room status` told you a room existed and silently would not tell you whether you
// could write in it. `room list` had surfaced "Can post" as a column all along — the
// data was one command away from the command that asks about ONE room.
//
// ⚠️ EVERY ASSERTION HERE USES true. A bool decodes to false when the key is absent,
// so asserting false would pass against a struct that dropped the field entirely —
// the test would be green for the exact defect it exists to catch.
func TestRoomStatusCarriesTheCallersPermissions(t *testing.T) {
	const payload = `{"exists":true,"archived":false,"messageCount":3703,
		"canPost":true,"canModerate":true}`

	var got RoomStatusResponse
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatal(err)
	}
	if !got.CanPost {
		t.Error("CanPost = false on a payload that says true — the field is not being " +
			"read from the wire, and an agent cannot tell whether it may post")
	}
	if !got.CanModerate {
		t.Error("CanModerate = false on a payload that says true")
	}
	if got.MessageCount != 3703 || !got.Exists {
		t.Errorf("existing fields regressed: %+v", got)
	}

	// ⛔ The two permissions must be INDEPENDENT. Membership grants posting without
	// moderation, so reading one from the other's key would look right in the common
	// case and be wrong for every ordinary member.
	var member RoomStatusResponse
	if err := json.Unmarshal([]byte(`{"exists":true,"canPost":true,"canModerate":false}`), &member); err != nil {
		t.Fatal(err)
	}
	if !member.CanPost || member.CanModerate {
		t.Errorf("an ordinary member decoded as canPost=%t canModerate=%t, want true/false — "+
			"the two fields are being read from one key", member.CanPost, member.CanModerate)
	}
}
