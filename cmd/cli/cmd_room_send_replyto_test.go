package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// 🛑 AN EMPTY --reply-to MUST NOT POST (#466).
//
// `--reply-to ""` used to be accepted: the message posted TOP-LEVEL, exit 0, with
// the ordinary success block. A scripted reply whose id extraction returned empty
// silently became a broadcast, and the only way to notice was reading
// `threadRootId` back off your own message afterwards.
func TestValidateReplyTo(t *testing.T) {
	for _, tc := range []struct {
		name    string
		changed bool
		value   string
		wantErr bool
	}{
		// ✅ NEGATIVE CONTROL, and the arm that stops this guard breaking every
		// ordinary post in the fleet: omitting the flag is a legitimate intent to
		// post top-level and must stay silent. Without it the check could "pass"
		// by refusing everything.
		{"omitted is a legitimate top-level post", false, "", false},

		// ⛔ The defect. Same "" as above, and the opposite meaning.
		{"given but empty is a failed lookup", true, "", true},

		// `mid=$(… | tr -d '\n')` yielding spaces is the same failure wearing
		// whitespace, and it reached the server as a silent broadcast too.
		{"given but whitespace-only", true, "   ", true},
		{"given but a newline", true, "\n", true},

		// A real id must pass, or the guard has eaten the feature.
		{"a real id is accepted", true, "1788199431155-0", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReplyTo(tc.changed, tc.value)
			if tc.wantErr && err == nil {
				t.Fatalf("changed=%v value=%q was accepted; a failed lookup would post a broadcast", tc.changed, tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("changed=%v value=%q was refused: %v", tc.changed, tc.value, err)
			}
			if tc.wantErr && !strings.Contains(err.Error(), "reply-to") {
				t.Errorf("error does not name the flag, so the caller cannot tell what to fix: %v", err)
			}
		})
	}
}

// ⚠️ ASSERTING THE ERROR IS NOT ENOUGH, and that is the whole point of this test.
// The defect was never "no error" — it was "posted anyway". A test that only
// checked the message would pass against the broken build, because the broken
// build still sent. This arm asserts what reached the SERVER.
func TestRoomSendWithAnEmptyReplyToSendsNothing(t *testing.T) {
	var sends int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			atomic.AddInt32(&sends, 1)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1-0"}`))
	}))
	defer srv.Close()

	cmd := newRoomSendCmd()
	cmd.Flags().String("url", srv.URL, "")
	cmd.Flags().String("token", "test-token", "")
	cmd.Flags().String("credentials-file", "", "")
	cmd.Flags().String("type", "conversation", "")
	cmd.SetArgs([]string{"room-1", "--reply-to", "", "--content", "must not be sent"})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true

	if err := cmd.Execute(); err == nil {
		t.Fatal("empty --reply-to was accepted")
	}
	if n := atomic.LoadInt32(&sends); n != 0 {
		t.Errorf("message WAS sent (%d requests) despite the guard — the broadcast still happened", n)
	}
}
