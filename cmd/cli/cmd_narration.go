package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// === story narrate ===

func newStoryNarrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "narrate <story-id>",
		Short: "Start narration generation for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrate,
	}
	cmd.Flags().String("voice", "", "TTS voice name (e.g., 'Kore'). Omit for server default.")
	return cmd
}

func runStoryNarrate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	voice, _ := cmd.Flags().GetString("voice")
	if err := svc.StartNarration(cmd.Context(), args[0], voice); err != nil {
		return err
	}

	fmt.Println("Narration started.")
	return nil
}

// === story narration (group) ===

func newStoryNarrationGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "narration",
		Short: "Narration management",
	}
	cmd.AddCommand(
		newStoryNarrationStatusCmd(),
		newStoryNarrationVoicesCmd(),
		newStoryNarrationRegenerateCmd(),
		newStoryNarrationRetryCmd(),
		newStoryNarrationRebuildCmd(),
		newStoryNarrationDeleteCmd(),
		newStoryNarrationResumeCmd(),
		newStoryNarrationCancelCmd(),
		newStoryNarrationSegmentsCmd(),
		newStoryNarrationSegmentRegenerateCmd(),
		newStoryNarrationPatchCmd(),
	)
	return cmd
}

func newStoryNarrationStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <story-id>",
		Short: "Get narration status for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarration,
	}
}

func runStoryNarration(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.GetNarration(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === story audiobook ===

func newStoryAudiobookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "audiobook <story-id>",
		Short: "Get audiobook download info for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryAudiobook,
	}
}

func runStoryAudiobook(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.GetAudiobook(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === story narration regenerate ===

func newStoryNarrationRegenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate <story-id> <chapter-id>",
		Short: "Regenerate narration for a specific chapter",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryNarrationRegenerate,
	}
	cmd.Flags().Bool("force", false, "Regenerate even if content hasn't changed")
	cmd.Flags().String("voice", "", "Voice override for this chapter (e.g., Puck, af_sarah)")
	return cmd
}

func runStoryNarrationRegenerate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	voice, _ := cmd.Flags().GetString("voice")

	if err := svc.RegenerateChapter(cmd.Context(), args[0], args[1], force, voice); err != nil {
		return err
	}

	fmt.Printf("Chapter %s regeneration started.\n", args[1])
	return nil
}

// === story narration voices ===

func newStoryNarrationVoicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "voices",
		Short: "List available TTS voices",
		RunE:  runStoryNarrationVoices,
	}
}

func runStoryNarrationVoices(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.ListVoices(cmd.Context())
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === story narration retry ===

func newStoryNarrationRetryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "retry <story-id> <chapter-id>",
		Short: "Retry a failed/stuck chapter narration",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryNarrationRetry,
	}
}

func runStoryNarrationRetry(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.RetryChapter(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}

	fmt.Printf("Chapter %s retry started.\n", args[1])
	return nil
}

// === story narration rebuild ===

func newStoryNarrationRebuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebuild <story-id>",
		Short: "Rebuild audiobook from existing chapter audio",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationRebuild,
	}
	cmd.Flags().Bool("chapter-announcements", false, "Insert TTS-generated chapter title announcements")
	return cmd
}

func runStoryNarrationRebuild(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	announcements, _ := cmd.Flags().GetBool("chapter-announcements")

	if err := svc.RebuildNarration(cmd.Context(), args[0], announcements); err != nil {
		return err
	}

	fmt.Println("Narration rebuild started.")
	return nil
}

// === story narration delete ===

func newStoryNarrationDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <story-id>",
		Short: "Delete all narration data for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationDelete,
	}
}

func runStoryNarrationDelete(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.DeleteNarration(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Println("Narration deleted.")
	return nil
}

// === story narration resume ===

func newStoryNarrationResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <story-id>",
		Short: "Resume a stuck narration",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationResume,
	}
}

func runStoryNarrationResume(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.ResumeNarration(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Println("Narration resumed.")
	return nil
}

// === story narration cancel ===

func newStoryNarrationCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <story-id> <chapter-id>",
		Short: "Cancel a specific chapter's narration",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryNarrationCancel,
	}
}

