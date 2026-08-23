package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/claytonharbour/proseforge-workbench/internal/room"
)

// Room commands (#231). Thin cobra wrappers over room.Service — the same
// service the MCP room_* tools use. The room methods return typed api structs,
// so the CLI renders them directly.

// roomType returns the --type flag (entity type), defaulting to "story".
func roomType(cmd *cobra.Command) string {
	t, _ := cmd.Flags().GetString("type")
	if t == "" {
		return "story"
	}
	return t
}

func newRoomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "room",
		Short: "Coordination room operations (read/send/status/archive)",
	}
	// ⚠️ Free-form, like the MCP `entity_type` arg — the SERVER owns the allowlist. So this
	// has accepted "conversation" since forge/proseforge#821 registered it, and the help text
	// named three types and stopped. Same defect as the eight MCP descriptions fixed in
	// 1724804: a flag value a user is never told about is, from their side, one that does not
	// exist. The description is the functional part here, not decoration.
	cmd.PersistentFlags().String("type", "story",
		"Entity type: story, series, bundle, or conversation (a standalone room with no story or series attached)")
	cmd.AddCommand(
		newRoomReadCmd(),
		newRoomListCmd(),
		newRoomSendCmd(),
		newRoomStatusCmd(),
		&cobra.Command{
			Use:   "archive <entity-id>",
			Short: "Archive a room (reads still work; writes are rejected)",
			Args:  cobra.ExactArgs(1),
			RunE:  runRoomArchive,
		},
		&cobra.Command{
			Use:   "unarchive <entity-id>",
			Short: "Unarchive a room, re-enabling writes",
			Args:  cobra.ExactArgs(1),
			RunE:  runRoomUnarchive,
		},
		newRoomCursorCmd(),
		newRoomWatchCmd(),
		newRoomHealthCmd(),
		newRoomWatchdogCmd(),
		newRoomDeleteCmd(),
	)
	return cmd
}

func newRoomListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rooms you can join and their members",
		Args:  cobra.NoArgs,
		RunE:  runRoomList,
	}
	cmd.Flags().String("handle", "", "Agent handle for the account's stored room cursor context")
	return cmd
}

func roomMemberLabel(member gen.HandlersRoomMember) string {
	if member.Name != nil && *member.Name != "" {
		return *member.Name
	}
	if member.VanityHandle != nil && *member.VanityHandle != "" {
		return "@" + *member.VanityHandle
	}
	if member.Id != nil {
		return *member.Id
	}
	return "(unnamed)"
}

func roomMemberSummary(room gen.HandlersRoomListEntry) string {
	if room.Members == nil || len(*room.Members) == 0 {
		return "0"
	}
	labels := make([]string, 0, len(*room.Members))
	for _, member := range *room.Members {
		labels = append(labels, roomMemberLabel(member))
	}
	return fmt.Sprintf("%d: %s", len(labels), strings.Join(labels, ", "))
}

func runRoomList(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	handle, _ := cmd.Flags().GetString("handle")
	rooms, err := svc.ListMine(cmd.Context(), handle)
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(rooms)
	}
	if len(*rooms) == 0 {
		fmt.Println("No rooms available.")
		return nil
	}
	var rows [][]string
	for _, room := range *rooms {
		rows = append(rows, []string{
			deref(room.EntityType), deref(room.EntityId), truncate(deref(room.Title), 28),
			roomMemberSummary(room), fmt.Sprintf("%d", derefInt(room.UnreadCount)),
			fmt.Sprintf("%t", derefBool(room.CanPost)),
		})
	}
	printTable([]string{"Type", "ID", "Title", "Members", "Unread", "Can post"}, rows)
	return nil
}

// === room read ===

func newRoomReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <entity-id>",
		Short: "Read messages from a room",
		Long:  "Read messages from a room — full history, a delta from --since/--handle, or a --match regex subset.",
		Args:  cobra.ExactArgs(1),
		RunE:  runRoomRead,
	}
	cmd.Flags().String("since", "", "Read messages after this message ID")
	cmd.Flags().String("handle", "", "Resume from this agent's server-stored cursor (when --since is empty)")
	// ⚠️ Says content AND target, because it matches both (filter.go: Content+" "+Target).
	// This read "on message content" for months and it was not cosmetic: four benches
	// concluded --match cannot see the target field and the room nearly removed a
	// working flag over it (#346). Keep this wording identical to `room watch`.
	cmd.Flags().String("match", "", "RE2 regex filter (case-insensitive) on message content AND target")
	// ⚠️ A VETO, not a third filter term (#336). It runs first and beats a
	// positive --match, which is the only way to express "messages naming me,
	// but not the ones I wrote".
	cmd.Flags().String("exclude-from", "", "VETO a sender: drop their messages even if --match/--from kept them. Beats a positive match.")
	cmd.Flags().String("order", "", "Sort order: asc (default, OLDEST first) or desc (newest first)")
	// 🛑 SCAN, not results. Every wrong conclusion in #346 came from here: with the
	// default asc order, --limit N reads the N OLDEST messages, so a pattern only
	// used recently returns 0 and looks like a broken filter. Use --order desc, or
	// --since, to test a recent pattern.
	cmd.Flags().Int("limit", 0, "Max messages to SCAN, not to return (0 = server default). With --order asc this scans the OLDEST N.")
	return cmd
}

func runRoomRead(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	since, _ := cmd.Flags().GetString("since")
	handle, _ := cmd.Flags().GetString("handle")
	match, _ := cmd.Flags().GetString("match")
	order, _ := cmd.Flags().GetString("order")
	limit, _ := cmd.Flags().GetInt("limit")
	from, _ := cmd.Flags().GetString("from")
	excludeFrom, _ := cmd.Flags().GetString("exclude-from")
	opts := api.ReadRoomMessagesOptions{
		Since:       since,
		AgentHandle: handle,
		Match:       nonEmpty(match),
		From:        nonEmpty(from),
		ExcludeFrom: nonEmpty(excludeFrom),
		Order:       order,
		Limit:       limit,
	}
	res, err := svc.Read(cmd.Context(), roomType(cmd), args[0], opts)
	if err != nil {
		return notFound(err, "room", args[0])
	}
	if isJSON(cmd) {
		return printJSON(res)
	}
	if len(res.Messages) == 0 {
		if !isBrief(cmd) {
			fmt.Println("No messages.")
		}
		return nil
	}
	if isBrief(cmd) {
		var rows [][]string
		for _, m := range res.Messages {
			rows = append(rows, []string{m.ID, m.Agent})
		}
		printBrief(rows)
		return nil
	}
	// ⚠️ Say when the result may not answer the question asked (#346). Three
	// benches drew confident false negatives from this read tonight; all three
	// were the defaults behaving as documented, and documentation is not a
	// warning at the moment of use.
	for _, w := range room.WindowWarnings(len(res.Messages), limit, order,
		match != "" || from != "" || excludeFrom != "", since != "" || handle != "") {
		status("WINDOW: %s\n", w)
	}
	status("Messages: %d  lastId: %s", len(res.Messages), res.LastID)
	fmt.Println()
	for _, m := range res.Messages {
		hdr := m.Agent
		if m.Perspective != "" {
			hdr += " / " + m.Perspective
		}
		fmt.Printf("── [%s] %s\n", hdr, m.Timestamp)
		fmt.Println(m.Content)
		fmt.Println()
	}
	return nil
}

// === room send ===

func newRoomSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <entity-id>",
		Short: "Post a message to a room",
		Args:  cobra.ExactArgs(1),
		RunE:  runRoomSend,
	}
	cmd.Flags().String("agent", "", "Your identity (title-of-the-moment)")
	cmd.Flags().String("perspective", "", "Craft lens (artificer, keeper, ...)")
	cmd.Flags().String("target", "", "Topic this relates to")
	cmd.Flags().String("content", "", "Message body (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the message body from stdin")
	// 🛑 "Am I where I think I am." `room watch` has had --expect-email since
	// #308; sending never had an equivalent, because room delivery has NO BOUNCE
	// — post to a wrong-but-valid id and it succeeds silently (@Sten, #356).
	//
	// ⚠️ Both are guards against a MISTAKE, never assurance of delivery. Neither
	// can tell you anyone read it.
	cmd.Flags().String("expect-title", "", "Refuse to post unless the room's title matches this regex. A TYPO GUARD against the wrong room id — not delivery assurance.")
	cmd.Flags().String("expect-member", "", "Refuse to post unless this person is in the room's roster. Membership is capability, NOT readership.")
	cmd.Flags().StringArray("image", nil, "Path to an image to post (jpeg/png/webp, max 10 MiB = 10,485,760 bytes). Repeatable. Place each with a numbered token — {{image:1}} is the first --image, {{image:2}} the second, bare {{image}} means the first; untokened images are appended in order. A token is replaced and does not survive into the message; wrap it in backticks to write about it literally. Any failed upload aborts the post")
	return cmd
}

