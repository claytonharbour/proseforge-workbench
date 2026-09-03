package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 🛑 force=true IS NOT IN THE SPEC FOR THIS ROUTE (#455, forge/proseforge#1310).
//
// /story/{id}/quality/assess declares a `force` query param; the /{sha} form
// declares only id and sha. So the generated CreateStoryQualityAssess takes no
// params struct, and the flag survives the migration to the generated client
// only because a request editor puts it back.
//
// Dropping it would not fail to compile, would not fail any other test, and
// would not error at runtime — it would quietly return CACHED scores where the
// caller asked for a re-assessment. That is the failure this test exists for,
// so it asserts the WIRE, not the call.
func TestAssessQualityAtVersionSendsForce(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := New(srv.URL, "tok", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AssessQualityAtVersion(context.Background(), "story-1", "abc123"); err != nil {
		t.Fatal(err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if want := "/api/v1/story/story-1/quality/assess/abc123"; gotPath != want {
		t.Errorf("path = %s, want %s", gotPath, want)
	}
	// The whole point. An empty query here means the editor did not fire.
	if want := "force=true"; gotQuery != want {
		t.Errorf("query = %q, want %q — force was dropped, so this now returns cached scores", gotQuery, want)
	}
}
