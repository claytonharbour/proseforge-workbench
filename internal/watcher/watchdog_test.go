package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 🛑 Every alarm path is proven to FIRE before any test trusts its silence.
// @Wayland's method, and the reason it matters here: an alarm nobody has seen
// speak is indistinguishable from an alarm that cannot speak, and this whole
// area keeps producing exactly that failure.

// watchedDir builds a watcher state dir whose last successful poll was `age` ago.
func watchedDir(t *testing.T, age time.Duration) string {
	t.Helper()
	d := t.TempDir()
	stamp := time.Now().Add(-age).UTC().Format(time.RFC3339)
	if err := os.WriteFile(filepath.Join(d, fileHealthOK), []byte(stamp), 0o644); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestWatchdogAlarmPathsAllFire(t *testing.T) {
	cases := []struct {
		name      string
		watched   func(*testing.T) string
		wantState string
		wantText  string
	}{
		{
			name:      "stale fires with the age and the limit",
			watched:   func(t *testing.T) string { return watchedDir(t, time.Hour) },
			wantState: LiveStale,
			wantText:  "STALE",
		},
		{
			// ⚠️ Distinct wording from stale on purpose: this one means "check
			// your config", the other means "check whether the process runs".
			// Evidence a watcher was here = an arm stamp with no success (#367).
			name: "never fires with distinct wording",
			watched: func(t *testing.T) string {
				d := t.TempDir()
				touchArm(d)
				return d
			},
			wantState: LiveNever,
			wantText:  "NEVER POLLED",
		},
		{
			// 🛑 #367: an empty dir is NOT a death claim — it may be a bench
			// that polls natively and leaves no stamp of ours. Still an alarm,
			// still exit-3 red, but it names a path to check rather than
			// accusing a healthy watcher of having died.
			name:      "no state at all fires without claiming a death",
			watched:   func(t *testing.T) string { return t.TempDir() },
			wantState: LiveUnknown,
			wantText:  "NO WATCHER STATE FOUND",
		},
		{
			// The pair's own alarm: running, and refused.
			name: "polling-but-failing gets its own wording",
			watched: func(t *testing.T) string {
				d := t.TempDir()
				if err := os.WriteFile(filepath.Join(d, fileHealthOK),
					[]byte(time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)), 0o644); err != nil {
					t.Fatal(err)
				}
				touchArm(d)
				return d
			},
			wantState: LiveFailing,
			wantText:  "POLLING BUT FAILING",
		},
		{
			name: "unknown fires when the state dir is unreadable",
			watched: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			wantState: LiveUnknown,
			wantText:  "UNREADABLE",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Watchdog(tc.watched(t), t.TempDir(), 5*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if res.State != tc.wantState {
				t.Errorf("state = %q, want %q", res.State, tc.wantState)
			}
			if res.Message == "" {
				t.Fatalf("%s produced NO message — a silent alarm is not an alarm", tc.wantState)
			}
			if !strings.Contains(res.Message, tc.wantText) {
				t.Errorf("message %q does not contain %q", res.Message, tc.wantText)
			}
			if !res.Alarming() {
				t.Error("Alarming() = false for a non-alive state")
			}
		})
	}
}

