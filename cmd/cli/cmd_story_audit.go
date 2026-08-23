package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Story-side parity holes (#234): delete, regenerate tagline/title,
// update-visibility, section delete/reorder, version restore. Thin wrappers
// over existing story.Service methods.

// === story delete ===

func newStoryDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <story-id>",
		Short: "Delete a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryDelete,
	}
}

func runStoryDelete(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Delete(cmd.Context(), args[0]); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println("Story deleted.")
	return nil
}

// === story update-visibility ===

func newStoryUpdateVisibilityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-visibility <story-id>",
		Short: "Change a published story's visibility",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryUpdateVisibility,
	}
	cmd.Flags().String("visibility", "", "Visibility: public or members")
	_ = cmd.MarkFlagRequired("visibility")
	return cmd
}

func runStoryUpdateVisibility(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	visibility, _ := cmd.Flags().GetString("visibility")
	if err := svc.UpdateVisibility(cmd.Context(), args[0], visibility); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Printf("Visibility updated to %s.\n", visibility)
	return nil
}

// === story regenerate (tagline | title) ===

func newStoryRegenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate",
		Short: "Regenerate a story's tagline or title (AI)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "tagline <story-id>",
			Short: "Regenerate the story tagline",
			Args:  cobra.ExactArgs(1),
			RunE:  runStoryRegenerateTagline,
		},
		&cobra.Command{
			Use:   "title <story-id>",
			Short: "Regenerate the story title",
			Args:  cobra.ExactArgs(1),
			RunE:  runStoryRegenerateTitle,
		},
	)
	return cmd
}

func runStoryRegenerateTagline(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.RegenerateTagline(cmd.Context(), args[0]); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println("Tagline regenerated.")
	return nil
}

func runStoryRegenerateTitle(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.RegenerateTitle(cmd.Context(), args[0]); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println("Title regenerated.")
	return nil
}

// === story section delete / reorder (added to the section subgroup) ===

func newStorySectionDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <story-id> <section-id>",
		Short: "Delete a section",
		Args:  cobra.ExactArgs(2),
		RunE:  runStorySectionDelete,
	}
}

func runStorySectionDelete(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.DeleteSection(cmd.Context(), args[0], args[1]); err != nil {
		return notFound(err, "section", args[1])
	}
	fmt.Println("Section deleted.")
	return nil
}

func newStorySectionReorderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reorder <story-id> <section-id>",
		Short: "Move a section to a new position",
		Args:  cobra.ExactArgs(2),
		RunE:  runStorySectionReorder,
	}
	cmd.Flags().Int("order", -1, "New 0-indexed position")
	_ = cmd.MarkFlagRequired("order")
	return cmd
}

func runStorySectionReorder(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	order, _ := cmd.Flags().GetInt("order")
	if order < 0 {
		return fmt.Errorf("--order must be 0 or greater")
	}
	data, err := svc.ReorderSection(cmd.Context(), args[0], args[1], order)
	if err != nil {
		return notFound(err, "section", args[1])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Section reordered.")
	return nil
}

// === story version restore (added to the version subgroup) ===

func newStoryVersionRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <story-id> <sha>",
		Short: "Restore a story to a previous version",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryVersionRestore,
	}
}

func runStoryVersionRestore(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.RestoreVersion(cmd.Context(), args[0], args[1])
	if err != nil {
		return notFound(err, "story", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Version restored.")
	return nil
}
