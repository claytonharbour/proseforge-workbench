package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gendocs",
		Short:  "Generate CLI reference documentation",
		Hidden: true,
		RunE:   runGenDocs,
	}
	cmd.Flags().String("dir", "docs/cli", "Output directory for generated docs")
	return cmd
}

func runGenDocs(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	rootCmd.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(rootCmd, dir); err != nil {
		return fmt.Errorf("generating docs: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Documentation generated in %s/\n", dir)
	return nil
}
