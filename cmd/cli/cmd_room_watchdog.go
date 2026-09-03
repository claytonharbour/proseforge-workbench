package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// `pfw room watchdog` — the thing that RUNS `room health` (#351).
//
// #248 shipped the check and left the reader missing: a dead watcher still
// reports nothing, because the thing that would report is the thing that died.
// @Wayland built this as shell, proved every alarm path fires before trusting
// its silence, and offered it to the tool rather than let nine benches each keep
// a copy. The subtle parts are all his.

func newRoomWatchdogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watchdog <watched-state-dir>",
		Short: "Alarm when a watcher stops polling — and when it recovers",
		Long: `Check a watcher's liveness and speak ONLY when the answer changes.

Single-shot: one check, one comparison, exit. A separate monitor, a cron, or a
human decides when it runs. It never schedules itself.

⚠️ RUN IT AS A SEPARATE PROCESS FROM YOUR POLL LOOP. That is the whole
mechanism. A sequential loop with a HUNG 'room watch' stalls every leg and emits
nothing, because a hang never produces a non-zero exit code — so a watchdog
living inside that loop dies with it and stays silent.

WHY NOT JUST 'room health' IN A LOOP:

  It compares the STATE WORD, not the exit code. Exit 3 covers stale, never AND
  unknown, so a watchdog built on the code alone cannot tell "check your config"
  from "check whether the process is running".

  It speaks only on a TRANSITION. Repeating every tick while stale floods the
  channel you are trying to alert on, and the alarm gets muted.

  It alarms on RECOVERY too. An alarm that goes red and never green is a stuck
  alarm, and it looks identical to a permanent outage.

A first run finding a healthy watcher is SILENT — announcing "alive" at startup
teaches the reader this channel carries routine noise. A first run finding an
alarm does speak; that is news.

🛑 IT ONLY UNDERSTANDS 'pfw room watch' STATE DIRS, and you must name them
explicitly. Point it at a watcher built some other way and it reports "never" —
a FALSE ALARM about a healthy process.

  Never auto-discover these directories. Scanning the filesystem for the stamp
  file finds only watchers that use this tool, and reports every other kind as
  dead. That is backwards: the watchers least like yours are the ones a watchdog
  most needs to be honest about, and a sweep like that raises its loudest alarm
  about a bench that is perfectly fine while staying silent about one that is
  genuinely gone but leaves no directory at all.

🛑 WHAT IT DOES NOT COVER: the session ending, or every monitor dying at once.
Same session, dies with them, says nothing. That is an accepted trade, not a
closed gap.

🛑 THE EXIT CODE DESCRIBES THIS WATCHDOG, NOT THE WATCHER IT WATCHES. Conflating
those is the category error this whole area keeps making.

  0  the watchdog did its job — INCLUDING when it found the watcher dead
  1  the watchdog could NOT do its job (misconfigured, cannot write its state)

A detected outage exits 0 on purpose. Returning non-zero there tells a supervisor
the check failed, so it restarts the one process that was working correctly and
throws away the state-change memory that stops the alarm repeating.

Misuse is LOUD and goes to STDOUT, never quietly to stderr. A silent broken
watchdog is indistinguishable from a healthy fleet, which is the exact failure
this command exists to prevent — so it says "NOTHING is being checked" rather
than nothing at all.

'pfw room health' is the gate and exits 0/3 on the WATCHER's state. This is the
reporter.`,
		// 🛑 NOT ExactArgs(1). Cobra's arity failure prints to STDERR and a monitor
		// raises no event from stderr — so a watchdog invoked wrongly inside a
		// `while true` loop emits one invisible line per tick and reports nothing,
		// which is exactly "a broken watchdog looks like a healthy fleet"
		// (@Sten, verifying bbbe3bb). Arity is checked in RunE so the misuse
		// banner goes where the alarm goes.
		Args: cobra.ArbitraryArgs,
		RunE: runRoomWatchdog,
	}
	cmd.Flags().Duration("max-age", 5*time.Minute,
		"Presume dead after this long without a successful poll. A small multiple of the poll interval.")
	cmd.Flags().String("state", "", "The WATCHDOG's own state dir, separate from the watcher's (required)")
	return cmd
}

func runRoomWatchdog(cmd *cobra.Command, args []string) error {
	maxAge, _ := cmd.Flags().GetDuration("max-age")
	state, _ := cmd.Flags().GetString("state")

	// 🛑 Misuse is LOUD, on STDOUT, and says NOTHING IS BEING CHECKED.
	//
	// @Wayland proved this path in shell before trusting the tool's silence, and
	// it is the one that matters most: a broken watchdog and a healthy fleet
	// produce identical quiet. Returning a cobra error prints to stderr, which a
	// monitor raises no event from — so the operator would see nothing, from the
	// one command whose whole job is to speak when nobody else can.
	// ⚠️ EVERY arity failure, not just the one that was reported. Fixing only
	// "zero args" would leave "two args" on cobra's stderr path — the same defect
	// one step over, and the next person to find it would be finding it in
	// production. Arity is checked here so misuse always reaches stdout.
	if len(args) != 1 {
		what := "no watcher state dir given"
		if len(args) > 1 {
			what = fmt.Sprintf("%d state dirs given, this checks ONE", len(args))
		}
		fmt.Printf("🛑 WATCHDOG BROKEN — %s. NOTHING is being checked.\n", what)
		fmt.Println("   This is NOT an all-clear.")
		fmt.Println("   Usage: pfw room watchdog <watched-state-dir> --state <watchdog-state-dir>")
		os.Exit(1)
	}
	if state == "" {
		fmt.Println("🛑 WATCHDOG BROKEN — --state is missing. NOTHING is being checked.")
		fmt.Println("   This is NOT an all-clear. The watchdog needs its own state dir,")
		fmt.Println("   separate from the watcher's, so it cannot corrupt what it observes.")
		os.Exit(1)
	}

	res, err := watcher.Watchdog(args[0], state, maxAge)
	if err != nil {
		// Same reasoning: the watchdog failing must not be quiet.
		fmt.Printf("🛑 WATCHDOG BROKEN — %v. NOTHING is being checked.\n", err)
		fmt.Println("   This is NOT an all-clear.")
		os.Exit(1)
	}

	if isJSON(cmd) {
		if err := printJSON(struct {
			PfwVersion string `json:"pfw_version"`
			*watcher.WatchdogResult
		}{Version, res}); err != nil {
			return err
		}
	} else if res.Message != "" {
		// ⚑ stdout, because a monitor turns stdout lines into events and stderr
		// into nothing. An alarm on stderr is an alarm nobody is woken by.
		fmt.Println(res.Message)
	}

	// ⚑ Deliberately NOT exiting non-zero on a detected outage (@Sten). The
	// watched watcher being dead is this command working, not failing. A
	// supervisor that restarts on non-zero would kill the healthy watchdog and
	// discard the state-change memory, turning one alarm into a repeating flood.
	return nil
}
