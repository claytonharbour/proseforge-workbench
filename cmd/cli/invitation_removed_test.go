package main

import "testing"

func TestLegacyInvitationCommandIsRemoved(t *testing.T) {
	for _, command := range rootCmd.Commands() {
		if command.Name() == "invitation" {
			t.Fatal("legacy invitation command is still registered")
		}
	}
}
