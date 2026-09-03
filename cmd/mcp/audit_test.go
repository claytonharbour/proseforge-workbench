package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestAuditEffectiveBackend covers the #246 observability fix: every audit line
// must record the effective backend a call hit and whether it came from a per-call
// override or the server default — so a silently-defaulted write is visible.
func TestAuditEffectiveBackend(t *testing.T) {
	newReq := func(args map[string]any) mcp.CallToolRequest {
		req := mcp.CallToolRequest{}
		req.Params.Name = "room_send"
		req.Params.Arguments = args
		return req
	}

	logLine := func(t *testing.T, defaultURL string, args map[string]any) map[string]any {
		t.Helper()
		var buf bytes.Buffer
		a := &auditLogger{
			logger:     slog.New(slog.NewJSONHandler(&buf, nil)),
			defaultURL: defaultURL,
		}
		a.logToolCall(newReq(args), "ok", time.Millisecond, nil)
		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("audit line not JSON: %v (%q)", err, buf.String())
		}
		return rec
	}

	t.Run("url override → effective_url is the override, source=override", func(t *testing.T) {
		rec := logLine(t, "http://example.com", map[string]any{
			"url":     "https://app.proseforge.ai",
			"token":   "secret-token",
			"content": "hi",
		})
		if rec["effective_url"] != "https://app.proseforge.ai" {
			t.Errorf("effective_url = %v, want the override", rec["effective_url"])
		}
		if rec["backend_source"] != "override" {
			t.Errorf("backend_source = %v, want override", rec["backend_source"])
		}
		// Token must stay redacted in the args blob.
		if args, _ := rec["args"].(string); strings.Contains(args, "secret-token") {
			t.Errorf("token leaked into audit args: %s", args)
		}
	})

	t.Run("no url → effective_url is the server default, source=default", func(t *testing.T) {
		rec := logLine(t, "http://example.com", map[string]any{"content": "hi"})
		if rec["effective_url"] != "http://example.com" {
			t.Errorf("effective_url = %v, want the default", rec["effective_url"])
		}
		if rec["backend_source"] != "default" {
			t.Errorf("backend_source = %v, want default", rec["backend_source"])
		}
	})

	t.Run("empty url arg is treated as default, not override", func(t *testing.T) {
		rec := logLine(t, "http://example.com", map[string]any{"url": "", "content": "hi"})
		if rec["backend_source"] != "default" {
			t.Errorf("backend_source = %v, want default for empty url", rec["backend_source"])
		}
	})
}

// TestAuditTokenHandling covers the #250 change: a literal credential stays
// redacted, but an environment reference is recorded verbatim. The reference is
// a variable name rather than a secret, and it is the only thing that lets the
// log say *which identity* made a call — "***" is identical for every bench.
func TestAuditTokenHandling(t *testing.T) {
	logLine := func(t *testing.T, token string) string {
		t.Helper()
		var buf bytes.Buffer
		a := &auditLogger{
			logger:     slog.New(slog.NewJSONHandler(&buf, nil)),
			defaultURL: "http://example.com",
		}
		req := mcp.CallToolRequest{}
		req.Params.Name = "room_send"
		req.Params.Arguments = map[string]any{"token": token, "content": "hi"}
		a.logToolCall(req, "ok", time.Millisecond, nil)

		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("audit line not JSON: %v (%q)", err, buf.String())
		}
		args, _ := rec["args"].(string)
		return args
	}

	t.Run("literal token is redacted", func(t *testing.T) {
		literal := "pf_EXAMPLE_LITERAL_NOT_A_REAL_KEY"
		args := logLine(t, literal)
		if strings.Contains(args, literal) {
			t.Errorf("literal token leaked into audit args: %s", args)
		}
		if !strings.Contains(args, "***") {
			t.Errorf("literal token not redacted: %s", args)
		}
	})

	t.Run("env reference is preserved for attribution", func(t *testing.T) {
		args := logLine(t, "${SMILEY_PROD_TOKEN}")
		if !strings.Contains(args, "SMILEY_PROD_TOKEN") {
			t.Errorf("env reference not recorded, attribution lost: %s", args)
		}
	})
}

// TestAuditEffectiveURLResolvesEnvRef guards the #246 guarantee against the
// #250 env-reference change: effective_url must name the backend the call
// actually hit, never the unresolved "${VAR}" placeholder.
func TestAuditEffectiveURLResolvesEnvRef(t *testing.T) {
	t.Setenv("PFW_TEST_URL", "https://example.com")

	var buf bytes.Buffer
	a := &auditLogger{
		logger:     slog.New(slog.NewJSONHandler(&buf, nil)),
		defaultURL: "https://default.example.com",
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = "room_read"
	req.Params.Arguments = map[string]any{"url": "${PFW_TEST_URL}"}
	a.logToolCall(req, "ok", time.Millisecond, nil)

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("audit line not JSON: %v", err)
	}
	if rec["effective_url"] != "https://example.com" {
		t.Errorf("effective_url = %v, want the resolved backend", rec["effective_url"])
	}
	if rec["backend_source"] != "override" {
		t.Errorf("backend_source = %v, want override", rec["backend_source"])
	}
}

// TestAuditIdentitySource covers the #250 attribution field: the log must say
// which channel supplied the identity, so a call that silently fell back to the
// server's default account is greppable rather than indistinguishable from a
// correct one.
func TestAuditIdentitySource(t *testing.T) {
	field := func(t *testing.T, args map[string]any) string {
		t.Helper()
		var buf bytes.Buffer
		a := &auditLogger{
			logger:     slog.New(slog.NewJSONHandler(&buf, nil)),
			defaultURL: "https://example.com",
		}
		req := mcp.CallToolRequest{}
		req.Params.Name = "room_send"
		req.Params.Arguments = args
		a.logToolCall(req, "ok", time.Millisecond, nil)

		var rec map[string]any
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("audit line not JSON: %v", err)
		}
		s, _ := rec["identity_source"].(string)
		return s
	}

	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"no credential falls back to the server default", map[string]any{"content": "hi"}, "default"},
		{"explicit token argument", map[string]any{"token": "pf_EXAMPLE_NOT_A_REAL_KEY"}, "token_argument"},
		{"credentials file", map[string]any{"credentials_file": "/tmp/creds.env"}, "credentials_file"},
		{"empty strings are not a credential", map[string]any{"token": "", "credentials_file": ""}, "default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := field(t, tc.args); got != tc.want {
				t.Errorf("identity_source = %q, want %q", got, tc.want)
			}
		})
	}
}
