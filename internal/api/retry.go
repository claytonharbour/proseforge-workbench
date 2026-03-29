package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
)

// retryTransport is an http.RoundTripper that retries failed requests with
// exponential backoff. It sits in the transport chain above loggingTransport:
//
//	http.Client → retryTransport → loggingTransport → http.DefaultTransport
type retryTransport struct {
	inner      http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
	logger     *slog.Logger
}

func newRetryTransport(inner http.RoundTripper, maxRetries int, logger *slog.Logger) *retryTransport {
	return &retryTransport{
		inner:      inner,
		maxRetries: maxRetries,
		baseDelay:  1 * time.Second,
		logger:     logger,
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Buffer the body on the first call so we can replay it on retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("buffering request body for retry: %w", err)
		}
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Reset body for retries.
		if attempt > 0 && bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = t.inner.RoundTrip(req)

		if attempt >= t.maxRetries {
			break
		}

		retry, reason := t.shouldRetry(resp, err)
		if !retry {
			break
		}

		// Close response body from failed attempt to prevent leaks.
		if resp != nil {
			resp.Body.Close()
		}

		delay := t.backoffDelay(attempt, resp)
		t.logger.Warn("retrying API request",
			"method", req.Method,
			"path", req.URL.Path,
			"attempt", attempt+1,
			"max_retries", t.maxRetries,
			"reason", reason,
			"delay", delay.Round(time.Millisecond),
		)

		if err := sleepWithContext(req.Context(), delay); err != nil {
			return nil, err
		}
	}

	return resp, err
}

// shouldRetry decides if a request should be retried based on the response or error.
func (t *retryTransport) shouldRetry(resp *http.Response, err error) (bool, string) {
	if err != nil {
		// Network-level errors are generally retryable.
		info := ClassifyError(err)
		if info != nil && info.Retryable {
			return true, info.Code
		}
		// Default: retry network errors even if we can't classify them,
		// unless it's a context cancellation.
		if info != nil && info.Code == "cancelled" {
			return false, ""
		}
		return true, "network_error"
	}

	if resp == nil {
		return false, ""
	}

	switch resp.StatusCode {
	case 429:
		return true, "rate_limited"
	case 500:
		return true, "internal_error"
	case 502:
		return true, "bad_gateway"
	case 503:
		return true, "service_unavailable"
	case 504:
		return true, "gateway_timeout"
	}

	return false, ""
}

// backoffDelay calculates the delay before the next retry attempt.
// Uses exponential backoff (1s, 2s, 4s, ...) but respects Retry-After header for 429.
func (t *retryTransport) backoffDelay(attempt int, resp *http.Response) time.Duration {
	// Check for Retry-After header on 429 responses.
	if resp != nil && resp.StatusCode == 429 {
		if d := parseRetryAfter(resp.Header.Get("Retry-After")); d > 0 {
			// Cap at 60s to prevent absurd waits.
			if d > 60*time.Second {
				d = 60 * time.Second
			}
			return d
		}
	}

	// Exponential backoff: baseDelay * 2^attempt
	return t.baseDelay * time.Duration(math.Pow(2, float64(attempt)))
}

// parseRetryAfter parses a Retry-After header value as seconds.
func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 0
	}
	secs, err := strconv.Atoi(val)
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// sleepWithContext waits for the given duration or until the context is done.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
