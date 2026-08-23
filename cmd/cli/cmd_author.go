package main

import (
	"github.com/spf13/cobra"
)

// === author (group) ===

func newAuthorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "author",
		Short: "Author-facing operations",
	}
	cmd.AddCommand(
		newAuthorBookshelfCmd(),
	)
	return cmd
}

func newAuthorBookshelfCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bookshelf <handle>",
		Short: "Get an author's complete bookshelf (titles, taglines, covers, series)",
		Long: "Get an author's complete bookshelf in a single call — stories with " +
			"titles, taglines, slugs, cover URLs, section images, and series " +
			"membership. Replaces the N+1 pattern of story list + per-story images.\n\n" +
			"Works unauthenticated for published stories; authenticated calls also " +
			"include drafts and pitches.",
		Args: cobra.ExactArgs(1),
		RunE: runAuthorBookshelf,
	}
	cmd.Flags().String("q", "", "Search title, tagline, or series name")
	cmd.Flags().String("series", "", "Filter by series slug")
	cmd.Flags().String("status", "", "Filter by status: draft, published, pitch")
	cmd.Flags().String("sort", "", "Sort order: series (default), title, newest, oldest")
	cmd.Flags().Int("limit", 0, "Max results (default 50, max 100)")
	cmd.Flags().Int("offset", 0, "Pagination offset")
	return cmd
}

func runAuthorBookshelf(cmd *cobra.Command, args []string) error {
	svc, err := newAuthorService(cmd)
	if err != nil {
		return err
	}
	q, _ := cmd.Flags().GetString("q")
	series, _ := cmd.Flags().GetString("series")
	status, _ := cmd.Flags().GetString("status")
	sort, _ := cmd.Flags().GetString("sort")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")

	result, err := svc.Bookshelf(cmd.Context(), args[0], q, series, status, sort, limit, offset)
	if err != nil {
		return notFound(err, "author", args[0])
	}
	return printJSON(result)
}
