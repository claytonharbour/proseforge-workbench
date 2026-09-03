package main

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

func TestDemoBannerToolsRegistered(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	registerAdminTools(s, newClientResolver("http://example.test", "token", nil, nil))
	for _, name := range []string{"demo_banner_get", "demo_banner_update"} {
		if s.GetTool(name) == nil {
			t.Fatalf("missing %s", name)
		}
	}
}
