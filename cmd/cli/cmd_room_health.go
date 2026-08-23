package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// `pfw room health` — is my watcher alive? (#248)
//
// @Gordon and @Sten each hand-rolled this in shell, which is the same signal
// the cursor file gave before #340: when two people independently reimplement
// something against a tool's own state files, the tool is missing a command.

func newRoomHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health <state-dir>[:max-age]...",
		Short: "Is this watcher still polling?",
		Long: `Report whether a watcher's last poll actually succeeded, and how long ago.

A watcher that has STOPPED looks exactly like a room with nothing to report.
Every quiet tick is indistinguishable from death, so the failure is invisible
precisely while it is happening.

This asks the health-poll stamp, which 'room watch' refreshes on every
SUCCESSFUL poll and never on a failure. So its age answers "when did a poll last
actually succeed" without depending on what any exit code means — which matters,
because enumerating exit codes is a losing game and a guard built on a stale
enumeration is silently wrong.

STATES, and they are deliberately not one boolean. "never" and "stale" need
different fixes — check your config, versus check whether the process is
running — so they must not collapse into "unhealthy":
  alive    a poll succeeded within --max-age
  failing  polls ARE being attempted and none succeeds — the process is fine,
           the backend or the credentials are not. Do NOT restart it.
  stale    nothing is polling at all — THE failure case
  never    attempts recorded, no success ever
  unknown  no pfw watcher state here, or the dir cannot be read

Exit: 0 alive · 3 anything else. The code is also PRINTED on line one, because
'health | head -1' returns HEAD's status, so an rc-only caller behind a pipe
reads stale, never and unreadable alike as alive.

🛑 THAT LAST READER IS STILL UNREACHABLE and no change here can fix it — $? is
shell semantics, not ours to grant. Printing rc into the payload rescues the
caller who reads the OUTPUT; a caller who reads only $? behind a pipe still
gets 0. If you must pipe, take the real code explicitly:

  pfw room health "$STATE" --max-age 5m | head -1
  rc=${PIPESTATUS[0]}        # bash/zsh — NOT $?, which is head's

  # or simply do not pipe:
  out=$(pfw room health "$STATE" --max-age 5m); rc=$?`,
		Args: cobra.MinimumNArgs(1),
		RunE: runRoomHealth,
	}
	// --loop (#372). The last three things keeping @Sten's watchdog.sh alive:
	// the loop, state-change-only emit, and the recovery announcement. All three
	// already exist for `room watch --loop`; this gives the watchdog the same
	// treatment rather than leaving 53 lines of shell to reimplement them.
	cmd.Flags().Bool("loop", false, "Check repeatedly instead of exiting. Requires --interval.")
	cmd.Flags().Duration("interval", 0, "Sleep between checks in --loop mode (e.g. 5m). Refused without --loop.")
	cmd.Flags().Bool("no-reexec", false,
		"Do not replace this process when the binary changes. The default swap is how a fix reaches a long-lived watchdog at all.")
	cmd.Flags().String("telemetry", "",
		"Where --loop writes per-cycle telemetry (default: <state>/telemetry.log). Rotated at 5 MiB.")
	// ⚠️ WITHOUT this, alarm state is in memory and dies with the process — so a
	// restart re-alarms about a leg that has been stale all along. This fleet
	// rebuilds the shared binary 20+ times a day and restarts watchers with it;
	// that is one false alarm per restart per failing leg, in the channel the
	// alarm exists to protect. @Sten's watchdog.sh kept its seen-state on disk
	// for exactly this reason, and it is the last thing it did that this did not.
	cmd.Flags().String("state", "",
		"Remember each leg's last state HERE, so a restart does not re-alarm about an unchanged failure. Without it, state is in-memory and lost on restart.")
	cmd.Flags().Duration("max-age", 5*time.Minute,
		"Default: presume dead after this long without a successful poll. MUST be at least 2x your poll "+
			"interval: one failed poll leaves a leg looking dead for ~2 intervals, because it sleeps a full "+
			"interval before retrying. Budget 2x interval + 60s to survive ONE failed poll; measured "+
			"overhead is 46-57s per failure cycle and is FIXED, not a multiple of the interval. "+
			"Override PER LEG with <dir>:<duration> — a 30m poller and a 60s poller cannot share a threshold.")
	return cmd
}

