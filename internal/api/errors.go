package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ErrorCategory classifies errors for consumer decision-making (retry, escalate, fix input).
type ErrorCategory int

const (
	CategoryUnknown     ErrorCategory = iota
	CategoryNetwork                   // connection refused, DNS, timeout
	CategoryAuth                      // 401, 403
	CategoryNotFound                  // 404
	CategoryConflict                  // 409
	CategoryValidation                // 400, 422
	CategoryRateLimit                 // 429
	CategoryServerError               // 500, 502, 503, 504
)

func (c ErrorCategory) String() string {
	switch c {
	case CategoryNetwork:
		return "network"
	case CategoryAuth:
		return "auth"
	case CategoryNotFound:
		return "not_found"
	case CategoryConflict:
		return "conflict"
	case CategoryValidation:
		return "validation"
	case CategoryRateLimit:
		return "rate_limit"
	case CategoryServerError:
		return "server_error"
	default:
		return "unknown"
	}
}

// ErrorInfo contains classified error details for structured consumption.
type ErrorInfo struct {
	Category   ErrorCategory
	Code       string // machine-readable code (e.g. "connection_refused", "internal_error")
	StatusCode int    // HTTP status code, 0 for non-HTTP errors
	Retryable  bool
	Message    string // user-friendly message
	RawError   error  // original error
}

// ClassifyError unwraps an error and returns structured classification.
func ClassifyError(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	// Check for APIError (HTTP response errors).
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return classifyAPIError(apiErr)
	}

	// Check for network errors.
	if info := classifyNetworkError(err); info != nil {
		return info
	}

	// Check for context errors.
	if errors.Is(err, errors.ErrUnsupported) || isContextError(err) {
		return &ErrorInfo{
			Category: CategoryUnknown,
			Code:     "cancelled",
			Message:  "Request was cancelled.",
			RawError: err,
		}
	}

	return &ErrorInfo{
		Category: CategoryUnknown,
		Code:     "unknown",
		Message:  err.Error(),
		RawError: err,
	}
}

// IsRetryableErr returns true if the error is classified as retryable.
func IsRetryableErr(err error) bool {
	info := ClassifyError(err)
	return info != nil && info.Retryable
}

// UserFriendlyError returns a concise message suitable for display.
func UserFriendlyError(err error) string {
	info := ClassifyError(err)
	if info == nil {
		return ""
	}
	return info.Message
}

// classifyAPIError classifies errors from HTTP responses.
func classifyAPIError(e *APIError) *ErrorInfo {
	info := &ErrorInfo{
		StatusCode: e.StatusCode,
		RawError:   e,
	}

	// Try to extract structured error from response body.
	serverMsg := parseServerError(e.Body)

	switch {
	case e.StatusCode == 400:
		info.Category = CategoryValidation
		info.Code = "bad_request"
		info.Message = withFallback(serverMsg, "Invalid request. Check your parameters.")
	case e.StatusCode == 401:
		info.Category = CategoryAuth
		info.Code = "unauthorized"
		info.Message = "Authentication failed. Check your API token."
	case e.StatusCode == 403:
		info.Category = CategoryAuth
		serverCode := codeFromServer(e.Body, "")
		if serverCode == "members_only" {
			info.Code = "members_only"
			info.Message = "This story is members-only. Authenticate with a registered account to read it."
		} else {
			info.Code = "forbidden"
			info.Message = "Access denied. You don't have permission for this operation."
		}
	case e.StatusCode == 404:
		info.Category = CategoryNotFound
		info.Code = "not_found"
		info.Message = withFallback(serverMsg, "Resource not found.")
	case e.StatusCode == 409:
		info.Category = CategoryConflict
		info.Code = "conflict"
		info.Message = withFallback(serverMsg, "Conflict. The resource may have been modified.")
	case e.StatusCode == 422:
		info.Category = CategoryValidation
		info.Code = "unprocessable"
		info.Message = withFallback(serverMsg, "Request could not be processed. Check input.")
	case e.StatusCode == 429:
		info.Category = CategoryRateLimit
		info.Code = "rate_limited"
		info.Retryable = true
		info.Message = "Rate limited. Try again shortly."
	case e.StatusCode == 500:
		info.Category = CategoryServerError
		info.Code = codeFromServer(e.Body, "internal_error")
		info.Message = withFallback(serverMsg, "Server error. This is not your fault — try again shortly.")
		info.Retryable = true
	case e.StatusCode == 502:
		info.Category = CategoryServerError
		info.Code = "bad_gateway"
		info.Retryable = true
		info.Message = "API server is temporarily unavailable. Try again shortly."
	case e.StatusCode == 503:
		info.Category = CategoryServerError
		info.Code = "service_unavailable"
		info.Retryable = true
		info.Message = "API server is temporarily unavailable. Try again shortly."
	case e.StatusCode == 504:
		info.Category = CategoryServerError
		info.Code = "gateway_timeout"
		info.Retryable = true
		info.Message = "API request timed out. Try again shortly."
	default:
		info.Category = CategoryUnknown
		info.Code = fmt.Sprintf("http_%d", e.StatusCode)
		info.Message = withFallback(serverMsg, e.Error())
	}

	return info
}

