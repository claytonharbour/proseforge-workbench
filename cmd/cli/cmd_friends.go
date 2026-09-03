package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/friends"
)

func newFriendsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "friends",
		Short: "Your friends directory — the people you can invite to collaborate",
		Long: `Your friends directory.

Friendship is consent to be asked: you invite collaborators to a story from the
people you are friends with. Adding a friend is a request the other person
accepts or declines.`,
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List your accepted friends",
		Args:  cobra.NoArgs,
		RunE:  runFriendsList,
	}
	list.Flags().Bool("include-ai", false, "Include the AI reviewers alongside people")

	candidates := &cobra.Command{
		Use:   "search <term>",
		Short: "Find people you could add as a friend",
		Long: `Search for people to add as a friend.

Match is against name and email. Omitting the term currently returns the whole
directory, so a term is required here.`,
		Args: cobra.ExactArgs(1),
		RunE: runFriendsSearch,
	}

	add := &cobra.Command{
		Use:   "add <user-id>",
		Short: "Send a friend request",
		Args:  cobra.ExactArgs(1),
		RunE:  runFriendsAdd,
	}

	remove := &cobra.Command{
		Use:   "remove <friend-id>",
		Short: "End a friendship",
		Args:  cobra.ExactArgs(1),
		RunE:  runFriendsRemove,
	}

	requests := &cobra.Command{
		Use:   "requests",
		Short: "Friend requests awaiting you — add --outgoing for ones you have sent",
		Args:  cobra.NoArgs,
		RunE:  runFriendsRequests,
	}
	requests.Flags().Bool("outgoing", false, "Show requests you sent instead of ones awaiting you")

	respond := &cobra.Command{
		Use:   "respond <request-id>",
		Short: "Accept or decline an incoming friend request",
		Args:  cobra.ExactArgs(1),
		RunE:  runFriendsRespond,
	}
	respond.Flags().Bool("accept", false, "Accept the request")
	respond.Flags().Bool("decline", false, "Decline the request")

	cmd.AddCommand(list, candidates, add, remove, requests, respond)
	return cmd
}

func newFriendsService(cmd *cobra.Command) (*friends.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return friends.NewService(client, friends.WithLogger(cliLogger)), nil
}

// friendRow is one entry in any of the friends payloads. The surface uses three
// different array keys and both vocabularies, so decode leniently: `friends`,
// `reviewers` (list and candidates) and `requests` (both request directions).
// Missing an alternative here silently prints "none" over real data — which is
// exactly what this renderer did to a pending request on its first outing.
type friendRow struct {
	ID             string `json:"id"`
	FriendID       string `json:"friendId"`
	RequesterID    string `json:"requesterId"`
	Status         string `json:"status"`
	Name           string `json:"name"`
	ReviewerName   string `json:"reviewerName"`
	ReviewerEmail  string `json:"reviewerEmail"`
	RequesterName  string `json:"requesterName"`
	RequesterEmail string `json:"requesterEmail"`
	Email          string `json:"email"`
	// UserKind separates a person from an AI reviewer ("user" / "system").
	// Worth a column: system accounts only appear when you ask for them with
	// --include-ai, and when they do you want to see which rows they are.
	UserKind string `json:"userKind"`

	// ── forge/proseforge#1017 C2: direction, stated instead of inferred ──────────
	//
	// 🛑 `FriendID` is only "the other person" on the accepted-friends list, and only
	// because the server normalises every row there to caller/other. Nothing in this
	// repo could verify that invariant — it lives in the API's SQL — and the same
	// field on an INCOMING request is the reader themselves, which this file already
	// learned the hard way (see modeIncoming).
	//
	// CounterpartyID/Name are the other party on EVERY endpoint by construction, and
	// InitiatorID is who actually asked. All three are absent from a pre-#1017 server,
	// so every use below falls back — see identity().
	InitiatorID      string `json:"initiatorId"`
	CounterpartyID   string `json:"counterpartyId"`
	CounterpartyName string `json:"counterpartyName"`
}

// rowMode selects which id and which party a row should show. The two are not
// cosmetic choices — each view's id is the one the *next* command consumes.
type rowMode int