func runRoomHealth(cmd *cobra.Command, args []string) error {
	maxAge, _ := cmd.Flags().GetDuration("max-age")

	loop, interval, err := validateLoopFlags(cmd, "5m")
	if err != nil {
		return err
	}
	if loop {
		return runRoomHealthLoop(cmd, args, maxAge, interval)
	}

	// ⚑ MULTI-LEG (#372). @Tuner deleted 346 lines of watcher shell and kept 53,
	// for exactly one reason: "pfw room health checks ONE state dir and you will
	// run several." @Sten kept his for the same reason rather than claim a clean
	// sweep. This is those last lines.
	//
	// ⚠️ One leg behaves EXACTLY as before — same single line, same rc, same
	// stderr detail. A bench with one leg must not have to learn a new output
	// format to keep working.
	if len(args) > 1 {
		return runRoomHealthMulti(cmd, args, maxAge)
	}

	// ⚠️ parseLegSpec on the SINGLE-leg path too. It was applied only to the
	// multi-leg path, so `room health <dir>:5m` with ONE leg treated the whole
	// string as a directory name and reported `unknown` — a FALSE DEAD REPORT
	// about a healthy watcher, in the most alarming state the command has.
	//
	// 🛑 That is #361 exactly, which I fixed this morning, reintroduced by me
	// this afternoon in a different function. The characterisation test missed
	// it because it used a plain directory — the syntax I had just documented
	// was the one path nothing exercised.
	dir, legAge := parseLegSpec(args[0], maxAge)
	l, err := watcher.CheckLiveness(dir, legAge)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		if err := printJSON(struct {
			PfwVersion string `json:"pfw_version"`
			*watcher.Liveness
		}{Version, l}); err != nil {
			return err
		}
	} else {
		// 🛑 THE EXIT CODE IS PRINTED, not only returned (#367 criterion 3).
		//
		// `pfw room health … | head -1` returns HEAD's status, not this
		// command's, so every rc-only watchdog behind a pipe reads ALIVE for
		// stale, never and unreadable alike. A pipe is exactly what a shell
		// author adds to tidy up output, and it silently converts a detector
		// into one that cannot go red.
		//
		// ⚑ Shell semantics are not ours to fix, so the fix is to put the
		// answer somewhere a pipe CANNOT strip: line one of stdout. rc and the
		// state word can no longer disagree, because they travel together.
		code := 0
		if !l.OK() {
			code = 3
		}
		// ⚑ `rc=` and not `(exit N)`. Two spellings for one concept landed in
		// this tool on the same day — `room watch --loop` emits `rc=12` in its
		// per-tick telemetry (#364) and this emitted `(exit 3)`. @Tate offered
		// to change his; his is the right one to keep. The primary reader here
		// is another agent, and `rc=` is what a grep is already written for.
		fmt.Printf("%-7s rc=%d %s\n", l.State, code, l.Detail)
		// ⚠️ A false-alarming threshold is announced even on an ALIVE tick —
		// that is the only moment the reader is relaxed enough to fix it.
		if l.ConfigWarning != "" {
			fmt.Printf("        %s\n", l.ConfigWarning)
		}
		if l.LastArm != "" && l.State != watcher.LiveAlive {
			status("last attempt %ds ago · last success %ds ago\n", l.ArmAgeSeconds, l.AgeSeconds)
		}
		if l.Cursor != "" {
			status("cursor %s\n", l.Cursor)
		}
	}

	// Non-zero for anything that is not demonstrably alive — including
	// "unknown". A check that cannot read its own input must not report OK.
	if !l.OK() {
		os.Exit(3)
	}
	return nil
}

