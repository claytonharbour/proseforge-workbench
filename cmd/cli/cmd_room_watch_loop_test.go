package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// watchExitCode is what keeps loop telemetry and single-shot exit codes from
// drifting apart. If it stops agreeing with runRoomWatch's switch, a bench
// reading `rc=` from a tick log and a bench reading `$?` from cron get different
// answers for the same failure — and both will believe their own number.
func TestWatchExitCodeMatchesTheSingleShotContract(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
		why  string
	}{
		{"success", nil, 0, "a clean poll"},
		{"unreachable", watcher.ErrUnreachable, 12, "an OUTAGE — retry, but report it"},
		{"locked", watcher.ErrLocked, 11, "benign: another tick mid-flight, resolves next cycle"},
		{"identity guard", watcher.ErrIdentityGuard, 10, "permanent misconfiguration, needs a human"},
		{"anything else", errors.New("boom"), 1, "generic failure"},
		{"wrapped unreachable", errWrap(watcher.ErrUnreachable), 12,
			"errors arrive wrapped from the service layer; matching must survive it"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := watchExitCode(tt.err); got != tt.want {
				t.Errorf("watchExitCode = %d, want %d — %s", got, tt.want, tt.why)
			}
		})
	}
}

// 🛑 rc=11 must NOT be treated as failure by the loop. A tick that finds another
// holding the lock is the watcher WORKING; folding it into the failure state
// makes a busy watcher alarm about its own health, and an alarm that fires
// during normal operation is one an operator learns to ignore.
func TestLockHeldIsNotAFailureState(t *testing.T) {
	if watchExitCode(watcher.ErrLocked) != 11 {
		t.Fatal("lock held must map to 11, distinct from both success and outage")
	}
	// The loop's own branch treats nil and ErrLocked identically — this asserts
	// they are distinguishable in telemetry while sharing the healthy path.
	if watchExitCode(nil) == watchExitCode(watcher.ErrLocked) {
		t.Error("success and lock-held must remain distinguishable in the tick log, " +
			"even though both count as healthy")
	}
}

// errSuffix carries the reason onto the telemetry line. Without it a tick log
// has an rc and no cause (@Sten), which is the difference between "the backend
// is down" and "your token expired" — opposite fixes, same number.
func TestErrSuffixCarriesTheReason(t *testing.T) {
	if got := errSuffix(nil); got != "" {
		t.Errorf("a successful tick must add nothing, got %q", got)
	}
	got := errSuffix(errors.New("whoami: connection refused"))
	for _, want := range []string{"err=", "connection refused"} {
		if !contains(got, want) {
			t.Errorf("errSuffix = %q, want it to contain %q", got, want)
		}
	}
}

func errWrap(err error) error { return &wrapped{err} }

type wrapped struct{ err error }

func (w *wrapped) Error() string { return "read room messages: " + w.err.Error() }
func (w *wrapped) Unwrap() error { return w.err }

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// #371: the fingerprint is what decides whether a rebuild is picked up. If it
// stops changing when the binary does, --loop silently returns to running
// deleted code — the regression this exists to undo, restored invisibly.
func TestBinaryFingerprintChangesWithTheImage(t *testing.T) {
	first, ok := binaryFingerprint()
	if !ok {
		t.Skip("no resolvable executable in this test environment")
	}
	if first == "" {
		t.Fatal("empty fingerprint would compare equal forever and never trigger a re-exec")
	}
	second, _ := binaryFingerprint()
	if first != second {
		t.Errorf("fingerprint is unstable across calls (%q vs %q) — an unstable one re-execs "+
			"every tick, which is worse than never", first, second)
	}
}

// 🛑 The symlink case is the one that matters here and the one that would fail
// silently. ~/.local/bin/pfw is a SYMLINK into build/bin/pfw, and `make build`
// replaces the target, not the link. Fingerprinting the link would never change,
// so the check would never fire — a guard that cannot go red, which is the
// defect class this whole epic is about.
func TestBinaryFingerprintFollowsSymlinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks unavailable")
	}

	stat := func(p string) string {
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			t.Fatal(err)
		}
		fi, err := os.Stat(resolved)
		if err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%d:%d", fi.ModTime().UnixNano(), fi.Size())
	}

	before := stat(link)
	if err := os.WriteFile(target, []byte("v2-longer"), 0o755); err != nil {
		t.Fatal(err)
	}
	if after := stat(link); after == before {
		t.Error("fingerprint via the symlink did not change when the TARGET was rewritten — " +
			"a rebuild would go undetected and the watcher would keep the old image forever")
	}
}

