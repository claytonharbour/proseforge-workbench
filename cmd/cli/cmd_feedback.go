package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/feedback"
)

func newFeedbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Feedback review operations",
	}
	cmd.AddCommand(
		newFeedbackListCmd(),
		newFeedbackGetCmd(),
		newFeedbackDiffCmd(),
		newFeedbackSuggestionsCmd(),
		newFeedbackCreateCmd(),
		newFeedbackItemCmd(),
		newFeedbackSectionCmd(),
		newFeedbackSubmitCmd(),
		newFeedbackIncorporateCmd(),
	)
	return cmd
}

// === feedback list ===

func newFeedbackListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <story-id>",
		Short: "List feedback reviews for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runFeedbackList,
	}
}

func runFeedbackList(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.List(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(result)
	}

	if result.Reviews == nil || len(*result.Reviews) == 0 {
		fmt.Println("No feedback reviews.")
		return nil
	}

	status("Feedback reviews: %d", derefInt(result.Total))
	fmt.Println()

	var rows [][]string
	for _, r := range *result.Reviews {
		rows = append(rows, []string{
			deref(r.Id),
			deref(r.Status),
			deref(r.ReviewType),
			deref(r.CreatedAt),
		})
	}

	printTable([]string{"ID", "Status", "Type", "Created"}, rows)
	return nil
}

// === feedback get ===

func newFeedbackGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <story-id> <review-id>",
		Short: "Get feedback review details",
		Args:  cobra.ExactArgs(2),
		RunE:  runFeedbackGet,
	}
}

func runFeedbackGet(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	result, err := svc.GetFull(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(result)
	}

	if result.Review != nil {
		review := result.Review
		fmt.Printf("Review ID:   %s\n", deref(review.Id))
		fmt.Printf("Status:      %s\n", deref(review.Status))
		fmt.Printf("Type:        %s\n", deref(review.ReviewType))
		fmt.Printf("Iterations:  %d\n", derefInt(review.TotalIterations))
		fmt.Printf("Created:     %s\n", deref(review.CreatedAt))
		if review.CompletedAt != nil {
			fmt.Printf("Completed:   %s\n", *review.CompletedAt)
		}
	}

	if result.Items != nil {
		fmt.Printf("\nFeedback Items: %d total suggestions\n", result.Items.TotalSuggestions)
		for _, sec := range result.Items.Sections {
			total := len(sec.Suggestions) + len(sec.Strengths) + len(sec.Opportunities) + len(sec.Comments) + len(sec.Context)
			if total > 0 {
				fmt.Printf("  %s: %d items\n", sec.SectionTitle, total)
			}
		}
	}

	return nil
}

// === feedback diff ===

func newFeedbackDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <story-id> <review-id>",
		Short: "Show diff of suggested changes",
		Args:  cobra.ExactArgs(2),
		RunE:  runFeedbackDiff,
	}
}

func runFeedbackDiff(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	diff, err := svc.GetDiff(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	return printJSON(diff)
}

// === feedback suggestions ===

func newFeedbackSuggestionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suggestions <story-id> <review-id>",
		Short: "List suggestions for a feedback review",
		Args:  cobra.ExactArgs(2),
		RunE:  runFeedbackSuggestions,
	}
}

func runFeedbackSuggestions(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	fb, err := svc.GetSuggestions(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(fb)
	}

	if fb.Sections == nil || len(*fb.Sections) == 0 {
		fmt.Println("No suggestions.")
		return nil
	}

	count := 0
	for _, sec := range *fb.Sections {
		sectionTitle := deref(sec.SectionTitle)
		if sec.SpecificRewrites == nil {
			continue
		}
		for _, s := range *sec.SpecificRewrites {
			count++
			fmt.Printf("--- Suggestion %d [%s] ---\n", count, sectionTitle)
			fmt.Printf("Type:      %s\n", deref(s.Type))
			if s.Original != nil && *s.Original != "" {
				fmt.Printf("Original:  %s\n", deref(s.Original))
			}
			if s.Suggested != nil && *s.Suggested != "" {
				fmt.Printf("Suggested: %s\n", deref(s.Suggested))
			}
			fmt.Printf("Rationale: %s\n", deref(s.Rationale))
			fmt.Println()
		}
	}

	if count == 0 {
		fmt.Println("No suggestions.")
	}

	return nil
}

// === feedback create ===

func newFeedbackCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <story-id>",
		Short: "Create a new feedback review",
		Args:  cobra.ExactArgs(1),
		RunE:  runFeedbackCreate,
	}
}

func runFeedbackCreate(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	req := api.StartAIReviewRequest{}
	review, err := svc.Create(cmd.Context(), args[0], req)
	if err != nil {
		return err
	}

	fmt.Printf("Review created: %s (status: %s)\n", deref(review.Id), deref(review.Status))
	return nil
}

// === feedback item (subcommand group) ===

func newFeedbackItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "item",
		Short: "Feedback item operations",
	}
	cmd.AddCommand(newFeedbackItemAddCmd())
	return cmd
}

func newFeedbackItemAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <story-id> <review-id>",
		Short: "Add a feedback item (reads JSON from stdin with --stdin)",
		Args:  cobra.ExactArgs(2),
		RunE:  runFeedbackItemAdd,
	}
	cmd.Flags().Bool("stdin", false, "Read feedback item JSON from stdin")
	cmd.Flags().Bool("batch", false, "Read multiple items, one JSON object per line")
	cmd.Flags().String("type", "", "Item type: replacement, strength, opportunity, suggestion, context")
	cmd.Flags().String("section", "", "Section ID")
	cmd.Flags().String("text", "", "Original text (replacement) or feedback text")
	cmd.Flags().String("suggested", "", "Suggested replacement text")
	cmd.Flags().String("rationale", "", "Why this improves the writing")
	cmd.Flags().String("context-type", "", "For type=context: characters, plot, tone, threads (default: general)")
	return cmd
}