// runRoomHealthMulti checks several legs in one invocation (#372).
//
// 🛑 The rule that governs the output: a healthy SET says almost nothing, and an
// unhealthy one names only what is wrong. @Wayland measured what the opposite
// costs — 12 wakes, 0 actionable — and the lesson generalises past that notice:
// a checker read by a cron every few minutes must be quiet when there is
// nothing to do, or it teaches its reader to skip it.
//
// ⚠️ A leg that cannot be READ is reported, never skipped. @Sten's case:
// `find -name health-poll` returns nothing for a bench using a native poller,
// so a sweep silently drops the one leg it exists to protect. Absence of a leg
// is a finding, not an empty result.
func runRoomHealthMulti(cmd *cobra.Command, dirs []string, maxAge time.Duration) error {
	type legResult struct {
		Dir string `json:"dir"`
		// MaxAge is the threshold ACTUALLY applied to this leg, echoed so a
		// report cannot hide which one it used. A per-leg override that is
		// silently ignored looks identical to one that worked.
		MaxAge   string            `json:"max_age"`
		Liveness *watcher.Liveness `json:"liveness"`
	}
	results := make([]legResult, 0, len(dirs))
	bad := 0

	for _, spec := range dirs {
		// ⚑ PER-LEG threshold: <dir>[:max-age] (#372). @Sten runs demo at 5m and
		// prod at 70m off one watchdog, and a single --max-age cannot express
		// that: 5m against a 30m poller false-alarms for 25 minutes in every 30,
		// and 70m against a 60s poller is an hour of silence before it notices.
		//
		// ⚠️ I wrote this as an acceptance criterion on #372 and then shipped
		// without it, closing the ticket on the strength of the criteria I had
		// written rather than the ones I had met.
		dir, legAge := parseLegSpec(spec, maxAge)
		l, err := watcher.CheckLiveness(dir, legAge)
		if err != nil {
			// Report, do not abort: one unreadable leg must not hide the state
			// of the others. Silently returning here would turn a four-leg check
			// into a one-leg check with no indication it had.
			l = &watcher.Liveness{State: watcher.LiveUnknown, Detail: err.Error()}
		}
		if !l.OK() {
			bad++
		}
		results = append(results, legResult{Dir: dir, MaxAge: legAge.String(), Liveness: l})
	}

	if isJSON(cmd) {
		if err := printJSON(struct {
			PfwVersion string      `json:"pfw_version"`
			Legs       []legResult `json:"legs"`
			Unhealthy  int         `json:"unhealthy"`
			Total      int         `json:"total"`
		}{Version, results, bad, len(results)}); err != nil {
			return err
		}
		if bad > 0 {
			os.Exit(3)
		}
		return nil
	}

	code := 0
	if bad > 0 {
		code = 3
	}
	// One verdict line first, so a reader piping to `head -1` gets the answer
	// rather than the first leg — the #367 lesson, applied to a set.
	fmt.Printf("%d/%d unhealthy rc=%d\n", bad, len(results), code)

	for _, r := range results {
		// ⚑ Every leg is NAMED, healthy or not. A summary that lists only
		// failures cannot be distinguished from one that failed to look:
		// "0/4 unhealthy" and "0/0 checked" both print as good news.
		fmt.Printf("  %-7s %-46s %s\n", r.Liveness.State, shortDir(r.Dir), r.MaxAge)
		if !r.Liveness.OK() {
			fmt.Printf("          %s\n", r.Liveness.Detail)
		}
	}

	if bad > 0 {
		os.Exit(3)
	}
	return nil
}

// shortDir trims the home prefix so a four-leg report stays readable without
// losing which leg is which.
func shortDir(dir string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, home) {
		return "~" + dir[len(home):]
	}
	return dir
}

