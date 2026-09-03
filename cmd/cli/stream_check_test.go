package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// capture swaps stdout/stderr for the duration of fn, so the check inspects the
// descriptors we set up rather than the test runner's.
func withStreams(t *testing.T, out, errF *os.File, fn func()) []string {
	t.Helper()
	oo, oe := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = out, errF
	defer func() { os.Stdout, os.Stderr = oo, oe }()
	var got []string
	fn2 := func(format string, a ...any) { got = append(got, format) }
	_ = fn2
	checkOutputStreams(func(format string, a ...any) { got = append(got, format) })
	return got
}

// 🛑 THE RULE EXISTED AND SIX BENCHES BROKE IT IN ONE NIGHT (#409). CLAUDE.md
// forbids `2>&1` on a room watcher because a room quotes text that looks like
// the watcher's own output — rc=, tick=, telemetry — so merging the streams
// removes the one channel whose provenance is still trustworthy.
//
// ⚑ A document cannot enforce a shell redirect. The process can see one.
func TestArmWarnsWhenStderrIsMergedIntoStdout(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "watch.log")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// `>>file 2>&1` — both descriptors on the SAME file.
	got := withStreams(t, f, f, nil)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "2>&1") {
		t.Errorf("merged streams did not warn: %q", joined)
	}
	// ⛔ The WORDING is the part that was wrong twice, so it gets its own arm.
	// v1 said "Drop the 2>&1" — @Crispin's Monitor leg would have gone silent on
	// startup failures. v2 called it a two-sided trade — @Tuner and @Crispin then
	// kept a redirect for alarm visibility they already had on stdout. Both
	// versions shipped, and a test asserting only that "2>&1" appears passed for
	// each. The claim has to be pinned, not just the topic.
	if strings.Contains(joined, "Drop the 2>&1") {
		t.Errorf("prescriptive v1 wording is back; it silences startup failures "+
			"on a Monitor-backed leg: %q", joined)
	}
	// ⛔ v3's defect: it asserted the operator wrote 2>&1. Measured false — this
	// harness merges fd1/fd2 for a backgrounded child with no redirect at all.
	// The warning must not accuse, and must give an escape that exists.
	// ⛔ v4: naming ONE cause is what churned live legs. Both must appear, and the
	// remedy must be the one that works either way.
	// ⛔ v5: the remedy must be CONDITIONAL. v4 told every reader to create a new
	// unrotated file, including those whose fix is a deletion — recreating the
	// hazard of three tickets filed the same night.
	for _, required := range []string{"buys nothing for alarms", "STARTUP",
		"TWO CAUSES", "own path", "just drop it", "does NOT rotate"} {
		if !strings.Contains(joined, required) {
			t.Errorf("warning must name WHICH path the merge covers, missing %q — "+
				"without it the reader cannot tell if their reason is the real one: %q",
				required, joined)
		}
	}

	// ⚠️ CONTROL, in the same test: separate files must NOT produce that
	// warning, or the check fires on everyone and gets ignored.
	g, err := os.Create(filepath.Join(dir, "err.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	got = withStreams(t, f, g, nil)
	if strings.Contains(strings.Join(got, "\n"), "2>&1") {
		t.Errorf("separate stdout/stderr falsely reported as merged: %q", got)
	}
}

// ⚠️ A regular-file stdout has no bound — pfw rotates its own telemetry and
// nothing else. Worth one line at arm rather than an hour of lsof later.
func TestArmWarnsWhenStdoutIsAnUnrotatedFile(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.Create(filepath.Join(dir, "watch.log"))
	defer f.Close()
	g, _ := os.Create(filepath.Join(dir, "err.log"))
	defer g.Close()

	got := strings.Join(withStreams(t, f, g, nil), "\n")
	if !strings.Contains(got, "does not rotate") {
		t.Errorf("unbounded file stdout did not warn: %q", got)
	}
}

// ⛔ stdout to /dev/null means the leg buffers and reports healthy while waking
// nobody — the deaf-but-green shape this whole ticket family exists to remove.
func TestArmWarnsWhenStdoutIsDevNull(t *testing.T) {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Skip("no /dev/null")
	}
	defer devnull.Close()
	g, _ := os.Create(filepath.Join(t.TempDir(), "err.log"))
	defer g.Close()

	got := strings.Join(withStreams(t, devnull, g, nil), "\n")
	if !strings.Contains(got, "/dev/null") {
		t.Errorf("discarded stdout did not warn: %q", got)
	}
}
