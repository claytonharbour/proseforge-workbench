package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 🛑 A FAILING TIER LOOKUP MUST NOT FAIL WhoAmI (#448), and this is not politeness.
//
// The watcher's identity guard calls WhoAmI inside Watch(), so it runs on every
// tick of every leg. checkIdentity classifies any NON-transport error from it as
// ErrIdentityGuard — which exits 10, ENDS the loop and "needs a human". So if a
// billing 500 could propagate out of WhoAmI, one wobble on an endpoint that has
// nothing to do with identity would permanently stop every watcher on the box.
//
// The second reason is the operator's: whoami is what you reach for WHILE
// diagnosing a 403. A version that dies on a 403 is useless exactly when needed.
func TestWhoAmISurvivesTheTierLookupFailing(t *testing.T) {
	const identity = `{"id":"u-1","email":"reviewer@example.test","name":"Reviewer"}`

	newServer := func(t *testing.T, billing func(w http.ResponseWriter)) *httptest.Server {
		t.Helper()
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/v1/users/me":
				_, _ = w.Write([]byte(identity))
			case "/api/v1/billing/subscription":
				billing(w)
			default:
				t.Errorf("unexpected request: %s", r.URL.RequestURI())
			}
		}))
	}

	decode := func(t *testing.T, raw json.RawMessage) map[string]any {
		t.Helper()
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		return m
	}

	// ⛔ THE ARM THAT MATTERS. Billing 500 must not become an identity failure.
	t.Run("billing 500 still returns the identity", func(t *testing.T) {
		srv := newServer(t, func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		})
		defer srv.Close()

		client, err := New(srv.URL, "test-token", WithRetry(0))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := client.WhoAmI(context.Background())
		if err != nil {
			t.Fatalf("WhoAmI returned an error on a billing failure: %v\n"+
				"⇒ the identity guard would classify this as ErrIdentityGuard, exit 10 "+
				"and stop every watcher permanently", err)
		}
		got := decode(t, raw)
		if got["email"] != "reviewer@example.test" {
			t.Errorf("identity lost: %v", got)
		}
		if got["tier"] != nil {
			t.Errorf("tier should be null when it could not be read, got %v", got["tier"])
		}
		// "could not ask" and "no entitlement" are different facts; a caller must
		// be able to tell them apart.
		if s, _ := got["tierError"].(string); s == "" {
			t.Error("tierError must say why the tier is missing, otherwise an unknown tier is indistinguishable from an empty one")
		}
	})

	// 403 specifically: the case an operator is most likely to be debugging when
	// they run this command.
	t.Run("billing 403 still returns the identity", func(t *testing.T) {
		srv := newServer(t, func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"forbidden"}`))
		})
		defer srv.Close()

		client, err := New(srv.URL, "test-token", WithRetry(0))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := client.WhoAmI(context.Background())
		if err != nil {
			t.Fatalf("WhoAmI died on a 403 — the exact situation it exists to diagnose: %v", err)
		}
		if decode(t, raw)["email"] != "reviewer@example.test" {
			t.Error("identity lost on a 403")
		}
	})

	// ✅ POSITIVE CONTROL. Without this the assertions above pass on a WhoAmI that
	// never reports a tier at all — the failure mode #448 exists to fix, rebuilt
	// inside its own regression test.
	t.Run("a readable tier is reported", func(t *testing.T) {
		srv := newServer(t, func(w http.ResponseWriter) {
			_, _ = w.Write([]byte(`{"subscription":{"tierId":"tier_trial","tierName":"Loremaster",` +
				`"tierFeatures":["narration_forge"],"isOverridden":true,"baseTierName":"Apprentice"}}`))
		})
		defer srv.Close()

		client, err := New(srv.URL, "test-token", WithRetry(0))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := client.WhoAmI(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		got := decode(t, raw)
		if got["tierError"] != nil {
			t.Errorf("tierError set on a successful lookup: %v", got["tierError"])
		}
		tier, ok := got["tier"].(map[string]any)
		if !ok {
			t.Fatalf("tier missing or wrong shape: %v", got["tier"])
		}
		// ⚠️ The RESOLVED name, not the raw row. An overridden account reports
		// tierId "tier_trial" beside tierName "Loremaster"; reading the former is
		// what produced three wrong capability claims in one morning.
		if tier["tierName"] != "Loremaster" {
			t.Errorf("tierName = %v, want the resolved value Loremaster", tier["tierName"])
		}
		if tier["tierId"] != "tier_trial" {
			t.Errorf("tierId = %v — the raw row is expected to disagree; if it stops "+
				"disagreeing, this test no longer proves the resolved value is used", tier["tierId"])
		}
	})

	// Entitlement is asked as a question, not by attempting the action.
	t.Run("Can answers without performing the capability", func(t *testing.T) {
		sub := &Subscription{TierFeatures: []string{"narration_forge", "collaborate"}}
		if !sub.Can("narration_forge") {
			t.Error("Can said no to a feature that is present")
		}
		if sub.Can("series_forge") {
			t.Error("Can said yes to a feature that is absent")
		}
		if (&Subscription{}).Can("narration_forge") {
			t.Error("an empty feature list must not grant anything")
		}
	})

}
