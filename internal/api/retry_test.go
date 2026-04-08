package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// mockTransport is a test RoundTripper that returns preconfigured responses.
type mockTransport struct {
	responses []*http.Response
	errors    []error
	calls     int
	bodies    []string // captures request bodies from each attempt
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	i := m.calls
	m.calls++

	// Capture request body.
	if req.Body != nil {
		data, _ := io.ReadAll(req.Body)
		m.bodies = append(m.bodies, string(data))
		// Restore body so it can be read again if needed.
		req.Body = io.NopCloser(bytes.NewReader(data))
	}

	if i < len(m.errors) && m.errors[i] != nil {
		return nil, m.errors[i]
	}
	if i < len(m.responses) {
		return m.responses[i], nil
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRetryTransport_SuccessNoRetry(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("ok")))},
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 1 * time.Millisecond

	req, _ := http.NewRequest("GET", "http://example.com/api/v1/stories", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if mock.calls != 1 {
		t.Errorf("calls = %d, want 1", mock.calls)
	}
}

func TestRetryTransport_NetworkRetryThenSuccess(t *testing.T) {
	mock := &mockTransport{
		errors: []error{
			errors.New("connection refused"),
			errors.New("connection refused"),
			nil,
		},
		responses: []*http.Response{
			nil,
			nil,
			{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))},
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 1 * time.Millisecond

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if mock.calls != 3 {
		t.Errorf("calls = %d, want 3", mock.calls)
	}
}

func TestRetryTransport_StatusRetryThenSuccess(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			{StatusCode: 503, Body: io.NopCloser(bytes.NewReader(nil))},
			{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("ok")))},
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 1 * time.Millisecond

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if mock.calls != 2 {
		t.Errorf("calls = %d, want 2", mock.calls)
	}
}

func TestRetryTransport_ExhaustedRetries(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			{StatusCode: 503, Body: io.NopCloser(bytes.NewReader(nil))},
			{StatusCode: 503, Body: io.NopCloser(bytes.NewReader(nil))},
			{StatusCode: 503, Body: io.NopCloser(bytes.NewReader(nil))},
			{StatusCode: 503, Body: io.NopCloser(bytes.NewReader(nil))},
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 1 * time.Millisecond

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("status = %d, want 503 (last attempt)", resp.StatusCode)
	}
	// 1 initial + 3 retries = 4 total
	if mock.calls != 4 {
		t.Errorf("calls = %d, want 4", mock.calls)
	}
}

func TestRetryTransport_BodyBuffering(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			{StatusCode: 500, Body: io.NopCloser(bytes.NewReader(nil))},
			{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))},
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 1 * time.Millisecond

	body := `{"text":"hello"}`
	req, _ := http.NewRequest("POST", "http://example.com/test", bytes.NewBufferString(body))
	_, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 2 {
		t.Fatalf("calls = %d, want 2", mock.calls)
	}
	// Both attempts should see the same body.
	for i, got := range mock.bodies {
		if got != body {
			t.Errorf("attempt %d body = %q, want %q", i, got, body)
		}
	}
}

func TestRetryTransport_ContextCancel(t *testing.T) {
	mock := &mockTransport{
		errors: []error{
			errors.New("connection refused"),
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 5 * time.Second // long delay so cancel wins

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com/test", nil)

	// Cancel after a short delay.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := rt.RoundTrip(req)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestRetryTransport_NoRetryOn4xx(t *testing.T) {
	statuses := []int{400, 401, 403, 404, 409, 422}
	for _, status := range statuses {
		mock := &mockTransport{
			responses: []*http.Response{
				{StatusCode: status, Body: io.NopCloser(bytes.NewReader(nil))},
			},
		}
		rt := newRetryTransport(mock, 3, testLogger())
		rt.baseDelay = 1 * time.Millisecond

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		resp, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("status %d: unexpected error: %v", status, err)
		}
		if resp.StatusCode != status {
			t.Errorf("status %d: got %d", status, resp.StatusCode)
		}
		if mock.calls != 1 {
			t.Errorf("status %d: calls = %d, want 1 (no retry)", status, mock.calls)
		}
	}
}

func TestRetryTransport_RetryAfterHeader(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "2")
	mock := &mockTransport{
		responses: []*http.Response{
			{StatusCode: 429, Header: header, Body: io.NopCloser(bytes.NewReader(nil))},
			{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))},
		},
	}
	rt := newRetryTransport(mock, 3, testLogger())
	rt.baseDelay = 1 * time.Millisecond

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	start := time.Now()
	resp, err := rt.RoundTrip(req)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	// Should have waited ~2 seconds for Retry-After.
	if elapsed < 1*time.Second {
		t.Errorf("elapsed %v too short — Retry-After should have caused ~2s wait", elapsed)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		val  string
		want time.Duration
	}{
		{"5", 5 * time.Second},
		{"0", 0},
		{"-1", 0},
		{"not a number", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseRetryAfter(tt.val)
		if got != tt.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

func TestBackoffDelay(t *testing.T) {
	rt := newRetryTransport(nil, 3, testLogger())
	rt.baseDelay = 1 * time.Second

	// Exponential: 1s, 2s, 4s
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	for i, want := range delays {
		got := rt.backoffDelay(i, nil)
		if got != want {
			t.Errorf("attempt %d: delay = %v, want %v", i, got, want)
		}
	}
}
