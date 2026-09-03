package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

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
		newRoomStatsCmd(),
		newRoomPresenceCmd(),
		newRoomReactCmd(),
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
		newRoomRosterCmd(),
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
		// ⚠️ SAYS "including your own" ON PURPOSE. `room watch` skips your posts by
		// default (--include-self), and that asymmetry is documented only on watch —
		// so a reader who knows the watcher self-skips has no signal that read does
		// not. A bench concluded from it that they were "structurally excluded" from
		// their own volume and could not measure it; the tool one command away
		// answers exactly that, and nothing said so.
		Long: "Read messages from a room — full history, a delta from --since/--handle, or a --match regex subset.\n\n" +
			"Returns every message including your own. `room watch` is the one that skips your posts by default; " +
			"this does not, so it is the right tool for \"what have I posted here\".",
		Args: cobra.ExactArgs(1),
		RunE: runRoomRead,
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
	// ⚠️ Say when the result may not answer the question asked (#346), and say it
	// BEFORE any early return. #414, @Crispin: this block used to sit below the
	// -o json and --brief exits, so the warning fired only in table mode.
	//
	// 🛑 That is backwards on severity. A table has visible dates and an obvious
	// end, so a human spots the truncation anyway; a JSON or --brief consumer is a
	// script computing on an array with no way to see it is holding a window. The
	// warning was loud exactly where it was least needed and silent exactly where
	// nothing else could catch it — and a script does not shrug, it publishes
	// "it is not there".
	//
	// ⚑ @Crispin nearly published that two messages did not exist; what caught it
	// was noticing their own last post was three days old, not this tool.
	for _, w := range room.WindowWarnings(len(res.Messages), limit, order,
		match != "" || from != "" || excludeFrom != "", since != "" || handle != "") {
		status("WINDOW: %s\n", w)
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
	status("Messages: %d  lastId: %s", len(res.Messages), res.LastID)
	fmt.Println()

	// Index the window so a reply can name its parent. ⚠️ Only messages IN THIS
	// WINDOW are resolvable — see replyLine for why that is stated rather than
	// papered over.
	byID := make(map[string]api.RoomMessage, len(res.Messages))
	for _, m := range res.Messages {
		byID[m.ID] = m
	}

	for _, m := range res.Messages {
		hdr := m.Agent
		if m.Perspective != "" {
			hdr += " / " + m.Perspective
		}
		fmt.Printf("── [%s] %s  %s\n", hdr, m.Timestamp, m.ID)
		if m.ReplyTo != "" {
			fmt.Println(replyLine(m.ReplyTo, byID))
		}
		fmt.Println(m.Content)
		if line := reactionLine(m.Reactions); line != "" {
			fmt.Println(line)
		}
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
	cmd.Flags().String("reply-to", "", "Thread this under another message (id from `room read`)")
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
	replyTo, _ := cmd.Flags().GetString("reply-to")
	if err := validateReplyTo(cmd.Flags().Changed("reply-to"), replyTo); err != nil {
		return err
	}
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

	res, err := svc.Send(cmd.Context(), roomType(cmd), args[0], api.SendRoomMessageRequest{
		Agent:       agent,
		Perspective: perspective,
		Target:      target,
		Content:     content,
		ReplyTo:     replyTo,
	})
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

// validateReplyTo rejects a --reply-to that was GIVEN but empty (#466).
//
// 🛑 A FAILED LOOKUP AND AN INTENT TO BROADCAST ARE DIFFERENT ACTS, and before
// this they were the same value. `--reply-to ""` posted TOP-LEVEL, exit 0, with
// the ordinary success block — so a scripted reply whose id extraction returned
// nothing silently became a broadcast, detectable only by reading threadRootId
// back off your own message afterwards. An empty variable is the NORMAL failure
// of any id-extraction pipeline against a fast-moving room, so this is the
// common case rather than an exotic one.
//
// ⚠️ It takes `changed` rather than reading the flag itself because that is the
// whole distinction: by the time ReplyTo is a string on the request struct, ""
// from an omitted flag and "" from a failed lookup are indistinguishable. Only
// the surface knows whether the flag was PRESENT — which is also why this cannot
// live in the service layer where CLI/MCP shared logic normally belongs (#226).
// The MCP handler carries the same check keyed on argument presence.
func validateReplyTo(changed bool, value string) error {
	if changed && strings.TrimSpace(value) == "" {
		return fmt.Errorf("--reply-to was given but empty: a reply target is required when the flag is present\n" +
			"       (omit --reply-to entirely to post top-level)")
	}
	return nil
}

// replyLine renders a reply's parent pointer.
//
// 🛑 THE PARENT MAY GENUINELY BE GONE. The room stream is capped (proseforge#579),
// so a long-lived thread outlives its own root and replies orphan BY DEFAULT. The
// API's own type says clients must render "no longer available" rather than assume
// a resolvable parent.
//
// ⚠️ But a client CANNOT distinguish "evicted from the stream" from "outside the
// window I asked for" — both look like a miss in this map. So the unresolved case
// says exactly what is known ("not in this window") and does NOT claim the parent
// is gone. Raising --limit is the reader's next move, and a message that asserted
// deletion would send them looking for a bug instead.
func replyLine(parentID string, byID map[string]api.RoomMessage) string {
	p, ok := byID[parentID]
	if !ok {
		return fmt.Sprintf("   ↳ replying to %s (not in this window — raise --limit or --since)", parentID)
	}
	preview := strings.ReplaceAll(p.Content, "\n", " ")
	return fmt.Sprintf("   ↳ replying to %s (%s: %q)", parentID, p.Agent, truncate(preview, 48))
}

// reactionLine renders a message's reactions, or "" when it has none.
//
// ⚠️ Shows the COUNT, not who — the principal IDs are on the wire but rendering a
// dozen UUIDs under every message would cost more attention than the reactions save,
// which is the opposite of the point. `-o json` carries principalIds for anyone who
// needs to answer "did THAT person see it".
func reactionLine(rs []api.MessageReaction) string {
	if len(rs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(rs))
	for _, r := range rs {
		mine := ""
		if r.Mine {
			mine = "*" // the caller is among the reactors
		}
		parts = append(parts, fmt.Sprintf("%s%s%d", r.Emoji, mine, r.Count))
	}
	return "   " + strings.Join(parts, "  ")
}

// === room react ===

func newRoomReactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "react <entity-id> <message-id> <emoji>",
		Short: "Acknowledge a message with an emoji instead of writing a reply",
		Long: "Acknowledge a message with an emoji instead of writing a reply.\n\n" +
			"A reaction is the cheapest way to say \"seen it\" — the alternative is a\n" +
			"paragraph that costs every reader in the room their attention.\n\n" +
			"Both directions are idempotent: reacting twice is one reaction, and removing\n" +
			"one that is not there is not an error. Message IDs come from `pfw room read`.",
		Args: cobra.ExactArgs(3),
		RunE: runRoomReact,
	}
	cmd.Flags().Bool("remove", false, "Withdraw this reaction instead of adding it")
	return cmd
}

