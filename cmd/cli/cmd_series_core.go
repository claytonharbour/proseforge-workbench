package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Series core + membership commands (#228). Thin cobra wrappers over the
// existing series.Service methods; reuse generated types, no hand-rolled structs.

// renderSeries prints a single series across the -o table|json|brief formats.
func renderSeries(cmd *cobra.Command, s *api.Series) error {
	if isJSON(cmd) {
		return printJSON(s)
	}
	if isBrief(cmd) {
		printBrief([][]string{{deref(s.Id), deref(s.Name)}})
		return nil
	}
	fmt.Printf("ID:          %s\n", deref(s.Id))
	fmt.Printf("Name:        %s\n", deref(s.Name))
	fmt.Printf("Status:      %s\n", deref(s.Status))
	fmt.Printf("Genre:       %s\n", deref(s.GenreName))
	fmt.Printf("Tone:        %s\n", deref(s.ToneName))
	fmt.Printf("Stories:     %d\n", derefInt(s.StoryCount))
	if s.Description != nil && *s.Description != "" {
		fmt.Printf("Description: %s\n", *s.Description)
	}
	return nil
}

// === series get ===

func newSeriesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <series-id>",
		Short: "Get a series' details",
		Args:  cobra.ExactArgs(1),
		RunE:  runSeriesGet,
	}
}

func runSeriesGet(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	s, err := svc.Get(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "series", args[0])
	}
	return renderSeries(cmd, s)
}

// === series create ===

func newSeriesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new series",
		Args:  cobra.NoArgs,
		RunE:  runSeriesCreate,
	}
	cmd.Flags().String("name", "", "Series name")
	cmd.Flags().String("description", "", "Series description")
	cmd.Flags().String("genre-id", "", "Genre ID")
	cmd.Flags().String("tone-id", "", "Tone ID")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runSeriesCreate(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	req := api.CreateSeriesReq{Name: &name}
	if v, _ := cmd.Flags().GetString("description"); v != "" {
		req.Description = &v
	}
	if v, _ := cmd.Flags().GetString("genre-id"); v != "" {
		req.GenreId = &v
	}
	if v, _ := cmd.Flags().GetString("tone-id"); v != "" {
		req.ToneId = &v
	}

	s, err := svc.Create(cmd.Context(), req)
	if err != nil {
		return err
	}
	status("Series created.")
	return renderSeries(cmd, s)
}

// === series update ===

func newSeriesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <series-id>",
		Short: "Update a series' metadata",
		Long:  "Update a series' metadata. Only the flags you pass are changed.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSeriesUpdate,
	}
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("genre-id", "", "New genre ID")
	cmd.Flags().String("tone-id", "", "New tone ID")
	return cmd
}

func runSeriesUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}

	req := api.UpdateSeriesReq{}
	if cmd.Flags().Changed("name") {
		v, _ := cmd.Flags().GetString("name")
		req.Name = &v
	}
	if cmd.Flags().Changed("description") {
		v, _ := cmd.Flags().GetString("description")
		req.Description = &v
	}
	if cmd.Flags().Changed("genre-id") {
		v, _ := cmd.Flags().GetString("genre-id")
		req.GenreId = &v
	}
	if cmd.Flags().Changed("tone-id") {
		v, _ := cmd.Flags().GetString("tone-id")
		req.ToneId = &v
	}
	if req.Name == nil && req.Description == nil && req.GenreId == nil && req.ToneId == nil {
		return fmt.Errorf("at least one of --name, --description, --genre-id, --tone-id is required")
	}

	if err := svc.Update(cmd.Context(), args[0], req); err != nil {
		return notFound(err, "series", args[0])
	}
	fmt.Println("Series updated.")
	return nil
}

// === series archive ===

func newSeriesArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <series-id>",
		Short: "Archive a series",
		Args:  cobra.ExactArgs(1),
		RunE:  runSeriesArchive,
	}
}

func runSeriesArchive(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Archive(cmd.Context(), args[0]); err != nil {
		return notFound(err, "series", args[0])
	}
	fmt.Println("Series archived.")
	return nil
}

// === series plan ===

func newSeriesPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <series-id>",
		Short: "Seed a Story Forge planning session for the next book in a series",
		Args:  cobra.ExactArgs(1),
		RunE:  runSeriesPlan,
	}
	cmd.Flags().Int("book-number", -1, "Override book number (omit to auto-detect next)")
	cmd.Flags().StringSlice("characters", nil, "Character slugs to include (default: all)")
	cmd.Flags().String("notes", "", "Author notes injected into AI context")
	return cmd
}

func runSeriesPlan(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}

	req := api.PlanStoryReq{}
	if cmd.Flags().Changed("book-number") {
		n, _ := cmd.Flags().GetInt("book-number")
		req.BookNumber = &n
	}
	if cmd.Flags().Changed("characters") {
		chars, _ := cmd.Flags().GetStringSlice("characters")
		req.IncludeCharacters = &chars
	}
	if v, _ := cmd.Flags().GetString("notes"); v != "" {
		req.Notes = &v
	}

	resp, err := svc.PlanStory(cmd.Context(), args[0], req)
	if err != nil {
		return notFound(err, "series", args[0])
	}

	if isJSON(cmd) {
		return printJSON(resp)
	}
	fmt.Printf("Planning session seeded.\n")
	fmt.Printf("Session ID: %s\n", deref(resp.SessionId))
	fmt.Printf("Series ID:  %s\n", deref(resp.SeriesId))
	return nil
}

// === series stories-add / stories-remove / stories-reorder ===

func newSeriesStoriesAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stories-add <series-id> <story-id>",
		Short: "Link an existing story to a series",
		Args:  cobra.ExactArgs(2),
		RunE:  runSeriesStoriesAdd,
	}
}

func runSeriesStoriesAdd(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.AddStory(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}
	fmt.Println("Story added to series.")
	return nil
}

func newSeriesStoriesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stories-remove <series-id> <story-id>",
		Short: "Unlink a story from a series",
		Args:  cobra.ExactArgs(2),
		RunE:  runSeriesStoriesRemove,
	}
}

func runSeriesStoriesRemove(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.RemoveStory(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}
	fmt.Println("Story removed from series.")
	return nil
}

func newSeriesStoriesReorderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stories-reorder <series-id> <story-id>...",
		Short: "Set the order of stories in a series",
		Long:  "Set the display order of stories in a series by listing the story IDs in the desired order.",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runSeriesStoriesReorder,
	}
}

func runSeriesStoriesReorder(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.ReorderSeriesStories(cmd.Context(), args[0], args[1:]); err != nil {
		return err
	}
	fmt.Println("Series stories reordered.")
	return nil
}
