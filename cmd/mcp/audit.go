package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/config"
)

// auditLogger writes tool call records to an audit log file.
type auditLogger struct {
	logger *slog.Logger
	// defaultURL is the server's fallback backend (PROSEFORGE_URL). It mirrors
	// clientResolver's fallback so the audit log can record the *effective*
	// backend a call hit even when no per-call url override was passed.
	defaultURL string
}

// newAuditLogger creates an audit logger writing to the given file path.
// defaultURL is the server's fallback backend, used to record the effective
// target of calls that omit a url override.
// Returns nil if the file can't be opened (audit is best-effort).
func newAuditLogger(logFile, defaultURL string) *auditLogger {
	if logFile == "" {
		return nil
	}

	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil
	}

	logger := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &auditLogger{logger: logger, defaultURL: defaultURL}
}

// auditMiddleware returns an MCP tool handler middleware that logs every call.
func auditMiddleware(a *auditLogger) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			start := time.Now()

			// Refuse arguments we never declared rather than dropping them. A
			// dropped argument still returns 200 and answers a different
			// question — room_read with `since_id` instead of `since` returned
			// the whole room instead of the delta (#279).
			if unknown := unknownArgs(req.Params.Name, req.GetArguments()); len(unknown) > 0 {
				msg := fmt.Sprintf("unknown argument(s) for %s: %s", req.Params.Name, strings.Join(unknown, ", "))
				if s := suggestArg(req.Params.Name, unknown[0]); s != "" {
					msg += fmt.Sprintf(" — did you mean %q?", s)
				}
				msg += ". The call was refused rather than run without them, because ignoring an " +
					"argument silently changes what the call returns."
				result := mcp.NewToolResultError(msg)
				a.logToolCall(req, "tool_error", time.Since(start), nil)
				return result, nil
			}

			result, err := next(ctx, req)
			duration := time.Since(start)

			status := "ok"
			if err != nil {
				status = "error"
			} else if result != nil && result.IsError {
				status = "tool_error"
			}

			a.logToolCall(req, status, duration, err)
			return result, err
		}
	}
}

// logToolCall records a tool invocation.
func (a *auditLogger) logToolCall(req mcp.CallToolRequest, result string, duration time.Duration, err error) {
	if a == nil {
		return
	}

	args := req.GetArguments()
	// Redact tokens from audit log. An environment reference is kept verbatim:
	// a variable name is not a secret, and recording which one was used is the
	// only way the log can say *which identity* made the call — "***" cannot.
	safeArgs := make(map[string]any, len(args))
	for k, v := range args {
		if k != "token" {
			safeArgs[k] = v
			continue
		}
		if s, ok := v.(string); ok {
			if _, isRef := config.EnvRefName(s); isRef {
				safeArgs[k] = s
				continue
			}
		}
		safeArgs[k] = "***"
	}

	// Which channel supplied the identity. A call that quietly fell back to the
	// server's default acts as whatever account that entry holds — which, when
	// one server serves several agents, is usually not the caller. Recording it
	// makes that greppable instead of indistinguishable from a correct call.
	identitySource := "default"
	if s, ok := args["token"].(string); ok && s != "" {
		identitySource = "token_argument"
	} else if s, ok := args["credentials_file"].(string); ok && s != "" {
		identitySource = "credentials_file"
	}

	// Record the effective backend the call hit. A url override wins; otherwise
	// the server default. backend_source flags the silent-default case — the one
	// that makes a write land on the wrong backend without anyone noticing.
	effectiveURL := a.defaultURL
	backendSource := "default"
	if u, ok := args["url"].(string); ok && u != "" {
		// Resolve an env reference before recording. Logging the raw
		// "${PROSEFORGE_URL}" would name a placeholder instead of the backend
		// the call actually hit, which is exactly the blindness #246 removed.
		if resolved, err := config.ExpandEnvRef(u, "url"); err == nil {
			effectiveURL = resolved
		} else {
			effectiveURL = u
		}
		backendSource = "override"
	}

	if err != nil {
		a.logger.Info("tool_call",
			"tool", req.Params.Name,
			"args", fmt.Sprintf("%v", safeArgs),
			"effective_url", effectiveURL,
			"backend_source", backendSource,
			"identity_source", identitySource,
			"result", "error",
			"error", err.Error(),
			"duration_ms", duration.Milliseconds(),
		)
	} else {
		a.logger.Info("tool_call",
			"tool", req.Params.Name,
			"args", fmt.Sprintf("%v", safeArgs),
			"effective_url", effectiveURL,
			"backend_source", backendSource,
			"identity_source", identitySource,
			"result", result,
			"duration_ms", duration.Milliseconds(),
		)
	}
}
