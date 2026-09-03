package main

import (
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
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

// reactionLine is the only place a reader learns a message was acknowledged, so its
// two jobs — the count, and whether the CALLER is among the reactors — are worth
// pinning. The `*` marker is what turns "3 people saw it" into "I already replied".
func TestReactionLineMarksTheCallersOwnReaction(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []api.MessageReaction
		want string
	}{
		{"no reactions renders nothing at all", nil, ""},
		{"empty slice is not an empty line", []api.MessageReaction{}, ""},
		{
			"a reaction the caller did not make carries no marker",
			[]api.MessageReaction{{Emoji: "👀", Count: 2, Mine: false}},
			"   👀2",
		},
		{
			"the caller's own reaction is marked",
			[]api.MessageReaction{{Emoji: "👀", Count: 2, Mine: true}},
			"   👀*2",
		},
		{
			"several emoji keep their own counts and markers",
			[]api.MessageReaction{
				{Emoji: "✅", Count: 1, Mine: true},
				{Emoji: "👀", Count: 3, Mine: false},
			},
			"   ✅*1  👀3",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := reactionLine(tc.in); got != tc.want {
				t.Errorf("reactionLine() = %q, want %q", got, tc.want)
			}
		})
	}
}