func runRoomReact(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	entityID, messageID, emoji := args[0], args[1], args[2]
	remove, _ := cmd.Flags().GetBool("remove")

	if remove {
		if err := svc.Unreact(cmd.Context(), roomType(cmd), entityID, messageID, emoji); err != nil {
			return reactErr(err, messageID)
		}
		fmt.Printf("Removed %s from %s\n", emoji, messageID)
		return nil
	}
	if err := svc.React(cmd.Context(), roomType(cmd), entityID, messageID, emoji); err != nil {
		return reactErr(err, messageID)
	}
	fmt.Printf("Reacted %s to %s\n", emoji, messageID)
	return nil
}

// reactErr distinguishes an older backend from a bad message id, for the same reason
// `presence` does: `pfw` is pointed at servers we do not control.
func reactErr(err error, messageID string) error {
	if routeMissing(err) {
		return fmt.Errorf("this server does not support reactions: the endpoint is " +
			"missing (needs a backend with forge/proseforge#1250)")
	}
	return notFound(err, "message", messageID)
}

// === room presence ===

func newRoomPresenceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "presence <entity-id>",
		Short: "Who has READ this room, and how long ago",
		Long: "Who has READ this room, and how long ago.\n\n" +
			"⛔ This is NOT an online/offline list and does not report one. Watchers in a\n" +
			"fleet poll at different cadences — 60s, 30m and 1h are all in use — so a\n" +
			"40-minute age is healthy for an hourly leg and stale for a per-minute one.\n" +
			"There is no threshold this command could apply that would be right for both,\n" +
			"so it reports the age and leaves the judgement to you.",
		Args: cobra.ExactArgs(1),
		RunE: runRoomPresence,
	}
}

