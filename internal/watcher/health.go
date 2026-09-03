package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Liveness for a watcher (forge/proseforge-workbench#248).
//
// 🛑 The failure this exists for: a watcher that has STOPPED looks exactly like
// a room with nothing to report. Every quiet tick is indistinguishable from
// death, so the failure is invisible precisely while it is happening (@Tuner,
// whose poll cron vanished from CronList with no error).
//
// ⚑ The signal is the health-poll stamp, and the reason it works is @Sten's:
// `room watch` refreshes it on every SUCCESSFUL poll and never on a failure, so
// its age answers "when did a poll last actually succeed" WITHOUT caring what
// any exit code means. That matters because enumerating exit codes is a losing
// game — he found a fourth by accident inside ten minutes, and until an
// enumeration is updated the guard built on it is silently wrong.
//
// So this asks for POSITIVE evidence of a recent success. It cannot be fooled by
// a mode nobody has enumerated yet.

// Liveness is what a watcher's state dir says about whether it is alive.
//
// There is deliberately no single boolean. "Never polled" and "polled and then
// stopped" are different problems with different fixes, and collapsing them is
// how a fresh install and a dead watcher come to look the same.
type Liveness struct {
	// State is one of: alive, failing, stale, never, unknown.
	State string `json:"state"`

	LastPoll   string `json:"last_poll,omitempty"`
	AgeSeconds int64  `json:"age_seconds,omitempty"`

	// LastArm is when a poll last STARTED, as against LastPoll which is when
	// one last SUCCEEDED. The gap between them is the diagnosis (#367).
	LastArm       string `json:"last_arm,omitempty"`
	ArmAgeSeconds int64  `json:"arm_age_seconds,omitempty"`

	// IntervalSeconds is the OBSERVED gap between the last two poll attempts —
	// measured, not configured. It is what --max-age has to be sane against.
	IntervalSeconds int64 `json:"interval_seconds,omitempty"`

	// NoStateFound marks the unknown that means "nothing of ours is here",
	// as against the unknown that means "this dir cannot be read". Same state
	// word, opposite next action: one is a path to check, the other is a
	// permission or a mount.
	NoStateFound bool `json:"no_state_found,omitempty"`

	// ConfigWarning names a --max-age that cannot work against the observed
	// interval. Populated regardless of state: a threshold that false-alarms
	// must be said out loud even on a tick where it happens to read alive.
	ConfigWarning string `json:"config_warning,omitempty"`

	Baselined  string `json:"baselined,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	QueueBytes int64  `json:"queue_bytes,omitempty"`

	// RunningVersion is the binary actually executing this leg, as the watcher
	// itself recorded it — not as any stream reported it (see fileVersion).
	RunningVersion string `json:"running_version,omitempty"`

	Detail string `json:"detail"`
}

const (
	// LiveAlive — a poll succeeded within the limit.
	LiveAlive = "alive"
	// LiveFailing — polls are STILL BEING ATTEMPTED and none is succeeding.
	//
	// ⚑ Distinct from stale on purpose, because the response is opposite:
	// failing means the process is alive and the BACKEND or the CREDENTIALS are
	// refusing it; stale means nothing is polling at all. Told only "unhealthy",
	// half the readers go and restart a process that was working correctly.
	//
	// 🛑 This is the state my own rc=10 exemption made unreachable for a day:
	// an unreachable backend and a held lock exited the same way, so I treated
	// both as benign and the loud alarm could never fire for the real one.
	LiveFailing = "failing"
	// LiveStale — nothing is polling. Not one attempt inside --max-age.
	LiveStale = "stale"
	// LiveNever — no poll has ever succeeded. A fresh install, or one that has
	// never worked; distinct from stale because the fix is different.
	LiveNever = "never"
	// LiveUnknown — the state dir cannot be read at all.
	LiveUnknown = "unknown"
	// LiveRetired — this leg was DELIBERATELY stopped (@Tuner, #363).
	//
	// 🛑 Without it a retired leg reports `stale` forever, and stale demands
	// action while retired demands the opposite. That is a permanent red nobody
	// can clear, and an alarm that is always red is one its reader learns to
	// skip — which costs more than having no alarm.
	//
	// ⚑ It is a DECLARATION, not a detection: the operator writes the marker,
	// because "was this stopped on purpose?" is not answerable from the state
	// dir. Nothing infers it.
	LiveRetired = "retired"
)

// fileRetired marks a leg as deliberately stopped. Its presence is the whole
// mechanism; contents are a free-text note for whoever reads it later.
const fileRetired = "retired"

// Alive reports whether the watcher is polling. Callers wanting a verdict
// should still print State — see the note on Liveness.
func (l *Liveness) Alive() bool { return l.State == LiveAlive }

// OK reports whether anything requires action — the question an exit code and
// an alarm are actually asking.
//
// ⚑ A retired leg is NOT alive and nothing is wrong with it. Collapsing those
// two into one boolean is what made a deliberately stopped leg indistinguishable
// from a dead one.
func (l *Liveness) OK() bool { return l.State == LiveAlive || l.State == LiveRetired }

// minMeasurableInterval is the floor below which an arm-to-arm gap is treated as
// a restart artefact rather than a cadence (#373). Deliberately not zero: gaps of
// exactly 0s and 1s are both produced by a re-exec, and neither is a measurement.
const minMeasurableInterval = 2 * time.Second

// CheckLiveness inspects a watcher's state dir.
//
// maxAge is the caller's policy: how long without a successful poll before the
// watcher is presumed dead. A sensible value is a small multiple of the poll
// interval — long enough that one slow tick is not an alarm, short enough that
// a genuine death is noticed while it still matters.
func CheckLiveness(stateDir string, maxAge time.Duration) (*Liveness, error) {
	if stateDir == "" {
		return nil, fmt.Errorf("no state dir")
	}
	if fi, err := os.Stat(stateDir); err != nil || !fi.IsDir() {
		return &Liveness{
			State:  LiveUnknown,
			Detail: fmt.Sprintf("state dir %s is unreadable — a watcher that never started leaves no trace", stateDir),
		}, nil
	}

	l := &Liveness{
		RunningVersion: readTrimmed(filepath.Join(stateDir, fileVersion)),
		Cursor:         readTrimmed(filepath.Join(stateDir, fileCursor)),
		Baselined:      readTrimmed(filepath.Join(stateDir, fileBaseline)),
	}

	// ⚑ Checked FIRST: a retired leg has stopped polling by definition, so every
	// check below would report it stale — correctly, and uselessly. A permanent
	// red nobody can clear is an alarm its reader learns to skip (@Tuner, #363).
	//
	// 🛑 It is a DECLARATION, not a detection. "Was this stopped on purpose?"
	// cannot be answered from the state dir, so the operator writes the marker
	// and nothing infers it.
	if retiredMarker(stateDir) != "" {
		l.State = LiveRetired
		l.Detail = "retired on purpose — this leg was deliberately stopped. Not a fault."
		if note := readTrimmed(filepath.Join(stateDir, retiredMarker(stateDir))); note != "" {
			l.Detail += " Note: " + note
		}
		return l, nil
	}

	// ── The ARM stamp: evidence that something TRIED (#367) ─────────────────
	//
	// Read before anything is decided, because it changes the meaning of every
	// other answer below. Absent on state dirs written by builds before #367 —
	// those fall through to exactly the pre-#367 behaviour, so upgrading does
	// not reclassify a fleet that is running fine.
	arm, hasArm := readStamp(stateDir, fileHealthArm)
	if hasArm {
		l.LastArm = arm.UTC().Format(time.RFC3339)
		l.ArmAgeSeconds = int64(time.Since(arm).Seconds())
		if prevArm, ok := readStamp(stateDir, fileHealthArmPrev); ok {
			// 🛑 A sub-second gap is a RESTART, not a cadence (#373, @Smiley).
			//
			// The process arms on start, moments after the previous process's
			// last tick armed — so a re-exec or a manual restart leaves two arm
			// stamps in the same second and the "observed interval" collapses to
			// 0 or 1s. On my own 60s leg after a self-re-exec, both stamps read
			// 13:09:27 exactly.
			//
			// ⚠️ ONE SAMPLE CANNOT TELL a 1s cadence from a restart. Reporting
			// the number anyway is worse than reporting nothing, because this
			// field exists to SIZE --max-age: a bogus 1s makes any threshold
			// look generous and silences the false-alarm warning entirely —
			// the check stops working in exactly the state that produced it.
			//
			// So: under-report. An absent interval says "not measured"; a wrong
			// one says "measured, and small".
			if gap := arm.Sub(prevArm); gap >= minMeasurableInterval {
				l.IntervalSeconds = int64(gap.Seconds())
				l.ConfigWarning = checkMaxAge(maxAge, gap)
			}
		}
	}
	armFresh := hasArm && maxAge > 0 && time.Since(arm) <= maxAge

	stamp, ok := readStamp(stateDir, fileHealthOK)
	if !ok {
		switch {
		// 🛑 NOTHING HERE AT ALL — no success, no attempt, no cursor.
		//
		// This is UNKNOWN, not never, and the difference is a false alarm about
		// a bench that is fine. @Angel polls natively and leaves no stamp of
		// ours; "has never polled" is a claim about THEIR watcher made from the
		// absence of OUR file, and it is the loudest thing this command can say
		// about the one bench it knows least about (#367 criterion 4).
		//
		// ⚠️ Still exit 3. Not-alive must stay red — the softening to guard
		// against is a detector that goes quiet, not one that stops guessing.
		case !hasArm && l.Cursor == "" && !fileExists(stateDir, fileHealthOK):
			l.State = LiveUnknown
			l.NoStateFound = true
			l.Detail = "no pfw watcher state here at all — no attempt, no success, no cursor. " +
				"Either the wrong --state dir, or a watcher that does not use `pfw room watch`. " +
				"This is NOT a claim that anything died."

		// Armed and never succeeded. Now evidenced rather than guessed: we know
		// a pfw watcher ran here, so "it has never worked" is a finding.
		case hasArm:
			l.State = LiveNever
			l.Detail = fmt.Sprintf(
				"polls ARE being attempted (last attempt %s ago) and not one has EVER succeeded. "+
					"The process is running — look at credentials, --url, or the backend, NOT at whether it is alive.",
				time.Since(arm).Round(time.Second))
			if !armFresh {
				l.Detail += " ⚠️ and the attempts have stopped too — nothing has tried recently either."
			}

		// A cursor or an unparseable stamp FILE: a watcher was here, from a
		// build before the arm stamp existed, or its stamp got truncated.
		//
		// 🛑 The truncated case is why this is not folded into the branch above.
		// An EMPTY health-poll file is positive evidence a watcher wrote here —
		// classifying it "no state at all" would report a corrupted watcher as
		// somebody else's business. I wrote that bug into the first version of
		// this switch and the existing truncated-stamp test caught it, which is
		// the whole argument for keeping tests that assert the negative pole.
		default:
			l.State = LiveNever
			l.Detail = "no successful poll has EVER been recorded. Fresh install, wrong --state dir, or it has never worked."
			if l.Cursor != "" {
				l.Detail += " ⚠️ a cursor exists without a poll stamp, which should not happen — check the state dir is the one the watcher uses."
			} else {
				l.Detail += " ⚠️ a poll stamp file EXISTS but cannot be parsed — it was truncated or half-written."
			}
		}

		// 🛑 GO AND LOOK before reporting a watcher dead. This said "wrong --state
		// dir" as one of three guesses and never checked which — so @Smiley got
		// `never` on a watcher that had delivered eight events in the previous
		// eighteen minutes and whose stamp was written four seconds earlier. He
		// had passed the parent; the stamp was one level down.
		//
		// ⚠️ A wrong path and a dead watcher are OPPOSITE emergencies — one is a
		// typo, the other is "you have been deaf for hours" — and this returned
		// the second for the first, during a fleet-wide liveness audit. Naming a
		// possible cause is not the same as ruling it out.
		if found, at := findStampNearby(stateDir); found {
			l.State = LiveUnknown
			l.Detail = fmt.Sprintf(
				"NO STAMP HERE, but one exists at %s (last poll %s ago) — you almost certainly "+
					"want that path. This is a wrong --state dir, NOT a dead watcher.",
				at.path, time.Since(at.stamp).Round(time.Second))
		}
		return l, nil
	}

	age := time.Since(stamp)
	l.LastPoll = stamp.UTC().Format(time.RFC3339)
	l.AgeSeconds = int64(age.Seconds())

	switch {
	case maxAge > 0 && age > maxAge && armFresh:
		// ⚑ THE PAIR EARNING ITS KEEP. Attempts are current, successes are not:
		// the loop is alive and something is refusing it. Reported as its own
		// state because "stale" would send the reader to restart a process that
		// is doing its job, and the restart would look like a fix — the alarm
		// clears on the next tick either way.
		l.State = LiveFailing
		l.Detail = fmt.Sprintf(
			"polls are RUNNING (last attempt %s ago) but none has succeeded for %s (limit %s). "+
				"The process is fine — this is the backend, the credentials, or the network.",
			time.Since(arm).Round(time.Second), age.Round(time.Second), maxAge)

	case maxAge > 0 && age > maxAge:
		l.State = LiveStale
		l.Detail = fmt.Sprintf(
			"no successful poll for %s (limit %s). THE WATCHER IS NOT RECEIVING — this is silence, not a quiet room.",
			age.Round(time.Second), maxAge)
		if hasArm {
			l.Detail += fmt.Sprintf(" Nothing has even ATTEMPTED a poll for %s — the loop is not running.",
				time.Since(arm).Round(time.Second))
		}

	default:
		l.State = LiveAlive
		l.Detail = fmt.Sprintf("last successful poll %s ago", age.Round(time.Second))
	}
	return l, nil
}

// checkMaxAge names a --max-age that cannot work against the interval actually
// observed, and returns "" when the threshold is sane (#367 criterion 2).
//
// 🛑 The trap is one-sided and it is @Smiley's: a threshold BELOW the poll
// interval is red for most of every interval even when the watcher is perfectly
// healthy. 5m against a 30m leg is stale for 25 minutes in every 30 — and an
// alarm that is usually red teaches its reader to ignore it, which costs more
// than having no alarm at all. @Tuner's leg went uncovered for 6.5 hours to
// exactly this.
//
// ⚠️ It reports rather than corrects. Silently widening the caller's threshold
// would be a monitor deciding on its own how deaf it is willing to let you be.
func checkMaxAge(maxAge, observed time.Duration) string {
	if maxAge <= 0 || observed <= 0 {
		return ""
	}
	if maxAge <= observed {
		return fmt.Sprintf(
			"🛑 --max-age %s is BELOW the observed poll interval %s. This reports stale between "+
				"healthy polls — a guaranteed false alarm. Use a small MULTIPLE of the interval (%s or more).",
			maxAge, observed.Round(time.Second), (observed * 3).Round(time.Second))
	}
	if maxAge < observed*2 {
		return fmt.Sprintf(
			"⚠️ --max-age %s is under 2× the observed interval %s. One slow poll will read as death.",
			maxAge, observed.Round(time.Second))
	}
	return ""
}

// fileExists reports whether a state file is present at all, regardless of
// whether its contents parse. The distinction matters: an unparseable stamp is
// evidence a watcher wrote here, and a missing one is evidence of nothing.
func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// retiredMarker returns the retirement marker's filename, or "" if absent.
//
// ⚠️ CASE-INSENSITIVE BY HAND, not by luck. @Gordon wrote `RETIRED` and it
// worked — because this volume happens to be case-insensitive. On a
// case-sensitive filesystem the same marker is silently ignored and the leg
// reports `stale` forever, which is the permanent red this whole feature
// exists to remove. The tool ships for linux as well as darwin, so relying on
// the developer's filesystem to normalise a filename is a defect that only
// appears for users we cannot see.
func retiredMarker(stateDir string) string {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), fileRetired) {
			return e.Name()
		}
	}
	return ""
}

// readStamp parses an RFC3339 timestamp file. Shared with the watch path so the
// two cannot disagree about what a stamp means.
func readStamp(dir, name string) (time.Time, bool) {
	raw := readTrimmed(filepath.Join(dir, name))
	if raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// nearbyStamp is a health stamp found somewhere other than where the caller
// looked, with the path to report back.
type nearbyStamp struct {
	path  string
	stamp time.Time
}

// findStampNearby probes the directories a caller most often means when they
// miss — one level down (`<dir>/state`, the layout every bench in the fleet
// uses) and one level up, for the caller who went too deep.
//
// ⚑ Deliberately shallow: two candidates, no walk. This runs only on the
// already-failed path, and a recursive search could surface some OTHER
// watcher's stamp and confidently report the wrong watcher as healthy — which
// would be a worse error than the one it fixes.
func findStampNearby(stateDir string) (bool, nearbyStamp) {
	for _, cand := range []string{
		filepath.Join(stateDir, "state"),
		filepath.Dir(stateDir),
	} {
		if cand == stateDir || cand == "." || cand == string(filepath.Separator) {
			continue
		}
		if stamp, ok := readStamp(cand, fileHealthOK); ok {
			return true, nearbyStamp{path: cand, stamp: stamp}
		}
	}
	return false, nearbyStamp{}
}

// failureOverhead is what a failing poll costs ON TOP of the interval, so a
// leg's "peak age" after N failures is (N+1)*interval + N*overhead.
//
// ⚑ MEASURED, not assumed: across 30 real failing ticks the gap was
// interval+30s on both 15s and 60s legs — so the cost is FIXED, not a multiple
// of the interval. That fact is what stops anyone extrapolating a RATIO: a
// single failure peaks at ~2.8x a 60s interval and ~2.0x a 30m one, purely
// because the same 30s is half of one and 1.7% of the other.
//
// ⚠️ DELIBERATELY CONSERVATIVE. @Aldric decomposed the recovery cycle at 46-57s
// (the failing tick costs ~30s, the recovering attempt ~27s more). Using the
// smaller value UNDER-states the peak, which under-states the excursion, which
// makes this guard warn MORE. For a check whose failure mode is silence, erring
// toward warning is the correct direction.
const failureOverhead = 30 * time.Second

// SamplingWarning reports when a watchdog's own --interval cannot resolve the
// threshold it is checking, and returns "" when the configuration is sound.
//
// 🛑 THIS IS THE INVERSE OF checkMaxAge AND IT IS WORSE (#400, @Vance).
// checkMaxAge catches a threshold that reports a healthy leg dead — loud, and
// self-limiting because someone investigates. This catches a threshold that
// reports a DEAD leg alive, and its output is byte-identical to health:
// `alive rc=0`. Nothing goes red, so nobody looks.
//
// The mechanism is aliasing. During an outage `age` climbs continuously and
// peaks at recovery, so it exceeds --max-age for only (peak - maxAge). If that
// excursion is narrower than the sampling interval, the check can step over it:
//
//	leg 60s · --max-age 5m · watchdog 2m · 3 failed polls
//	  peak 5m30s, exceeds the limit for 30s, sampled every 120s → seen ~25% of the time
//
// ⛔ The two failure modes pull OPPOSITE ways on --max-age: raising it silences
// checkMaxAge and WIDENS this one. There is no safe threshold until the
// watchdog's cadence is right, which is why the structural check comes first.
func SamplingWarning(observed, maxAge, loopInterval time.Duration) string {
	if observed <= 0 || maxAge <= 0 || loopInterval <= 0 {
		return ""
	}
	spacing := observed + failureOverhead

	// ① STRUCTURAL. Peaks are `spacing` apart, so a sampler coarser than that
	// always has some N landing in a sub-sample excursion — for EVERY threshold.
	// Tuning --max-age here sends the operator after a number that cannot exist.
	if loopInterval > spacing {
		return fmt.Sprintf(
			"🛑 sample cadence %s is coarser than this leg's poll spacing %s. Some outage lengths are "+
				"undetectable at ANY --max-age — a breach can fall entirely between samples. "+
				"Use %s or less.",
			loopInterval, spacing.Round(time.Second), spacing.Round(time.Second))
	}

	// ② SPECIFIC. The cadence is fine; this threshold happens to sit just under
	// a peak, so the first outage that breaches it does so too briefly to see.
	for n := 1; n <= 16; n++ {
		peak := time.Duration(n+1)*observed + time.Duration(n)*failureOverhead
		// 🛑 STRICTLY LESS. A peak exactly EQUAL to maxAge is the most fragile
		// position available, not a safe one: the age reaches the threshold
		// precisely at recovery, so whether it registers depends on jitter and
		// on > versus >= somewhere else. @Vance warned about exact multiples;
		// @Sten acknowledged it and then wrote a sweep whose hole test was
		// `0 < exc < cadence`, which passes a zero excursion. I had the same
		// bug here, in the checker, on the same day — `<=` skipped the equal
		// case as "tolerated by design" when nothing about it is by design.
		if peak < maxAge {
			continue // genuinely tolerated — recovery lands before the threshold
		}
		if excursion := peak - maxAge; excursion < loopInterval {
			return fmt.Sprintf(
				"⚠️ --max-age %s sits %s below the %d-failure peak %s, under the %s sample cadence — "+
					"that outage is detected only some of the time. Use --max-age %s: it tolerates %d "+
					"with %s of drift headroom and still leaves a %s breach, longer than your cadence.",
				maxAge, excursion.Round(time.Second), n, peak.Round(time.Second),
				loopInterval, bandOptimum(peak, spacing, loopInterval).Round(time.Second), n,
				((spacing - loopInterval) / 2).Round(time.Second),
				(loopInterval + (spacing-loopInterval)/2).Round(time.Second))
		}
		break // first breaching N is cleanly detectable; later ones are wider still
	}
	return ""
}

// bandOptimum returns where to sit in the gap ABOVE `peak`, given that peaks are
// `spacing` apart and the watchdog samples every `cadence`.
//
// 🛑 NOT THE PLAIN MIDPOINT, and I shipped that first. @Sten, @Vance and @Aldric
// all converged on "midpoint between consecutive peaks" within minutes, I agreed,
// implemented it — and my own test caught that it FAILS THIS GUARD on the very
// leg the fleet runs. The two margins are not symmetric quantities:
//
//	tolerate-margin = maxAge - peak(n)     must beat interval DRIFT   (~1-2s)
//	detect-excursion = peak(n+1) - maxAge  must beat the CADENCE      (30-60s)
//	                                       ─────────────────────────
//	                       they sum to spacing, so they trade 1:1
//
// ⚑ Splitting them evenly gives each side half of `spacing` and asks a 45s
// half to cover a 60s cadence. Measured: 60s leg, 90s spacing, 60s cadence —
// the "optimum" midpoint 4m45s leaves a 45s breach and IS STILL A HOLE.
//
// The cadence is a hard floor and drift is a soft one, so the honest answer is
// to satisfy the floor first and centre in what remains: the feasible band is
// a tolerate-margin of 0..(spacing-cadence), and we take the middle of THAT.
//
// ⚠️ The general shape, which is the same error three times in one afternoon:
// an optimum found by symmetry is only optimal if the axes are commensurable.
// peak+1s optimised nothing; the plain midpoint optimised the wrong thing.
func bandOptimum(peak, spacing, cadence time.Duration) time.Duration {
	if spacing <= cadence {
		return peak // caller's structural check ① owns this case
	}
	return peak + (spacing-cadence)/2
}

// SamplingMargin describes how far a sound threshold sits from the nearest peak
// in EITHER direction, and returns "" only when the inputs are unusable.
//
// 🛑 SILENCE IS NOT A MEASUREMENT, and this is the gap @Aldric named and @Sten
// then demonstrated on his own leg. SamplingWarning is binary, so a threshold
// 3s clear of a peak and one 33s clear produce the same empty string — and the
// first is one tick of scheduling drift away from being the thing the guard
// exists to catch. Observed drift on this fleet is ±1s at 13-14% of ticks
// (@Aldric 1257 gaps, @Sten 237), so single-digit margins are not theoretical.
//
// ⚠️ BOTH SIDES MATTER, and they fail differently (@Sten):
//
//	just BELOW a peak   you DETECT that outage, and the excursion is narrow
//	just ABOVE a peak   you TOLERATE it, and drift walks the peak past you
//
// Which side you are on decides which direction of drift hurts you, so the
// margin is reported per-side rather than collapsed into one number — and
// bandOptimum names the best point in the gap, which is NOT the midpoint.
//
// ⛔ This deliberately does NOT model jitter. Turning ±1s into a pass/fail
// threshold would replace a number the operator can weigh with a verdict
// derived from one bench's measurement of one leg.
func SamplingMargin(observed, maxAge, loopInterval time.Duration) string {
	if observed <= 0 || maxAge <= 0 {
		return ""
	}
	spacing := observed + failureOverhead

	var below, above time.Duration // nearest peak under / over maxAge
	var belowN, aboveN int
	for n := 1; n <= 64; n++ {
		peak := time.Duration(n+1)*observed + time.Duration(n)*failureOverhead
		if peak < maxAge {
			below, belowN = peak, n
			continue
		}
		above, aboveN = peak, n
		break
	}
	if above == 0 {
		return "" // maxAge is beyond every peak we enumerate — nothing useful to say
	}

	if below == 0 { // maxAge sits under the very first peak: no lower neighbour
		return fmt.Sprintf("margin: %s to the %d-failure peak %s (detect side). No peak below it.",
			(above - maxAge).Round(time.Second), aboveN, above.Round(time.Second))
	}
	best := ""
	if loopInterval > 0 {
		if opt := bandOptimum(below, spacing, loopInterval); opt != below {
			best = fmt.Sprintf(" Best in this band is %s (%s of drift headroom, %s breach).",
				opt.Round(time.Second), (opt - below).Round(time.Second),
				(above - opt).Round(time.Second))
		}
	}
	return fmt.Sprintf(
		"margin: %s above the %d-failure peak %s (tolerate side — beat interval drift), %s below "+
			"the %d-failure peak %s (detect side — beat your cadence).%s",
		(maxAge - below).Round(time.Second), belowN, below.Round(time.Second),
		(above - maxAge).Round(time.Second), aboveN, above.Round(time.Second), best)
}