// runRoomHealthLoop is the watchdog: check every leg on a timer, and speak only
// when a leg CHANGES state (#372).
//
// 🛑 One alarm PER LEG, not one for the set. A shared alarm would announce
// "recovered" when a different leg recovered than the one that failed, and stay
// silent about a second leg failing while the first was already down. @Sten runs
// two legs; the states are independent and the reporting has to be too.
//
// ⚠️ This is the piece that lets a watchdog script be deleted, and it is the one
// I keep having to re-derive: a checker that reports every tick is a checker its
// operator learns to skip. Silence IS the healthy output.
func runRoomHealthLoop(cmd *cobra.Command, specs []string, maxAge, interval time.Duration) error {
	stateDir, _ := cmd.Flags().GetString("state")

	// 🛑 ONE WATCHDOG PER STATE DIR (@Tuner, #363). `room watch` has refused a
	// second instance since day one (rc=11); this loop did not, and @Tuner ran
	// two over the same legs. The older one held a STALE LEG LIST and alarmed
	// about a leg he had already retired.
	//
	// ⚠️ What makes it worth a guard rather than a note: NEITHER PROCESS'S
	// OUTPUT SHOWED ANYTHING WRONG. The only signal was `pgrep` returning two
	// pids — a thing nobody checks until afterwards. Same exit code as
	// `room watch` so the condition has one vocabulary, not two.
	if stateDir != "" {
		// ⚠️ CREATE IT FIRST. `room watch --state` creates its dir; this did not,
		// so adding the lock turned a missing dir into "take lock: open
		// <dir>/.lock: no such file or directory" — an error naming a file the
		// caller never mentioned, for a dir they simply hadn't made yet
		// (@Tate, found within minutes of the guard landing).
		if err := os.MkdirAll(stateDir, 0o755); err != nil {
			return fmt.Errorf("create --state dir %s: %w", stateDir, err)
		}
		// 🛑 REWRITTEN after it broke two benches within minutes (@Wayland #392,
		// @Tuner). My first version inferred "somebody is running" from leg-file
		// freshness alone, and resolved every ambiguity toward "refuse".
		//
		// ⚑ That is safe for a duplicate check and UNSAFE FOR AVAILABILITY: it
		// turned an unclean exit into a permanent outage of the thing whose job
		// is detecting outages. @Wayland could not recover by doing what the
		// message said — "stop it first" is impossible when nothing is running,
		// and every retry re-refused off mtimes written by the corpse. @Tuner
		// stopped a watchdog cleanly and could not start another for 5 minutes.
		//
		// ✅ Both cases are DECIDABLE. Neither needs a timeout:
		//   .lock holds a LIVE pid   → a real duplicate. Refuse, name the pid.
		//   .lock holds a DEAD pid   → unclean exit. Say so, take over.
		//   no .lock, leg files      → sample twice. STILL ADVANCING means a
		//                              pre-guard incumbent; static means a corpse.
		if pid, alive, found := lockHolder(stateDir); found {
			// 🛑 OUR OWN LOCK, INHERITED ACROSS exec (@Tuner, urgent).
			//
			// syscall.Exec REPLACES the process image and KEEPS THE PID. Deferred
			// releases never run, so the lock survives holding our own pid — and
			// the new image then reads it, confirms that pid is alive (it is: it
			// is us), and refuses to start. The watchdog locks ITSELF out on
			// every published build, then exits, and the thing that would have
			// reported the outage is the thing that died.
			//
			// ⚑ Both features worked exactly as specified. The combination was
			// lethal, and it only appears when re-exec and the lock live in the
			// same command — room watch re-execs and takes no lock, which is why
			// this never showed there.
			if pid == os.Getpid() {
				status("resuming after re-exec — reclaiming our own lock (pid %d)\n", pid)
			} else if alive {
				fmt.Printf("🛑 ANOTHER WATCHDOG IS ALREADY RUNNING (pid %d) on %s — refusing to start a second.\n"+
					"   Two watchdogs over one state dir drift apart silently: the older keeps the leg\n"+
					"   list it started with, so it alarms about legs you have since retired.\n", pid, stateDir)
				os.Exit(11)
			}
			// ⚠️ Take over rather than refuse. A dead holder is the ordinary
			// aftermath of a kill, and refusing here is what left @Wayland with
			// a state dir he could only abandon.
			status("stale lock from pid %d (not running) — taking over\n", pid)
			_ = os.Remove(filepath.Join(stateDir, ".lock"))
		} else if advancing, age := legsAdvancing(stateDir); advancing {
			fmt.Printf("🛑 A WATCHDOG IS RUNNING HERE and cannot be locked out — refusing to start a second.\n"+
				"   %s has no .lock, but its leg files are STILL BEING WRITTEN (last %s ago).\n"+
				"   That is an incumbent on a build older than this guard. Stop it, then start.\n",
				stateDir, age.Round(time.Second))
			os.Exit(11)
		}

		release, err := watcher.AcquireStateLock(stateDir)
		if err != nil {
			if errors.Is(err, watcher.ErrLocked) {
				fmt.Printf("🛑 ANOTHER WATCHDOG IS ALREADY RUNNING on %s — refusing to start a second.\n"+
					"   Two watchdogs over one state dir drift apart silently: the older keeps the leg\n"+
					"   list it started with, so it alarms about legs you have since retired.\n", stateDir)
				os.Exit(11)
			}
			return err
		}
		defer release()
	}

	// #382 consistency: `room watch --loop` routes per-tick telemetry to a file;
	// this loop still put its per-cycle lines on stderr, which is the terminal
	// for anyone not under a Monitor. @Wayland measured 4 lines every 2 minutes
	// with nothing wrong — 120 lines/hour of "everything is fine".
	//
	// ⚑ His diagnosis was that `room health --loop` needs a --quiet. Measured:
	// STDOUT is already silent when healthy (0 lines across 3 cycles), so it was
	// never waking anyone. The defect is that routine telemetry was on stderr at
	// all — the same inconsistency #382 removed one command over.
	telPath, _ := cmd.Flags().GetString("telemetry")
	tel := openTelemetry(stateDir, telPath)
	defer tel.Close()
	defer tel.capturePanic()
	tel.setHeader("watchdog — %d leg(s) every %s — pfw=%s", len(specs), interval, Version)

	// 🛑 SELF-RE-EXEC, same as room watch (#371). @Tuner found the gap: a lock is
	// only taken at STARTUP, so a watchdog that predates the duplicate guard
	// never acquires one — and with no re-exec it runs its original image
	// forever, so the guard could never arrive. That is true of EVERY future fix
	// to this command, not just the lock.
	//
	// ⚑ "The feature that removes the need for restarts requires exactly one
	// restart to arrive" (@Tate, about re-exec this morning). A lock has the
	// same property and worse odds — self-re-exec at least propagates on the
	// next tick. Without this, a long-lived watchdog holds the hole open
	// indefinitely and nothing tells you which processes those are.
	armedFingerprint, _ := binaryFingerprint()
	noReexec, _ := cmd.Flags().GetBool("no-reexec")

	seeded := make(map[string]bool, len(specs))
	alarms := make(map[string]*roomAlarm, len(specs))
	tel.warn("watchdog up — %d leg(s), checking every %s — telemetry: %s",
		len(specs), interval, tel.Path())

	return loopUntilSignal(cmd, interval, func(context.Context) error {
		for _, spec := range specs {
			dir, legAge := parseLegSpec(spec, maxAge)
			label := filepath.Base(filepath.Dir(dir))
			if label == "." || label == "/" {
				label = dir
			}
			if alarms[dir] == nil {
				alarms[dir] = newRoomAlarm(1) // watchdog alarms on the first bad check; its own cadence is --interval
			}

			l, err := watcher.CheckLiveness(dir, legAge)
			if err != nil {
				l = &watcher.Liveness{State: watcher.LiveUnknown, Detail: err.Error()}
			}
			// 🛑 STDOUT: a Monitor raises events from stdout only, and this line
			// is the entire reason the watchdog exists.
			// ⚑ Seed from disk on the first observation of a leg, so a restart
			// inherits what the previous process already announced.
			if !seeded[dir] {
				seedAlarmFromDisk(alarms[dir], stateDir, label, l.State)
				seeded[dir] = true

				// #400 (@Vance): say ONCE per leg whether this watchdog's own
				// cadence can resolve the threshold it is checking.
				//
				// 🛑 On STDERR and once, not stdout and not per cycle. It is a
				// statement about the CONFIG, which cannot change while we run,
				// so repeating it every interval would be the alarm-fatigue this
				// command exists to avoid — and putting it on stdout would wake
				// a Monitor for something that is not an outage.
				//
				// ⚠️ It needs the leg's observed interval, which only exists
				// after a leg has armed twice — so it is silent on a leg that
				// has not yet established a cadence, and that silence is honest.
				if l.IntervalSeconds > 0 {
					if w := watcher.SamplingWarning(
						time.Duration(l.IntervalSeconds)*time.Second, legAge, interval); w != "" {
						tel.warn("%s: %s", label, w)
					}
				}
			}
			if line := alarms[dir].observeLeg(l.OK(), label, l.State, l.Detail); line != "" {
				fmt.Println(line)
			}
			persistLegState(stateDir, label, l.State)
			// FILE, not stderr (#382): this is the per-cycle log. A log must be
			// complete; a wake must be rare. Alarms below still go to stdout.
			tel.log("wd %s=%s max-age=%s", label, l.State, legAge)
		}

		// ⚠️ Only when every leg is healthy. Alarm state ("already announced")
		// lives in memory and does not survive an exec, so swapping mid-alarm
		// would re-announce an outage that never ended.
		if !noReexec && !anyAlarming(alarms) {
			reexecIfBinaryChanged(&armedFingerprint, tel)
		}

		return nil
	})
}

