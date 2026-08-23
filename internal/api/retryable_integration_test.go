package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestRetryableHint_EndToEnd drives the real Client + retry transport against a
// controlled HTTP server to prove the forge/proseforge#578 fix end-to-end: an
// explicit retryable:false on a 5xx is honored (surfaced on the first attempt,
// not retried), while retryable:true / no-hint 5xx keep the blanket-retry
// behavior. The dev backend cannot emit retryable:false yet (that's the backend
// half of #578), so a local server is the only faithful way to exercise this
// path today.
func TestRetryableHint_EndToEnd(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		maxRetries int
		wantCalls  int32
		wantInErr  string
	}{
		{
			name:       "deterministic 500 with retryable:false surfaces on first attempt",
			body:       `{"error":"reservation_exists","message":"reservation exists for job","retryable":false}`,
			maxRetries: 3,
			wantCalls:  1, // no retries
			wantInErr:  "reservation exists for job",
		},
		{
			name:       "500 with retryable:true exhausts retries (backward compatible)",
			body:       `{"error":"internal_error","retryable":true}`,
			maxRetries: 1,
			wantCalls:  2, // 1 initial + 1 retry
			wantInErr:  "API 500",
		},
		{
			name:       "500 with no hint exhausts retries (backward compatible)",
			body:       `{"error":"internal_error","message":"boom"}`,
			maxRetries: 1,
			wantCalls:  2,
			wantInErr:  "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client, err := New(srv.URL, "test-token", WithRetry(tt.maxRetries))
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			// Keep the exhaustion cases fast — shrink the transport backoff.
			if rt, ok := client.httpClient.Transport.(*retryTransport); ok {
				rt.baseDelay = 1 * time.Millisecond
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = client.RegenerateSection(ctx, "story-id", "section-id", true, "")
			if err == nil {
				t.Fatal("expected an error from a 500 response, got nil")
			}
			if got := atomic.LoadInt32(&calls); got != tt.wantCalls {
				t.Errorf("backend calls = %d, want %d", got, tt.wantCalls)
			}
			if !strings.Contains(err.Error(), tt.wantInErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantInErr)
			}
		})
	}
}
