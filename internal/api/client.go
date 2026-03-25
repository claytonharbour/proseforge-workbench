// Package api provides a hand-written wrapper around the generated ProseForge
// API client. It adds authentication, error handling, and ergonomic methods
// for the endpoints the workbench uses.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Client is a ProseForge API client authenticated with a single API key.
// The workbench creates two of these: one for the author, one for the reviewer.
type Client struct {
	raw     *gen.ClientWithResponses
	apiKey  string
	baseURL string
}

// BaseURL returns the API base URL this client is configured to talk to.
func (c *Client) BaseURL() string { return c.baseURL }

// Option configures a Client during construction.
type Option func(*clientOptions)

type clientOptions struct {
	logger     *slog.Logger
	maxRetries int
}

// WithLogger sets the logger for HTTP request/response logging.
// If not provided, the client uses slog.Default().
func WithLogger(logger *slog.Logger) Option {
	return func(o *clientOptions) {
		o.logger = logger
	}
}

// WithRetry sets the maximum number of retries for transient failures.
// Default is 3. Set to 0 to disable retries.
func WithRetry(n int) Option {
	return func(o *clientOptions) {
		o.maxRetries = n
	}
}

// New creates a new ProseForge API client for the given base URL and API key.
func New(baseURL, apiKey string, opts ...Option) (*Client, error) {
	o := &clientOptions{
		logger:     slog.Default(),
		maxRetries: 3,
	}
	for _, opt := range opts {
		opt(o)
	}

	authEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		return nil
	}

	// Transport chain: retryTransport → loggingTransport → http.DefaultTransport
	logging := &loggingTransport{
		inner:  http.DefaultTransport,
		logger: o.logger,
	}
	var transport http.RoundTripper = logging
	if o.maxRetries > 0 {
		transport = newRetryTransport(logging, o.maxRetries, o.logger)
	}

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	raw, err := gen.NewClientWithResponses(
		baseURL+"/api/v1",
		gen.WithHTTPClient(httpClient),
		gen.WithRequestEditorFn(authEditor),
	)
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}

	return &Client{raw: raw, apiKey: apiKey, baseURL: baseURL}, nil
}

// APIError is returned when the server responds with a non-success status.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("API %s: %s", e.Status, e.Body)
	}
	return fmt.Sprintf("API %s", e.Status)
}

// checkResponse reads the response body and returns an APIError if the status
// code is not in the 2xx range.
func checkResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}

	return nil, &APIError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(body),
	}
}

// withInclude returns a RequestEditorFn that appends ?include=val1,val2 to the request URL.
func withInclude(vals ...string) gen.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("include", strings.Join(vals, ","))
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// decode is a helper that checks the response and JSON-decodes into dst.
func decode(resp *http.Response, dst any) error {
	body, err := checkResponse(resp)
	if err != nil {
		return err
	}
	if dst == nil {
		return nil
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
