package main

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

func TestLegacyInvitationToolsAreNotRegistered(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerAllTools(s, newClientResolver("http://example.test", "token", nil, nil), nil)

	for _, name := range []string{"invitation_send", "invitation_list", "invitation_mine", "invitation_respond"} {
		if tool := s.GetTool(name); tool != nil {
			t.Errorf("legacy invitation tool %q is still registered", name)
		}
	}
}
