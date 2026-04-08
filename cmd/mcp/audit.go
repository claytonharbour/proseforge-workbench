package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// auditLogger writes tool call records to an audit log file.
type auditLogger struct {
	logger *slog.Logger
}

// newAuditLogger creates an audit logger writing to the given file path.
// Returns nil if the file can't be opened (audit is best-effort).
func newAuditLogger(logFile string) *auditLogger {
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

	return &auditLogger{logger: logger}
}

// auditMiddleware returns an MCP tool handler middleware that logs every call.
func auditMiddleware(a *auditLogger) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			start := time.Now()
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
	// Redact tokens from audit log
	safeArgs := make(map[string]any, len(args))
	for k, v := range args {
		if k == "token" {
			safeArgs[k] = "***"
		} else {
			safeArgs[k] = v
		}
	}

	if err != nil {
		a.logger.Info("tool_call",
			"tool", req.Params.Name,
			"args", fmt.Sprintf("%v", safeArgs),
			"result", "error",
			"error", err.Error(),
			"duration_ms", duration.Milliseconds(),
		)
	} else {
		a.logger.Info("tool_call",
			"tool", req.Params.Name,
			"args", fmt.Sprintf("%v", safeArgs),
			"result", result,
			"duration_ms", duration.Milliseconds(),
		)
	}
}
