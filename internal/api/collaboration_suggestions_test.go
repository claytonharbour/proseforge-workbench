package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestUpdateContributionSuggestionStatusPassesThrough checks the status reaches
// the wire unaltered, whatever it is.
//
// The vocabulary is being simplified upstream, so this client deliberately does
// not police the set: a hard-coded list would reject values the server accepts
// the moment it changes, and only a workbench release could clear that. The
// generated model comment already disagrees with the route documentation (three
// values against five), which is the same drift arriving early.
//
// "ready" is in the list precisely because it is not valid today — it is where
// the vocabulary is heading, and it must not need a code change to work.
func TestUpdateContributionSuggestionStatusPassesThrough(t *testing.T) {
	var sent []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		sent = append(sent, string(buf))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}

	statuses := []string{"accepted", "rejected", "acknowledged", "dismissed", "pending", "ready", "draft"}
	for _, status := range statuses {
		if _, err := client.UpdateContributionSuggestionStatus(
			context.Background(), "story-1", "contrib-1", "sug-1", status); err != nil {
			t.Errorf("status %q rejected by the client: %v", status, err)
		}
	}
	if len(sent) != len(statuses) {
		t.Fatalf("reached the wire %d times, want %d", len(sent), len(statuses))
	}
	for i, status := range statuses {
		if !strings.Contains(sent[i], status) {
			t.Errorf("status %q did not reach the body: %s", status, sent[i])
		}
	}
}

// TestUpdateContributionSuggestionStatusRequiresStatus checks the one case
// still refused locally: an empty status is a caller mistake, not a vocabulary
// question, and sending it would ask the server to move a suggestion to nothing.
func TestUpdateContributionSuggestionStatusRequiresStatus(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.UpdateContributionSuggestionStatus(
		context.Background(), "story-1", "contrib-1", "sug-1", ""); err == nil {
		t.Fatal("expected an error for an empty status")
	}
	if called {
		t.Error("empty status reached the server; it must be refused locally")
	}
}