func runRoomPresence(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	entries, err := svc.Presence(cmd.Context(), roomType(cmd), args[0])
	if err != nil {
		if routeMissing(err) {
			return fmt.Errorf("this server does not provide room presence: the endpoint "+
				"is missing, not the room (needs a backend with forge/proseforge#1270). "+
				"Verify the room itself with: pfw room status %s", args[0])
		}
		return notFound(err, "room", args[0])
	}
	if isJSON(cmd) {
		return printJSON(entries)
	}
	if len(entries) == 0 {
		// ⚠️ The endpoint returns an empty array for an UNKNOWN room as well as an
		// unread one — it does not 404 — so this message must not assert the room
		// exists. Measured on dev: a bogus UUID yields [] rather than an error.
		fmt.Println("No reads recorded for this room.")
		status("(an unknown room id also returns no reads — `pfw room status %s` tells them apart)", args[0])
		return nil
	}

	// Most recently seen first: the reader is usually asking "did anyone see it".
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastSeenAt.After(entries[j].LastSeenAt)
	})

	// The endpoint returns principal UUIDs only. Join them against the room's member
	// list so the table reads as people rather than identifiers. Best-effort: if the
	// lookup fails the UUID is still correct, so a degraded table beats no table.
	names := presenceNames(cmd, svc, roomType(cmd), args[0])

	now := time.Now()
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		who := e.PrincipalID
		if n, ok := names[e.PrincipalID]; ok && n != "" {
			who = n
		}
		rows = append(rows, []string{
			who,
			e.LastSeenAt.Local().Format("2006-01-02 15:04:05"),
			humanAge(now.Sub(e.LastSeenAt)),
		})
	}
	printTable([]string{"Who", "Last read", "Age"}, rows)

	// stderr: the rows are on stdout for pipes, the caveats are here for humans.
	//
	// ⚠️ Give the reader a DENOMINATOR. Presence only knows about reads it recorded,
	// so a room whose tracking is younger than its traffic shows a handful of readers
	// out of a full membership — which looks like a dead room and is not one. Naming
	// the member count reframes "1 reader" as "1 of 13 recorded", which is the true
	// and much less alarming statement.
	out := cmd.ErrOrStderr()
	if len(names) > 0 {
		fmt.Fprintf(out, "\n%d of %d members have a recorded read.\n", len(entries), len(names))

		// The second population — who has NOT been recorded reading. This is the
		// question a sender actually has ("did N see it"), and it is a set
		// subtraction on data already in hand, not another call.
		//
		// ⛔ It degrades with the roster for the same reason names do, and the
		// failure is WORSE here: a silently-empty roster would print "no read
		// recorded: 0", which reads as "everyone has seen it" — the inverse of the
		// misreading this line exists to prevent, and far more convincing. Guarded
		// by the same len(names) check above, so it cannot print alone.
		read := make(map[string]bool, len(entries))
		for _, e := range entries {
			read[e.PrincipalID] = true
		}
		var silent []string
		for id, n := range names {
			if !read[id] {
				silent = append(silent, n)
			}
		}
		if len(silent) > 0 {
			sort.Strings(silent)
			const cap = 12
			shown := silent
			suffix := ""
			if len(shown) > cap {
				shown, suffix = shown[:cap], fmt.Sprintf(" (+%d more)", len(silent)-cap)
			}
			fmt.Fprintf(out, "No read recorded: %s%s\n", strings.Join(shown, ", "), suffix)
		}
	}
	fmt.Fprintln(out,
		"\n⚠️  Ages, not availability. Poll cadences in this fleet span 60s to 1h,\n"+
			"   so no single \"active\" window is correct for every reader.\n"+
			"   Absence means NO READ RECORDED — which includes reads that happened\n"+
			"   before this server started tracking presence.")
	return nil
}

// presenceNames maps principal id -> display name using the caller's room list,
// which is the only surface that carries both. Returns an empty map on any failure:
// this is a display nicety and must never turn a working `presence` into an error.
//
// ⚠️ Only members the CALLER can see are resolvable. A principal who has read the
// room but is not in the caller's member view stays a UUID — which is the honest
// rendering, not a gap to paper over.
func presenceNames(cmd *cobra.Command, svc *room.Service, entityType, entityID string) map[string]string {
	names := map[string]string{}
	rooms, err := svc.ListMine(cmd.Context(), "")
	if err != nil || rooms == nil {
		return names
	}
	for _, r := range *rooms {
		if r.EntityId == nil || *r.EntityId != entityID {
			continue
		}
		if r.EntityType == nil || *r.EntityType != entityType {
			continue
		}
		if r.Members == nil {
			continue
		}
		for _, m := range *r.Members {
			if m.Id != nil && m.Name != nil {
				names[*m.Id] = *m.Name
			}
		}
	}
	return names
}

