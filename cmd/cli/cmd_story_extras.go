package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// More #234 holes: narration drift/acknowledge/regenerate-stale + the credits
// estimate/history commands. Thin wrappers over existing story.Service methods.

// === narration drift / acknowledge / regenerate-stale ===

func newStoryNarrationDriftCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drift <story-id>",
		Short: "Cheap drift check — just the drift envelope from narration status",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationDrift,
	}
}

func runStoryNarrationDrift(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	full, err := svc.GetNarration(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "story", args[0])
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(full, &m); err != nil {
		fmt.Println(string(full))
		return nil
	}
	drift, ok := m["drift"]
	if !ok {
		fmt.Println(`{"drift":null}`)
		return nil
	}
	fmt.Println(string(drift))
	return nil
}

func newStoryNarrationAcknowledgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "acknowledge <story-id>",
		Short: "Acknowledge narration staleness (clear the stale flag)",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationAcknowledge,
	}
}

func runStoryNarrationAcknowledge(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	if err := svc.AcknowledgeNarrationStale(cmd.Context(), args[0]); err != nil {
		return notFound(err, "story", args[0])
	}
	fmt.Println("Narration staleness acknowledged.")
	return nil
}

func newStoryNarrationRegenerateStaleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "regenerate-stale <story-id>",
		Short: "Re-narrate only the sections whose content drifted",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationRegenerateStale,
	}
}

func runStoryNarrationRegenerateStale(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.RegenerateStaleNarration(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "story", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Stale narration sections queued for regeneration.")
	return nil
}

// === story credits (group: balance | estimate | history) ===

func newStoryCreditsGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credits",
		Short: "Credit balance, cost estimates, and transaction history",
	}
	cmd.AddCommand(
		newCreditsBalanceCmd(),
		newStoryCreditsEstimateCmd(),
		newStoryCreditsHistoryCmd(),
	)
	return cmd
}

func newStoryCreditsEstimateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Estimate the credit cost of an operation",
		Args:  cobra.NoArgs,
		RunE:  runStoryCreditsEstimate,
	}
	cmd.Flags().String("operation", "", "Operation: narrate, generate, rewrite, image, avatar, patch, insights")
	cmd.Flags().Int("sections", 0, "Number of sections (narrate/generate/rewrite)")
	cmd.Flags().Int("segments", 0, "Number of segments (patch)")
	cmd.Flags().Bool("images", false, "Include image generation (generate)")
	_ = cmd.MarkFlagRequired("operation")
	return cmd
}

func runStoryCreditsEstimate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	operation, _ := cmd.Flags().GetString("operation")
	params := &gen.GetCreditsEstimateParams{Operation: operation}
	if cmd.Flags().Changed("sections") {
		n, _ := cmd.Flags().GetInt("sections")
		params.Sections = &n
	}
	if cmd.Flags().Changed("segments") {
		n, _ := cmd.Flags().GetInt("segments")
		params.Segments = &n
	}
	if cmd.Flags().Changed("images") {
		b, _ := cmd.Flags().GetBool("images")
		params.Images = &b
	}
	data, err := svc.EstimateCredits(cmd.Context(), params)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func newStoryCreditsHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show credit transaction history",
		Args:  cobra.NoArgs,
		RunE:  runStoryCreditsHistory,
	}
	cmd.Flags().Int("limit", 0, "Max transactions (0 = server default)")
	return cmd
}

func runStoryCreditsHistory(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}
	limit, _ := cmd.Flags().GetInt("limit")
	data, err := svc.GetCreditHistory(cmd.Context(), limit)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
