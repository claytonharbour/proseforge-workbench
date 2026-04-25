package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"
)

func TestClassifyError_Nil(t *testing.T) {
	if ClassifyError(nil) != nil {
		t.Fatal("expected nil for nil error")
	}
}

func TestClassifyError_APIErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *APIError
		wantCat    ErrorCategory
		wantCode   string
		wantRetry  bool
		wantStatus int
	}{
		{
			name:       "400 bad request",
			err:        &APIError{StatusCode: 400, Status: "400 Bad Request", Body: `{"error":"bad_request","message":"invalid story_id"}`},
			wantCat:    CategoryValidation,
			wantCode:   "bad_request",
			wantRetry:  false,
			wantStatus: 400,
		},
		{
			name:       "401 unauthorized",
			err:        &APIError{StatusCode: 401, Status: "401 Unauthorized"},
			wantCat:    CategoryAuth,
			wantCode:   "unauthorized",
			wantRetry:  false,
			wantStatus: 401,
		},
		{
			name:       "403 forbidden",
			err:        &APIError{StatusCode: 403, Status: "403 Forbidden"},
			wantCat:    CategoryAuth,
			wantCode:   "forbidden",
			wantRetry:  false,
			wantStatus: 403,
		},
		{
			name:       "403 members_only",
			err:        &APIError{StatusCode: 403, Status: "403 Forbidden", Body: `{"error":"members_only","message":"This story is available to registered members."}`},
			wantCat:    CategoryAuth,
			wantCode:   "members_only",
			wantRetry:  false,
			wantStatus: 403,
		},
		{
			name:       "404 not found",
			err:        &APIError{StatusCode: 404, Status: "404 Not Found"},
			wantCat:    CategoryNotFound,
			wantCode:   "not_found",
			wantRetry:  false,
			wantStatus: 404,
		},
		{
			name:       "409 conflict",
			err:        &APIError{StatusCode: 409, Status: "409 Conflict"},
			wantCat:    CategoryConflict,
			wantCode:   "conflict",
			wantRetry:  false,
			wantStatus: 409,
		},
		{
			name:       "422 unprocessable",
			err:        &APIError{StatusCode: 422, Status: "422 Unprocessable Entity"},
			wantCat:    CategoryValidation,
			wantCode:   "unprocessable",
			wantRetry:  false,
			wantStatus: 422,
		},
		{
			name:       "429 rate limit",
			err:        &APIError{StatusCode: 429, Status: "429 Too Many Requests"},
			wantCat:    CategoryRateLimit,
			wantCode:   "rate_limited",
			wantRetry:  true,
			wantStatus: 429,
		},
		{
			name:       "500 internal error",
			err:        &APIError{StatusCode: 500, Status: "500 Internal Server Error", Body: `{"error":"internal_error","message":"An unexpected error occurred"}`},
			wantCat:    CategoryServerError,
			wantCode:   "internal_error",
			wantRetry:  true,
			wantStatus: 500,
		},
		{
			name:       "502 bad gateway",
			err:        &APIError{StatusCode: 502, Status: "502 Bad Gateway"},
			wantCat:    CategoryServerError,
			wantCode:   "bad_gateway",
			wantRetry:  true,
			wantStatus: 502,
		},
		{
			name:       "503 service unavailable",
			err:        &APIError{StatusCode: 503, Status: "503 Service Unavailable"},
			wantCat:    CategoryServerError,
			wantCode:   "service_unavailable",
			wantRetry:  true,
			wantStatus: 503,
		},
		{
			name:       "504 gateway timeout",
			err:        &APIError{StatusCode: 504, Status: "504 Gateway Timeout"},
			wantCat:    CategoryServerError,
			wantCode:   "gateway_timeout",
			wantRetry:  true,
			wantStatus: 504,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ClassifyError(tt.err)
			if info.Category != tt.wantCat {
				t.Errorf("category = %v, want %v", info.Category, tt.wantCat)
			}
			if info.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", info.Code, tt.wantCode)
			}
			if info.Retryable != tt.wantRetry {
				t.Errorf("retryable = %v, want %v", info.Retryable, tt.wantRetry)
			}
			if info.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", info.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestClassifyError_WrappedAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 503, Status: "503 Service Unavailable"}
	wrapped := fmt.Errorf("getting story: %w", apiErr)

	info := ClassifyError(wrapped)
	if info.Category != CategoryServerError {
		t.Errorf("category = %v, want ServerError", info.Category)
	}
	if !info.Retryable {
		t.Error("wrapped 503 should be retryable")
	}
}

