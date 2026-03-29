package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

func newReviewerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviewer",
		Short: "Reviewer pool operations",
	}
	cmd.AddCommand(
		newReviewerListCmd(),
		newReviewerRequestCmd(),
		newReviewerRespondCmd(),
	)
	return cmd
}

// === reviewer list ===

func newReviewerListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your accepted reviewers",
		RunE:  runReviewerList,
	}
}

func runReviewerList(cmd *cobra.Command, args []string) error {
	svc, err := newReviewerService(cmd)
	if err != nil {
		return err
	}

	reviewers, err := svc.ListMy(cmd.Context())
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(reviewers)
	}

	if len(reviewers) == 0 {
		fmt.Println("No reviewers.")
		return nil
	}

	var rows [][]string
	for _, r := range reviewers {
		rows = append(rows, []string{
			deref(r.Id),
			deref(r.ReviewerName),
			deref(r.Status),
		})
	}

	printTable([]string{"ID", "Name", "Status"}, rows)
	return nil
}

// === reviewer request ===

func newReviewerRequestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "request <user-id>",
		Short: "Request someone as a reviewer",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewerRequest,
	}
}

func runReviewerRequest(cmd *cobra.Command, args []string) error {
	svc, err := newReviewerService(cmd)
	if err != nil {
		return err
	}

	userID := args[0]
	req := api.CreateReviewerRequestReq{ReviewerId: &userID}

	if err := svc.Request(cmd.Context(), req); err != nil {
		return err
	}

	fmt.Printf("Reviewer request sent to %s.\n", userID)
	return nil
}

// === reviewer respond ===

func newReviewerRespondCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "respond <request-id>",
		Short: "Accept or decline a reviewer request",
		Args:  cobra.ExactArgs(1),
		RunE:  runReviewerRespond,
	}
	cmd.Flags().Bool("accept", false, "Accept the request")
	cmd.Flags().Bool("decline", false, "Decline the request")
	cmd.MarkFlagsMutuallyExclusive("accept", "decline")
	return cmd
}

func runReviewerRespond(cmd *cobra.Command, args []string) error {
	accept, _ := cmd.Flags().GetBool("accept")
	decline, _ := cmd.Flags().GetBool("decline")

	if !accept && !decline {
		return fmt.Errorf("specify --accept or --decline")
	}

	svc, err := newReviewerService(cmd)
	if err != nil {
		return err
	}

	action := "accepted"
	acceptVal := true
	if decline {
		action = "declined"
		acceptVal = false
	}

	req := api.RespondToReviewerReq{Accept: &acceptVal}
	if err := svc.Respond(cmd.Context(), args[0], req); err != nil {
		return err
	}

	fmt.Printf("Reviewer request %s %s.\n", args[0], action)
	return nil
}
