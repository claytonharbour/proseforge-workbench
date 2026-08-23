package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Story planning / lifecycle commands (#232): pitch → meta → promote → draft.
// Thin wrappers over the existing story.Service + storyforge.Service.

// === story pitch ===

func newStoryPitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pitch",
		Short: "Pre-writing story pitches (idea + planning data, no sections yet)",
	}
	cmd.AddCommand(newStoryPitchCreateCmd())
	return cmd
}

func newStoryPitchCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pitch (a pre-writing story idea)",
		Long: "Create a pitch — a pre-writing story idea with planning data but no " +
			"sections. Next: story meta upsert (premise/characters/plot), then story promote.",
		Args: cobra.NoArgs,
		RunE: runStoryPitchCreate,
	}
	cmd.Flags().String("genre", "", "Genre name (e.g., \"Historical Fiction\")")
	cmd.Flags().String("title", "", "Story title (optional)")
	cmd.Flags().String("tagline", "", "Story tagline (optional)")
	_ = cmd.MarkFlagRequired("genre")
	return cmd
}

func runStoryPitchCreate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	genreName, _ := cmd.Flags().GetString("genre")
	genreID, err := svc.ResolveGenreID(cmd.Context(), genreName)
	if err != nil {
		return err
	}

	req := api.CreateStoryRequest{GenreId: &genreID}
	if t, _ := cmd.Flags().GetString("title"); t != "" {
		req.Title = &t
	}

	result, err := svc.CreatePitch(cmd.Context(), req)
	if err != nil {
		return err
	}

	if tagline, _ := cmd.Flags().GetString("tagline"); tagline != "" && result.Id != nil {
		if err := svc.Update(cmd.Context(), *result.Id, api.UpdateStoryRequest{Tagline: &tagline}); err != nil {
			return fmt.Errorf("pitch created (%s) but failed to set tagline: %w", deref(result.Id), err)
		}
	}

	if isJSON(cmd) {
		return printJSON(result)
	}
	fmt.Printf("Pitch created.\n")
	fmt.Printf("ID:     %s\n", deref(result.Id))
	fmt.Printf("Title:  %s\n", deref(result.Title))
	fmt.Printf("Status: %s\n", deref(result.Status))
	return nil
}

// === story promote ===

func newStoryPromoteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "promote <story-id>",
		Short: "Promote a pitch to draft (ready to write sections)",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryPromote,
	}
}

func runStoryPromote(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Promote(cmd.Context(), args[0]); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println("Story promoted to draft.")
	return nil
}

// === story meta ===

func newStoryMetaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta",
		Short: "Story planning data (premise, characters, plot outline)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get <story-id>",
			Short: "Read all story planning data (story / characters / plot)",
			Args:  cobra.ExactArgs(1),
			RunE:  runStoryMetaGet,
		},
		newStoryMetaUpsertCmd(),
		&cobra.Command{
			Use:   "stale <story-id>",
			Short: "List sections affected by meta changes since last generation",
			Args:  cobra.ExactArgs(1),
			RunE:  runStoryMetaStale,
		},
		&cobra.Command{
			Use:   "acknowledge <story-id>",
			Short: "Acknowledge meta staleness (clear the stale flag)",
			Args:  cobra.ExactArgs(1),
			RunE:  runStoryMetaAcknowledge,
		},
	)
	return cmd
}

func runStoryMetaGet(cmd *cobra.Command, args []string) error {
	svc, err := newStoryForgeService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.GetMeta(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "story", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	// Planning data is three markdown documents — render them with headers.
	var docs map[string]string
	if err := json.Unmarshal(data, &docs); err != nil {
		fmt.Println(string(data))
		return nil
	}
	for _, key := range []string{"story", "characters", "plot"} {
		if v := docs[key]; v != "" {
			fmt.Printf("===== %s =====\n%s\n\n", key, v)
		}
	}
	return nil
}

func newStoryMetaUpsertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upsert <story-id>",
		Short: "Write a story planning document (story | characters | plot)",
		Long:  "Write story planning data as markdown. --type selects which document; use --stdin for large content.",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryMetaUpsert,
	}
	cmd.Flags().String("type", "", "Document type: story, characters, or plot")
	cmd.Flags().String("content", "", "Planning data content (markdown)")
	cmd.Flags().Bool("stdin", false, "Read the content from stdin")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func runStoryMetaUpsert(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	metaType, _ := cmd.Flags().GetString("type")
	switch metaType {
	case "story", "characters", "plot":
		// valid
	default:
		return fmt.Errorf("type must be one of: story, characters, plot")
	}

	content, _ := cmd.Flags().GetString("content")
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		b, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			return fmt.Errorf("reading stdin: %w", rerr)
		}
		content = string(b)
	}
	if content == "" {
		return fmt.Errorf("content is required: use --content or --stdin")
	}

	data, err := svc.UpsertMeta(cmd.Context(), args[0], metaType, content)
	if err != nil {
		return notFound(err, "story", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Meta updated (%s).\n", metaType)
	return nil
}

func runStoryMetaStale(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.GetMetaStale(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println(string(data))
	return nil
}

func runStoryMetaAcknowledge(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.AcknowledgeMetaStale(cmd.Context(), args[0]); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println("Meta staleness acknowledged.")
	return nil
}
