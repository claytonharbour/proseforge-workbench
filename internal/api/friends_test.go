package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListFriendsOmitsIncludeByDefault pins the default: the workbench must not
// ask for system accounts unless the caller opts in.
//
// The `include` parameter is work in progress upstream, so the shape at risk is
// the server default changing to include AI reviewers. If that happens, a
// caller listing "my friends" silently starts getting six AI reviewers mixed in
// with people — a bigger, plausible-looking answer to a question nobody asked,
// which is the failure mode that produced #279 and #280.
//
// Sending the parameter explicitly is what makes us immune to that flip, so
// this asserts on the wire, not on the response.
func TestListFriendsOmitsIncludeByDefault(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		// #1159 REMOVE (97a91d94): the server emits `friends`, never `reviewers`.
		// Nothing here parses the body — this test asserts on r.URL.RawQuery — but the
		// payload names the server it models so a future sweep does not have to guess.
		_, _ = w.Write([]byte(`{"friends":[],"total":0}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.ListFriends(context.Background(), ""); err != nil {
		t.Fatalf("list friends: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("default listing sent query %q, want none — system accounts must be opt-in", gotQuery)
	}

	if _, err := client.ListFriends(context.Background(), "system"); err != nil {
		t.Fatalf("list friends (system): %v", err)
	}
	if gotQuery != "include=system" {
		t.Errorf("opt-in sent query %q, want include=system", gotQuery)
	}
}

// TestListFriendCandidatesAlwaysSendsSearch pins that the search term reaches
// the wire. It became required upstream (#853/#856) because an empty query
// returned the entire user directory.
func TestListFriendCandidatesAlwaysSendsSearch(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		// #1159 REMOVE (97a91d94): the server emits `friends`, never `reviewers`.
		// Nothing here parses the body — this test asserts on r.URL.RawQuery — but the
		// payload names the server it models so a future sweep does not have to guess.
		_, _ = w.Write([]byte(`{"friends":[],"total":0}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.ListFriendCandidates(context.Background(), "gordon"); err != nil {
		t.Fatalf("list candidates: %v", err)
	}
	if gotQuery != "search=gordon" {
		t.Errorf("query = %q, want search=gordon", gotQuery)
	}
}