// checkRoomExpectations enforces --expect-title / --expect-member (#356).
//
// No flags means no API call, so an unguarded send costs exactly what it did.
func checkRoomExpectations(cmd *cobra.Command, svc *room.Service, entityType, entityID string) error {
	wantTitle, _ := cmd.Flags().GetString("expect-title")
	wantMember, _ := cmd.Flags().GetString("expect-member")
	if wantTitle == "" && wantMember == "" {
		return nil
	}

	facts, err := svc.LookupForSend(cmd.Context(), entityType, entityID)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotListable) {
			// 🛑 Say THIS, not "member not found". You have no roster, not an
			// empty one — reporting absence would be a confident wrong answer.
			//
			// ⚠️ And do NOT suggest dropping the flag. Measured: a room you
			// cannot list 403s on status, read AND send — access is
			// membership-gated throughout. So this state only occurs where the
			// post would fail anyway, and "try without the guard" would send
			// someone to a second, more confusing failure (@Sten, #356).
			return fmt.Errorf("refusing to post: you are not a member of %s/%s — it is not in "+
				"your room list. The post would be refused as well; membership is required to "+
				"read, post, or inspect a room", entityType, entityID)
		}
		return err
	}

	if wantTitle != "" {
		re, cerr := regexp.Compile("(?i)" + wantTitle)
		if cerr != nil {
			return fmt.Errorf("invalid --expect-title %q: %w", wantTitle, cerr)
		}
		if !re.MatchString(facts.Title) {
			return fmt.Errorf("refusing to post: room %s is titled %q, which does not match "+
				"--expect-title %q", entityID, facts.Title, wantTitle)
		}
	}

	if wantMember != "" {
		found := false
		for _, m := range facts.Members {
			if strings.EqualFold(m, wantMember) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("refusing to post: %q is not in %q. Roster: %s.\n"+
				"  (membership is capability, not readership — this cannot tell you anyone read it)",
				wantMember, facts.Title, strings.Join(facts.Members, ", "))
		}
	}
	return nil
}

func runRoomSend(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	// Guards run BEFORE any upload or post: a refusal must cost nothing and
	// leave nothing behind.
	if err := checkRoomExpectations(cmd, svc, roomType(cmd), args[0]); err != nil {
		return err
	}

	agent, _ := cmd.Flags().GetString("agent")
	perspective, _ := cmd.Flags().GetString("perspective")
	target, _ := cmd.Flags().GetString("target")
	content, _ := cmd.Flags().GetString("content")
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			return fmt.Errorf("reading stdin: %w", rerr)
		}
		content = string(data)
	}
	if content == "" {
		return fmt.Errorf("message body is required: use --content or --stdin")
	}
	// Upload first: if it fails the message is not posted, because a post whose
	// image quietly went missing reads as a complete message and hides it.
	if imagePaths, _ := cmd.Flags().GetStringArray("image"); len(imagePaths) > 0 {
		urls, uerr := svc.UploadImages(cmd.Context(), roomType(cmd), args[0], imagePaths)
		if uerr != nil {
			return uerr
		}
		content = room.EmbedImages(content, urls)
		for i, url := range urls {
			status("Uploaded {{image:%d}} → %s\n", i+1, url)
		}
	}

	res, err := svc.Send(cmd.Context(), roomType(cmd), args[0], agent, perspective, target, content)
	if err != nil {
		return err
	}
	// Echo WHERE it went, not just that it went (#355). Resolved after the send
	// so a lookup failure can never make a delivered message report an error.
	dest := svc.ResolveDestination(cmd.Context(), roomType(cmd), args[0])

	if isJSON(cmd) {
		return printJSON(room.SendResult{SendRoomMessageResponse: res, Destination: dest})
	}
	// ⚠️ Destination on the FIRST line, before the id. The id is what you passed
	// in; the title is the only part that can contradict you, and a confirmation
	// that leads with the thing you already knew is what let four misroutes in a
	// row read as success.
	if dest.Resolved() {
		fmt.Printf("Message sent to %q (%s, %d members)\n", dest.Title, dest.EntityType, dest.Members)
	} else {
		fmt.Printf("Message sent to %s %s\n", dest.EntityType, dest.EntityID)
		status("room title unavailable — it is not in your room list, so the destination could not be confirmed\n")
	}
	fmt.Printf("  ID:      %s\n  Backend: %s\n", res.ID, res.Backend)
	return nil
}

