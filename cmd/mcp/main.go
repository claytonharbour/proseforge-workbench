package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// loadWorkerConfig reads key=value pairs from ~/.proseforge-workbench/config.
// Returns a map of the values. Missing file is not an error.
func loadWorkerConfig() map[string]string {
	cfg := make(map[string]string)
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	path := filepath.Join(home, ".proseforge-workbench", "config")
	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			cfg[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return cfg
}

// parseLogLevel maps a string to slog.Level. Defaults to info.
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// setupLogger creates a logger from worker config. Falls back to stderr.
// Returns the logger and a cleanup function to close the log file.
func setupLogger(workerCfg map[string]string) (*slog.Logger, func()) {
	// Level: worker config > env var > default (info)
	levelStr := workerCfg["PROSEFORGE_WORKER_LOG_LEVEL"]
	if levelStr == "" {
		levelStr = os.Getenv("PROSEFORGE_LOG_LEVEL")
	}
	level := parseLogLevel(levelStr)

	// Output: worker config log file > stderr
	logFile := workerCfg["PROSEFORGE_WORKER_LOG_FILE"]
	var writer io.Writer = os.Stderr
	var cleanup func()

	if logFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0o755); err == nil {
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err == nil {
				writer = f
				cleanup = func() { f.Close() }
			}
		}
	}

	if cleanup == nil {
		cleanup = func() {}
	}

	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: level,
	}))

	return logger, cleanup
}

// clientResolver manages API clients, caching by url+token pair.
type clientResolver struct {
	defaultURL   string
	defaultToken string
	logger       *slog.Logger
	audit        *auditLogger
	mu           sync.Mutex
	cache        map[string]*api.Client
}

func newClientResolver(url, token string, logger *slog.Logger, audit *auditLogger) *clientResolver {
	return &clientResolver{
		defaultURL:   url,
		defaultToken: token,
		logger:       logger,
		audit:        audit,
		cache:        make(map[string]*api.Client),
	}
}

// resolve returns an API client for the given request. If url/token overrides
// are provided in the tool arguments, a separate client is created (and cached).
func (r *clientResolver) resolve(req mcp.CallToolRequest) (*api.Client, error) {
	url := optionalArg(req, "url")
	token := optionalArg(req, "token")

	if url == "" {
		url = r.defaultURL
	}
	if token == "" {
		token = r.defaultToken
	}

	if url == "" {
		return nil, fmt.Errorf("API URL required: set PROSEFORGE_URL or pass url parameter")
	}
	if token == "" {
		return nil, fmt.Errorf("API token required: set PROSEFORGE_TOKEN or pass token parameter")
	}

	key := url + "|" + token

	r.mu.Lock()
	defer r.mu.Unlock()

	if c, ok := r.cache[key]; ok {
		return c, nil
	}

	c, err := api.New(url, token, api.WithLogger(r.logger))
	if err != nil {
		return nil, err
	}
	r.cache[key] = c
	return c, nil
}

func run() error {
	workerCfg := loadWorkerConfig()
	logger, cleanup := setupLogger(workerCfg)
	defer cleanup()

	// Audit log — structured JSON of every tool call
	auditFile := workerCfg["PROSEFORGE_WORKER_AUDIT_FILE"]
	if auditFile == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			auditFile = filepath.Join(home, ".proseforge-workbench", "logs", "audit.log")
		}
	}
	audit := newAuditLogger(auditFile)

	logger.Info("proseforge-workbench MCP server starting",
		"version", Version,
		"log_file", workerCfg["PROSEFORGE_WORKER_LOG_FILE"],
		"audit_file", auditFile,
	)

	resolver := newClientResolver(
		os.Getenv("PROSEFORGE_URL"),
		os.Getenv("PROSEFORGE_TOKEN"),
		logger,
		audit,
	)

	s := server.NewMCPServer(
		"proseforge-workbench",
		Version,
		server.WithToolCapabilities(true),
		server.WithToolHandlerMiddleware(auditMiddleware(audit)),
	)

	registerAllTools(s, resolver)

	return server.ServeStdio(s)
}

// registerAllTools registers all MCP tools from the domain-specific files.
func registerAllTools(s *server.MCPServer, r *clientResolver) {
	registerStoryTools(s, r)
	registerReviewTools(s, r)
	registerFeedbackTools(s, r)
	registerReviewerTools(s, r)
}
