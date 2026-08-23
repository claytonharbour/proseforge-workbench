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
		{"@Aldric's :4m also clean at 60s", 60 * sec, 4 * min, 60 * sec, false, ""},

		// @Aldric swept this: :5m still holds a hole at N=3 even at 60s
		// sampling — peak 5m30s, excursion 30s < 60s.
		{"specific: :5m holes at N=3", 60 * sec, 5 * min, 60 * sec,
			true, "below the 3-failure peak"},

		// @Vance's prod: 30m leg, :90m, watchdog 2m. Peak for N=2 is 91m, so
		// the excursion is 60s against a 120s sampler — his coin flip.
		{"specific: prod :90m is a coin flip at N=2", 30 * min, 90 * min, 2 * min,
			true, "below the 2-failure peak"},

		// Just above that peak: deterministic, per his own recommendation.
		{"prod :91m is deterministic", 30 * min, 91 * min, 2 * min, false, ""},

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
		{60 * time.Second, 3 * time.Minute, 60 * time.Second},  // @Aldric demo, fixed
		{30 * time.Minute, 91 * time.Minute, 2 * time.Minute},  // @Vance prod, fixed
		{15 * time.Second, 46 * time.Second, 10 * time.Second}, // @Rowan's fast leg
	}
	for _, c := range sound {
		if w := SamplingWarning(c.observed, c.maxAge, c.loop); w != "" {
			t.Errorf("false positive on a sound config (observed=%s maxAge=%s loop=%s): %s",
				c.observed, c.maxAge, c.loop, w)
		}
	}
}