// === room delete ===

func newRoomDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <entity-id> <message-id>",
		Short: "Delete one message from a room",
		Long: `Delete a single message from a room.

You may remove your own; the room's owner or an admin may remove any.

Irreversible, and there is no edit — a correction is a new message. Take the
message id from 'pfw room read'.`,
		Args: cobra.ExactArgs(2),
		RunE: runRoomDelete,
	}
}

func runRoomDelete(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	if err := svc.DeleteMessage(cmd.Context(), roomType(cmd), args[0], args[1]); err != nil {
		return err
	}
	fmt.Printf("Message %s deleted.\n", args[1])
	return nil
}

// === room status ===

func newRoomStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <entity-id>",
		Short: "Check whether a room exists, is archived, and its message count",
		Args:  cobra.ExactArgs(1),
		RunE:  runRoomStatus,
	}
}

func runRoomStatus(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	res, err := svc.Status(cmd.Context(), roomType(cmd), args[0])
	if err != nil {
		return notFound(err, "room", args[0])
	}
	if isJSON(cmd) {
		return printJSON(res)
	}
	fmt.Printf("Exists:   %t\n", res.Exists)
	fmt.Printf("Archived: %t\n", res.Archived)
	fmt.Printf("Messages: %d\n", res.MessageCount)
	return nil
}

// === room archive / unarchive ===

func runRoomArchive(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Archive(cmd.Context(), roomType(cmd), args[0]); err != nil {
		return notFound(err, "room", args[0])
	}
	fmt.Println("Room archived.")
	return nil
}

func runRoomUnarchive(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Unarchive(cmd.Context(), roomType(cmd), args[0]); err != nil {
		return notFound(err, "room", args[0])
	}
	fmt.Println("Room unarchived.")
	return nil
}

// === room cursor ===

func newRoomCursorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor",
		Short: "Get or set an agent's stored room cursor",
	}
	get := &cobra.Command{
		Use:   "get <entity-id>",
		Short: "Read an agent's stored cursor",
		Args:  cobra.ExactArgs(1),
		RunE:  runRoomCursorGet,
	}
	get.Flags().String("handle", "", "Agent's lineage handle")

	set := &cobra.Command{
		Use:   "set <entity-id>",
		Short: "Advance an agent's stored cursor",
		Args:  cobra.ExactArgs(1),
		RunE:  runRoomCursorSet,
	}
	set.Flags().String("handle", "", "Agent's lineage handle")
	set.Flags().String("last-id", "", "Message ID to advance the cursor to")
	_ = set.MarkFlagRequired("last-id")

	cmd.AddCommand(get, set)
	return cmd
}

func runRoomCursorGet(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	handle, _ := cmd.Flags().GetString("handle")
	res, err := svc.GetCursor(cmd.Context(), roomType(cmd), args[0], handle)
	if err != nil {
		return notFound(err, "room", args[0])
	}
	if isJSON(cmd) {
		return printJSON(res)
	}
	if res.LastID == "" {
		fmt.Println("No cursor set.")
		return nil
	}
	fmt.Println(res.LastID)
	return nil
}

func runRoomCursorSet(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	handle, _ := cmd.Flags().GetString("handle")
	lastID, _ := cmd.Flags().GetString("last-id")
	res, err := svc.SetCursor(cmd.Context(), roomType(cmd), args[0], handle, lastID)
	if err != nil {
		return notFound(err, "room", args[0])
	}
	if isJSON(cmd) {
		return printJSON(res)
	}
	fmt.Printf("Cursor set.\nBackend: %s\n", res.Backend)
	return nil
}

// nonEmpty wraps a single pattern for the repeatable server params, dropping
// the empty case so an unset flag sends nothing rather than an empty regex
// (which would match everything).
func nonEmpty(v string) []string {
	if v == "" {
		return nil
	}
	return []string{v}
}
