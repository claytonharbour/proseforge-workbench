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
		case r.URL.Path == "/api/v1/rooms/series/series-1/messages":
			_, _ = w.Write([]byte(`{"messages":[],"lastId":"message-1"}`))
		case r.URL.Path == "/api/v1/rooms/series/series-1/cursor":
			_, _ = w.Write([]byte(`{"lastId":"message-1"}`))
		case r.URL.Path == "/api/v1/rooms/series/series-1/status":
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
