package main

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// toolErrorResponse is the structured JSON returned in MCP tool errors.
// AI consumers can parse this to decide whether to retry, fix input, or escalate.
type toolErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable"`
	StatusCode int    `json:"status_code,omitempty"`
}

// toolError classifies an error and returns a structured MCP tool error result.
// Drop-in replacement for mcp.NewToolResultError(err.Error()).
// If a Client is provided, the resolved base URL is appended to the error
// message so agents can see which environment was targeted.
func toolError(err error, clients ...*api.Client) *mcp.CallToolResult {
	var baseURL string
	if len(clients) > 0 && clients[0] != nil {
		baseURL = clients[0].BaseURL()
	}
	return toolErrorWithURL(err, baseURL)
}

// toolErrorWithURL is like toolError but takes an explicit base URL string.
func toolErrorWithURL(err error, baseURL string) *mcp.CallToolResult {
	info := api.ClassifyError(err)
	if info == nil {
		return mcp.NewToolResultError("unknown error")
	}

	msg := info.Message
	if baseURL != "" {
		msg = msg + " (" + baseURL + ")"
	}

	resp := toolErrorResponse{
		Error:      info.Code,
		Message:    msg,
		Retryable:  info.Retryable,
		StatusCode: info.StatusCode,
	}

	data, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		return mcp.NewToolResultError(msg)
	}

	return mcp.NewToolResultError(string(data))
}
