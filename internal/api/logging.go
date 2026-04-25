package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// loggingTransport is an http.RoundTripper that logs every request/response
// at debug level. It wraps an inner transport (typically http.DefaultTransport).
type loggingTransport struct {
	inner  http.RoundTripper
	logger *slog.Logger
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		t.logger.Debug("api request failed",
			"method", req.Method,
			"uri", req.URL.RequestURI(),
			"error", err,
			"duration", time.Since(start).Round(time.Millisecond),
		)
		return nil, err
	}

	duration := time.Since(start).Round(time.Millisecond)
	size := resp.ContentLength

	uri := req.URL.RequestURI()
	t.logger.Debug(fmt.Sprintf("api: %s %s → %d (%s, %s)",
		req.Method,
		uri,
		resp.StatusCode,
		duration,
		formatBytes(size),
	),
		"method", req.Method,
		"uri", uri,
		"status", resp.StatusCode,
		"duration", duration,
		"size", size,
	)

	return resp, nil
}

// formatBytes formats a byte count into a human-readable string.
// A negative value (unknown content length) returns "?".
func formatBytes(b int64) string {
	if b < 0 {
		return "?"
	}
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(b)/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.1fKB", float64(b)/1024)
	default:
		return fmt.Sprintf("%dB", b)
	}
}

