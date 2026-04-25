package main

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// requireArg gets a required string argument from the request.
func requireArg(req mcp.CallToolRequest, name string) (string, error) {
	args := req.GetArguments()
	v, ok := args[name]
	if !ok {
		return "", fmt.Errorf("missing required argument: %s", name)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("argument %s must be a string", name)
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

// jsonResult marshals v as indented JSON and wraps it in a tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(data)), nil
}

// authParams returns tool options for optional credential overrides.
func authParams() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("url", mcp.Description("API base URL override (default: from PROSEFORGE_URL)")),
		mcp.WithString("token", mcp.Description("API token override (default: from PROSEFORGE_TOKEN)")),
	}
}

// tool creates a new tool with auth params appended.
func tool(name string, opts ...mcp.ToolOption) mcp.Tool {
	return mcp.NewTool(name, append(opts, authParams()...)...)
}
