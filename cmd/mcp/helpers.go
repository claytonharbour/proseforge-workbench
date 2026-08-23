package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/mark3labs/mcp-go/mcp"
)

// requireArg gets a required string argument from the request.
func requireArg(req mcp.CallToolRequest, name string) (string, error) {
	args := req.GetArguments()
	v, ok := args[name]
	if !ok {
		return "", invalidInputErrorf("missing required argument: %s", name)
	}
	s, ok := v.(string)
	if !ok {
		return "", invalidInputErrorf("argument %s must be a string", name)
	}
	return s, nil
}

// optionalArg gets an optional string argument from the request.
func optionalArg(req mcp.CallToolRequest, name string) string {
	args := req.GetArguments()
	v, ok := args[name]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// optionalIntArg gets an optional int argument (JSON numbers are float64).
func optionalIntArg(req mcp.CallToolRequest, name string, defaultVal int) int {
	args := req.GetArguments()
	v, ok := args[name]
	if !ok {
		return defaultVal
	}
	f, ok := v.(float64)
	if !ok {
		return defaultVal
	}
	return int(f)
}

// optionalBoolArg gets an optional boolean argument.
func optionalBoolArg(req mcp.CallToolRequest, name string) bool {
	args := req.GetArguments()
	v, ok := args[name]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// optionalBoolPtrArg gets an optional boolean argument as a tri-state
// pointer: nil when the caller omitted it (or passed a non-bool), otherwise
// the provided value. Use when "not sent" must be distinguishable from false.
func optionalBoolPtrArg(req mcp.CallToolRequest, name string) *bool {
	args := req.GetArguments()
	v, ok := args[name]
	if !ok {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

// jsonResult marshals v as indented JSON and wraps it in a tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(data)), nil
}

// jsonResultWithBackend marshals v like jsonResult but adds a top-level
// "backend" field naming the host the call was routed to. The error path already
// surfaces the backend (toolErrorWithURL); this does the same on success so a
// write that lands on the wrong backend is visible in the response, not silent.
// Additive: existing fields are preserved. If v doesn't marshal to a JSON object
// (e.g. an array or scalar), it falls back to a plain jsonResult.
func jsonResultWithBackend(v any, client *api.Client) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return jsonResult(v)
	}
	m["backend"] = client.BaseURL()
	return jsonResult(m)
}

// audiobookMetadata projects the audiobook download fields out of a raw
// narration-status payload into a small envelope. It deliberately avoids the
// /narration/audiobook endpoint, which streams the raw M4B binary (marshaling
// those bytes as JSON always fails on the first null byte — see #244). The
// download URL and stats already live in the status response.
func audiobookMetadata(raw json.RawMessage) (map[string]any, error) {
	var st struct {
		NarrationID        string `json:"narration_id"`
		Status             string `json:"status"`
		Voice              string `json:"voice"`
		AudiobookURL       string `json:"audiobook_url"`
		TotalDurationMs    *int64 `json:"total_duration_ms"`
		TotalFileSizeBytes *int64 `json:"total_file_size_bytes"`
		CompletedSections  *int   `json:"completed_sections"`
		TotalSections      *int   `json:"total_sections"`
	}
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, fmt.Errorf("parse narration status: %w", err)
	}
	out := map[string]any{
		"ready":                 st.Status == "complete" && st.AudiobookURL != "",
		"status":                st.Status,
		"audiobook_url":         st.AudiobookURL,
		"total_duration_ms":     st.TotalDurationMs,
		"total_file_size_bytes": st.TotalFileSizeBytes,
		"voice":                 st.Voice,
		"narration_id":          st.NarrationID,
		"completed_sections":    st.CompletedSections,
		"total_sections":        st.TotalSections,
	}
	if st.Status != "complete" || st.AudiobookURL == "" {
		out["note"] = "audiobook not yet assembled — poll narration_status until status is 'complete'"
	}
	return out, nil
}

// authParams returns tool options for optional credential overrides.
func authParams() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("url", mcp.Description("API base URL override (default: from PROSEFORGE_URL). Accepts an environment reference like ${PROSEFORGE_URL}, resolved from the server's own environment.")),
		mcp.WithString("token", mcp.Description("API token override (default: from PROSEFORGE_TOKEN). Accepts an environment reference like ${PROSEFORGE_TOKEN}, resolved from the server's own environment — note that one server may be shared by several agents, so a shared variable yields the shared identity. To act as yourself, use credentials_file.")),
		mcp.WithString("credentials_file", mcp.Description("Path to your credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read on each call, so a rotated key applies immediately with no restart. This is how you act as your own account when one server serves several agents — omit it and the call uses the server's default identity, which may not be you. Check with whoami.")),
	}
}

// declaredArgs records the argument names each tool accepts, so a call carrying
// an argument we never declared can be refused instead of silently dropped.
//
// Written once during registration, read-only afterwards.
var declaredArgs = map[string]map[string]bool{}

// tool creates a new tool with auth params appended.
func tool(name string, opts ...mcp.ToolOption) mcp.Tool {
	t := mcp.NewTool(name, append(opts, authParams()...)...)

	names := make(map[string]bool, len(t.InputSchema.Properties))
	for prop := range t.InputSchema.Properties {
		names[prop] = true
	}
	declaredArgs[name] = names

	return t
}

// unknownArgs returns any arguments the caller sent that the tool does not
// declare, sorted for a stable message.
//
// An undeclared argument used to be dropped in silence, which is worse than an
// error: the call still succeeds and answers a different question. Passing
// `since_id` to room_read instead of `since` returned the entire room — 488
// messages and 1.1MB — with a 200 and no indication anything was ignored
// (forge/proseforge-workbench#279). A wrong answer that looks bigger and more
// complete than the right one is the hardest kind to notice.
func unknownArgs(toolName string, args map[string]any) []string {
	declared, ok := declaredArgs[toolName]
	if !ok {
		return nil // tool registered without the helper; nothing to check against
	}
	var unknown []string
	for name := range args {
		if !declared[name] {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// suggestArg finds the closest declared argument name, so the error can point at
// the likely intent rather than only naming the mistake.
func suggestArg(toolName, given string) string {
	best, bestScore := "", 0
	for candidate := range declaredArgs[toolName] {
		score := commonAffix(candidate, given)
		// Require a real overlap; two unrelated names share a letter by chance.
		if score > bestScore && score >= len(given)/2 {
			best, bestScore = candidate, score
		}
	}
	return best
}

// commonAffix scores similarity as the longer of the shared prefix and suffix.
// Enough to catch the cases that actually happen — since/since_id,
// agent_handle/agentHandle, section/section_id.
func commonAffix(a, b string) int {
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix &&
		a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	if prefix > suffix {
		return prefix
	}
	return suffix
}

// optionalStringSliceArg gets an optional array-of-strings argument. JSON
// decodes arrays as []any, so each element is asserted individually; a
// non-string element is skipped rather than aborting, because the tool's own
// error for the resulting path is clearer than a type complaint here.
func optionalStringSliceArg(req mcp.CallToolRequest, name string) []string {
	raw, ok := req.GetArguments()[name]
	if !ok {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}
