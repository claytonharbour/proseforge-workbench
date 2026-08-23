package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollaborationEndpointsUseGeneratedContract(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/story/story-1/grants":
			_, _ = w.Write([]byte(`{"grants":[{"email":"reader@example.com","capability":"story:view","principalId":"principal-1","createdAt":"2026-08-12T00:00:00Z"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/story/story-1/grants":
			_, _ = w.Write([]byte(`{"email":"reader@example.com","capability":"story:view","principalId":"principal-1"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/story/story-1/grants":
			_, _ = w.Write([]byte(`{"revoked":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/story/story-1/grants/mine":
			_, _ = w.Write([]byte(`{"left":true}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/series/series-1/grants":
			_, _ = w.Write([]byte(`{"grants":[{"email":"reader@example.com","capability":"room:enter","principalId":"principal-1"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/series/series-1/grants":
			_, _ = w.Write([]byte(`{"email":"reader@example.com","capability":"room:enter","principalId":"principal-1"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/series/series-1/grants":
			_, _ = w.Write([]byte(`{"revoked":true}`))
		case r.URL.Path == "/api/v1/users/me/shared-stories":
			_, _ = w.Write([]byte(`{"stories":[]}`))
		case r.URL.Path == "/api/v1/story/story-1/contributions":
			_, _ = w.Write([]byte(`{"contributions":[]}`))
		case r.URL.Path == "/api/v1/story/story-1/contributions/accounting":
			_, _ = w.Write([]byte(`{"contributors":[{"contributorEmail":"reader@example.com","netWords":12,"details":[]}]}`))
		case r.URL.Path == "/api/v1/series/series-1/contributions/accounting":
			_, _ = w.Write([]byte(`{"contributors":[]}`))
		case r.URL.Path == "/api/v1/story/story-1/contributions/mine":
			_, _ = w.Write([]byte(`{"contribution":{"id":"contrib-1","status":"draft"}}`))
		case r.URL.Path == "/api/v1/story/story-1/contributions/contrib-1/diff":
			_, _ = w.Write([]byte(`{"contributionId":"contrib-1","files":[]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/story/story-1/contributions/contrib-1/ready-for-owner":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/story/story-1/contributions/contrib-1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost || r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"contribution":{"id":"contrib-1","status":"ready"}}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.GetSharedStories(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListStoryGrants(ctx, "story-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GrantStoryCapability(ctx, "story-1", "reader@example.com", "story:view"); err != nil {
		t.Fatal(err)
	}
	if err := client.RevokeStoryCapability(ctx, "story-1", "reader@example.com", "story:view"); err != nil {
		t.Fatal(err)
	}
	if err := client.LeaveStory(ctx, "story-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListSeriesGrants(ctx, "series-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GrantSeriesCapability(ctx, "series-1", "reader@example.com", "room:enter"); err != nil {
		t.Fatal(err)
	}
	if err := client.RevokeSeriesCapability(ctx, "series-1", "reader@example.com", "room:enter"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListContributions(ctx, "story-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.StoryContributionAccounting(ctx, "story-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.SeriesContributionAccounting(ctx, "series-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetMyContribution(ctx, "story-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetContributionDiff(ctx, "story-1", "contrib-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.MarkContributionReady(ctx, "story-1", "contrib-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.NotifyOwnerContributionReady(ctx, "story-1", "contrib-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.DiscardContribution(ctx, "story-1", "contrib-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RequestContributionChanges(ctx, "story-1", "contrib-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.MergeContribution(ctx, "story-1", "contrib-1", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PullContribution(ctx, "story-1", "contrib-1", "contrib-branch", map[string]string{"content/1.md": "resolved"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.SuggestContributionSection(ctx, "story-1", "contrib-1", "section-1", "suggested"); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 20 {
		t.Fatalf("requests = %d, want 20: %v", len(requests), requests)
	}
}

func TestPullContributionResolutionsAreJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["resolutions"]["content/1.md"] != "resolved" {
			t.Fatalf("body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"autoMerged":[],"filesUpdated":1,"resolved":1}`))
	}))
	defer srv.Close()
	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.PullContribution(context.Background(), "story-1", "contrib-1", "branch", map[string]string{"content/1.md": "resolved"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Resolved == nil || *result.Resolved != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestSyncContributionPullsMasterIntoContribution(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.EscapedPath())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/story/story-1/contributions/contrib-1/pull/master" {
			_, _ = w.Write([]byte(`{"autoMerged":[],"filesUpdated":0,"resolved":0}`))
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.SyncContribution(context.Background(), "story-1", "contrib-1", nil); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths = %v, want one pull from master", paths)
	}
}

// The wire contract for partial accept (proseforge#812, workbench#275 B1).
//
// 🛑 The case that matters is `nil`: every existing caller merges everything, and
// if nil serialised as `"selections": {}` a fail-closed backend would read it as
// "no paths named" and 400 every merge in the fleet. So this asserts the KEY IS
// ABSENT, not merely that it decodes to nil — those differ on the wire and only
// one of them is safe.
func TestMergeContributionSelectionsWireContract(t *testing.T) {
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"contribution":{"id":"contrib-1","status":"merged"}}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if _, err := client.MergeContribution(ctx, "story-1", "contrib-1", nil); err != nil {
		t.Fatal(err)
	}
	if _, present := bodies[0]["selections"]; present {
		t.Errorf("nil selections must omit the key entirely; body was %v", bodies[0])
	}

	sel := map[string]bool{"content/a.md": true, "content/b.md": false}
	if _, err := client.MergeContribution(ctx, "story-1", "contrib-1", sel); err != nil {
		t.Fatal(err)
	}
	got, present := bodies[1]["selections"].(map[string]any)
	if !present {
		t.Fatalf("selections missing from body: %v", bodies[1])
	}
	if got["content/a.md"] != true || got["content/b.md"] != false {
		t.Errorf("selections not transmitted faithfully: %v", got)
	}

	// An EMPTY map is not nil. It must reach the server so the fail-closed rule
	// is the server's to apply — this layer does not get to invent "accept none".
	if _, err := client.MergeContribution(ctx, "story-1", "contrib-1", map[string]bool{}); err != nil {
		t.Fatal(err)
	}
	if _, present := bodies[2]["selections"]; !present {
		t.Errorf("an empty selections map must still be sent, distinct from nil; body was %v", bodies[2])
	}
}
