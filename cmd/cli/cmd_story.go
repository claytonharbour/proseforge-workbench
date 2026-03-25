package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

func newStoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Story operations",
	}
	cmd.AddCommand(
		newStoryListCmd(),
		newStoryGetCmd(),
		newStoryCreateCmd(),
		newStoryUpdateCmd(),
		newStoryPublishCmd(),
		newStoryUnpublishCmd(),
		newStoryExportCmd(),
		newStorySectionGroupCmd(),
		newStoryQualityCmd(),
		newStoryAssessCmd(),
		newStoryInsightsCmd(),
	)
	return cmd
}

// === story list ===

func newStoryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stories",
		RunE:  runStoryList,
	}
	cmd.Flags().String("status", "", "Filter by status (published, unpublished, generating, failed)")
	cmd.Flags().Int("limit", 25, "Max results (1-100)")
	return cmd
}

func runStoryList(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	filterStatus, _ := cmd.Flags().GetString("status")
	limit, _ := cmd.Flags().GetInt("limit")

	params := &gen.GetStoriesParams{Limit: &limit}
	if filterStatus != "" {
		params.Status = &filterStatus
	}

	result, err := svc.List(cmd.Context(), params)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(result)
	}

	if result.Stories == nil || len(*result.Stories) == 0 {
		fmt.Println("No stories found.")
		return nil
	}

	status("Stories: %d of %d", len(*result.Stories), derefInt(result.Total))
	fmt.Println()

	var rows [][]string
	for _, s := range *result.Stories {
		rows = append(rows, []string{
			deref(s.Id),
			truncate(deref(s.Title), 40),
			deref(s.Status),
			deref(s.GenreName),
		})
	}

	printTable([]string{"ID", "Title", "Status", "Genre"}, rows)
	return nil
}

// === story get ===

func newStoryGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get story details",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryGet,
	}
}

func runStoryGet(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	story, err := svc.Get(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(story)
	}

	fmt.Printf("Title:   %s\n", deref(story.Title))
	fmt.Printf("ID:      %s\n", deref(story.Id))
	fmt.Printf("Status:  %s\n", deref(story.Status))
	fmt.Printf("Genre:   %s\n", deref(story.GenreName))
	fmt.Printf("Summary: %s\n", deref(story.Summary))

	if story.Sections != nil && len(*story.Sections) > 0 {
		fmt.Printf("\nSections: %d\n\n", len(*story.Sections))
		var rows [][]string
		for _, sec := range *story.Sections {
			rows = append(rows, []string{
				deref(sec.Id),
				truncate(deref(sec.Name), 40),
				deref(sec.Status),
			})
		}
		printTable([]string{"ID", "Name", "Status"}, rows)
	}

	return nil
}

// === story section (group) ===

func newStorySectionGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "section",
		Short: "Section operations",
	}
	cmd.AddCommand(
		newStorySectionListCmd(),
		newStorySectionGetCmd(),
		newStorySectionCreateCmd(),
		newStorySectionWriteCmd(),
	)
	return cmd
}

func newStorySectionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <story-id>",
		Short: "List sections in a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStorySectionList,
	}
}

func runStorySectionList(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	story, err := svc.Get(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(story.Sections)
	}

	if story.Sections == nil || len(*story.Sections) == 0 {
		fmt.Println("No sections.")
		return nil
	}

	var rows [][]string
	for i, sec := range *story.Sections {
		rows = append(rows, []string{
			deref(sec.Id),
			fmt.Sprintf("%d", i),
			truncate(deref(sec.Name), 40),
			deref(sec.Status),
		})
	}
	printTable([]string{"ID", "Order", "Name", "Status"}, rows)
	return nil
}

func newStorySectionGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <story-id> <section-id>",
		Short: "Get a single section's content",
		Args:  cobra.ExactArgs(2),
		RunE:  runStorySection,
	}
}

func runStorySection(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	data, err := svc.GetSection(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}

	// For table mode, extract and print just the content
	var section struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &section); err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("# %s\n\n%s\n", section.Name, section.Content)
	return nil
}

// === story export ===

func newStoryExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <id>",
		Short: "Export story to stdout (json or markdown)",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryExport,
	}
	cmd.Flags().String("format", "json", "Export format: json, markdown, pdf")
	return cmd
}

func runStoryExport(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	content, err := svc.Export(cmd.Context(), args[0], format)
	if err != nil {
		return err
	}

	fmt.Print(content)
	return nil
}
