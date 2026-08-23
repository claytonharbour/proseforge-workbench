package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Bundle commands (#230). Thin cobra wrappers over bundle.Service — the same
// service the MCP server uses. Render via the generated bundle response types.

// stdinOrFlag returns content from stdin (if --stdin) or the named string flag.
func stdinOrFlag(cmd *cobra.Command, flag string) (string, error) {
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}
	v, _ := cmd.Flags().GetString(flag)
	return v, nil
}

func newBundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Bundle operations (package stories into EPUB/PDF/markdown/JSON)",
	}
	cmd.AddCommand(
		newBundleListCmd(),
		newBundleGetCmd(),
		newBundleCreateCmd(),
		newBundleUpdateCmd(),
		newBundleDeleteCmd(),
		newBundleEntryCmd(),
		newBundleExportCmd(),
	)
	return cmd
}

// === bundle list ===

func newBundleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your bundles",
		Args:  cobra.NoArgs,
		RunE:  runBundleList,
	}
}

func runBundleList(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.List(cmd.Context())
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	var resp gen.HandlersBundleListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Println(string(data))
		return nil
	}
	var bundles []gen.HandlersBundleResponse
	if resp.Bundles != nil {
		bundles = *resp.Bundles
	}
	if len(bundles) == 0 {
		if !isBrief(cmd) {
			fmt.Println("No bundles found.")
		}
		return nil
	}
	if isBrief(cmd) {
		var rows [][]string
		for _, b := range bundles {
			rows = append(rows, []string{deref(b.Id), deref(b.Name)})
		}
		printBrief(rows)
		return nil
	}
	status("Bundles: %d", len(bundles))
	fmt.Println()
	var rows [][]string
	for _, b := range bundles {
		rows = append(rows, []string{deref(b.Id), deref(b.Name), fmt.Sprintf("%d", derefInt(b.EntryCount))})
	}
	printTable([]string{"ID", "Name", "Entries"}, rows)
	return nil
}

// === bundle get ===

func newBundleGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <bundle-id>",
		Short: "Get a bundle with its entries",
		Args:  cobra.ExactArgs(1),
		RunE:  runBundleGet,
	}
}

func runBundleGet(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.Get(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "bundle", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	var b gen.HandlersBundleDetailResponse
	if err := json.Unmarshal(data, &b); err != nil {
		fmt.Println(string(data))
		return nil
	}
	if isBrief(cmd) {
		printBrief([][]string{{deref(b.Id), deref(b.Name)}})
		return nil
	}
	fmt.Printf("ID:      %s\n", deref(b.Id))
	fmt.Printf("Name:    %s\n", deref(b.Name))
	fmt.Printf("Entries: %d\n", derefInt(b.EntryCount))
	if b.Intro != nil && *b.Intro != "" {
		fmt.Printf("Intro:   %s\n", *b.Intro)
	}
	var entries []gen.HandlersBundleEntryResponse
	if b.Entries != nil {
		entries = *b.Entries
	}
	if len(entries) > 0 {
		fmt.Println()
		var rows [][]string
		for _, e := range entries {
			rows = append(rows, []string{
				fmt.Sprintf("%d", derefInt(e.SortOrder)),
				deref(e.Id),
				deref(e.StoryTitle),
				truncate(deref(e.Transition), 50),
			})
		}
		printTable([]string{"Order", "EntryID", "Story", "Transition"}, rows)
	}
	return nil
}

// === bundle create ===

func newBundleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new bundle",
		Args:  cobra.NoArgs,
		RunE:  runBundleCreate,
	}
	cmd.Flags().String("name", "", "Bundle name")
	cmd.Flags().String("intro", "", "Introduction text (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the intro from stdin")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runBundleCreate(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	name, _ := cmd.Flags().GetString("name")
	intro, err := stdinOrFlag(cmd, "intro")
	if err != nil {
		return err
	}
	data, err := svc.Create(cmd.Context(), name, intro)
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	var b gen.HandlersBundleResponse
	if err := json.Unmarshal(data, &b); err == nil && b.Id != nil {
		status("Bundle created.")
		fmt.Printf("ID:   %s\n", deref(b.Id))
		fmt.Printf("Name: %s\n", deref(b.Name))
		return nil
	}
	fmt.Println(string(data))
	return nil
}

// === bundle update ===

func newBundleUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <bundle-id>",
		Short: "Update a bundle's name or intro",
		Long:  "Update a bundle's name or intro. Only the flags you pass are changed.",
		Args:  cobra.ExactArgs(1),
		RunE:  runBundleUpdate,
	}
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("intro", "", "New introduction text (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the intro from stdin")
	return cmd
}

func runBundleUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	name, _ := cmd.Flags().GetString("name")
	intro, err := stdinOrFlag(cmd, "intro")
	if err != nil {
		return err
	}
	if name == "" && intro == "" {
		return fmt.Errorf("at least one of --name or --intro (or --stdin) is required")
	}
	if err := svc.Update(cmd.Context(), args[0], name, intro); err != nil {
		return notFound(err, "bundle", args[0])
	}
	fmt.Println("Bundle updated.")
	return nil
}

// === bundle delete ===

func newBundleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <bundle-id>",
		Short: "Delete a bundle (the stories are not affected)",
		Args:  cobra.ExactArgs(1),
		RunE:  runBundleDelete,
	}
}

func runBundleDelete(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Delete(cmd.Context(), args[0]); err != nil {
		return notFound(err, "bundle", args[0])
	}
	fmt.Println("Bundle deleted.")
	return nil
}

// === bundle entry ===

func newBundleEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry",
		Short: "Manage bundle entries",
	}
	cmd.AddCommand(
		newBundleEntryAddCmd(),
		newBundleEntryUpdateCmd(),
		&cobra.Command{
			Use:   "remove <bundle-id> <entry-id>",
			Short: "Remove an entry from a bundle",
			Args:  cobra.ExactArgs(2),
			RunE:  runBundleEntryRemove,
		},
		&cobra.Command{
			Use:   "reorder <bundle-id> <entry-id>...",
			Short: "Set the order of entries by listing entry IDs",
			Args:  cobra.MinimumNArgs(2),
			RunE:  runBundleEntryReorder,
		},
	)
	return cmd
}

func newBundleEntryAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <bundle-id> <story-id>",
		Short: "Add a story to a bundle",
		Long:  "Add a story to a bundle. --comment sets the interstitial transition text (the \"Transition\" field).",
		Args:  cobra.ExactArgs(2),
		RunE:  runBundleEntryAdd,
	}
	cmd.Flags().String("comment", "", "Interstitial transition text (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the transition text from stdin")
	return cmd
}

func runBundleEntryAdd(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	comment, err := stdinOrFlag(cmd, "comment")
	if err != nil {
		return err
	}
	data, err := svc.AddEntry(cmd.Context(), args[0], args[1], comment)
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Story added to bundle.")
	return nil
}

func newBundleEntryUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <bundle-id> <entry-id>",
		Short: "Update an entry's transition text",
		Args:  cobra.ExactArgs(2),
		RunE:  runBundleEntryUpdate,
	}
	cmd.Flags().String("comment", "", "New transition text (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the transition text from stdin")
	return cmd
}

func runBundleEntryUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	comment, err := stdinOrFlag(cmd, "comment")
	if err != nil {
		return err
	}
	if comment == "" {
		return fmt.Errorf("transition text is required: use --comment or --stdin")
	}
	if err := svc.UpdateEntry(cmd.Context(), args[0], args[1], comment); err != nil {
		return err
	}
	fmt.Println("Bundle entry updated.")
	return nil
}

func runBundleEntryRemove(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	if err := svc.RemoveEntry(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}
	fmt.Println("Bundle entry removed.")
	return nil
}

func runBundleEntryReorder(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	if err := svc.ReorderEntries(cmd.Context(), args[0], args[1:]); err != nil {
		return err
	}
	fmt.Println("Bundle entries reordered.")
	return nil
}

// === bundle export ===

func newBundleExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <bundle-id>",
		Short: "Export a bundle as epub, json, markdown, or pdf",
		Args:  cobra.ExactArgs(1),
		RunE:  runBundleExport,
	}
	cmd.Flags().String("format", "", "Export format: epub, json, markdown, pdf")
	cmd.Flags().String("out", "", "Write to this file instead of stdout (recommended for epub/pdf)")
	_ = cmd.MarkFlagRequired("format")
	return cmd
}

func runBundleExport(cmd *cobra.Command, args []string) error {
	svc, err := newBundleService(cmd)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "epub", "json", "markdown", "pdf":
		// valid
	default:
		return fmt.Errorf("format must be one of: epub, json, markdown, pdf")
	}
	data, err := svc.Export(cmd.Context(), args[0], format)
	if err != nil {
		return notFound(err, "bundle", args[0])
	}
	if out, _ := cmd.Flags().GetString("out"); out != "" {
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", out, err)
		}
		status("Wrote %d bytes to %s", len(data), out)
		return nil
	}
	_, err = os.Stdout.Write(data)
	return err
}