func TestClassifyError_NetworkErrors(t *testing.T) {
	t.Run("connection refused", func(t *testing.T) {
		err := &url.Error{
			Op:  "Get",
			URL: "http://localhost:8080/api/v1/stories",
			Err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: errors.New("connection refused"),
			},
		}
		info := ClassifyError(err)
		if info.Category != CategoryNetwork {
			t.Errorf("category = %v, want Network", info.Category)
		}
		if !info.Retryable {
			t.Error("connection refused should be retryable")
		}
	})

	t.Run("DNS error temporary", func(t *testing.T) {
		err := &net.DNSError{
			Name:        "api.example.com",
			IsTemporary: true,
		}
		info := ClassifyError(err)
		if info.Category != CategoryNetwork {
			t.Errorf("category = %v, want Network", info.Category)
		}
		if !info.Retryable {
			t.Error("temporary DNS error should be retryable")
		}
	})

	t.Run("DNS error permanent", func(t *testing.T) {
		err := &net.DNSError{
			Name:        "api.example.com",
			IsTemporary: false,
		}
		info := ClassifyError(err)
		if info.Category != CategoryNetwork {
			t.Errorf("category = %v, want Network", info.Category)
		}
		if info.Retryable {
			t.Error("permanent DNS error should not be retryable")
		}
	})
}

func TestClassifyError_ContextErrors(t *testing.T) {
	info := ClassifyError(context.Canceled)
	if info.Code != "cancelled" {
		t.Errorf("code = %q, want cancelled", info.Code)
	}
	if info.Retryable {
		t.Error("context cancelled should not be retryable")
	}
}

func TestIsRetryableErr(t *testing.T) {
	if IsRetryableErr(&APIError{StatusCode: 404}) {
		t.Error("404 should not be retryable")
	}
	if !IsRetryableErr(&APIError{StatusCode: 503}) {
		t.Error("503 should be retryable")
	}
	if IsRetryableErr(nil) {
		t.Error("nil should not be retryable")
	}
}

func TestUserFriendlyError(t *testing.T) {
	msg := UserFriendlyError(&APIError{StatusCode: 401})
	if msg != "Authentication failed. Check your API token." {
		t.Errorf("unexpected message: %s", msg)
	}

	if UserFriendlyError(nil) != "" {
		t.Error("expected empty for nil")
	}
}

func TestParseServerError(t *testing.T) {
	tests := []struct {
		body string
		want string
	}{
		{`{"error":"bad_request","message":"invalid story_id"}`, "invalid story_id"},
		{`{"error":"internal_error","message":"An unexpected error occurred"}`, ""},
		{`not json`, ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := parseServerError(tt.body)
		if got != tt.want {
			t.Errorf("parseServerError(%q) = %q, want %q", tt.body, got, tt.want)
		}
	}
}

func TestServerErrorMessageInClassification(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Status:     "400 Bad Request",
		Body:       `{"error":"bad_request","message":"story_id must be a UUID"}`,
	}
	info := ClassifyError(err)
	if info.Message != "story_id must be a UUID" {
		t.Errorf("expected server message, got: %s", info.Message)
	}
}

func TestErrorCategoryString(t *testing.T) {
	tests := []struct {
		cat  ErrorCategory
		want string
	}{
		{CategoryNetwork, "network"},
		{CategoryAuth, "auth"},
		{CategoryNotFound, "not_found"},
		{CategoryConflict, "conflict"},
		{CategoryValidation, "validation"},
		{CategoryRateLimit, "rate_limit"},
		{CategoryServerError, "server_error"},
		{CategoryUnknown, "unknown"},
	}
	for _, tt := range tests {
		if got := tt.cat.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.cat, got, tt.want)
		}
	}
}
