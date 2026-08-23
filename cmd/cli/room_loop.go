package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// Shared scaffolding for the two commands that poll on a timer (#377).
//
// 🛑 Why this file exists: building `room health --loop` I wrote a second copy of
// `room watch --loop`'s flag validation and signal loop, and a third copy of the
// validation survived as dead code. Three near-identical validators is how one
// of them drifts and nobody notices — the same duplication #363 exists to end,
// committed by whoever is running #363.

// validateLoopFlags enforces the --loop/--interval pair.
//
// ⚠️ Called BEFORE any config is read. An operator who mistypes a flag must get
// the flag error, not "PROSEFORGE_URL is required" — otherwise they fix the
// config and hit the real error on a second run.
//
// ⚑ Refusals print to STDOUT, deliberately. A Monitor raises events from stdout
// only, so a refusal on stderr reaches nobody — and a watcher that silently
// declined to start is exactly the failure this whole subsystem exists to catch.
func validateLoopFlags(cmd *cobra.Command, example string) (loop bool, interval time.Duration, err error) {
	loop, _ = cmd.Flags().GetBool("loop")
	interval, _ = cmd.Flags().GetDuration("interval")
	switch {
	case loop && interval <= 0:
		fmt.Printf("--loop requires --interval (e.g. --interval %s). Refusing to spin.\n", example)
		return loop, interval, fmt.Errorf("--loop requires --interval")
	case !loop && interval > 0:
		// A no-op that LOOKS configured: someone who sets --interval believes
		// they are polling repeatedly. Running once and exiting silently is the
		// failure class this subsystem exists to remove.
		fmt.Printf("--interval %s has no effect without --loop.\n", interval)
		return loop, interval, fmt.Errorf("--interval requires --loop")
	}
	return loop, interval, nil
}

// loopUntilSignal runs tick on a timer until SIGINT/SIGTERM.
//
// tick returning an error ENDS the loop; returning nil continues. That choice is
// per-caller and load-bearing: `room watch` ends only on a permanent identity
// failure and survives outages, because a watcher that dies on the event it
// exists to report is worse than no watcher.
//
// ⚠️ The context is passed to tick so a poll in flight is cancelled on Ctrl-C
// rather than the loop waiting a full interval to notice.
func loopUntilSignal(cmd *cobra.Command, interval time.Duration, tick func(ctx context.Context) error) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		if err := tick(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
		}
	}
}

// parseLegSpec splits "<dir>[:max-age]" (#372).
//
// ⚠️ LastIndex, not Index: a colon inside a path must not be mistaken for the
// separator. An unparseable suffix stays part of the path rather than being
// silently eaten as a threshold — guessing there would resolve a real directory
// to a different one and report on a watcher the caller never asked about.
func parseLegSpec(spec string, fallback time.Duration) (dir string, maxAge time.Duration) {
	if i := strings.LastIndex(spec, ":"); i > 0 {
		if parsed, err := time.ParseDuration(spec[i+1:]); err == nil {
			return spec[:i], parsed
		}
	}
	return spec, fallback
}

// seedAlarmFromDisk makes a restarted watchdog inherit what the previous process
// already said (#377).
//
// 🛑 Without this, alarm state dies with the process: a restart re-announces a
// leg that has been failing all along, and the operator learns that a restart
// means noise rather than news. @Sten's shell watchdog kept this on disk and
// mine did not — the one thing it still did better.
//
// ⚠️ Seeds ONLY the failing state, never the healthy one. Seeding "was alive"
// would suppress the entry alarm for a leg that broke while the watchdog was
// down — silence about a failure that happened in the gap, which is the worst
// thing a watchdog can do.
func seedAlarmFromDisk(a *roomAlarm, stateDir, label, current string) {
	if stateDir == "" || a == nil {
		return
	}
	prev, err := os.ReadFile(filepath.Join(stateDir, "leg-"+label))
	if err != nil {
		return
	}
	if strings.TrimSpace(string(prev)) == current && current != "alive" {
		// Already announced by a previous process, and nothing has changed.
		a.failing, a.firstFail = true, a.now()
	}
}

// persistLegState records a leg's state for the next process. Errors are
// swallowed: this is diagnostic continuity, and a watchdog that refuses to watch
// because it could not write a file has turned a nicety into an outage.
func persistLegState(stateDir, label, state string) {
	if stateDir == "" {
		return
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(stateDir, "leg-"+label), []byte(state+"\n"), 0o644)
}
