package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

func newSeriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "series",
		Short: "Series operations",
		// Series-contributor management lives under `pfw contributor --series`,
		// not here, and calls the same service methods as the MCP
		// series_contributor_* tools. Without this signpost an agent reading
		// only `pfw series --help` concludes the capability is missing (#349).
		Long: `Series operations.

Granting and revoking access to a series is handled by the contributor command
with the --series flag, not by a subcommand here:

  pfw contributor grant  <series-id> <email> --series --capability room:enter
  pfw contributor list   <series-id>         --series
  pfw contributor revoke <series-id> <email> --series --capability room:enter

Both parties must already be accepted friends, or the grant returns 400.
Creating a series requires a subscription; a bench account gets 403 tier_required.`,
	}
	cmd.AddCommand(newSeriesListCmd())
	cmd.AddCommand(newSeriesGetCmd())
	cmd.AddCommand(newSeriesCreateCmd())
	cmd.AddCommand(newSeriesUpdateCmd())
	cmd.AddCommand(newSeriesArchiveCmd())
	cmd.AddCommand(newSeriesPlanCmd())
	cmd.AddCommand(newSeriesStoriesCmd())
	cmd.AddCommand(newSeriesStoriesAddCmd())
	cmd.AddCommand(newSeriesStoriesRemoveCmd())
	cmd.AddCommand(newSeriesStoriesReorderCmd())
	cmd.AddCommand(newSeriesWorldCmd())
	cmd.AddCommand(newSeriesCharacterCmd())
	cmd.AddCommand(newSeriesTimelineCmd())
	return cmd
}

func newSeriesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your series",
		Long: "List the series you own, so you can find a series ID without " +
			"leaving the CLI. Supports -o table|json|brief.",
		Args: cobra.NoArgs,
		RunE: runSeriesList,
	}
}

func newSeriesStoriesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stories <series-id>",
		Short: "List stories in a series",
		Long: "List the stories linked to a series, by actual series membership " +
			"(not title heuristics). Supports -o json|table|brief.",
		Args: cobra.ExactArgs(1),
		RunE: runSeriesStories,
	}
}

func runSeriesList(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}

	list, err := svc.List(cmd.Context())
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(list)
	}

	var series []gen.HandlersSeriesResponse
	if list.Series != nil {
		series = *list.Series
	}
	if len(series) == 0 {
		if !isBrief(cmd) {
			fmt.Println("No series found.")
		}
		return nil
	}

	if isBrief(cmd) {
		var rows [][]string
		for _, s := range series {
			rows = append(rows, []string{deref(s.Id), deref(s.Name)})
		}
		printBrief(rows)
		return nil
	}

	status("Series: %d", derefInt(list.Total))
	fmt.Println()

	var rows [][]string
	for _, s := range series {
		rows = append(rows, []string{
			deref(s.Id),
			deref(s.Name),
			deref(s.Status),
			fmt.Sprintf("%d", derefInt(s.StoryCount)),
		})
	}
	printTable([]string{"ID", "Name", "Status", "Stories"}, rows)
	return nil
}

func runSeriesStories(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}

	// ListStories returns the raw JSON body; the table view unmarshals it into
	// the generated response type (HandlersStoryListInSeriesResponse) rather than
	// a hand-rolled struct, so the CLI stays in sync with the Swagger schema.
	data, err := svc.ListStories(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "series", args[0])
	}

	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}

	var resp gen.HandlersStoryListInSeriesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		// Unexpected shape — emit the raw payload rather than swallowing it.
		fmt.Println(string(data))
		return nil
	}

	var stories []gen.HandlersStoryInSeriesResponse
	if resp.Stories != nil {
		stories = *resp.Stories
	}
	if len(stories) == 0 {
		if !isBrief(cmd) {
			fmt.Println("No stories found in this series.")
		}
		return nil
	}

	if isBrief(cmd) {
		var rows [][]string
		for _, s := range stories {
			rows = append(rows, []string{deref(s.StoryId), deref(s.Title)})
		}
		printBrief(rows)
		return nil
	}

	status("Stories: %d", derefInt(resp.Total))
	fmt.Println()

	// Title is left untruncated — printTable auto-sizes the column to content,
	// so full titles stay distinguishable (e.g. "...: Part 1" vs "...: Part 2").
	var rows [][]string
	for _, s := range stories {
		rows = append(rows, []string{
			fmt.Sprintf("%d", derefInt(s.BookNumber)),
			deref(s.StoryId),
			deref(s.Title),
			deref(s.Status),
		})
	}
	printTable([]string{"Book", "ID", "Title", "Status"}, rows)
	return nil
}
