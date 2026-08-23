package main

import (
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/spf13/cobra"
)

func TestRoomIdentityFlagsAreOptional(t *testing.T) {
	send := newRoomSendCmd().Flag("agent")
	if send == nil {
		t.Fatal("room send has no agent flag")
	}
	if _, required := send.Annotations[cobra.BashCompOneRequiredFlag]; required {
		t.Error("room send agent flag must be optional")
	}

	cursor := newRoomCursorCmd()
	for _, command := range cursor.Commands() {
		flag := command.Flag("handle")
		if flag == nil {
			t.Fatalf("room cursor %s has no handle flag", command.Name())
		}
		if _, required := flag.Annotations[cobra.BashCompOneRequiredFlag]; required {
			t.Errorf("room cursor %s handle flag must be optional", command.Name())
		}
	}
}

func TestRoomListMemberSummaryIncludesSilentMembers(t *testing.T) {
	rooms := gen.HandlersRoomListEntry{
		Members: &[]gen.HandlersRoomMember{
			{Id: stringPtr("owner-1"), Name: stringPtr("Owner")},
			{Id: stringPtr("silent-1"), Name: stringPtr("Silent Collaborator")},
		},
	}
	if got := roomMemberSummary(rooms); got != "2: Owner, Silent Collaborator" {
		t.Fatalf("summary = %q", got)
	}
}

func stringPtr(value string) *string { return &value }