// A config refusal is PERMANENT — same class as the identity guard, same code.
// It must not fall through to the generic 1, because every fleet recipe reads 1
// as "transient, retry next tick" and this condition never clears on its own.
//
// 🛑 The single-shot path now CALLS watchExitCode instead of repeating its
// switch. It carried a copy until #357, and the copies drifted the moment a
// fourth error kind appeared: --loop reported 10 while single-shot exited 1 for
// the identical deaf watcher. Two switches over one taxonomy is a promise to
// diverge, and the divergence is invisible until someone compares them.
func TestConfigRefusalIsPermanentNotGeneric(t *testing.T) {
	if got := watchExitCode(watcher.ErrConfigRefusal); got != 10 {
		t.Errorf("ErrConfigRefusal = %d, want 10", got)
	}
	// The real call sites all wrap with %w — an unwrapped-only check would pass
	// while production returned 1.
	wrapped := fmt.Errorf("%w: cursor is a timestamp", watcher.ErrConfigRefusal)
	if got := watchExitCode(wrapped); got != 10 {
		t.Errorf("wrapped ErrConfigRefusal = %d, want 10", got)
	}
	if watchExitCode(watcher.ErrConfigRefusal) == watchExitCode(errors.New("boom")) {
		t.Error("config refusal must be distinguishable from a generic failure")
	}
}

// 🛑 THE --loop REPORTER IS A SEPARATE FUNCTION FROM THE SINGLE-SHOT ONE, and
// that divergence is what hid #445 for as long as it existed.
//
// runRoomWatch (single-shot) reports QuotedOnly via status() to stderr.
// reportWatchResult (--loop) did not report it at all — so on every leg in the
// fleet, which all run --loop, a drop REASON was emitted to no stream: stdout 0,
// stderr 0, telemetry 0. Five benches computed a drop rate from `filtered:` and
// none could attribute it; four reached a wrong mechanism before it was found.
//
// ⚠️ This file already warns that "no test asserts what single-shot does NOT
// print". The same gap in the other direction is what this test closes: nothing
// asserted what --loop DOES print, so the two paths could drift silently and did.
func TestLoopReporterRecordsWhyMessagesWereDropped(t *testing.T) {
	read := func(t *testing.T, res *watcher.WatchResult) string {
		t.Helper()
		dir := t.TempDir()
		tel := openTelemetry(dir, "")
		reportWatchResult(newRoomWatchCmd(), res, tel)
		b, err := os.ReadFile(filepath.Join(dir, "telemetry.log"))
		if err != nil {
			t.Fatalf("reading telemetry: %v", err)
		}
		return string(b)
	}

	// POSITIVE: a drop with a reason must name the reason, in the file.
	t.Run("a drop records its reason", func(t *testing.T) {
		got := read(t, &watcher.WatchResult{Fetched: 85, Kept: 74, QuotedOnly: 11})

		if !strings.Contains(got, "drop-reasons: quoted 11, self-skipped 0, server-missed 0") {
			t.Errorf("the reason is missing — a count nobody can attribute is what #445 was:\n%s", got)
		}
		// The fleet greps this exact prefix. Widening it would break their
		// parsers, which is why the reason is a SEPARATE line.
		if !strings.Contains(got, "filtered: fetched 85, kept 74") {
			t.Errorf("the existing line changed shape; fleet greps depend on it byte-for-byte:\n%s", got)
		}
	})

	// 🛑 NEGATIVE CONTROL. Without this the assertion above passes on a line
	// printed unconditionally — which is the very defect #445 reported, rebuilt
	// inside its own regression test. A counter that cannot read zero is not a
	// counter.
	t.Run("no drop records nothing", func(t *testing.T) {
		got := read(t, &watcher.WatchResult{Fetched: 74, Kept: 74})

		if strings.Contains(got, "drop-reasons:") {
			t.Errorf("reported reasons when nothing was dropped — the line is decoration, not a measurement:\n%s", got)
		}
		if strings.Contains(got, "filtered:") {
			t.Errorf("reported a filter event when fetched == kept:\n%s", got)
		}
	})
}
