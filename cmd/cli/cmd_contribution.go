package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/claytonharbour/proseforge-workbench/internal/collaboration"
)

// The collaboration surface had sixteen MCP tools and no CLI at all, which made
// it the one area that could not be exercised without a live MCP session and a
// reconnect between every change (#226). These commands close that: the CLI and
// the MCP tools share internal/collaboration, so testing here tests the same
// code path the tools run.

func newContributionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contribution",
		Short: "Work on a story you don't own, and review work on one you do",
		Long: `Contributions are branch-based edits to someone else's story.

A grant (see 'pfw contributor') lets you edit; your writes land on your own
branch, never the owner's trunk. Mark it ready when it is worth looking at, and
the owner reviews, suggests, and merges.`,
	}

	list := &cobra.Command{
		Use:   "list <story-id>",
		Short: "List contributions on a story you own",
		Args:  cobra.ExactArgs(1),
		RunE:  runContributionList,
	}
	mine := &cobra.Command{
		Use:   "mine <story-id>",
		Short: "Show your own contribution to a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runContributionMine,
	}

	diff := &cobra.Command{
		Use:   "diff <story-id> <contribution-id>",
		Short: "Show a contribution's changes against trunk",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributionDiff,
	}

	ready := &cobra.Command{
		Use:   "ready <story-id> <contribution-id>",
		Short: "Submit a contribution for review",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributionReady,
	}

	readyForOwner := &cobra.Command{
		Use:   "ready-for-owner <story-id> <contribution-id>",
		Short: "Tell the owner a delegated review is finished",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributionReadyForOwner,
	}

	requestChanges := &cobra.Command{
		Use:   "request-changes <story-id> <contribution-id>",
		Short: "Ask the contributor to revise",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributionRequestChanges,
	}

	merge := &cobra.Command{
		Use:   "merge <story-id> <contribution-id>",
		Short: "Merge a contribution into your story",
		Long: `Merge a contribution into your story.

Resolve outstanding suggestions first — only accepted text reaches trunk.

PARTIAL ACCEPT: --selections takes a JSON object of file path -> true/false, so
you can take two of three files instead of all or nothing.

  pfw contribution merge <story> <contrib> \
      --selections '{"content/<sectionId>.md": true, "content/<other>.md": false}'

🛑 IT IS FAIL-CLOSED. If you supply --selections it must name EVERY changed path.
An absent path is a 400 invalid_selections, NOT an implicit accept — the opposite
of the feedback merge. List the contribution's changed paths first (contribution
diff) and name all of them.

Keys are FILE PATHS (content/<sectionId>.md), not bare section ids.

Omit --selections entirely to merge everything, which is the default and
unchanged.`,
		Args: cobra.ExactArgs(2),
		RunE: runContributionMerge,
	}
	merge.Flags().String("selections", "", "JSON object mapping each changed file path to true (accept) or false (reject). Must name EVERY changed path.")
	merge.Flags().Bool("stdin", false, "Read the selections JSON from stdin")

	discard := &cobra.Command{
		Use:   "discard <story-id> <contribution-id>",
		Short: "Throw away an unmerged contribution branch",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributionDiscard,
	}

	sync := &cobra.Command{
		Use:   "sync <story-id> <contribution-id>",
		Short: "Merge trunk into a contribution branch",
		Long: `Merge the owner's trunk into a contribution branch.

On conflict the backend reports the conflicted paths; pass resolved content with
--resolutions as a JSON object of path to content, or --stdin for the same JSON.`,
		Args: cobra.ExactArgs(2),
		RunE: runContributionSync,
	}
	sync.Flags().String("resolutions", "", "JSON object mapping conflicted paths to resolved content")
	sync.Flags().Bool("stdin", false, "Read the resolutions JSON from stdin")

	suggest := &cobra.Command{
		Use:   "suggest <story-id> <contribution-id> <section-id>",
		Short: "Suggest a section rewrite on a contribution",
		Long: `Write suggested section content to the contributor's branch.

Content comes from --content or, for anything longer than a line, --stdin.`,
		Args: cobra.ExactArgs(3),
		RunE: runContributionSuggest,
	}
	suggest.Flags().String("content", "", "Suggested section content")
	suggest.Flags().Bool("stdin", false, "Read the content from stdin")

	suggestions := &cobra.Command{
		Use:   "suggestions <story-id> <contribution-id>",
		Short: "List suggestions raised against a contribution",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributionSuggestions,
	}

	resolve := &cobra.Command{
		Use:   "resolve <story-id> <contribution-id> <suggestion-id>",
		Short: "Accept or reject one suggestion",
		Long: `Move one suggestion's status.

The status vocabulary is set by the API and is being simplified upstream, so it
is passed through rather than checked here — an unknown one comes back as an
API error naming it.`,
		Args: cobra.ExactArgs(3),
		RunE: runContributionResolve,
	}
	resolve.Flags().String("status", "", "New status (required)")
	_ = resolve.MarkFlagRequired("status")

	accounting := &cobra.Command{
		Use:   "accounting <id>",
		Short: "Show contribution totals for a story, or a whole series",
		Args:  cobra.ExactArgs(1),
		RunE:  runContributionAccounting,
	}
	accounting.Flags().Bool("series", false, "Treat the id as a series rather than a story")

	cmd.AddCommand(list, mine, diff, ready, readyForOwner, requestChanges,
		merge, discard, sync, suggest, suggestions, resolve, accounting)
	return cmd
}

func newContributorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contributor",
		Short: "Who may work on your story, and at what level",
		Long: `Grants control access to a story you own.

Capabilities: story:view, story:edit, story:review, story:merge, room:enter,
room:post. Granting is idempotent.`,
	}

	list := &cobra.Command{
		Use:   "list <entity-id>",
		Short: "List the grants on your story or series",
		Args:  cobra.ExactArgs(1),
		RunE:  runContributorList,
	}
	list.Flags().Bool("series", false, "Treat the id as a series rather than a story")

	grant := &cobra.Command{
		Use:   "grant <entity-id> <email>",
		Short: "Grant one capability to an existing user on a story or series",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributorGrant,
	}
	grant.Flags().String("capability", "story:edit", "Capability to grant")
	grant.Flags().Bool("series", false, "Treat the id as a series rather than a story")

	revoke := &cobra.Command{
		Use:   "revoke <entity-id> <email>",
		Short: "Revoke one capability",
		Args:  cobra.ExactArgs(2),
		RunE:  runContributorRevoke,
	}
	revoke.Flags().String("capability", "story:edit", "Capability to revoke")
	revoke.Flags().Bool("series", false, "Treat the id as a series rather than a story")

	leave := &cobra.Command{
		Use:   "leave <story-id>",
		Short: "Leave a story by removing your grants",
		Args:  cobra.ExactArgs(1),
		RunE:  runContributorLeave,
	}

	shared := &cobra.Command{
		Use:   "shared",
		Short: "List stories other people have shared with you",
		Args:  cobra.NoArgs,
		RunE:  runSharedStories,
	}

	cmd.AddCommand(list, grant, revoke, leave, shared)
	return cmd
}

func newCollaborationService(cmd *cobra.Command) (*collaboration.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return collaboration.NewService(client, collaboration.WithLogger(cliLogger)), nil
}

// readStdinOrFlag takes content from --stdin or the named flag. Large content
// through argv hits shell limits, so writes accept stdin as the primary path.
func readStdinOrFlag(cmd *cobra.Command, flag string) (string, error) {
	useStdin, _ := cmd.Flags().GetBool("stdin")
	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), nil
	}
	value, _ := cmd.Flags().GetString(flag)
	if value == "" {
		return "", fmt.Errorf("pass --%s or --stdin", flag)
	}
	return value, nil
}

func runContributionList(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.Contributions(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(result)
	}
	if result.Contributions == nil || len(*result.Contributions) == 0 {
		fmt.Println("No contributions on this story.")
		return nil
	}
	var rows [][]string
	for _, c := range *result.Contributions {
		rows = append(rows, []string{deref(c.Id), deref(c.ContributorId), deref(c.Status), deref(c.GitBranch)})
	}
	printTable([]string{"ID", "Contributor", "Status", "Branch"}, rows)
	return nil
}

