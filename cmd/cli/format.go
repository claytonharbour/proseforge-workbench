package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
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

// isBrief returns true if output format is brief.
func isBrief(cmd *cobra.Command) bool {
	return outputFormat(cmd) == "brief"
}

// printBrief renders rows for -o brief: one line per record, fields
// tab-separated, no header or separator row. Terse and pipe-friendly
// (cut -f1, awk -F'\t'). Convention: id + primary label, see
// docs/internal/DEVELOPMENT.md → CLI Command Conventions.
func printBrief(rows [][]string) {
	for _, row := range rows {
		fmt.Println(strings.Join(row, "\t"))
	}
}

// notFound maps a 404 API error to a clean "<entity> not found: <id>" message
// so a missing resource reads as a user error, not a raw HTTP failure. Non-404
// errors pass through unchanged.
func notFound(err error, entity, id string) error {
	// 🛑 A MISSING ROUTE IS NOT A MISSING RECORD, and reporting one as the other
	// sends people to look for a data problem that does not exist.
	//
	// Measured 2026-08-29: demo deleted /authors/{handle}/books mid-migration and
	// `pfw author bookshelf clayton` said "author not found: clayton". The author
	// was fine; the endpoint had moved. @Sten hit it and went looking for the
	// author.
	//
	// ⚠️ routeMissing() has existed since this morning and this function did not
	// call it — the discriminator was built and never wired to the 51 callers that
	// needed it. Distinguishing here fixes all of them at once.
	if routeMissing(err) {
		return fmt.Errorf("this server has no such endpoint, so the %s %q was never looked up — "+
			"the ROUTE is missing, not the data. The server may be older than this client, "+
			"or the route may have moved; check `pfw --version` against the server's /api/v1/version",
			entity, id)
	}
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
		return fmt.Errorf("%s not found: %s", entity, id)
	}
	return err
}

// routeMissing reports whether err is a 404 from the ROUTER rather than from the API
// — i.e. this backend predates the endpoint the command needs.
//
// ⚑ The two are distinguishable and the distinction matters to a user: a mux 404 is
// plain text ("404 page not found"), while an API 404 is JSON carrying an error code.
// Treating the first as "not found" tells someone their ID is wrong when the truth is
// that their SERVER is older — and `pfw` ships to people pointing it at their own
// deployments, so that is a real audience rather than a hypothetical one.
func routeMissing(err error) bool {
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		return false
	}
	return !json.Valid([]byte(strings.TrimSpace(apiErr.Body)))
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

// derefBool safely dereferences a bool pointer.
func derefBool(b *bool) bool {
	return b != nil && *b
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
