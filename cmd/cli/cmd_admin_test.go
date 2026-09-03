package main

import "testing"

func TestAdminDemoBannerCommands(t *testing.T) {
	cmd := newAdminDemoBannerCmd()
	for _, name := range []string{"get", "update", "enable", "suppress"} {
		found := false
		for _, child := range cmd.Commands() {
			if child.Name() == name {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing admin demo-banner %s command", name)
		}
	}
	suppress, _, err := cmd.Find([]string{"suppress"})
	if err != nil {
		t.Fatal(err)
	}
	if suppress.Flag("expires-at") == nil {
		t.Fatal("suppress missing --expires-at")
	}
}
