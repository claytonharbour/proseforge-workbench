package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review operations",
	}
	cmd.AddCommand(
		newReviewListCmd(),
		newReviewActiveCmd(),
		newReviewRequestCmd(),
		newReviewAcceptCmd(),
		newReviewDeclineCmd(),
		newReviewApproveCmd(),
		newReviewRejectCmd(),
	)
	return cmd
}

// === review list ===

func newReviewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pending reviews assigned to you",
		RunE:  runReviewList,
	}
	cmd.Flags().Int("limit", 25, "Max results (1-100)")
	return cmd
}

func runReviewList(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	params := &gen.GetReviewsPendingParams{Limit: &limit}

	result, err := svc.ListPending(cmd.Context(), params)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(result)
	}

	if result.Reviews == nil || len(*result.Reviews) == 0 {
		fmt.Println("No pending reviews.")
		return nil
	}

	var rows [][]string
	for _, r := range *result.Reviews {
		rows = append(rows, []string{
			deref(r.Id),
			deref(r.ReviewId),
			deref(r.StoryId),
			deref(r.Status),
		})
	}

	printTable([]string{"ID", "Review ID", "Story ID", "Status"}, rows)
	return nil
}

// === review active ===

func newReviewActiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "active <story-id>",
		Short: "Show the active (running) review for a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewActive,
	}
}

func runReviewActive(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	review, err := svc.ActiveReview(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if review == nil {
		fmt.Println("No active review for this story.")
		return nil
	}

	if isJSON(cmd) {
		return printJSON(review)
	}

	fmt.Printf("Review ID: %s\nStatus:    %s\nType:      %s\nCreated:   %s\n",
		deref(review.Id), deref(review.Status), deref(review.ReviewType), deref(review.CreatedAt))
	return nil
}

// === review request ===

func newReviewRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request <story-id>",
		Short: "Add a reviewer to a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewRequest,
	}
	cmd.Flags().String("reviewer", "", "Reviewer user ID")
	cmd.Flags().String("email", "", "Reviewer email (alternative to --reviewer)")
	return cmd
}

func runReviewRequest(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	reviewerID, _ := cmd.Flags().GetString("reviewer")
	email, _ := cmd.Flags().GetString("email")
	if reviewerID == "" && email == "" {
		return fmt.Errorf("provide --reviewer (user ID) or --email")
	}

	req := api.AddReviewerRequest{}
	if reviewerID != "" {
		req.ReviewerId = &reviewerID
	}
	if email != "" {
		req.Email = &email
	}

	reviewer, err := svc.AddReviewer(cmd.Context(), args[0], req)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(reviewer)
	}
	fmt.Printf("Reviewer added: %s (status: %s, reviewId: %s)\n", deref(reviewer.Id), deref(reviewer.Status), deref(reviewer.ReviewId))
	return nil
}

// === review accept ===

func newReviewAcceptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accept <review-id>",
		Short: "Accept a review assignment",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewAccept,
	}
}

func runReviewAccept(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	if err := svc.Accept(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Printf("Review %s accepted.\n", args[0])
	return nil
}

// === review decline ===

func newReviewDeclineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decline <review-id>",
		Short: "Decline a review assignment",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewDecline,
	}
}

func runReviewDecline(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	if err := svc.Decline(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Printf("Review %s declined.\n", args[0])
	return nil
}

// === review approve ===

func newReviewApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <review-id>",
		Short: "Approve a story after review",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewApprove,
	}
}

func runReviewApprove(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	if err := svc.Approve(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Printf("Review %s approved.\n", args[0])
	return nil
}

// === review reject ===

func newReviewRejectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject <review-id>",
		Short: "Reject a story after review",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewReject,
	}
	cmd.Flags().String("reason", "", "Reason for rejection")
	return cmd
}

func runReviewReject(cmd *cobra.Command, args []string) error {
	svc, err := newReviewService(cmd)
	if err != nil {
		return err
	}

	reason, _ := cmd.Flags().GetString("reason")
	req := api.ReviewFeedbackRequest{}
	if reason != "" {
		req.Feedback = &reason
	}

	if err := svc.Reject(cmd.Context(), args[0], req); err != nil {
		return err
	}

	fmt.Printf("Review %s rejected.\n", args[0])
	return nil
}