// classifyNetworkError checks for common network-level errors.
func classifyNetworkError(err error) *ErrorInfo {
	base := &ErrorInfo{
		Category:  CategoryNetwork,
		Retryable: true,
		RawError:  err,
	}

	// Check for URL errors (wraps net.OpError, etc.)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err // unwrap for inner checks
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		base.Code = "connection_error"
		base.Message = fmt.Sprintf("Cannot reach API server: %s.", opErr.Op)
		return base
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		base.Code = "dns_error"
		if dnsErr.IsTemporary {
			base.Message = "DNS lookup failed temporarily. Try again shortly."
		} else {
			base.Retryable = false
			base.Message = fmt.Sprintf("DNS lookup failed for %s. Check your API URL.", dnsErr.Name)
		}
		return base
	}

	// Check error message strings for common patterns.
	msg := err.Error()
	switch {
	case strings.Contains(msg, "connection refused"):
		base.Code = "connection_refused"
		base.Message = "API server is not reachable. Check if the server is running."
		return base
	case strings.Contains(msg, "connection reset"):
		base.Code = "connection_reset"
		base.Message = "Connection was reset. Try again shortly."
		return base
	case strings.Contains(msg, "no such host"):
		base.Code = "dns_error"
		base.Retryable = false
		base.Message = "API host not found. Check your API URL."
		return base
	case strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "deadline exceeded"):
		base.Code = "timeout"
		base.Message = "Request timed out. The server may be overloaded."
		return base
	case strings.Contains(msg, "EOF"):
		base.Code = "connection_closed"
		base.Message = "Connection closed unexpectedly. Try again shortly."
		return base
	}

	return nil
}

// serverErrorBody represents the JSON structure returned by the ProseForge API on errors.
type serverErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// parseServerError tries to extract a human-readable message from a JSON error body.
func parseServerError(body string) string {
	if body == "" {
		return ""
	}
	var seb serverErrorBody
	if err := json.Unmarshal([]byte(body), &seb); err != nil {
		return ""
	}
	if seb.Message != "" && seb.Message != "An unexpected error occurred" {
		return seb.Message
	}
	return ""
}

// codeFromServer extracts the "error" field from a JSON body, falling back to def.
func codeFromServer(body, def string) string {
	if body == "" {
		return def
	}
	var seb serverErrorBody
	if err := json.Unmarshal([]byte(body), &seb); err != nil || seb.Error == "" {
		return def
	}
	return seb.Error
}

func withFallback(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func isContextError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "context canceled") || strings.Contains(msg, "context deadline exceeded")
}
