package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// outputFormat returns the --output flag value.
func outputFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("output")
	return f
}

// isJSON returns true if output format is json.
func isJSON(cmd *cobra.Command) bool {
	return outputFormat(cmd) == "json"
}

// status prints a status message to stderr (not stdout).
func status(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// printJSON marshals v to stdout as indented JSON.
func printJSON(v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// truncate returns at most n characters of the input string.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

// deref safely dereferences a string pointer.
func deref(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

// derefInt safely dereferences an int pointer.
func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// printTable prints a formatted table with header and rows.
func printTable(header []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No results.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	// Print header
	for i, h := range header {
		fmt.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i := range header {
		fmt.Print(strings.Repeat("-", widths[i]))
		fmt.Print("  ")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) {
				fmt.Printf("%-*s  ", widths[i], col)
			}
		}
		fmt.Println()
	}
}
