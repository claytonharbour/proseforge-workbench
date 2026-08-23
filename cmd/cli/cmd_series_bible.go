package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Series bible commands (#229): world doc, characters, and timeline.
// Thin cobra wrappers over series.Service. World/timeline return raw JSON
// (passthrough, mirrors the MCP tools); characters use the generated types.

// contentFromFlags reads body content from --stdin or --content.
func contentFromFlags(cmd *cobra.Command) (string, error) {
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}
	c, _ := cmd.Flags().GetString("content")
	return c, nil
}

// === series world ===

func newSeriesWorldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "world",
		Short: "Series world doc (direction, doctrine, canon)",
		Long: "Read and update the series world doc. The world doc holds series " +
			"direction — tone, core truths, revision doctrine, anti-patterns — and " +
			"is the read agents should make before any series work.",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get <series-id>",
			Short: "Read the series world doc",
			Args:  cobra.ExactArgs(1),
			RunE:  runSeriesWorldGet,
		},
		func() *cobra.Command {
			c := &cobra.Command{
				Use:   "update <series-id>",
				Short: "Update the series world doc",
				Args:  cobra.ExactArgs(1),
				RunE:  runSeriesWorldUpdate,
			}
			c.Flags().Bool("stdin", false, "Read the world doc from stdin")
			c.Flags().String("content", "", "World doc content (markdown)")
			return c
		}(),
	)
	return cmd
}

func runSeriesWorldGet(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.GetWorld(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "series", args[0])
	}
	fmt.Println(string(data))
	return nil
}

func runSeriesWorldUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	content, err := contentFromFlags(cmd)
	if err != nil {
		return err
	}
	if content == "" {
		return fmt.Errorf("content is required: use --stdin or --content")
	}
	if err := svc.UpdateWorld(cmd.Context(), args[0], content); err != nil {
		return notFound(err, "series", args[0])
	}
	fmt.Println("World doc updated.")
	return nil
}

// === series character ===

func renderCharacter(cmd *cobra.Command, c *api.Character) error {
	if isJSON(cmd) {
		return printJSON(c)
	}
	if isBrief(cmd) {
		printBrief([][]string{{deref(c.Slug), deref(c.Name)}})
		return nil
	}
	fmt.Printf("Slug:    %s\n", deref(c.Slug))
	fmt.Printf("Name:    %s\n", deref(c.Name))
	fmt.Printf("Role:    %s\n", deref(c.Role))
	fmt.Printf("Status:  %s\n", deref(c.Status))
	if c.Profile != nil && *c.Profile != "" {
		fmt.Printf("\n%s\n", *c.Profile)
	}
	return nil
}

func newSeriesCharacterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "character",
		Short: "Series characters",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list <series-id>",
			Short: "List characters in a series",
			Args:  cobra.ExactArgs(1),
			RunE:  runSeriesCharacterList,
		},
		&cobra.Command{
			Use:   "get <series-id> <slug>",
			Short: "Get a character by slug",
			Args:  cobra.ExactArgs(2),
			RunE:  runSeriesCharacterGet,
		},
		newSeriesCharacterCreateCmd(),
		newSeriesCharacterUpdateCmd(),
		&cobra.Command{
			Use:   "delete <series-id> <slug>",
			Short: "Delete a character",
			Args:  cobra.ExactArgs(2),
			RunE:  runSeriesCharacterDelete,
		},
	)
	return cmd
}

func runSeriesCharacterList(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	list, err := svc.ListCharacters(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "series", args[0])
	}
	if isJSON(cmd) {
		return printJSON(list)
	}
	var chars []api.Character
	if list.Characters != nil {
		chars = *list.Characters
	}
	if len(chars) == 0 {
		if !isBrief(cmd) {
			fmt.Println("No characters found.")
		}
		return nil
	}
	if isBrief(cmd) {
		var rows [][]string
		for _, c := range chars {
			rows = append(rows, []string{deref(c.Slug), deref(c.Name)})
		}
		printBrief(rows)
		return nil
	}
	status("Characters: %d", derefInt(list.Total))
	fmt.Println()
	var rows [][]string
	for _, c := range chars {
		rows = append(rows, []string{deref(c.Slug), deref(c.Name), deref(c.Role), deref(c.Status)})
	}
	printTable([]string{"Slug", "Name", "Role", "Status"}, rows)
	return nil
}

func runSeriesCharacterGet(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	c, err := svc.GetCharacter(cmd.Context(), args[0], args[1])
	if err != nil {
		return notFound(err, "character", args[1])
	}
	return renderCharacter(cmd, c)
}

func newSeriesCharacterCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <series-id>",
		Short: "Create a character in a series",
		Args:  cobra.ExactArgs(1),
		RunE:  runSeriesCharacterCreate,
	}
	cmd.Flags().String("name", "", "Character name")
	cmd.Flags().String("role", "", "Character role")
	cmd.Flags().String("status", "", "Character status")
	cmd.Flags().String("profile", "", "Character profile (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the profile from stdin")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runSeriesCharacterCreate(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	name, _ := cmd.Flags().GetString("name")
	req := api.CreateCharacterReq{Name: &name}
	if v, _ := cmd.Flags().GetString("role"); v != "" {
		req.Role = &v
	}
	if v, _ := cmd.Flags().GetString("status"); v != "" {
		req.Status = &v
	}
	profile, _ := cmd.Flags().GetString("profile")
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		profile = string(data)
	}
	if profile != "" {
		req.Profile = &profile
	}
	c, err := svc.CreateCharacter(cmd.Context(), args[0], req)
	if err != nil {
		return notFound(err, "series", args[0])
	}
	status("Character created.")
	return renderCharacter(cmd, c)
}

func newSeriesCharacterUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <series-id> <slug>",
		Short: "Update a character",
		Long:  "Update a character. Only the flags you pass are changed.",
		Args:  cobra.ExactArgs(2),
		RunE:  runSeriesCharacterUpdate,
	}
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("profile", "", "New profile (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the profile from stdin")
	return cmd
}

func runSeriesCharacterUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	req := api.UpdateCharacterReq{}
	if cmd.Flags().Changed("name") {
		v, _ := cmd.Flags().GetString("name")
		req.Name = &v
	}
	if cmd.Flags().Changed("role") {
		v, _ := cmd.Flags().GetString("role")
		req.Role = &v
	}
	if cmd.Flags().Changed("status") {
		v, _ := cmd.Flags().GetString("status")
		req.Status = &v
	}
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		s := string(data)
		req.Profile = &s
	} else if cmd.Flags().Changed("profile") {
		v, _ := cmd.Flags().GetString("profile")
		req.Profile = &v
	}
	if req.Name == nil && req.Role == nil && req.Status == nil && req.Profile == nil {
		return fmt.Errorf("at least one of --name, --role, --status, --profile is required")
	}
	if err := svc.UpdateCharacter(cmd.Context(), args[0], args[1], req); err != nil {
		return notFound(err, "character", args[1])
	}
	fmt.Println("Character updated.")
	return nil
}

func runSeriesCharacterDelete(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.DeleteCharacter(cmd.Context(), args[0], args[1]); err != nil {
		return notFound(err, "character", args[1])
	}
	fmt.Println("Character deleted.")
	return nil
}

// === series timeline ===

func newSeriesTimelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Series canon timeline",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get <series-id>",
			Short: "Get the canon timeline for a series",
			Args:  cobra.ExactArgs(1),
			RunE:  runSeriesTimelineGet,
		},
		&cobra.Command{
			Use:   "sections <series-id>",
			Short: "List timeline sections (slugs, titles, order)",
			Args:  cobra.ExactArgs(1),
			RunE:  runSeriesTimelineSections,
		},
		&cobra.Command{
			Use:   "reorder <series-id> <slug>...",
			Short: "Set the order of timeline sections by slug",
			Args:  cobra.MinimumNArgs(2),
			RunE:  runSeriesTimelineReorder,
		},
		newSeriesTimelineSectionCmd(),
	)
	return cmd
}

func runSeriesTimelineGet(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.GetTimeline(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "series", args[0])
	}
	fmt.Println(string(data))
	return nil
}

func runSeriesTimelineSections(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.ListTimelineSections(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "series", args[0])
	}
	fmt.Println(string(data))
	return nil
}

func runSeriesTimelineReorder(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.ReorderTimelineSections(cmd.Context(), args[0], args[1:]); err != nil {
		return err
	}
	fmt.Println("Timeline reordered.")
	return nil
}

func newSeriesTimelineSectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "section",
		Short: "A single timeline section (per-book events, by slug)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get <series-id> <slug>",
			Short: "Get a timeline section by slug",
			Args:  cobra.ExactArgs(2),
			RunE:  runSeriesTimelineSectionGet,
		},
		newSeriesTimelineSectionUpdateCmd(),
		&cobra.Command{
			Use:   "delete <series-id> <slug>",
			Short: "Delete a timeline section",
			Args:  cobra.ExactArgs(2),
			RunE:  runSeriesTimelineSectionDelete,
		},
	)
	return cmd
}

func runSeriesTimelineSectionGet(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.GetTimelineSection(cmd.Context(), args[0], args[1])
	if err != nil {
		return notFound(err, "timeline section", args[1])
	}
	fmt.Println(string(data))
	return nil
}

func newSeriesTimelineSectionUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <series-id> <slug>",
		Short: "Create or update a timeline section",
		Args:  cobra.ExactArgs(2),
		RunE:  runSeriesTimelineSectionUpdate,
	}
	cmd.Flags().String("title", "", "Section title")
	cmd.Flags().Bool("stdin", false, "Read the section content from stdin")
	cmd.Flags().String("content", "", "Section content (markdown)")
	return cmd
}

func runSeriesTimelineSectionUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	title, _ := cmd.Flags().GetString("title")
	content, err := contentFromFlags(cmd)
	if err != nil {
		return err
	}
	if title == "" && content == "" {
		return fmt.Errorf("at least one of --title or content (--stdin/--content) is required")
	}
	data, err := svc.UpdateTimelineSection(cmd.Context(), args[0], args[1], title, content)
	if err != nil {
		return notFound(err, "series", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Timeline section updated.")
	return nil
}

func runSeriesTimelineSectionDelete(cmd *cobra.Command, args []string) error {
	svc, err := newSeriesService(cmd)
	if err != nil {
		return err
	}
	if err := svc.DeleteTimelineSection(cmd.Context(), args[0], args[1]); err != nil {
		return notFound(err, "timeline section", args[1])
	}
	fmt.Println("Timeline section deleted.")
	return nil
}