// humanAge renders a duration at one significant unit — enough to judge staleness
// against a known cadence, without implying precision the timestamp does not have.
func humanAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
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
	// ⚠️ Printed for the same reason it is carried: an agent deciding whether to
	// post should be able to ask, rather than find out by posting into a shared room.
	fmt.Printf("Can post: %t\n", res.CanPost)
	fmt.Printf("Moderate: %t\n", res.CanModerate)
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

// newRoomStatsCmd answers "how much am I saying in here", which #458 exists because
// nobody can answer about themselves.
//
// ⚑ Measured 2026-08-29: self-estimates in this fleet were wrong by roughly 5x in
// BOTH directions — one bench guessed ~25 against an actual 117, another called
// themselves "a heavy consumer" and ranked 9th of 12. Volume is not merely
// unmeasured, it is unobservable to the person generating it.
func newRoomStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats <entity-id>",
		Short: "Per-author message and character volume for a room",
		Long: "Group a room's messages by author and report count, characters and average length.\n\n" +
			"Computed entirely client-side from `room read`, so it needs no server support and " +
			"counts your OWN messages — which `room watch` skips by default, and which is exactly " +
			"the half you cannot otherwise see.\n\n" +
			"Characters are RUNES, not bytes: these rooms are full of box-drawing and emoji, and a " +
			"byte count ranks formatting style rather than volume.",
		Args: cobra.ExactArgs(1),
		RunE: runRoomStats,
	}
	cmd.Flags().String("since", "", "Only count messages after this message ID")
	cmd.Flags().String("from", "", "Only count messages from this sender")
	cmd.Flags().Bool("by-day", false, "Also break the totals down by day")
	cmd.Flags().String("order", "desc", "Sort order to SCAN in: desc (newest first, the default here) or asc")
	// 🛑 Defaults to desc, unlike `room read`. A stats window that starts at the room's
	// beginning describes ancient history, and the question is almost always "lately".
	cmd.Flags().Int("limit", 0, "Max messages to SCAN (0 = server default). This bounds the window the numbers describe.")
	return cmd
}

func runRoomStats(cmd *cobra.Command, args []string) error {
	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	since, _ := cmd.Flags().GetString("since")
	from, _ := cmd.Flags().GetString("from")
	order, _ := cmd.Flags().GetString("order")
	limit, _ := cmd.Flags().GetInt("limit")
	byDay, _ := cmd.Flags().GetBool("by-day")

	res, err := svc.Read(cmd.Context(), roomType(cmd), args[0], api.ReadRoomMessagesOptions{
		Since: since, From: nonEmpty(from), Order: order, Limit: limit,
	})
	if err != nil {
		return notFound(err, "room", args[0])
	}

	// ⚠️ BEFORE the -o json return, deliberately. A truncated window turns a total
	// into a lower bound, and the JSON consumer is the one with no visible dates to
	// notice it with (#414).
	for _, w := range room.WindowWarnings(len(res.Messages), limit, order, from != "", since != "") {
		status("WINDOW: %s\n", w)
	}

	st := room.Stats(res.Messages, byDay)
	if isJSON(cmd) {
		return printJSON(st)
	}
	if st.Messages == 0 {
		fmt.Println("No messages.")
		return nil
	}

	fmt.Printf("%-20s %8s %10s %8s\n", "AGENT", "MSGS", "CHARS", "AVG")
	fmt.Printf("%-20s %8s %10s %8s\n", strings.Repeat("-", 20), "--------", "----------", "--------")
	for _, a := range st.Authors {
		fmt.Printf("%-20s %8d %10d %8d\n", a.Agent, a.Messages, a.Chars, a.AvgChars)
	}
	fmt.Printf("%-20s %8d %10d %8d\n", "TOTAL", st.Messages, st.Chars, st.Chars/st.Messages)

	if byDay && len(st.Days) > 0 {
		fmt.Printf("\n%-12s %8s %10s\n", "DAY", "MSGS", "CHARS")
		for _, d := range st.Days {
			fmt.Printf("%-12s %8d %10d\n", d.Day, d.Messages, d.Chars)
		}
	}

	// 🛑 The window bound is printed WITH the numbers, never left to the caller to
	// remember. Every total above is "within this window"; without the dates it reads
	// as the room's lifetime volume.
	if st.Earliest != "" {
		status("\nwindow: %s .. %s (%d messages scanned)\n", st.Earliest, st.Latest, st.Messages)
	}
	return nil
}
