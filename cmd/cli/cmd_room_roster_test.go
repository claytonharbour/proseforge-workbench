package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// stamp writes the watcher state files a leg leaves behind.
func stamp(t *testing.T, dir string, pollAgo, intervalGap time.Duration) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for name, ago := range map[string]time.Duration{
		"health-poll": pollAgo, "health-arm": pollAgo,
		"health-arm-prev": pollAgo + intervalGap,
	} {
		if err := os.WriteFile(filepath.Join(dir, name),
			[]byte(now.Add(-ago).Format("2006-01-02T15:04:05Z")), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// 🛑 THE FIRST VERSION OF THIS COMMAND EXITED 3 ON A DIRECTORY THAT WAS NOT A
// LEG (#380). It gated discovery on Liveness.NoStateFound, which distinguishes
// "this dir holds nothing of ours" from "this dir cannot be read" — and a path
// that does not EXIST is the second case, so the field is false for it. Every
// non-existent candidate path was therefore admitted as a leg.
//
// ⚑ Measured against a real tree on the first run: it listed a watchdog state
// directory as a stopped watcher. A discovery tool whose debut output is a
// false alarm trains its reader to discount the next real one.
func TestRosterDoesNotInventLegsFromDirectoriesThatHaveNone(t *testing.T) {
	root := t.TempDir()
	// A watchdog directory: real, ours, and NOT a watcher leg. It has
	// subdirectories but no health stamps anywhere.
	if err := os.MkdirAll(filepath.Join(root, "watchdog-state", "leg-demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	// An ordinary non-leg directory.
	if err := os.MkdirAll(filepath.Join(root, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}

	legs, err := scanRoster(root, 10)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(legs) != 0 {
		t.Fatalf("invented %d leg(s) from directories with no watcher state: %+v", len(legs), legs)
	}

	// ⚠️ POSITIVE CONTROL, in the same tree. Without it a scanner that always
	// returns nothing passes the assertion above — which is the null instrument
	// this whole ticket family is about.
	stamp(t, filepath.Join(root, "realleg", "state"), 30*time.Second, 60*time.Second)
	legs, err = scanRoster(root, 10)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(legs) != 1 || legs[0].Name != "realleg" {
		t.Fatalf("the scan finds nothing even when a leg IS present: %+v", legs)
	}
}

// A forgotten leg is the whole point: its threshold cannot come from a
// declaration, because the declaration is the thing that went missing.
func TestRosterDerivesTheThresholdFromTheLegsOwnObservedInterval(t *testing.T) {
	root := t.TempDir()
	// Polling every 60s, silent for 13h — the shape of the six legs that were
	// found 13 hours late.
	stamp(t, filepath.Join(root, "stopped", "state"), 13*time.Hour, 60*time.Second)
	// Polling every 30m, silent for 12m — well inside its own cadence.
	stamp(t, filepath.Join(root, "slowbutfine", "state"), 12*time.Minute, 30*time.Minute)

	legs, err := scanRoster(root, 10)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	byName := map[string]rosterLeg{}
	for _, l := range legs {
		byName[l.Name] = l
	}
	if got := byName["stopped"].State; got != "stale" {
		t.Errorf("a 60s leg silent for 13h reported %q, not stale", got)
	}
	// 🛑 The 30m leg is the one that matters. A single fleet-wide threshold —
	// any threshold under 12 minutes — reports it dead while it is healthy,
	// and that false alarm is why the number has to come from the leg.
	if got := byName["slowbutfine"].State; got != "alive" {
		t.Errorf("a 30m leg silent for 12m reported %q — a per-leg threshold must not "+
			"fire inside the leg's own cadence", got)
	}
	if r := byName["stopped"].Ratio; r < 700 {
		t.Errorf("ratio %.0f does not convey 13h against a 60s cadence", r)
	}
}

// ⛔ A missing interval is a missing INPUT, not a passing check (#401, same
// shape). The caveat belongs beside the diagnosis, never on top of it.
func TestRosterKeepsTheDiagnosisWhenTheRatioCannotBeComputed(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "onearm", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Armed once, long ago: a real stamp, but no second arm to measure against.
	old := time.Now().UTC().Add(-67 * time.Hour).Format("2006-01-02T15:04:05Z")
	for _, n := range []string{"health-poll", "health-arm"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(old), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	legs, err := scanRoster(root, 10)
	if err != nil || len(legs) != 1 {
		t.Fatalf("expected one leg, got %d (%v)", len(legs), err)
	}
	d := legs[0].Detail
	if !strings.Contains(d, "NOT RECEIVING") {
		t.Errorf("the actual diagnosis was discarded; detail is only the caveat: %q", d)
	}
	if !strings.Contains(d, "ratio") {
		t.Errorf("the ratio caveat is missing, so an empty column looks like a measurement: %q", d)
	}
}

// 🛑 THE FATAL POLARITY (@Aldric). Discovery matched the literal directory name
// `state`, so a node with two LIVE legs returned an empty table and rc=0 — the
// legs were on `state-v2`, which is exactly what this repo's own cutover recipe
// produces: arm the new dir so the old one keeps running with no deaf window.
//
// ⛔ A false positive is noisy and self-correcting. A false NEGATIVE from a tool
// whose purpose is "finds the legs you would not think to name" returns a clean
// bill for the case it exists to catch, and rc=0 on an empty table is
// indistinguishable from rc=0 on a healthy node.
func TestRosterFindsLegsWhoseStateDirIsNotCalledState(t *testing.T) {
	root := t.TempDir()
	// The cutover shape: new leg live on state-v2, old one retired beside it.
	stamp(t, filepath.Join(root, "demo-fleet-room", "state-v2"), 30*time.Second, 60*time.Second)
	stamp(t, filepath.Join(root, "prod-rc-room", "state-v2"), 20*time.Minute, 30*time.Minute)
	// And the conventional shape, which must keep working.
	stamp(t, filepath.Join(root, "plain-room", "state"), 10*time.Second, 60*time.Second)

	legs, err := scanRoster(root, 10)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(legs) != 3 {
		t.Fatalf("found %d of 3 legs — a name-keyed scan silently drops the cutover shape: %+v",
			len(legs), legs)
	}
	names := map[string]bool{}
	for _, l := range legs {
		names[l.Name] = true
	}
	// ⚑ Two state dirs can live under one leg directory, so the label has to
	// carry the path — "demo-fleet-room" alone cannot distinguish v2 from the
	// retired one beside it.
	for _, want := range []string{"demo-fleet-room/state-v2", "prod-rc-room/state-v2", "plain-room"} {
		if !names[want] {
			t.Errorf("missing %q; got %v", want, names)
		}
	}
}

// 🛑 A STATE DIR PROVED A LEG RUNS, NEVER WHAT IT RUNS AGAINST (@Sten). They
// found a live PROD watcher on their node, 25.5 hours old, and no artifact
// could say which room or which environment — so "should this exist" was
// unanswerable from disk, and CLAUDE.md already names that as the gap nothing
// checks. An inventory is the precondition for asking the question at all.
func TestRosterReportsWhatALegWatchesWhenTheLegRecordedIt(t *testing.T) {
	root := t.TempDir()
	dir := stamp(t, filepath.Join(root, "prod-leg", "state"), 20*time.Minute, 30*time.Minute)
	if err := os.WriteFile(filepath.Join(dir, "watching"),
		[]byte("conversation 2eff7b50-de20 https://app.proseforge.ai\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A leg armed before the stamp existed: it must still be FOUND, and must
	// not be described.
	stamp(t, filepath.Join(root, "legacy", "state"), 30*time.Second, 60*time.Second)

	legs, err := scanRoster(root, 10)
	if err != nil || len(legs) != 2 {
		t.Fatalf("expected 2 legs, got %d (%v)", len(legs), err)
	}
	by := map[string]rosterLeg{}
	for _, l := range legs {
		by[l.Name] = l
	}
	if got := by["prod-leg"].BaseURL; got != "https://app.proseforge.ai" {
		t.Errorf("prod leg's environment not reported: %q", got)
	}
	if got := by["prod-leg"].EntityID; got != "2eff7b50-de20" {
		t.Errorf("prod leg's room not reported: %q", got)
	}
	// ⛔ An unstamped leg must report EMPTY, never a guess. "Watches nothing" and
	// "armed before we recorded this" are different facts and only one is true.
	if by["legacy"].EntityID != "" || by["legacy"].BaseURL != "" {
		t.Errorf("invented a subject for a leg that never recorded one: %+v", by["legacy"])
	}
	if by["legacy"].State != "alive" {
		t.Errorf("an unstamped leg must still be found and judged: %q", by["legacy"].State)
	}
}

// 🛑 A NARROWED LEG PASSES EVERY OTHER CHECK (#406). After the 2026-08-23 reboot
// every bench re-armed from a saved command — the pre-crash argv dies with the
// process — and one leg came back narrower than it went down. It stayed `alive`,
// `rc=0`, advancing its cursor and delivering messages; it simply stopped
// delivering some it used to.
//
// ⛔ `fetched=N queued=0` (#403) cannot see it either: a narrowed filter yields
// an entirely ordinary `fetched=8 queued=3`.
func TestRosterReportsTheScopeALegWasArmedWith(t *testing.T) {
	root := t.TempDir()
	dir := stamp(t, filepath.Join(root, "scoped", "state"), 30*time.Second, 60*time.Second)
	if err := os.WriteFile(filepath.Join(dir, "watching"),
		[]byte("conversation 2882d62c https://rooms.example.test\n"+
			"match=@Tate|@Team from=Clayton exclude-from=- limit=60\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Armed before the scope line existed: found, described by room, and NOT
	// described by scope.
	older := stamp(t, filepath.Join(root, "legacy", "state"), 30*time.Second, 60*time.Second)
	if err := os.WriteFile(filepath.Join(older, "watching"),
		[]byte("conversation deadbeef https://rooms.example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	legs, err := scanRoster(root, 10)
	if err != nil || len(legs) != 2 {
		t.Fatalf("expected 2 legs, got %d (%v)", len(legs), err)
	}
	by := map[string]rosterLeg{}
	for _, l := range legs {
		by[l.Name] = l
	}
	if got := by["scoped"].Scope; !strings.Contains(got, "match=@Tate|@Team") ||
		!strings.Contains(got, "from=Clayton") || !strings.Contains(got, "limit=60") {
		t.Errorf("scope not reported, so a re-arm cannot be diffed against it: %q", got)
	}
	// ⛔ The one-line leg must report NO scope, never an invented empty one:
	// "no filters" and "armed before we recorded them" are different claims and
	// only the second is true.
	if by["legacy"].Scope != "" {
		t.Errorf("invented a scope for a leg that never recorded one: %q", by["legacy"].Scope)
	}
	// ⚑ @Aldric: an empty scope must READ as pending, not as silence. A room line
	// with no scope line is distinguishable — a current binary always writes the
	// second line, even when every filter is empty — so the tool can say which
	// case it is instead of leaving the reader to guess.
	var out strings.Builder
	renderScope(&out, by["legacy"])
	if !strings.Contains(out.String(), "pending") {
		t.Errorf("a pre-scope leg renders as silence, indistinguishable from "+
			"a leg with no filters: %q", out.String())
	}
	out.Reset()
	renderScope(&out, by["scoped"])
	if strings.Contains(out.String(), "pending") {
		t.Errorf("a leg WITH a recorded scope reported pending: %q", out.String())
	}
	if by["legacy"].EntityID != "deadbeef" {
		t.Errorf("a pre-scope leg must still report its room: %q", by["legacy"].EntityID)
	}
}
