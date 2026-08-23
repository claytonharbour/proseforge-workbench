package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/activework"
)

// active-work commands (#313). The record a wake is handed alongside the room
// batch, so a fresh session resumes the TASK rather than only answering the room.
//
// Content is the bench's; structure is the product's. These commands validate
// shape, bound size and enforce the version check. They never write a field's
// content and never judge whether an objective is any good.

func newActiveWorkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active-work",
		Short: "Standing task that survives a session boundary",
		Long: `Read and write the standing-work record a watcher hands to a worker.

A turn-based runtime has no background task to resume: a session handed only a
room batch will do the room batch and whatever it was halfway through simply
stops. This record is what makes a fresh session as good as a resumed one.

Keep it small. It should POINT AT evidence — a ticket, a commit, a path — never
contain it.`,
	}

	show := &cobra.Command{
		Use:   "show <path>",
		Short: "Print the current record",
		Args:  cobra.ExactArgs(1),
		RunE:  runActiveWorkShow,
	}

	set := &cobra.Command{
		Use:   "set <path>",
		Short: "Create or update the record",
		Args:  cobra.ExactArgs(1),
		RunE:  runActiveWorkSet,
	}
	set.Flags().String("objective", "", "What is being accomplished")
	set.Flags().String("next", "", "The next CONCRETE action — a thing someone can do, not a status")
	set.Flags().String("ticket", "", "Ticket URL this work is authorized by")
	set.Flags().String("owner", "", "Bench that owns this work")
	set.Flags().String("work-id", "", "Identifier for this piece of work")
	set.Flags().StringArray("evidence", nil, "Pointer to evidence (repeatable)")
	set.Flags().StringArray("blocker", nil, "Something blocking progress (repeatable)")
	set.Flags().Int("version", -1, "Expected current version; refuses if it has moved (-1 skips the check)")
	set.Flags().String("holder", "", "Write through a lease you hold")
	set.Flags().Bool("clear-evidence", false, "Drop existing evidence before applying --evidence")
	set.Flags().Bool("clear-blockers", false, "Drop existing blockers before applying --blocker")

	claim := &cobra.Command{
		Use:   "claim <path>",
		Short: "Take the lease, or find out who holds it",
		Args:  cobra.ExactArgs(1),
		RunE:  runActiveWorkClaim,
	}
	claim.Flags().String("holder", "", "Who is claiming (required)")
	claim.Flags().Duration("ttl", 15*time.Minute, "How long the claim lasts")

	release := &cobra.Command{
		Use:   "release <path>",
		Short: "Drop a lease you hold",
		Args:  cobra.ExactArgs(1),
		RunE:  runActiveWorkRelease,
	}
	release.Flags().String("holder", "", "Who is releasing (required)")

	cmd.AddCommand(show, set, claim, release)
	return cmd
}

func runActiveWorkShow(cmd *cobra.Command, args []string) error {
	rec, err := activework.New(args[0]).Load()
	if errors.Is(err, activework.ErrNotFound) {
		// Exit 0. An empty desk is a legitimate state, and a watcher whose
		// first tick failed because it had no standing work yet would be
		// unusable for every new bench.
		if isJSON(cmd) {
			return printJSON(map[string]any{"active_work": nil})
		}
		fmt.Println("No active work.")
		return nil
	}
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(rec)
	}
	printActiveWork(rec)
	return nil
}

func printActiveWork(r *activework.Record) {
	fmt.Printf("Objective:  %s\n", r.Objective)
	fmt.Printf("Next:       %s\n", r.NextConcreteAction)
	if r.Ticket != "" {
		fmt.Printf("Ticket:     %s\n", r.Ticket)
	}
	if r.Owner != "" {
		fmt.Printf("Owner:      %s\n", r.Owner)
	}
	for i, e := range r.LatestEvidence {
		label := "Evidence:"
		if i > 0 {
			label = "         "
		}
		fmt.Printf("%s   %s\n", label, e)
	}
	for i, b := range r.Blockers {
		label := "Blockers:"
		if i > 0 {
			label = "         "
		}
		fmt.Printf("%s   %s\n", label, b)
	}
	fmt.Printf("Version:    %d   updated %s\n", r.Version, r.UpdatedAt)
	if r.Lease != nil {
		fmt.Printf("Lease:      %s until %s\n", r.Lease.Holder, r.Lease.ExpiresAt)
	}
}

func runActiveWorkSet(cmd *cobra.Command, args []string) error {
	var (
		objective, _ = cmd.Flags().GetString("objective")
		next, _      = cmd.Flags().GetString("next")
		ticket, _    = cmd.Flags().GetString("ticket")
		owner, _     = cmd.Flags().GetString("owner")
		workID, _    = cmd.Flags().GetString("work-id")
		evidence, _  = cmd.Flags().GetStringArray("evidence")
		blockers, _  = cmd.Flags().GetStringArray("blocker")
		version, _   = cmd.Flags().GetInt("version")
		holder, _    = cmd.Flags().GetString("holder")
		clearEv, _   = cmd.Flags().GetBool("clear-evidence")
		clearBl, _   = cmd.Flags().GetBool("clear-blockers")
	)

	store := activework.New(args[0])
	rec, err := store.Update(version, holder, func(r *activework.Record) {
		if objective != "" {
			r.Objective = objective
		}
		if next != "" {
			r.NextConcreteAction = next
		}
		if ticket != "" {
			r.Ticket = ticket
		}
		if owner != "" {
			r.Owner = owner
		}
		if workID != "" {
			r.WorkID = workID
		}
		if clearEv {
			r.LatestEvidence = nil
		}
		if clearBl {
			r.Blockers = nil
		}
		r.LatestEvidence = append(r.LatestEvidence, evidence...)
		r.Blockers = append(r.Blockers, blockers...)
	})
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(rec)
	}
	status("Active work updated to version %d.\n", rec.Version)

	// A record with no next action is the failure this field exists to prevent:
	// a wake can read it, post an update, and count as success. Warn rather than
	// refuse — the first write of a new piece of work legitimately arrives in
	// pieces, and refusing would push people to fill it with a placeholder.
	if rec.NextConcreteAction == "" {
		status("WARNING: no --next set. A wake handed this record has nothing to advance,\n" +
			"         so it will answer the room and call that done.\n")
	}
	return nil
}

func runActiveWorkClaim(cmd *cobra.Command, args []string) error {
	holder, _ := cmd.Flags().GetString("holder")
	if holder == "" {
		return fmt.Errorf("--holder is required: an anonymous lease cannot be told apart from anyone else's")
	}
	ttl, _ := cmd.Flags().GetDuration("ttl")

	rec, err := activework.New(args[0]).Claim(holder, ttl)
	if errors.Is(err, activework.ErrNotFound) {
		return fmt.Errorf("no active work at %s — nothing to claim", args[0])
	}
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(rec)
	}
	fmt.Printf("Claimed by %s until %s (version %d).\n", rec.Lease.Holder, rec.Lease.ExpiresAt, rec.Version)
	return nil
}

func runActiveWorkRelease(cmd *cobra.Command, args []string) error {
	holder, _ := cmd.Flags().GetString("holder")
	if holder == "" {
		return fmt.Errorf("--holder is required")
	}
	if err := activework.New(args[0]).Release(holder); err != nil {
		if errors.Is(err, activework.ErrNotFound) {
			fmt.Println("No active work — nothing to release.")
			return nil
		}
		return err
	}
	fmt.Println("Released.")
	return nil
}