func runFeedbackItemAdd(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	storyID := args[0]
	reviewID := args[1]
	useStdin, _ := cmd.Flags().GetBool("stdin")
	batch, _ := cmd.Flags().GetBool("batch")

	if useStdin {
		return addItemsFromStdin(cmd, svc, storyID, reviewID, batch)
	}

	return addItemFromFlags(cmd, svc, storyID, reviewID)
}

func addItemsFromStdin(cmd *cobra.Command, svc *feedback.Service, storyID, reviewID string, batch bool) error {
	if batch {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB per line
		count := 0
		for scanner.Scan() {
			var item api.AddFeedbackItemRequest
			if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
				return fmt.Errorf("parsing item %d: %w", count+1, err)
			}
			if err := svc.AddItem(cmd.Context(), storyID, reviewID, item); err != nil {
				return fmt.Errorf("adding item %d: %w", count+1, err)
			}
			count++
			fmt.Fprintf(os.Stderr, "Added item %d\n", count)
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		fmt.Printf("Added %d feedback items.\n", count)
		return nil
	}

	// Single JSON object from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	var item api.AddFeedbackItemRequest
	if err := json.Unmarshal(data, &item); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	if err := svc.AddItem(cmd.Context(), storyID, reviewID, item); err != nil {
		return err
	}

	fmt.Println("Feedback item added.")
	return nil
}

func addItemFromFlags(cmd *cobra.Command, svc *feedback.Service, storyID, reviewID string) error {
	itemType, _ := cmd.Flags().GetString("type")
	sectionID, _ := cmd.Flags().GetString("section")
	text, _ := cmd.Flags().GetString("text")
	suggested, _ := cmd.Flags().GetString("suggested")
	rationale, _ := cmd.Flags().GetString("rationale")
	contextType, _ := cmd.Flags().GetString("context-type")

	if itemType == "" {
		return fmt.Errorf("--type is required (replacement, strength, opportunity, suggestion, context)")
	}
	if text == "" {
		return fmt.Errorf("--text is required")
	}

	item := api.AddFeedbackItemRequest{
		Type: &itemType,
		Text: &text,
	}
	if sectionID != "" {
		item.SectionId = &sectionID
	}
	if suggested != "" {
		item.Suggested = &suggested
	}
	if rationale != "" {
		item.Rationale = &rationale
	}
	if itemType == "context" {
		if contextType == "" {
			return fmt.Errorf("--context-type is required when type=context (valid: characters, plot, tone, threads)")
		}
		item.ContextType = &contextType
	}

	if err := svc.AddItem(cmd.Context(), storyID, reviewID, item); err != nil {
		return err
	}

	fmt.Println("Feedback item added.")
	return nil
}

// === feedback section (subcommand group) ===

func newFeedbackSectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "section",
		Short: "Section-level feedback operations",
	}
	cmd.AddCommand(newFeedbackSectionUpdateCmd())
	return cmd
}

func newFeedbackSectionUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <story-id> <review-id> <section-id>",
		Short: "Rewrite a section's content (reads from stdin with --stdin)",
		Args:  cobra.ExactArgs(3),
		RunE:  runFeedbackSectionUpdate,
	}
	cmd.Flags().Bool("stdin", false, "Read section content from stdin")
	cmd.Flags().String("content", "", "Section content (for short content; prefer --stdin for full sections)")
	return cmd
}

func runFeedbackSectionUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	storyID, reviewID, sectionID := args[0], args[1], args[2]
	useStdin, _ := cmd.Flags().GetBool("stdin")
	contentFlag, _ := cmd.Flags().GetString("content")

	var content string
	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		content = string(data)
	} else if contentFlag != "" {
		content = contentFlag
	} else {
		return fmt.Errorf("provide content via --stdin or --content")
	}

	if err := svc.UpdateSection(cmd.Context(), storyID, reviewID, sectionID, content); err != nil {
		return err
	}

	status("Section %s updated (%d characters).", sectionID, len(content))
	return nil
}

// === feedback submit ===

func newFeedbackSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "submit <review-id>",
		Short: "Submit a review for author",
		Args:  cobra.ExactArgs(1),
		RunE:  runFeedbackSubmit,
	}
}

func runFeedbackSubmit(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	if err := svc.Submit(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Printf("Review %s submitted.\n", args[0])
	return nil
}

// === feedback incorporate ===

func newFeedbackIncorporateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incorporate <story-id> <review-id>",
		Short: "Incorporate feedback changes (author)",
		Args:  cobra.ExactArgs(2),
		RunE:  runFeedbackIncorporate,
	}
	cmd.Flags().Bool("all", false, "Accept all changes")
	cmd.Flags().String("selections", "", "JSON map of path->bool selections")
	return cmd
}

func runFeedbackIncorporate(cmd *cobra.Command, args []string) error {
	svc, err := newFeedbackService(cmd)
	if err != nil {
		return err
	}

	acceptAll, _ := cmd.Flags().GetBool("all")
	selectionsJSON, _ := cmd.Flags().GetString("selections")

	if acceptAll {
		if err := svc.IncorporateAll(cmd.Context(), args[0], args[1]); err != nil {
			return err
		}
	} else if selectionsJSON != "" {
		var selections map[string]bool
		if err := json.Unmarshal([]byte(selectionsJSON), &selections); err != nil {
			return fmt.Errorf("parsing selections: %w", err)
		}
		if err := svc.IncorporateSelective(cmd.Context(), args[0], args[1], selections); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("specify --all or --selections")
	}

	fmt.Println("Feedback incorporated.")
	return nil
}
