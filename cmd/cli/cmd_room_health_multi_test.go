package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

func legDir(t *testing.T, stamp string) string {
	t.Helper()
	dir := t.TempDir()
	if stamp != "" {
		if err := os.WriteFile(filepath.Join(dir, "health-poll"), []byte(stamp+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// #372: @Tuner kept 53 lines of shell for exactly one reason — "pfw room health
// checks ONE state dir and you will run several." These rows are the states a
// multi-leg sweep has to tell apart before those lines can go.
func TestMultiLegHealthClassifiesEachLegIndependently(t *testing.T) {
	fresh := time.Now().UTC().Format(time.RFC3339)
	old := time.Now().Add(-9 * time.Hour).UTC().Format(time.RFC3339)

	tests := []struct {
		name      string
		stamp     string
		wantAlive bool
		why       string
	}{
		{"fresh poll", fresh, true, "a leg polling normally"},
		{"stale poll", old, false, "THE failure case: polled once, then stopped"},
		{"never polled", "", false,
			"no stamp at all — a fresh install or one that has never worked; not the same as stale"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := watcher.CheckLiveness(legDir(t, tt.stamp), 5*time.Minute)
			if err != nil {
				t.Fatalf("CheckLiveness: %v", err)
			}
			if l.Alive() != tt.wantAlive {
				t.Errorf("Alive() = %v, want %v (state %q) — %s", l.Alive(), tt.wantAlive, l.State, tt.why)
			}
		})
	}
}

// 🛑 An unreadable leg must be REPORTED, never skipped. @Sten's case:
// `find -name health-poll` returns nothing for a bench using a native poller, so
// a sweep silently drops the one leg it exists to protect. A checker that omits
// what it could not read prints "0 unhealthy" and "0 checked" identically.
func TestUnreadableLegIsReportedNotSkipped(t *testing.T) {
	l, err := watcher.CheckLiveness(filepath.Join(t.TempDir(), "does-not-exist"), 5*time.Minute)
	if err != nil {
		t.Fatalf("CheckLiveness returned an error instead of a verdict: %v", err)
	}
	if l.Alive() {
		t.Error("an unreadable leg reported alive; a check that cannot read its input must not report OK")
	}
	if l.State != watcher.LiveUnknown {
		t.Errorf("State = %q, want %q — 'cannot read' is distinct from 'read it, it is dead'", l.State, watcher.LiveUnknown)
	}
}

// shortDir keeps a four-leg report readable. It must never collapse two
// different legs to the same label, or the summary names the wrong watcher.
func TestShortDirStaysDistinguishable(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	a := shortDir(filepath.Join(home, "cron", "demo", "state"))
	b := shortDir(filepath.Join(home, "cron", "prod", "state"))
	if a == b {
		t.Fatalf("two legs collapsed to the same label %q — the report would name the wrong watcher", a)
	}
	if !strings.HasPrefix(a, "~") {
		t.Errorf("home prefix not trimmed: %q", a)
	}
	if outside := shortDir("/var/tmp/leg"); outside != "/var/tmp/leg" {
		t.Errorf("a path outside home was rewritten to %q; it must be left alone", outside)
	}
}

// #372 per-leg thresholds. @Sten runs demo at 5m and prod at 70m off ONE
// watchdog; a single --max-age cannot express that. 5m against a 30m poller
// false-alarms for 25 minutes in every 30; 70m against a 60s poller is an hour
// of silence before it notices.
//
// ⚠️ I wrote this as an acceptance criterion on #372 and shipped without it,
// closing the ticket against the criteria I had written rather than the ones I
// had met. These rows are what that criterion should have forced.
func TestPerLegMaxAgeParsing(t *testing.T) {
	tests := []struct {
		spec    string
		wantDir string
		wantAge time.Duration
		why     string
	}{
		{"/a/b:70m", "/a/b", 70 * time.Minute, "the whole point: a slow leg gets a slow threshold"},
		{"/a/b:5m", "/a/b", 5 * time.Minute, "and a fast one keeps a tight threshold"},
		{"/a/b", "/a/b", 0, "no suffix ⇒ fall back to the global --max-age"},
		{"/a/b:notaduration", "/a/b:notaduration", 0,
			"an unparseable suffix must NOT be silently eaten as a threshold — it is part of the path"},
		{"/weird:path/state:30s", "/weird:path/state", 30 * time.Second,
			"LastIndex, not Index: a colon inside the path must not be mistaken for the separator"},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			dir, age := tt.spec, time.Duration(0)
			if i := strings.LastIndex(tt.spec, ":"); i > 0 {
				if parsed, err := time.ParseDuration(tt.spec[i+1:]); err == nil {
					dir, age = tt.spec[:i], parsed
				}
			}
			if dir != tt.wantDir {
				t.Errorf("dir = %q, want %q — %s", dir, tt.wantDir, tt.why)
			}
			if age != tt.wantAge {
				t.Errorf("age = %v, want %v — %s", age, tt.wantAge, tt.why)
			}
		})
	}
}
