package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/activework"
	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// watch commands (#312). The notify seam — the one place a runtime is allowed to
// exist. Everything about how a worker is invoked and authenticated lives behind
// an external command, so this product never names claude, codex, or anything
// else. A public BYOAI tool cannot shell out to a subscription product.

func newWatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Room watcher: buffer messages and hand them to a worker",
	}
	cmd.AddCommand(newWatchNotifyCmd())
	return cmd
}

func newWatchNotifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Run an adapter against the wake contract and report the outcome",
		Long: `Invoke a notify adapter and classify what came back.

This is the seam a watcher wakes a worker through, exposed on its own so an
adapter can be verified against the contract without running a whole watcher —
and without waiting for a real room message to arrive.

THE CONTRACT an adapter must satisfy:

  argv    $1 batch file   $2 prompt file   $3 active-work file ("" if none)
  stdout  the worker's reply, or the literal NO_POST
  exit    0  ok — the handoff was accepted
          10 CANNOT WAKE — no runtime, no credential, nothing to talk to
          1  woke and broke

Exit 10 must never be folded into 1. A deaf bench and a broken bench are
different problems, and only one of them announces itself: broken is loud by
nature, deaf stays silent and silence looks like calm.

Exit 0 means the handoff was accepted. It does NOT acknowledge anything — only
the consumer's own explicit ack retires a message.`,
		RunE: runWatchNotify,
	}
	cmd.Flags().String("adapter", "", "Adapter command to run (empty = buffering-only)")
	cmd.Flags().String("batch", "", "Batch file handed to the adapter as $1 (required)")
	cmd.Flags().String("prompt", "", "Prompt file handed to the adapter as $2")
	cmd.Flags().String("active-work", "", "Active-work file handed to the adapter as $3")
	cmd.Flags().String("wake-id", "", "Identifier for this wake (reaches the adapter as WATCHER_WAKE_ID)")
	cmd.Flags().Duration("timeout", watcher.DefaultTimeout, "Bound on the wake")
	cmd.Flags().StringArray("env", nil, "Extra environment for the adapter, KEY=VALUE (repeatable)")
	return cmd
}

func runWatchNotify(cmd *cobra.Command, args []string) error {
	var (
		adapter, _  = cmd.Flags().GetString("adapter")
		batch, _    = cmd.Flags().GetString("batch")
		prompt, _   = cmd.Flags().GetString("prompt")
		active, _   = cmd.Flags().GetString("active-work")
		wakeID, _   = cmd.Flags().GetString("wake-id")
		timeout, _  = cmd.Flags().GetDuration("timeout")
		extraEnv, _ = cmd.Flags().GetStringArray("env")
	)
	if batch == "" {
		return errors.New("--batch is required")
	}
	if wakeID == "" {
		wakeID = fmt.Sprintf("wake-%d", time.Now().UnixNano())
	}

	// Validate the active-work record before handing it over. A corrupt record
	// would otherwise reach the worker as an empty standing task, which reads as
	// "nothing to do" — the work is silently dropped and every signal stays green.
	var activeRaw []byte
	if active != "" {
		if _, err := activework.New(active).Load(); err != nil && !errors.Is(err, activework.ErrNotFound) {
			return fmt.Errorf("refusing to wake with an unusable active-work record: %w", err)
		}
		activeRaw, _ = os.ReadFile(active)
	}

	// Inline the batch for adapters that read stdin rather than argv. A read
	// failure here is not fatal: the path is still on argv, so an argv-style
	// adapter works regardless.
	batchRaw, _ := os.ReadFile(batch)

	res, err := watcher.Notify(cmd.Context(), watcher.Notification{
		Command:        adapter,
		BatchPath:      batch,
		PromptPath:     prompt,
		ActiveWorkPath: active,
		WakeID:         wakeID,
		Env:            extraEnv,
		Timeout:        timeout,
		Messages:       batchRaw,
		ActiveWork:     activeRaw,
	})
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		if err := printJSON(map[string]any{
			"wake_id":     wakeID,
			"outcome":     res.Outcome,
			"exit_code":   res.ExitCode,
			"should_post": res.ShouldPost(),
			"timed_out":   res.TimedOut,
			"duration_ms": res.Duration.Milliseconds(),
			"reply":       res.Reply,
			"stderr":      res.Stderr,
		}); err != nil {
			return err
		}
	} else {
		// Content to stdout, status to stderr: the reply is the useful output and
		// should survive a pipe on its own.
		status("wake %s — %s (exit %d, %s)\n", wakeID, res.Outcome, res.ExitCode, res.Duration.Round(time.Millisecond))
		if res.Stderr != "" {
			status("adapter stderr: %s\n", truncate(res.Stderr, 500))
		}
		if res.ShouldPost() {
			fmt.Println(res.Reply)
		} else if res.Outcome == watcher.OutcomeOK {
			status("worker chose not to post\n")
		}
	}

	// The outcome IS the exit status, so cron and any caller can branch on it
	// without parsing output. This is the distinction the whole seam exists for.
	switch res.Outcome {
	case watcher.OutcomeCannotWake:
		os.Exit(10)
	case watcher.OutcomeFailed:
		os.Exit(1)
	}
	return nil
}
