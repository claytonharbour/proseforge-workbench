package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAuthTools(s *server.MCPServer, r *clientResolver) {
	// whoami
	s.AddTool(
		tool("whoami",
			mcp.WithDescription("Report which account the workbench is authenticated as AND what it is entitled to — email, id, admin flag, plus the resolved tier and its feature list under \"tier\". Check tier.tierFeatures before attempting a gated action (e.g. \"narration_forge\") instead of calling it and interpreting a 403. tier.tierName is the EFFECTIVE tier with any admin override applied; tier.tierId is NOT resolved and will disagree. If the tier could not be read, \"tier\" is null and \"tierError\" says why — identity is still returned, so this remains usable while diagnosing a 403. With one server shared by several agents, a call that omits credentials_file acts as the server's default account, not you."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{ReadOnlyHint: mcp.ToBoolPtr(true)}),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			data, err := client.WhoAmI(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)
	// terms_accept — forge/proseforge#1025 tier 2.
	//
	// Google Play's UGC policy asks for terms acceptance BEFORE a user can create
	// or upload content, by a process that "cannot be skipped by the user". A gate
	// nobody can satisfy is an outage, so every surface that can post needs a way
	// to accept — and MCP is a surface that can post (room_send, section_write).
	//
	// ⚑ This tool exists so the gate can be turned on WITHOUT stranding agents.
	// It is deliberately shipped ahead of the gate, not with it.
	s.AddTool(
		tool("terms_accept",
			mcp.WithDescription("Accept the current ProseForge terms of service for the authenticated account. Records consent server-side; the server chooses the version, you cannot name one. Idempotent — re-accepting the same version does not move the recorded date, so it is safe to call without checking first. Use whoami to confirm WHICH account you are accepting for: with one server shared by several agents, a call that omits credentials_file accepts as the server's default account, not you."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := r.resolve(req)
			if err != nil {
				return toolError(err), nil
			}
			data, err := client.AcceptTerms(ctx)
			if err != nil {
				return toolError(err, client), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)
}