const (
	// modeDirectory: friends list and candidate search. The id is the person's
	// user id, which is what 'friends add' takes.
	modeDirectory rowMode = iota
	// modeIncoming: requests awaiting you. The id is the request id, which is
	// what 'friends respond' takes, and the party shown is whoever is asking.
	//
	// Using friendId here was a real bug: every incoming request carries your
	// own id in that field, so distinct requests rendered as identical rows and
	// the id shown could not be passed to respond.
	modeIncoming
	// modeOutgoing: requests you sent. Request id again, but the party shown is
	// the person you asked.
	modeOutgoing
)

// identity returns the id to display and the party to name, per mode.
//
// ⚑ #1017 C2 — CounterpartyName leads in every mode, because it is the other party on
// every endpoint by construction. The LEGACY fallback behind it is mode-specific and
// must stay that way:
//
//	incoming   the other party is the REQUESTER side — the reviewer side is YOU
//	outgoing   the other party is the REVIEWER side
//	directory  normalised to the other party, but only on that one list
//
// 🛑 A single legacy fallback is WRONG and @Gordon shipped exactly that on the web
// client an hour ago: one `|| reviewerName` renders the reader's own name as the person
// asking, on incoming rows, against any server without #1017. The fallback is the live
// path on demo and prod, so this is not hypothetical.
//
// ⛔ ONLY the directory id changes. `modeIncoming`/`modeOutgoing` return `r.ID` — the
// REQUEST id, which is what `friends respond` consumes. Swapping a person's id in there
// would break the next command, which is the bug recorded against modeIncoming above.
func (r friendRow) identity(mode rowMode) (id, name, email string) {
	switch mode {
	case modeIncoming:
		return r.ID, firstNonEmpty(r.CounterpartyName, r.RequesterName, r.Name),
			firstNonEmpty(r.RequesterEmail, r.Email)
	case modeOutgoing:
		return r.ID, firstNonEmpty(r.CounterpartyName, r.ReviewerName, r.Name),
			firstNonEmpty(r.ReviewerEmail, r.Email)
	default:
		return firstNonEmpty(r.CounterpartyID, r.FriendID, r.ID),
			firstNonEmpty(r.CounterpartyName, r.Name, r.ReviewerName),
			firstNonEmpty(r.Email, r.ReviewerEmail)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// friendsEnvelope is the shape the friends endpoints return. It is a named type
// rather than an inline struct so that rows() below can be pinned by a test.
type friendsEnvelope struct {
	Count     int         `json:"count"`
	Total     int         `json:"total"`
	Friends   []friendRow `json:"friends"`
	Reviewers []friendRow `json:"reviewers"`
	Requests  []friendRow `json:"requests"`
}

// rows picks which key to render, and the ORDER is the load-bearing part.
//
// ⚠️ `friends` must be read before `reviewers` (forge/proseforge#1159). During
// the ADD phase the server emitted both keys carrying IDENTICAL rows, so a
// reversed preference rendered correctly and was invisible; REMOVE (97a91d94)
// dropped `reviewers`, and from that point a reversed preference renders an
// EMPTY LIST rather than an error. That is why this is pinned rather than left
// to read correctly — the failure has no symptom at the call site.
func (e friendsEnvelope) rows() []friendRow {
	if len(e.Friends) > 0 {
		return e.Friends
	}
	if len(e.Reviewers) > 0 {
		return e.Reviewers
	}
	return e.Requests
}

func printFriendRows(data json.RawMessage, emptyMsg string, asJSON bool) error {
	return printFriendRowsMode(data, emptyMsg, asJSON, modeDirectory)
}

func printFriendRowsMode(data json.RawMessage, emptyMsg string, asJSON bool, mode rowMode) error {
	if asJSON {
		fmt.Println(string(data))
		return nil
	}

	var payload friendsEnvelope
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Println(string(data))
		return nil
	}

	rowsIn := payload.rows()
	if len(rowsIn) == 0 {
		fmt.Println(emptyMsg)
		return nil
	}

	// Email was dropped from every friendship payload upstream
	// (forge/proseforge#911) — a friendship is not a reason to hand someone's
	// address to whoever can see the row. Rendering the column anyway would
	// print an empty one forever, so the views now show what they actually
	// carry: a person's kind in the directory, nothing extra on a request.
	var header []string
	var rows [][]string

	if mode == modeDirectory {
		header = []string{"ID", "Name", "Kind", "Status"}
		for _, r := range rowsIn {
			id, name, _ := r.identity(mode)
			rows = append(rows, []string{id, truncate(name, 28), friendKind(r.UserKind), r.Status})
		}
	} else {
		header = []string{"Request ID", "Name", "Status"}
		for _, r := range rowsIn {
			id, name, _ := r.identity(mode)
			rows = append(rows, []string{id, truncate(name, 28), r.Status})
		}
	}

	printTable(header, rows)
	return nil
}

// friendKind renders userKind for a human reader. "system" is spelled out as AI
// because that is the distinction anyone scanning the list actually cares about
// — these rows only appear when --include-ai was passed.
func friendKind(kind string) string {
	switch kind {
	case "system":
		return "AI"
	case "user":
		return "person"
	case "":
		return ""
	default:
		return kind
	}
}

func runFriendsList(cmd *cobra.Command, args []string) error {
	svc, err := newFriendsService(cmd)
	if err != nil {
		return err
	}
	include := ""
	if ai, _ := cmd.Flags().GetBool("include-ai"); ai {
		include = "system"
	}
	data, err := svc.List(cmd.Context(), include)
	if err != nil {
		return err
	}
	return printFriendRows(data, "No friends yet. Use 'pfw friends search <term>' to find people.", isJSON(cmd))
}

func runFriendsSearch(cmd *cobra.Command, args []string) error {
	svc, err := newFriendsService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.Candidates(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	return printFriendRows(data, "No matches.", isJSON(cmd))
}

func runFriendsAdd(cmd *cobra.Command, args []string) error {
	svc, err := newFriendsService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Request(cmd.Context(), args[0]); err != nil {
		return err
	}
	fmt.Printf("Friend request sent to %s.\n", args[0])
	return nil
}

func runFriendsRemove(cmd *cobra.Command, args []string) error {
	svc, err := newFriendsService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Remove(cmd.Context(), args[0]); err != nil {
		return err
	}
	fmt.Printf("Friendship with %s ended.\n", args[0])
	return nil
}

func runFriendsRequests(cmd *cobra.Command, args []string) error {
	svc, err := newFriendsService(cmd)
	if err != nil {
		return err
	}

	outgoing, _ := cmd.Flags().GetBool("outgoing")
	var data json.RawMessage
	if outgoing {
		data, err = svc.Outgoing(cmd.Context())
	} else {
		data, err = svc.Incoming(cmd.Context())
	}
	if err != nil {
		return err
	}

	empty, mode := "No incoming friend requests.", modeIncoming
	if outgoing {
		empty, mode = "No outgoing friend requests.", modeOutgoing
	}

	// ⚑ #423 (@Wayland): this command shows ONE direction per invocation, and its
	// own help says "requests awaiting you, and ones you have sent" — which reads
	// as though one call covers both. Eight benches audited one direction of two
	// and reported themselves clear.
	//
	// 🛑 The dangerous shape is a CLEAN result: "No incoming friend requests" is
	// what you see whether or not a request you SENT is still pending, and the
	// global rule tells benches to check with this command. So the gap has to
	// announce itself at the moment of use rather than wait to be remembered.
	//
	// ⚠️ On stderr and in EVERY output mode, deliberately. Table-only would put
	// it exactly where a human would notice the truncation anyway and hide it
	// from the scripted agent that cannot — the same inversion as #414.
	other := "--outgoing (requests you have SENT)"
	if outgoing {
		other = "the default view (requests awaiting YOU)"
	}
	status("Showing ONE direction. Not shown: %s\n", other)

	return printFriendRowsMode(data, empty, isJSON(cmd), mode)
}

func runFriendsRespond(cmd *cobra.Command, args []string) error {
	accept, _ := cmd.Flags().GetBool("accept")
	decline, _ := cmd.Flags().GetBool("decline")
	if accept == decline {
		return fmt.Errorf("pass exactly one of --accept or --decline")
	}

	svc, err := newFriendsService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Respond(cmd.Context(), args[0], accept); err != nil {
		return err
	}

	verb := "declined"
	if accept {
		verb = "accepted"
	}
	fmt.Printf("Friend request %s %s.\n", args[0], verb)
	return nil
}
