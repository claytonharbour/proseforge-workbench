package watcher

import (
	"strings"
	"testing"
	"time"
)

// #400 (@Vance). Cases are HIS measured configurations and @Aldric's sweep, so
// a change here is a change against real fleet state, not against my model.
func TestSamplingWarning(t *testing.T) {
	const (
		min = time.Minute
		sec = time.Second
	)
	for _, tc := range []struct {
		name                           string
		observed, maxAge, loopInterval time.Duration
		wantFire                       bool
		wantContains                   string
	}{
		// @Vance's demo leg: 60s poll, watchdog 2m. Spacing is 90s, so the
		// sampler is coarser than the peaks — NO threshold can work.
		{"structural: watchdog coarser than spacing", 60 * sec, 5 * min, 2 * min,
			true, "coarser than this leg's poll spacing"},

		// @Aldric's fix: same leg, watchdog 60s. Cadence now fine, and :3m
		// clears every peak by more than a sample.
		{"@Aldric's applied fix is silent", 60 * sec, 3 * min, 60 * sec, false, ""},
		// ⚠️ WAS asserted clean. It is not: a 60s leg's 2-failure peak is 4m0s
		// exactly, so :4m sits ON it. Corrected 2026-08-23 when the <= became <.
		{"@Aldric's :4m sits ON the 2-failure peak", 60 * sec, 4 * min, 60 * sec,
			true, "below the 2-failure peak"},
		{"@Aldric's :4m1s clears it", 60 * sec, 4*min + sec, 60 * sec, false, ""},

		// @Aldric swept this: :5m still holds a hole at N=3 even at 60s
		// sampling — peak 5m30s, excursion 30s < 60s.
		{"specific: :5m holes at N=3", 60 * sec, 5 * min, 60 * sec,
			true, "below the 3-failure peak"},

		// @Vance's prod: 30m leg, :90m, watchdog 2m. Peak for N=2 is 91m, so
		// the excursion is 60s against a 120s sampler — his coin flip.
		{"specific: prod :90m is a coin flip at N=2", 30 * min, 90 * min, 2 * min,
			true, "below the 2-failure peak"},

		// Just above that peak: deterministic, per his own recommendation.
		// ⚠️ @Vance recommended :91m as deterministic and it is EXACTLY the
		// 2-failure peak. The tool's own suggestion is peak+1s, which is right;
		// the recommendation in the ticket was off by the same exact-multiple
		// error @Vance himself had warned about.
		{"prod :91m sits ON the peak", 30 * min, 91 * min, 2 * min,
			true, "below the 2-failure peak"},
		{"prod :91m1s is deterministic", 30 * min, 91*min + sec, 2 * min, false, ""},

		// 🛑 EXACTLY on a peak. @Sten's sweep counted a zero excursion as clean
		// and so did this checker until 2026-08-23. A 60s leg's 2-failure peak
		// is 4m0s, so :4m sits precisely on it — maximally fragile, and it must
		// NOT read as tolerated.
		{"exactly on a peak is fragile, not safe", 60 * sec, 4 * min, 30 * sec,
			true, "below the 2-failure peak"},

		// Degenerate inputs must not manufacture a warning.
		{"no observed interval", 0, 5 * min, 2 * min, false, ""},
		{"not looping", 60 * sec, 5 * min, 0, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := SamplingWarning(tc.observed, tc.maxAge, tc.loopInterval)
			if tc.wantFire && got == "" {
				t.Fatalf("wanted a warning, got silence — a guard whose failure mode is silence "+
					"must not be silent here (observed=%s maxAge=%s loop=%s)",
					tc.observed, tc.maxAge, tc.loopInterval)
			}
			if !tc.wantFire && got != "" {
				t.Fatalf("wanted silence on a sound config, got: %s", got)
			}
			if tc.wantContains != "" && !strings.Contains(got, tc.wantContains) {
				t.Errorf("warning does not name the case:\n  got:  %s\n  want it to contain: %q",
					got, tc.wantContains)
			}
		})
	}
}