func runStoryNarrationCancel(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.CancelChapter(cmd.Context(), args[0], args[1]); err != nil {
		return err
	}

	fmt.Printf("Chapter %s cancelled.\n", args[1])
	return nil
}

// === story narration segments ===

func newStoryNarrationSegmentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "segments <story-id> <chapter-id>",
		Short: "List segments for a chapter with text content and voice info",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoryNarrationSegments,
	}
}

func runStoryNarrationSegments(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.ListSegments(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	return printJSON(result)
}

// === story narration segment-regenerate ===

func newStoryNarrationSegmentRegenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "segment-regenerate <story-id> <chapter-id> <segment-id>",
		Short: "Regenerate a single segment's audio",
		Args:  cobra.ExactArgs(3),
		RunE:  runStoryNarrationSegmentRegenerate,
	}
	cmd.Flags().String("voice", "", "Voice override for this segment (e.g., Kore, af_sarah)")
	return cmd
}

func runStoryNarrationSegmentRegenerate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	voice, _ := cmd.Flags().GetString("voice")

	if err := svc.RegenerateSegment(cmd.Context(), args[0], args[1], args[2], voice); err != nil {
		return err
	}

	fmt.Printf("Segment %s regeneration started.\n", args[2])
	return nil
}

// === story narration patch ===

func newStoryNarrationPatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch <story-id>",
		Short: "Patch multiple segments/chapters with voice changes and rebuild",
		Long:  "Batch re-voice segments and chapters in one operation. Rebuilds audiobook once when done.\nUse --segment chapterID:segmentID:voice (repeatable) and --chapter chapterID:voice (repeatable).",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryNarrationPatch,
	}
	cmd.Flags().StringArray("segment", nil, "Segment to patch: chapterID:segmentID:voice (repeatable)")
	cmd.Flags().StringArray("chapter", nil, "Chapter to patch: chapterID:voice (repeatable)")
	return cmd
}

func runStoryNarrationPatch(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	segFlags, _ := cmd.Flags().GetStringArray("segment")
	chFlags, _ := cmd.Flags().GetStringArray("chapter")

	if len(segFlags) == 0 && len(chFlags) == 0 {
		return fmt.Errorf("specify at least one --segment or --chapter")
	}

	req := gen.HandlersPatchNarrationRequest{}

	if len(segFlags) > 0 {
		segs := make([]gen.NarrationPatchSegmentEntry, 0, len(segFlags))
		for _, s := range segFlags {
			parts := splitPatchArg(s, 3)
			if parts == nil {
				return fmt.Errorf("invalid --segment format %q, expected chapterID:segmentID:voice", s)
			}
			segs = append(segs, gen.NarrationPatchSegmentEntry{
				ChapterId: &parts[0],
				SegmentId: &parts[1],
				Voice:     &parts[2],
			})
		}
		req.Segments = &segs
	}

	if len(chFlags) > 0 {
		chs := make([]gen.NarrationPatchChapterEntry, 0, len(chFlags))
		for _, c := range chFlags {
			parts := splitPatchArg(c, 2)
			if parts == nil {
				return fmt.Errorf("invalid --chapter format %q, expected chapterID:voice", c)
			}
			chs = append(chs, gen.NarrationPatchChapterEntry{
				ChapterId: &parts[0],
				Voice:     &parts[1],
			})
		}
		req.Chapters = &chs
	}

	result, err := svc.PatchNarration(cmd.Context(), args[0], req)
	if err != nil {
		return err
	}

	return printJSON(result)
}

func splitPatchArg(s string, n int) []string {
	parts := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, ':')
		if idx < 0 {
			return nil
		}
		parts = append(parts, s[:idx])
		s = s[idx+1:]
	}
	if s == "" {
		return nil
	}
	parts = append(parts, s)
	return parts
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// === credits balance ===

func newCreditsBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "credits",
		Short: "Show credit balance",
		RunE:  runCreditsBalance,
	}
}

func runCreditsBalance(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.GetCredits(cmd.Context())
	if err != nil {
		return err
	}

	return printJSON(result)
}
