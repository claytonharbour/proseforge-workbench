package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func stateWith(t *testing.T, pollAge time.Duration, hasStamp bool, cursor string) string {
	t.Helper()
	dir := t.TempDir()
	if hasStamp {
		stamp := time.Now().Add(-pollAge).UTC().Format(time.RFC3339)
		if err := os.WriteFile(filepath.Join(dir, fileHealthOK), []byte(stamp+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if cursor != "" {
		if err := os.WriteFile(filepath.Join(dir, fileCursor), []byte(cursor), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// 🛑 THE case. A watcher that stopped looks exactly like a quiet room, so the
// failure is invisible while it is happening (@Tuner's cron vanished from
// CronList with no error).
func TestStoppedWatcherIsStaleNotQuiet(t *testing.T) {
	l, err := CheckLiveness(stateWith(t, time.Hour, true, "123-0"), 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if l.State != LiveStale {
		t.Fatalf("state = %q, want %q", l.State, LiveStale)
	}
	if l.Alive() {
		t.Error("a watcher an hour dead reported alive")
	}
}

func TestRecentPollIsAlive(t *testing.T) {
	l, _ := CheckLiveness(stateWith(t, 30*time.Second, true, "123-0"), 5*time.Minute)
	if !l.Alive() || l.State != LiveAlive {
		t.Fatalf("state = %q alive=%v, want alive/true", l.State, l.Alive())
	}
	if l.AgeSeconds < 25 || l.AgeSeconds > 40 {
		t.Errorf("age = %ds, want ~30", l.AgeSeconds)
	}
}

// 🛑 "never" and "stale" must not collapse. They need OPPOSITE fixes — check
// your config, versus check whether the process is running — so a single
// "unhealthy" sends you down the wrong one.
func TestNeverPolledIsDistinctFromStale(t *testing.T) {
	// ⚑ #367 changed what an EMPTY dir means. It used to report `never`, which
	// asserts a fact about a watcher — "it ran and never succeeded" — from the
	// absence of our own file. @Angel polls natively and leaves no stamp of
	// ours, so that claim was loudest about the bench it knew least about.
	// `never` is now reserved for dirs carrying EVIDENCE a watcher was here.
	armed := stateWith(t, 0, false, "")
	touchArm(armed)
	never, _ := CheckLiveness(armed, 5*time.Minute)
	stale, _ := CheckLiveness(stateWith(t, time.Hour, true, "123-0"), 5*time.Minute)

	if never.State != LiveNever {
		t.Fatalf("attempted-but-never-succeeded → %q, want %q", never.State, LiveNever)
	}
	if never.State == stale.State {
		t.Fatal("never and stale collapsed into one state")
	}
	if never.Alive() || stale.Alive() {
		t.Error("neither should be alive")
	}
}

// A cursor without a poll stamp is impossible for a working watcher, so it
// almost always means the --state dir is not the one the watcher uses. Saying
// so beats reporting a bare "never".
func TestCursorWithoutStampNamesTheLikelyCause(t *testing.T) {
	l, _ := CheckLiveness(stateWith(t, 0, false, "123-0"), 5*time.Minute)
	if l.State != LiveNever {
		t.Fatalf("state = %q, want %q", l.State, LiveNever)
	}
	if !contains(l.Detail, "state dir") {
		t.Errorf("detail should name the likely cause: %q", l.Detail)
	}
}

// ⚠️ A check that cannot read its own input must never report OK. That is the
// failure shape this whole area exists to delete.
func TestUnreadableStateDirIsNotHealthy(t *testing.T) {
	l, err := CheckLiveness(filepath.Join(t.TempDir(), "does-not-exist"), 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if l.State != LiveUnknown || l.Alive() {
		t.Fatalf("state = %q alive = %v, want unknown/false", l.State, l.Alive())
	}
}

// maxAge is policy, not a constant. A slow tick must not alarm and a real death
// must not hide, and only the caller knows their poll interval.
func TestMaxAgeIsTheCallersPolicy(t *testing.T) {
	dir := stateWith(t, 10*time.Minute, true, "123-0")
	if l, _ := CheckLiveness(dir, 5*time.Minute); l.State != LiveStale {
		t.Errorf("10m old with a 5m limit = %q, want stale", l.State)
	}
	if l, _ := CheckLiveness(dir, time.Hour); l.State != LiveAlive {
		t.Errorf("10m old with a 1h limit = %q, want alive", l.State)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

// @Smiley ran `pfw room health <parent>` during the fleet liveness audit and got
// `never — no successful poll has EVER been recorded`, on a watcher that had
// delivered eight events in the previous eighteen minutes and whose stamp had
// been written four seconds earlier. He had passed the parent directory.
//
// 🛑 A wrong path and a dead watcher are OPPOSITE emergencies. Returning the
// second for the first, during an audit whose entire purpose is telling live
// watchers from dead ones, is the worst possible moment for that confusion.
func TestCheckLivenessFindsAStampOneLevelDown(t *testing.T) {
	parent := t.TempDir()
	real := filepath.Join(parent, "state")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().Add(-4 * time.Second).UTC().Format(time.RFC3339)
	if err := os.WriteFile(filepath.Join(real, fileHealthOK), []byte(stamp), 0o644); err != nil {
		t.Fatal(err)
	}

	l, err := CheckLiveness(parent, 10*time.Minute)
	if err != nil {
		t.Fatalf("CheckLiveness: %v", err)
	}

	if l.State == LiveNever {
		t.Errorf("reported %q for a wrong path — this is the bug: it accuses a live watcher of never having polled", LiveNever)
	}
	if !strings.Contains(l.Detail, real) {
		t.Errorf("detail does not name the real stamp path %q, so the operator still has to guess:\n  %s", real, l.Detail)
	}
	if !strings.Contains(l.Detail, "NOT a dead watcher") {
		t.Errorf("detail must say plainly this is not a death; an operator mid-audit reads the state word and acts:\n  %s", l.Detail)
	}
}

// The negative arm, and it is the one that must not soften. A genuinely fresh
// dir with nothing near it must still be RED — trading a false death for a
// false life is strictly worse during an audit, because the audit then reports
// all-clear.
//
// ⚑ #367 changed the WORD from never to unknown (see above) and deliberately
// not the verdict: this asserts the exit-code pole, so the reclassification
// cannot quietly turn a fleet-wide check green.
func TestCheckLivenessStillNotAliveWhenNothingIsNearby(t *testing.T) {
	dir := t.TempDir()
	l, err := CheckLiveness(dir, 10*time.Minute)
	if err != nil {
		t.Fatalf("CheckLiveness: %v", err)
	}
	if l.Alive() {
		t.Fatal("an empty dir with no neighbour stamp reported ALIVE — the false-life failure")
	}
	if l.State != LiveUnknown || !l.NoStateFound {
		t.Errorf("State = %q (NoStateFound=%v), want %q with NoStateFound — no evidence of a watcher is not a claim that one died",
			l.State, l.NoStateFound, LiveUnknown)
	}
}

// 🛑 THE STATE THE PAIR EXISTS FOR (#367). Attempts are current, successes are
// not: the loop is running and something is refusing it. Reported as `failing`
// because `stale` sends the reader to restart a process that is working.
func TestPollingButFailingIsNotStale(t *testing.T) {
	dir := stateWith(t, time.Hour, true, "123-0") // last SUCCESS an hour ago
	touchArm(dir)                                 // but an attempt just now

	l, _ := CheckLiveness(dir, 5*time.Minute)
	if l.State != LiveFailing {
		t.Fatalf("State = %q, want %q — attempts fresh + successes stale is the backend failing, not the watcher dying", l.State, LiveFailing)
	}
	if l.Alive() {
		t.Error("failing must not be alive")
	}
	if !strings.Contains(l.Detail, "RUNNING") {
		t.Errorf("detail %q does not tell the reader the process is fine", l.Detail)
	}
}

// The other half: no attempt either. Nothing is polling, and that IS stale.
func TestNoAttemptsEitherIsStaleNotFailing(t *testing.T) {
	dir := stateWith(t, time.Hour, true, "123-0")
	// arm stamp deliberately old — write it directly rather than via touchArm
	if err := os.WriteFile(filepath.Join(dir, fileHealthArm),
		[]byte(time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)), 0o644); err != nil {
		t.Fatal(err)
	}
	l, _ := CheckLiveness(dir, 5*time.Minute)
	if l.State != LiveStale {
		t.Fatalf("State = %q, want %q — no attempts and no successes is a dead loop", l.State, LiveStale)
	}
	if !strings.Contains(l.Detail, "not running") {
		t.Errorf("detail %q does not say the loop stopped", l.Detail)
	}
}

// ⚑ ARM BEFORE THE FIRST TICK COMPLETES (@Crispin). His first version stamped
// inside the loop body, so nothing existed until after the first sleep — the
// exact ambiguity the stamp was added to remove. He caught it by noticing the
// FILE WAS ABSENT, so absence-then-presence is what this asserts.
func TestArmStampExistsBeforeAnyPollCanSucceed(t *testing.T) {
	dir := t.TempDir()
	if fileExists(dir, fileHealthArm) {
		t.Fatal("precondition: arm stamp already present")
	}
	touchArm(dir)
	if !fileExists(dir, fileHealthArm) {
		t.Fatal("no arm stamp after arming — a freshly started watcher is indistinguishable from one that never started")
	}
	if fileExists(dir, fileHealthOK) {
		t.Error("arming wrote a SUCCESS stamp — that would report a watcher healthy before it has polled once")
	}
}

// A --max-age below the observed interval is red for most of every interval
// even when the watcher is perfectly healthy. @Smiley's trap; @Tuner's leg went
// uncovered for 6.5h to it. It must be SAID, not silently tolerated.
func TestMaxAgeBelowObservedIntervalIsAnnounced(t *testing.T) {
	dir := stateWith(t, time.Second, true, "123-0")
	// two attempts 30 minutes apart ⇒ observed interval 30m
	if err := os.WriteFile(filepath.Join(dir, fileHealthArmPrev),
		[]byte(time.Now().Add(-30*time.Minute).UTC().Format(time.RFC3339)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, fileHealthArm),
		[]byte(time.Now().UTC().Format(time.RFC3339)), 0o644); err != nil {
		t.Fatal(err)
	}

	l, _ := CheckLiveness(dir, 5*time.Minute)
	if !l.Alive() {
		t.Fatalf("state = %q — this leg IS healthy; the config is what is wrong", l.State)
	}
	if l.ConfigWarning == "" {
		t.Fatal("a 5m threshold on a 30m leg produced NO warning — it will read stale for 25 minutes in every 30 and nobody was told")
	}
	if l.IntervalSeconds < 1700 || l.IntervalSeconds > 1900 {
		t.Errorf("IntervalSeconds = %d, want ~1800 — the interval must be MEASURED, not assumed", l.IntervalSeconds)
	}

	// ⚠️ And the sane case must stay quiet, or the warning is noise.
	if sane, _ := CheckLiveness(dir, 90*time.Minute); sane.ConfigWarning != "" {
		t.Errorf("90m against a 30m interval warned anyway: %q", sane.ConfigWarning)
	}
}

// 🛑 A RESTART IS NOT A CADENCE (#373, @Smiley).
//
// The watcher arms on process start, moments after the previous process's last
// tick armed — so a re-exec or a manual restart leaves two arm stamps in the
// same second. On my own 60s leg after a self-re-exec both stamps read
// 13:09:27 exactly, and the "observed interval" became 0.
//
// ⚠️ Reporting it anyway is worse than reporting nothing, because this field
// exists to SIZE --max-age: a bogus 1s makes every threshold look generous and
// silences the false-alarm warning — the check stops working in exactly the
// state that produced the bad reading.
func TestARestartDoesNotMasqueradeAsAOneSecondCadence(t *testing.T) {
	for _, gap := range []time.Duration{0, time.Second} {
		dir := stateWith(t, time.Second, true, "123-0")
		now := time.Now().UTC()
		writeStampAt(t, dir, fileHealthArmPrev, now.Add(-gap))
		writeStampAt(t, dir, fileHealthArm, now)

		l, _ := CheckLiveness(dir, 5*time.Minute)
		if l.IntervalSeconds != 0 {
			t.Errorf("gap %s reported IntervalSeconds=%d — a restart was measured as a cadence",
				gap, l.IntervalSeconds)
		}
		if l.ConfigWarning != "" {
			t.Errorf("gap %s produced a config warning from a non-measurement: %q", gap, l.ConfigWarning)
		}
	}
}

// ...but a genuine cadence at or above the floor still measures.
func TestAGenuineCadenceIsStillMeasured(t *testing.T) {
	dir := stateWith(t, time.Second, true, "123-0")
	now := time.Now().UTC()
	writeStampAt(t, dir, fileHealthArmPrev, now.Add(-60*time.Second))
	writeStampAt(t, dir, fileHealthArm, now)

	l, _ := CheckLiveness(dir, 5*time.Minute)
	if l.IntervalSeconds != 60 {
		t.Fatalf("IntervalSeconds = %d, want 60 — and it must be SECONDS, not minutes (#373's title reading)", l.IntervalSeconds)
	}
}

func writeStampAt(t *testing.T, dir, name string, at time.Time) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(at.Format(time.RFC3339)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// 🛑 A RETIRED LEG IS NOT A DEAD ONE (@Tuner, #363). Both stopped polling; one
// demands action and the other demands the opposite. Reported as `stale`, a
// deliberately retired leg is a permanent red nobody can clear — and an alarm
// that is always red is one its reader learns to skip, which costs more than
// having no alarm at all.
func TestRetiredLegIsNotAFault(t *testing.T) {
	dir := stateWith(t, 6*time.Hour, true, "123-0") // long dead by every other measure
	if l, _ := CheckLiveness(dir, 5*time.Minute); l.State != LiveStale {
		t.Fatalf("precondition: want %q without a marker, got %q", LiveStale, l.State)
	}

	if err := os.WriteFile(filepath.Join(dir, fileRetired),
		[]byte("dev room, retired when the fleet moved to demo"), 0o644); err != nil {
		t.Fatal(err)
	}

	l, _ := CheckLiveness(dir, 5*time.Minute)
	if l.State != LiveRetired {
		t.Fatalf("State = %q, want %q", l.State, LiveRetired)
	}
	if !l.OK() {
		t.Error("a retired leg reported NOT OK — that is the permanent red this removes")
	}
	if l.Alive() {
		t.Error("retired must not report ALIVE — it is not polling, it is just not a fault")
	}
	if !strings.Contains(l.Detail, "moved to demo") {
		t.Errorf("the marker's note was dropped: %q — it is the only record of WHY", l.Detail)
	}
}

// ⚠️ The negative pole. Retiring one leg must not make a genuinely dead one
// look fine — the marker is per-leg and must not generalise.
func TestRetiringOneLegDoesNotExcuseAnother(t *testing.T) {
	retired := stateWith(t, 6*time.Hour, true, "1-0")
	if err := os.WriteFile(filepath.Join(retired, fileRetired), []byte("gone"), 0o644); err != nil {
		t.Fatal(err)
	}
	dead := stateWith(t, 6*time.Hour, true, "2-0") // no marker

	if l, _ := CheckLiveness(retired, 5*time.Minute); !l.OK() {
		t.Error("retired leg should be OK")
	}
	l, _ := CheckLiveness(dead, 5*time.Minute)
	if l.OK() || l.State != LiveStale {
		t.Fatalf("the unmarked leg reported %q/OK=%v — a retirement leaked across legs", l.State, l.OK())
	}
}

// ⚑ OK() and Alive() answer different questions and must not collapse: "is it
// polling" versus "does anything need action". Conflating them is what made a
// deliberately stopped leg indistinguishable from a dead one.
func TestOKAndAliveAreDifferentQuestions(t *testing.T) {
	alive := stateWith(t, time.Second, true, "1-0")
	l, _ := CheckLiveness(alive, 5*time.Minute)
	if !l.Alive() || !l.OK() {
		t.Fatal("a polling leg must be both alive and OK")
	}
	dead := stateWith(t, 6*time.Hour, true, "2-0")
	d, _ := CheckLiveness(dead, 5*time.Minute)
	if d.Alive() || d.OK() {
		t.Fatal("a dead leg must be neither")
	}
}

// ⚠️ The marker must be found whatever case it is written in (@Gordon).
//
// He wrote `RETIRED` and it worked — because his volume is case-INSENSITIVE.
// On a case-sensitive filesystem the same file is silently ignored and the leg
// reports `stale` forever, which is the permanent red this feature removes.
// `pfw` ships for linux as well as darwin, so a filename normalised by the
// author's filesystem is a defect only remote users meet.
func TestRetiredMarkerIsFoundInAnyCase(t *testing.T) {
	for _, name := range []string{"retired", "RETIRED", "Retired"} {
		dir := stateWith(t, 6*time.Hour, true, "1-0")
		if err := os.WriteFile(filepath.Join(dir, name), []byte("moved to demo"), 0o644); err != nil {
			t.Fatal(err)
		}
		l, _ := CheckLiveness(dir, 5*time.Minute)
		if l.State != LiveRetired {
			t.Errorf("marker %q → state %q, want %q", name, l.State, LiveRetired)
		}
		if !strings.Contains(l.Detail, "moved to demo") {
			t.Errorf("marker %q: note not surfaced: %q", name, l.Detail)
		}
	}
}