// 🛑 The negative control. A guard that fires on everything is as useless as one
// that fires on nothing, and this one's whole purpose is to be trusted when it
// stays quiet — so prove it CAN stay quiet across a realistic spread.
func TestSamplingWarningIsSilentOnSoundConfigs(t *testing.T) {
	sound := []struct{ observed, maxAge, loop time.Duration }{
		{60 * time.Second, 3 * time.Minute, 60 * time.Second},             // @Aldric demo, fixed
		{30 * time.Minute, 91*time.Minute + time.Second, 2 * time.Minute}, // @Vance prod, peak+1s
		{15 * time.Second, 46 * time.Second, 10 * time.Second},            // @Rowan's fast leg
	}
	for _, c := range sound {
		if w := SamplingWarning(c.observed, c.maxAge, c.loop); w != "" {
			t.Errorf("false positive on a sound config (observed=%s maxAge=%s loop=%s): %s",
				c.observed, c.maxAge, c.loop, w)
		}
	}
}

// 🛑 The remedy a guard prints is itself a configuration recommendation, and
// this one was wrong in the same family as the bug it was written to catch
// (@Vance, @Sten, and @Aldric all landing on it within minutes of each other).
//
//	peak + 1s     satisfies the rule, and is the MOST FRAGILE position the rule
//	              permits — one second of scheduling drift from being a hazard
//	midpoint      the fleet's agreed fix, and STILL WRONG here: it splits a 90s
//	              gap evenly and hands 45s to a side that needs 60s
//	band centre   satisfy the cadence floor first, centre in what is left
//
// ⚑ The rule that decides whether a config is safe and the rule for where to
// actually sit are DIFFERENT RULES, and only the first was in the tool.
func TestRemedyCentresTheFeasibleBandNotTheGap(t *testing.T) {
	// 60s leg: peaks at 4m0s (N=2) and 5m30s (N=3), 90s apart.
	got := SamplingWarning(60*time.Second, 4*time.Minute, 60*time.Second)
	if got == "" {
		t.Fatal("a threshold exactly ON a peak must warn")
	}
	// spacing 90s, cadence 60s ⇒ feasible tolerate-margin is 0..30s, centre 15s
	// ⇒ 4m0s + 15s. NOT 4m45s: the plain midpoint leaves a 45s breach against a
	// 60s cadence and is itself a hole.
	if !strings.Contains(got, "4m15s") {
		t.Errorf("remedy should centre the FEASIBLE band at 4m15s:\n%s", got)
	}
	for _, bad := range []string{"4m1s", "4m45s"} {
		if strings.Contains(got, bad) {
			t.Errorf("remedy %s optimises the wrong axis:\n%s", bad, got)
		}
	}
}

// ⚠️ @Aldric: SamplingWarning is binary, so a 1s margin and a 45s margin are the
// same empty string. Reporting the distance is what lets an operator disagree
// with the tool — @Aldric read his own margin as 14x his measured drift and
// declined the optimum, which is a decision the binary verdict cannot support.
func TestMarginIsReportedOnBothSidesSoSilenceIsNotTheAnswer(t *testing.T) {
	const iv = 60 * time.Second
	for _, tc := range []struct {
		name   string
		maxAge time.Duration
		want   []string
	}{{
		name: "band optimum names its own headroom", maxAge: 4*time.Minute + 15*time.Second,
		want: []string{"15s above", "1m15s below", "4m0s", "5m30s", "Best in this band is 4m15s"},
	}, {
		// ⛔ Silent under SamplingWarning and one drift-tick from the hazard.
		// This is @Sten's tolerate-side case and the reason margin exists.
		name: "one second above a peak is silent but not safe", maxAge: 4*time.Minute + time.Second,
		want: []string{"1s above", "2-failure peak 4m0s", "tolerate side"},
	}, {
		name: "below the first peak has no lower neighbour", maxAge: 30 * time.Second,
		want: []string{"detect side", "No peak below it"},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			if w := SamplingWarning(iv, tc.maxAge, iv); w != "" {
				t.Fatalf("fixture is meant to be silent under the warn path, got:\n%s", w)
			}
			got := SamplingMargin(iv, tc.maxAge, iv)
			for _, sub := range tc.want {
				if !strings.Contains(got, sub) {
					t.Errorf("margin missing %q:\n%s", sub, got)
				}
			}
		})
	}
}