// A healthy first run must be SILENT. Announcing "alive" at startup teaches the
// reader the channel carries routine noise, and then the real alarm is skimmed.
func TestWatchdogHealthyFirstRunIsSilent(t *testing.T) {
	res, err := Watchdog(watchedDir(t, time.Second), t.TempDir(), 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if res.Message != "" {
		t.Errorf("healthy first run spoke: %q", res.Message)
	}
	if res.Alarming() {
		t.Error("Alarming() = true for a fresh healthy watcher")
	}
}

// 🛑 A repeated alarm must NOT re-fire. A watchdog that speaks every tick while
// stale floods the channel it is alerting on, and gets muted.
func TestWatchdogDoesNotFloodOnRepeatedAlarm(t *testing.T) {
	watched, state := watchedDir(t, time.Hour), t.TempDir()

	first, err := Watchdog(watched, state, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if first.Message == "" {
		t.Fatal("first alarm was silent")
	}
	second, err := Watchdog(watched, state, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if second.Message != "" {
		t.Errorf("repeated alarm spoke again: %q", second.Message)
	}
	if !second.Alarming() {
		t.Error("still stale, but Alarming() = false — the exit code must reflect CURRENT state, not the transition")
	}
}

// ⚑ The one most people skip. An alarm that goes red and never green is a stuck
// alarm, and it looks identical to a permanent outage (@Wayland).
func TestWatchdogFiresOnRecovery(t *testing.T) {
	watched, state := watchedDir(t, time.Hour), t.TempDir()

	if res, _ := Watchdog(watched, state, 5*time.Minute); res.Message == "" {
		t.Fatal("setup: expected the stale alarm to fire")
	}
	// Restore the stamp, exactly as backdating-then-restoring does live.
	stamp := time.Now().UTC().Format(time.RFC3339)
	if err := os.WriteFile(filepath.Join(watched, fileHealthOK), []byte(stamp), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Watchdog(watched, state, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Recovered {
		t.Error("Recovered = false after stale -> alive")
	}
	if res.Message == "" {
		t.Fatal("RECOVERY WAS SILENT — the alarm is stuck red and indistinguishable from a real outage")
	}
	if !strings.Contains(res.Message, "RECOVERED") {
		t.Errorf("recovery message %q does not say RECOVERED", res.Message)
	}
	if res.Alarming() {
		t.Error("Alarming() = true after recovery")
	}
}

// A truncated stamp must read as `never`, never as `alive`. @Wayland ran this
// against room health as a control before trusting a watchdog built on it; it is
// a regression test now that one depends on it.
func TestWatchdogTruncatedStampIsNeverNotAlive(t *testing.T) {
	watched := t.TempDir()
	if err := os.WriteFile(filepath.Join(watched, fileHealthOK), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Watchdog(watched, t.TempDir(), 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if res.State == LiveAlive {
		t.Fatal("an empty stamp reported ALIVE — the watchdog would trust a watcher it cannot verify")
	}
	if res.State != LiveNever {
		t.Errorf("state = %q, want %q — an EMPTY stamp file is evidence a watcher wrote here, "+
			"so this is a corrupted watcher and not an empty dir", res.State, LiveNever)
	}
}

// The watchdog must not write into the dir it observes, or it corrupts its own
// evidence.
func TestWatchdogDoesNotWriteToTheWatchedDir(t *testing.T) {
	watched := watchedDir(t, time.Hour)
	before, err := os.ReadDir(watched)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Watchdog(watched, t.TempDir(), 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadDir(watched)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Errorf("watched dir gained files: %d -> %d", len(before), len(after))
	}
}

// A detected outage is the watchdog WORKING. The exit-code semantics live in the
// CLI, but the invariant belongs here: Alarming() describes the WATCHED watcher,
// and callers must not read it as "the watchdog failed" (@Sten).
func TestAlarmingDescribesTheWatchedWatcherNotTheWatchdog(t *testing.T) {
	res, err := Watchdog(watchedDir(t, time.Hour), t.TempDir(), 5*time.Minute)
	if err != nil {
		t.Fatalf("a dead WATCHER must not produce a watchdog ERROR: %v", err)
	}
	if !res.Alarming() {
		t.Error("Alarming() = false for a stale watcher")
	}
	if res.Message == "" {
		t.Error("a detected outage produced no message")
	}
}

// The watchdog must fail loudly when it cannot do its job, which is a different
// thing from the watcher being dead.
func TestWatchdogErrorsWhenItCannotDoItsJob(t *testing.T) {
	if _, err := Watchdog("", t.TempDir(), time.Minute); err == nil {
		t.Error("no watched dir: want an error, got nil")
	}
	if _, err := Watchdog(t.TempDir(), "", time.Minute); err == nil {
		t.Error("no watchdog state dir: want an error, got nil")
	}
}