// anyAlarming reports whether any leg is currently in an announced failure, so
// a binary swap can wait for quiet.
func anyAlarming(alarms map[string]*roomAlarm) bool {
	for _, a := range alarms {
		if a != nil && a.failing {
			return true
		}
	}
	return false
}

// lockHolder reports the pid in <stateDir>/.lock and whether it is still alive.
//
// ⚑ A pid is DECIDABLE where freshness is a guess. The lock already records
// one; nothing was reading it.
func lockHolder(stateDir string) (pid int, alive, found bool) {
	raw, err := os.ReadFile(filepath.Join(stateDir, ".lock"))
	if err != nil {
		return 0, false, false
	}
	pid, err = strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || pid <= 0 {
		return 0, false, false
	}
	// Signal 0 tests for existence without delivering anything. EPERM means the
	// process exists and belongs to someone else — still alive.
	err = syscall.Kill(pid, 0)
	return pid, err == nil || errors.Is(err, syscall.EPERM), true
}

// legsAdvancing samples the newest leg-* mtime twice and reports whether it
// MOVED — the difference between an incumbent and a corpse.
//
// 🛑 Freshness alone cannot tell those apart, and my first version treated them
// as one. A watchdog stopped 70 seconds ago leaves files exactly as fresh as one
// running now (@Tuner), so a timeout can only choose which error to make. Two
// samples decide it: a live incumbent writes again; a corpse never does.
func legsAdvancing(stateDir string) (bool, time.Duration) {
	first, ok := newestLegStamp(stateDir)
	if !ok || first > 5*time.Minute {
		return false, first
	}
	time.Sleep(legSampleGap)
	second, ok := newestLegStamp(stateDir)
	if !ok {
		return false, first
	}
	// If the newest stamp got NEWER, someone wrote during our sample.
	return second < first, second
}

// legSampleGap must exceed the fastest plausible watchdog cycle so one write
// lands inside it. Paid only when leg files look fresh — never on a cold start.
const legSampleGap = 3 * time.Second

// newestLegStamp reports how long ago this watchdog state dir was last written.
// Only leg-* files count: telemetry.log is touched by US before this check.
func newestLegStamp(stateDir string) (time.Duration, bool) {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return 0, false
	}
	var newest time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "leg-") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.ModTime().After(newest) {
			newest = fi.ModTime()
		}
	}
	if newest.IsZero() {
		return 0, false
	}
	return time.Since(newest), true
}
