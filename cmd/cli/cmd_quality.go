package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

func newStoryQualityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quality <story-id>",
		Short: "Get story quality assessment scores",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryQuality,
	}
}

func runStoryQuality(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	data, err := svc.GetQuality(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return printJSON(json.RawMessage(data))
}

func newStoryAssessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assess <story-id>",
		Short: "Trigger a quality assessment for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryAssess,
	}
	cmd.Flags().Bool("force", false, "Force re-assessment even if content unchanged")
	return cmd
}

func runStoryAssess(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	data, err := svc.AssessQuality(cmd.Context(), args[0], force)
	if err != nil {
		return err
	}

	return printJSON(json.RawMessage(data))
}

func newStoryAssessVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "assess-version <story-id> <sha>",
		Short: "Assess quality at a specific version SHA (synchronous)",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryAssessVersion,
	}
}

func runStoryAssessVersion(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	data, err := svc.AssessQualityAtVersion(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	return printJSON(json.RawMessage(data))
}

func newStoryInsightsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "insights <story-id>",
		Short: "Get combined quality and AI analysis insights",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryInsights,
	}
}

func runStoryInsights(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	data, err := svc.GetInsights(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return printJSON(json.RawMessage(data))
}
