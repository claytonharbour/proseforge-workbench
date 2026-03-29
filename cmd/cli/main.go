package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/config"
	"github.com/claytonharbour/proseforge-workbench/internal/feedback"
	"github.com/claytonharbour/proseforge-workbench/internal/review"
	"github.com/claytonharbour/proseforge-workbench/internal/reviewer"
	"github.com/claytonharbour/proseforge-workbench/internal/story"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "pfw",
	Short: "ProseForge Workbench — CLI for AI-assisted story review",
	Example: `  # Set credentials via environment (recommended)
  export PROSEFORGE_URL=https://app.proseforge.ai
  export PROSEFORGE_TOKEN=pf_your_token
  pfw story list

  # Or pass inline
  pfw --url https://app.proseforge.ai --token pf_your_token story list

  # Review workflow
  pfw story get <story-id>
  pfw feedback list <story-id>
  pfw feedback diff <story-id> <review-id>`,
	Version: Version,
}

// cliLogger is the shared logger for the CLI process, initialised in init().
var cliLogger *slog.Logger

func init() {
	rootCmd.PersistentFlags().String("url", "", "API base URL (env: PROSEFORGE_URL)")
	rootCmd.PersistentFlags().String("token", "", "API token (env: PROSEFORGE_TOKEN)")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, brief")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	// PersistentPreRun sets up the logger before any subcommand runs.
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		level := slog.LevelWarn
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			level = slog.LevelDebug
		}
		cliLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}))
	}

	rootCmd.AddCommand(newGenreCmd())
	rootCmd.AddCommand(newStoryCmd())
	rootCmd.AddCommand(newReviewCmd())
	rootCmd.AddCommand(newFeedbackCmd())
	rootCmd.AddCommand(newReviewerCmd())
	rootCmd.AddCommand(newDocsCmd())
}

// newClient creates an API client from flags + env vars.
// No hardcoded defaults — both URL and token must be configured.
func newClient(cmd *cobra.Command) (*api.Client, error) {
	url, _ := cmd.Flags().GetString("url")
	token, _ := cmd.Flags().GetString("token")

	cfg := config.FromEnv().WithOverrides(url, token)
	return cfg.NewClient(api.WithLogger(cliLogger))
}

// newStoryService creates a story.Service from CLI flags + env vars.
func newStoryService(cmd *cobra.Command) (*story.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return story.NewService(client, story.WithLogger(cliLogger)), nil
}

// newReviewService creates a review.Service from CLI flags + env vars.
func newReviewService(cmd *cobra.Command) (*review.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return review.NewService(client, review.WithLogger(cliLogger)), nil
}

// newFeedbackService creates a feedback.Service from CLI flags + env vars.
func newFeedbackService(cmd *cobra.Command) (*feedback.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return feedback.NewService(client, feedback.WithLogger(cliLogger)), nil
}

// newReviewerService creates a reviewer.Service from CLI flags + env vars.
func newReviewerService(cmd *cobra.Command) (*reviewer.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return reviewer.NewService(client, reviewer.WithLogger(cliLogger)), nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
