package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The watchdog stage (forge/proseforge-workbench#351).
//
// `room health` answers "is this watcher alive?" once. Nothing ran it, which was
// the open half of #248 — a dead watcher still reports nothing, because the thing
// that would report is the thing that died. @Wayland built and PROVED the missing
// piece as a shell script, then offered it to the tool rather than let nine
// benches each keep a copy of subtle logic. This is that logic.
//
// Three things a naive `room health` loop gets wrong, all of them his:
//
//  1. 🛑 It compares the STATE WORD, not the exit code. rc=3 covers stale, never
//     AND unknown, so a watchdog built on rc cannot tell "check your config" from
//     "check whether the process is running" — the exact distinction room health
//     exists to draw.
//
//  2. It speaks only on a TRANSITION. A watchdog that repeats every tick while
//     stale floods the channel it is alerting on, and gets muted.
//
//  3. ⚑ It alarms on RECOVERY too. His line, and the one most people skip:
//     "an alarm that goes red and never green is a stuck alarm, and it looks
//     identical to a permanent outage."
//
// ⚠️ It must run as a SEPARATE process from the poll loop. Wayland found this
// live: a sequential loop with a HUNG `room watch` stalls every leg and emits
// nothing, because a hang never produces a non-zero exit code. A watchdog inside
// that loop dies with it and stays silent.
//
// 🛑 What it still does not cover: the session ending, or every monitor dying at
// once. Same session, dies with them, says nothing. That is an accepted trade,
// not a closed gap.

// fileWatchdogState records the last state this watchdog reported on, so it can
// speak on transitions instead of every tick.
const fileWatchdogState = "watchdog-last"

// WatchdogResult is one check.
type WatchdogResult struct {
	// State is the current liveness state word.
	State string `json:"state"`
	// Previous is what was reported last time; empty on first run.
	Previous string `json:"previous,omitempty"`
	// Changed is whether this is a transition worth speaking about.
	Changed bool `json:"changed"`
	// Recovered marks the alarm->alive edge specifically.
	Recovered bool `json:"recovered,omitempty"`

	Liveness *Liveness `json:"liveness"`

	// Message is what to print. Empty when there is nothing to say.
	Message string `json:"message,omitempty"`
}

// Alarming reports whether the watched watcher is currently unhealthy.
func (r *WatchdogResult) Alarming() bool { return r.State != LiveAlive }

// AcquireStateLock takes the state-dir lock for a caller that owns a state dir
// but is not `room watch` — today, `room health --loop` (@Tuner, #363).
//
// 🛑 THE GAP IT CLOSES: `room watch` refuses to run twice on one state dir
// (rc=11), and `room health --loop` did not. @Tuner ran two watchdogs over the
// same legs; the older held a stale leg list and alarmed about a leg he had
// retired. **Neither process's output showed anything wrong** — the only signal
// was `pgrep` returning two pids, which is exactly the kind of thing nobody
// looks at until afterwards.
//
// ⚠️ Returns ErrLocked so a caller can exit 11 and match `room watch`'s
// contract rather than inventing a second vocabulary for the same condition.
func AcquireStateLock(stateDir string) (func(), error) {
	return acquireLock(stateDir)
}

// Watchdog checks one watcher's liveness and reports only on a state change.
//
// watchedStateDir is the state dir of the watcher being checked; stateDir is the
// watchdog's own, kept separate so a watchdog cannot corrupt what it observes.
func Watchdog(watchedStateDir, stateDir string, maxAge time.Duration) (*WatchdogResult, error) {
	if watchedStateDir == "" {
		return nil, fmt.Errorf("no watched state dir")
	}
	if stateDir == "" {
		return nil, fmt.Errorf("no watchdog state dir")
	}

	live, err := CheckLiveness(watchedStateDir, maxAge)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, fmt.Errorf("watchdog state dir: %w", err)
	}
	prev := readTrimmed(filepath.Join(stateDir, fileWatchdogState))

	res := &WatchdogResult{State: live.State, Previous: prev, Liveness: live}

	// ⚠️ First run with a healthy watcher is deliberately SILENT. Announcing
	// "alive" on startup trains the reader that this channel carries routine
	// noise, and a watchdog is only useful if everything it says is worth
	// reading. A first run that finds an ALARM does speak — that is news.
	switch {
	case prev == "":
		res.Changed = live.State != LiveAlive
	case prev != live.State:
		res.Changed = true
		res.Recovered = live.State == LiveAlive
	}

	if res.Changed {
		res.Message = watchdogMessage(res)
	}

	// Record after deciding, so a crash between the two re-reports rather than
	// silently swallowing the transition. Re-reporting is noisy; missing an
	// alarm is not recoverable.
	if err := os.WriteFile(filepath.Join(stateDir, fileWatchdogState),
		[]byte(live.State+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("record watchdog state: %w", err)
	}
	return res, nil
}

// watchdogMessage renders a transition. Alarm text names the state word AND the
// detail, because "unhealthy" alone sends the reader down the wrong of two very
// different investigations.
func watchdogMessage(r *WatchdogResult) string {
	var b strings.Builder
	if r.Recovered {
		b.WriteString("✅ RECOVERED — ")
		b.WriteString(r.Liveness.Detail)
		return b.String()
	}
	switch r.State {
	case LiveFailing:
		// ⚠️ Named as its own alarm, not folded into STALE. The reader's next
		// action differs: this one says do NOT restart the process, which is
		// the reflex a "stale" alarm produces and the wrong move here (#367).
		b.WriteString("🛑 WATCHER POLLING BUT FAILING — ")
	case LiveStale:
		b.WriteString("🛑 WATCHER STALE — ")
	case LiveNever:
		b.WriteString("🛑 WATCHER NEVER POLLED — ")
	case LiveUnknown:
		if r.Liveness.NoStateFound {
			b.WriteString("🛑 NO WATCHER STATE FOUND — ")
		} else {
			b.WriteString("🛑 WATCHER STATE UNREADABLE — ")
		}
	default:
		b.WriteString("🛑 WATCHER UNHEALTHY — ")
	}
	b.WriteString(r.Liveness.Detail)
	return b.String()
}
