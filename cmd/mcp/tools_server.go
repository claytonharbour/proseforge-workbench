package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerServerTools exposes version facts about the running Workbench process
// and its configured ProseForge backend.
func registerServerTools(s *server.MCPServer, r *clientResolver, workerCfg map[string]string) {
	s.AddTool(
		tool("server_version",
			mcp.WithDescription("Report version metadata for the running Workbench MCP process and its ProseForge backend. Returns the MCP build and public backend API build time/commit. If PROSEFORGE_WORKER_VERSION is configured, it is included as worker metadata. The backend version check does not require admin credentials."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			backend, err := client.GetPublicVersionInfo(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			result := map[string]any{
				"mcp": map[string]string{
					"name":    "proseforge-workbench",
					"version": Version,
				},
				"server": backend,
			}
			if workerVersion := workerCfg["PROSEFORGE_WORKER_VERSION"]; workerVersion != "" {
				result["worker"] = map[string]string{"version": workerVersion}
			}
			return jsonResultWithBackend(result, client)
		},
	)
}
