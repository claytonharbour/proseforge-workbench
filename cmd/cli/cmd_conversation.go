package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Conversation commands (proseforge-workbench#329). Thin cobra wrappers over
// conversation.Service — the same service the MCP conversation_* tools use.
//
// ⚑ CLOSES A PARITY GAP, NOT A FEATURE GAP. `pfw room send/read --type conversation`
// already worked, so a human at a terminal could talk in a conversation they had no way
// to create. Every other row in the room reference table has both an MCP tool and a
// command; these four had a blank.
//
// ⛔ There is deliberately no `conversation send` or `read`. Messaging is `pfw room
// send/read --type conversation` — the same cursor, filtering and archive semantics as any
// other room. A parallel pair would have to re-implement all of it.
func newConversationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "conversation",
		Aliases: []string{"convo"},
		Short:   "Standalone rooms with no story or series attached (proseforge#821)",
		Long: "Create and manage conversations — rooms you create empty and add specific\n" +
			"people to, rather than inheriting whoever can see a story.\n\n" +
			"Talk in one with: pfw room send <id> --type conversation --content \"...\"",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "Conversations you started or were added to",
			Args:  cobra.NoArgs,
			RunE:  runConversationList,
		},
		newConversationCreateCmd(),
		newConversationAddCmd(),
		newConversationRemoveCmd(),
		&cobra.Command{
			Use:   "leave <conversation-id>",
			Short: "Leave a conversation — also how you refuse an invitation",
			Long: "Leaving and refusing an invitation are the same act at different moments,\n" +
				"so there is one command for both.\n\n" +
				"The creator cannot leave: they hold the conversation through ownership\n" +
				"rather than a grant, so leaving would revoke nothing. Archive the room instead.",
			Args: cobra.ExactArgs(1),
			RunE: runConversationLeave,
		},
	)
	return cmd
}

func newConversationCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Start a new, empty conversation",
		Long: "Creates the conversation with no members — add them afterwards with\n" +
			"`pfw conversation add`. Title is optional and best omitted for a 1:1,\n" +
			"where the members identify it better than a name would.",
		Args: cobra.NoArgs,
		RunE: runConversationCreate,
	}
	cmd.Flags().String("title", "", "Optional name; omit for a 1:1")
	return cmd
}

func newConversationAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <conversation-id>",
		Short: "Add a friend, by principal ID",
		// ⚠️ --principal, NOT --email. forge/proseforge#911 stopped the API disclosing
		// friends' addresses, so an id is the only name a client can supply. An --email
		// flag would look friendlier and could not be populated from `pfw friends list`.
		Long: "Names the person by PRINCIPAL ID — the id from `pfw friends list`, not an\n" +
			"email address. The API does not disclose friends' addresses, so an id is the\n" +
			"only name you can supply.\n\n" +
			"Creator only, and only people you are already friends with: a refusal here is\n" +
			"usually about the friendship rather than the capability.",
		Args: cobra.ExactArgs(1),
		RunE: runConversationAdd,
	}
	cmd.Flags().String("principal", "", "Principal ID of the person to add (required)")
	cmd.Flags().String("capability", "room:post", "room:post (read+write) or room:enter (read only)")
	_ = cmd.MarkFlagRequired("principal")
	return cmd
}

func newConversationRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <conversation-id>",
		Short: "Remove someone else, by principal ID",
		// 🛑 NOT `leave` with an argument. `leave` revokes YOUR OWN grant and any member
		// may run it; this revokes ANOTHER PERSON'S and only the creator may. Before
		// forge/proseforge#1173 the only exit from a room was for the person being
		// removed to remove themselves, which is the wrong party for every reason you
		// would want this.
		Long: "Names the person by PRINCIPAL ID — from `pfw conversation list` members or\n" +
			"`pfw friends list`, not an email address.\n\n" +
			"Creator only. Removal revokes access; it does not delete anything they wrote.",
		Args: cobra.ExactArgs(1),
		RunE: runConversationRemove,
	}
	cmd.Flags().String("principal", "", "Principal ID of the person to remove (required)")
	_ = cmd.MarkFlagRequired("principal")
	return cmd
}

func runConversationList(cmd *cobra.Command, args []string) error {
	svc, err := newConversationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.List(cmd.Context())
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(result)
	}
	if len(result.Conversations) == 0 {
		fmt.Println("No conversations. Start one with: pfw conversation create")
		return nil
	}
	var rows [][]string
	for _, c := range result.Conversations {
		title := "(untitled)"
		if c.Title != nil && *c.Title != "" {
			title = *c.Title
		}
		owner := ""
		if c.IsOwner {
			owner = "yours"
		}
		rows = append(rows, []string{c.ID, truncate(title, 32), owner, c.UpdatedAt})
	}
	printTable([]string{"ID", "TITLE", "", "UPDATED"}, rows)
	return nil
}

func runConversationCreate(cmd *cobra.Command, args []string) error {
	svc, err := newConversationService(cmd)
	if err != nil {
		return err
	}
	title, _ := cmd.Flags().GetString("title")
	c, err := svc.Create(cmd.Context(), title)
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(c)
	}
	fmt.Printf("Created conversation %s\n", c.ID)
	fmt.Printf("  add someone:  pfw conversation add %s --principal <id>\n", c.ID)
	// ⚠️ VERIFIED BY RUNNING IT, not by reading the flag list. My first version printed
	// `--entity <id>`, which is not a flag `room send` has — the entity id is POSITIONAL.
	// Help text that names a command nobody can run is the same defect as a doc naming a
	// type nobody can use; it reads fine and fails on paste.
	fmt.Printf("  talk in it:   pfw room send %s --type conversation --content \"...\"\n", c.ID)
	return nil
}

func runConversationAdd(cmd *cobra.Command, args []string) error {
	svc, err := newConversationService(cmd)
	if err != nil {
		return err
	}
	principal, _ := cmd.Flags().GetString("principal")
	capability, _ := cmd.Flags().GetString("capability")
	if err := svc.AddMember(cmd.Context(), args[0], principal, capability); err != nil {
		return err
	}
	fmt.Printf("Added %s to conversation %s (%s)\n", principal, args[0], capability)
	return nil
}

func runConversationRemove(cmd *cobra.Command, args []string) error {
	svc, err := newConversationService(cmd)
	if err != nil {
		return err
	}
	principal, _ := cmd.Flags().GetString("principal")
	if err := svc.RemoveMember(cmd.Context(), args[0], principal); err != nil {
		return err
	}
	fmt.Printf("Removed %s from conversation %s\n", principal, args[0])
	return nil
}

func runConversationLeave(cmd *cobra.Command, args []string) error {
	svc, err := newConversationService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Leave(cmd.Context(), args[0]); err != nil {
		return err
	}
	fmt.Printf("Left conversation %s — its messages are no longer readable by you.\n", args[0])
	return nil
}