func runContributionMine(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.Mine(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runContributionDiff(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.Diff(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runContributionReady(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.Ready(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	status("Contribution %s marked ready.\n", args[1])
	return printJSON(result)
}

func runContributionReadyForOwner(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	if err := svc.ReadyForOwner(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}
	fmt.Printf("Owner notified that %s is reviewed.\n", args[1])
	return nil
}

func runContributionRequestChanges(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.RequestChanges(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	status("Changes requested on %s.\n", args[1])
	return printJSON(result)
}

func runContributionMerge(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}

	// nil selections = take everything, the long-standing behaviour. An EMPTY
	// map is deliberately different: it names no paths, and against a fail-closed
	// backend that is a 400, not "accept nothing". Keeping nil and {} distinct is
	// what lets the server's rule be the one that decides.
	var selections map[string]bool
	useStdin, _ := cmd.Flags().GetBool("stdin")
	raw, _ := cmd.Flags().GetString("selections")
	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		raw = string(data)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &selections); err != nil {
			return fmt.Errorf("parse selections JSON: %w", err)
		}
	}

	result, err := svc.Merge(cmd.Context(), args[0], args[1], selections)
	if err != nil {
		return err
	}
	if selections == nil {
		status("Contribution %s merged (all changed files).\n", args[1])
	} else {
		// ⚑ Say what was taken. The contributor cannot see their own diff
		// (proseforge#765), so this line is where anyone learns what landed.
		accepted := make([]string, 0, len(selections))
		for path, take := range selections {
			if take {
				accepted = append(accepted, path)
			}
		}
		sort.Strings(accepted)
		status("Contribution %s merged — %d of %d paths accepted:\n", args[1], len(accepted), len(selections))
		for _, p := range accepted {
			status("  + %s\n", p)
		}
	}
	return printJSON(result)
}

func runContributionDiscard(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Discard(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}
	fmt.Printf("Contribution %s discarded.\n", args[1])
	return nil
}

func runContributionSync(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}

	var resolutions map[string]string
	useStdin, _ := cmd.Flags().GetBool("stdin")
	raw, _ := cmd.Flags().GetString("resolutions")
	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		raw = string(data)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &resolutions); err != nil {
			return fmt.Errorf("parse resolutions JSON: %w", err)
		}
	}

	result, err := svc.Sync(cmd.Context(), args[0], args[1], resolutions)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runContributionSuggest(cmd *cobra.Command, args []string) error {
	content, err := readStdinOrFlag(cmd, "content")
	if err != nil {
		return err
	}
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.SuggestSection(cmd.Context(), args[0], args[1], args[2], content)
	if err != nil {
		return err
	}
	status("Suggestion written to %s.\n", args[2])
	return printJSON(result)
}

func runContributionSuggestions(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.Suggestions(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	return printSuggestionRows(data)
}

// printSuggestionRows renders the review surface's per-section bucketing as one
// flat table. The payload shape is still moving upstream, so an unrecognised
// one falls back to raw JSON rather than printing a confidently empty table.
func printSuggestionRows(data json.RawMessage) error {
	var payload struct {
		TotalSuggestions int  `json:"totalSuggestions"`
		HasConflicts     bool `json:"hasConflicts"`
		Sections         []struct {
			SectionID   string `json:"sectionId"`
			SectionName string `json:"sectionName"`
			Suggestions []struct {
				ID        string `json:"id"`
				Status    string `json:"status"`
				Source    string `json:"source"`
				CanApply  bool   `json:"canApply"`
				Suggested string `json:"suggested"`
			} `json:"suggestions"`
		} `json:"sections"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Println(string(data))
		return nil
	}
	if payload.TotalSuggestions == 0 {
		fmt.Println("No suggestions on this contribution.")
		return nil
	}

	var rows [][]string
	for _, sec := range payload.Sections {
		for _, s := range sec.Suggestions {
			apply := ""
			if s.CanApply {
				apply = "yes"
			}
			rows = append(rows, []string{
				s.ID, truncate(sec.SectionName, 20), s.Status, s.Source, apply,
				truncate(s.Suggested, 40),
			})
		}
	}
	if len(rows) == 0 {
		fmt.Println(string(data))
		return nil
	}
	printTable([]string{"ID", "Section", "Status", "Source", "Apply", "Suggested"}, rows)
	if payload.HasConflicts {
		status("This contribution has conflicts — sync before merging.\n")
	}
	return nil
}

func runContributionResolve(cmd *cobra.Command, args []string) error {
	statusValue, _ := cmd.Flags().GetString("status")
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.ResolveSuggestion(cmd.Context(), args[0], args[1], args[2], statusValue)
	if err != nil {
		return err
	}
	status("Suggestion %s set to %s.\n", args[2], statusValue)
	fmt.Println(string(data))
	return nil
}

func runContributionAccounting(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	isSeries, _ := cmd.Flags().GetBool("series")
	if isSeries {
		result, err := svc.SeriesAccounting(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}
	result, err := svc.StoryAccounting(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runContributorList(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	series, _ := cmd.Flags().GetBool("series")
	var result *gen.HandlersGrantListResponse
	if series {
		result, err = svc.SeriesGrants(cmd.Context(), args[0])
	} else {
		result, err = svc.Grants(cmd.Context(), args[0])
	}
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(result)
	}
	if result.Grants == nil || len(*result.Grants) == 0 {
		fmt.Println("Nobody has been granted access to this entity.")
		return nil
	}
	var rows [][]string
	for _, g := range *result.Grants {
		rows = append(rows, []string{deref(g.Email), deref(g.Capability), deref(g.PrincipalId)})
	}
	printTable([]string{"Email", "Capability", "Principal"}, rows)
	return nil
}

func runContributorGrant(cmd *cobra.Command, args []string) error {
	capability, _ := cmd.Flags().GetString("capability")
	series, _ := cmd.Flags().GetBool("series")
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	var grantErr error
	if series {
		_, grantErr = svc.SeriesGrant(cmd.Context(), args[0], args[1], capability)
	} else {
		_, grantErr = svc.Grant(cmd.Context(), args[0], args[1], capability)
	}
	if grantErr != nil {
		return grantErr
	}
	fmt.Printf("Granted %s to %s.\n", capability, args[1])
	return nil
}

func runContributorRevoke(cmd *cobra.Command, args []string) error {
	capability, _ := cmd.Flags().GetString("capability")
	series, _ := cmd.Flags().GetBool("series")
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	var revokeErr error
	if series {
		revokeErr = svc.SeriesRevoke(cmd.Context(), args[0], args[1], capability)
	} else {
		revokeErr = svc.Revoke(cmd.Context(), args[0], args[1], capability)
	}
	if revokeErr != nil {
		return revokeErr
	}
	fmt.Printf("Revoked %s from %s.\n", capability, args[1])
	return nil
}

func runContributorLeave(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Leave(cmd.Context(), args[0]); err != nil {
		return err
	}
	fmt.Println("Story access left.")
	return nil
}

func runSharedStories(cmd *cobra.Command, args []string) error {
	svc, err := newCollaborationService(cmd)
	if err != nil {
		return err
	}
	result, err := svc.SharedStories(cmd.Context())
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(result)
	}
	if result.Stories == nil || len(*result.Stories) == 0 {
		fmt.Println("No stories have been shared with you.")
		return nil
	}
	var rows [][]string
	for _, s := range *result.Stories {
		// Capabilities is the column that answers "what am I allowed to do
		// here?", which is the only reason to run this command.
		caps := ""
		if s.Capabilities != nil {
			caps = strings.Join(*s.Capabilities, ",")
		}
		rows = append(rows, []string{
			deref(s.StoryId), truncate(deref(s.Title), 30), deref(s.OwnerName), truncate(caps, 34),
		})
	}
	printTable([]string{"Story ID", "Title", "Owner", "You may"}, rows)
	return nil
}
