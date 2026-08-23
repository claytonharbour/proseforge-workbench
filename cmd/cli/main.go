package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/author"
	"github.com/claytonharbour/proseforge-workbench/internal/bundle"
	"github.com/claytonharbour/proseforge-workbench/internal/config"
	"github.com/claytonharbour/proseforge-workbench/internal/conversation"
	"github.com/claytonharbour/proseforge-workbench/internal/feedback"
	"github.com/claytonharbour/proseforge-workbench/internal/listing"
	"github.com/claytonharbour/proseforge-workbench/internal/review"
	"github.com/claytonharbour/proseforge-workbench/internal/reviewer"
	"github.com/claytonharbour/proseforge-workbench/internal/room"
	"github.com/claytonharbour/proseforge-workbench/internal/series"
	"github.com/claytonharbour/proseforge-workbench/internal/story"
	"github.com/claytonharbour/proseforge-workbench/internal/storyforge"
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
	// Don't dump the full usage/flags block after a runtime error — it makes a
	// simple failure (bad ID, API error) look like a wall of noise. The error
	// message itself is still printed.
	SilenceUsage: true,
}

// cliLogger is the shared logger for the CLI process, initialised in init().
var cliLogger *slog.Logger

func init() {
	rootCmd.PersistentFlags().String("url", "", "API base URL (env: PROSEFORGE_URL). Accepts a quoted env reference, e.g. --url '${PROSEFORGE_URL}'")
	rootCmd.PersistentFlags().String("token", "", "API token (env: PROSEFORGE_TOKEN). Accepts a quoted env reference, e.g. --token '${PROSEFORGE_TOKEN}', which keeps the key out of argv")
	rootCmd.PersistentFlags().String("credentials-file", "", "Path to a credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read per invocation, so a rotated key applies immediately")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, brief")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	// PersistentPreRun sets up the logger before any subcommand runs.
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Default to Error so transient-retry WARN lines don't clutter the
		// console and make a user error (e.g. a bad ID) look like an outage.
		// Errors surface through returned errors, not the logger; --debug
		// restores full WARN/DEBUG visibility for troubleshooting.
		level := slog.LevelError
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			level = slog.LevelDebug
		}
		cliLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}))
	}

	rootCmd.AddCommand(newGenreCmd())
	rootCmd.AddCommand(newStoryCmd())
	rootCmd.AddCommand(newSeriesCmd())
	rootCmd.AddCommand(newBundleCmd())
	rootCmd.AddCommand(newRoomCmd())
	rootCmd.AddCommand(newConversationCmd()) // proseforge#821
	rootCmd.AddCommand(newActiveWorkCmd())
	rootCmd.AddCommand(newWatchCmd())
	rootCmd.AddCommand(newListingCmd())
	rootCmd.AddCommand(newReviewCmd())
	rootCmd.AddCommand(newFeedbackCmd())
	rootCmd.AddCommand(newFriendsCmd())
	rootCmd.AddCommand(newContributionCmd())
	rootCmd.AddCommand(newContributorCmd())
	rootCmd.AddCommand(newReviewerCmd())
	rootCmd.AddCommand(newAuthorCmd())
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newServerCmd()) // #349 parity with server_version
	rootCmd.AddCommand(newTermsCmd())
	rootCmd.AddCommand(newDocsCmd())
}

// newClient creates an API client from flags + env vars.
// No hardcoded defaults — both URL and token must be configured.
func newClient(cmd *cobra.Command) (*api.Client, error) {
	url, _ := cmd.Flags().GetString("url")
	token, _ := cmd.Flags().GetString("token")

	// Quote the reference to reach this — an unquoted $VAR is expanded by the
	// shell before the flag is parsed, which is fine but puts the credential in
	// argv (and so in `ps`). Quoting keeps it out.
	url, err := config.ExpandEnvRef(url, "--url")
	if err != nil {
		return nil, err
	}
	token, err = config.ExpandEnvRef(token, "--token")
	if err != nil {
		return nil, err
	}

	if credFile, _ := cmd.Flags().GetString("credentials-file"); credFile != "" {
		creds, err := config.LoadCredentialsFile(credFile)
		if err != nil {
			return nil, err
		}
		if token == "" {
			token = creds.Token
		}
		if url == "" {
			url = creds.URL
		}
	}

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

// newSeriesService creates a series.Service from CLI flags + env vars.
func newSeriesService(cmd *cobra.Command) (*series.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return series.NewService(client, series.WithLogger(cliLogger)), nil
}

// newBundleService creates a bundle.Service from CLI flags + env vars.
func newBundleService(cmd *cobra.Command) (*bundle.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return bundle.NewService(client, bundle.WithLogger(cliLogger)), nil
}

// newRoomService creates a room.Service from CLI flags + env vars.
func newRoomService(cmd *cobra.Command) (*room.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return room.NewService(client, room.WithLogger(cliLogger)), nil
}

// newConversationService creates a conversation.Service from CLI flags + env vars.
func newConversationService(cmd *cobra.Command) (*conversation.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return conversation.NewService(client, conversation.WithLogger(cliLogger)), nil
}

// newListingService creates a listing.Service from CLI flags + env vars.
func newListingService(cmd *cobra.Command) (*listing.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return listing.NewService(client, listing.WithLogger(cliLogger)), nil
}

// newStoryForgeService creates a storyforge.Service from CLI flags + env vars.
func newStoryForgeService(cmd *cobra.Command) (*storyforge.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return storyforge.NewService(client, storyforge.WithLogger(cliLogger)), nil
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

// newAuthorService creates an author.Service from CLI flags + env vars.
func newAuthorService(cmd *cobra.Command) (*author.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return author.NewService(client, author.WithLogger(cliLogger)), nil
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
