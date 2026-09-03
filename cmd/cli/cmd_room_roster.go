package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// `pfw room roster` — what legs exist here, and is each one still polling? (#380)
//
// 🛑 THE GAP THIS FILLS IS "A LEG NOBODY NAMES". `room health` answers "is THIS
// state dir being polled", and to ask it you must name the dir — so a leg that
// fell off someone's list is a leg no check ever runs against. Six legs stopped
// inside a 22-minute window on 2026-08-22 and were found 13 HOURS later, by
// accident, during an unrelated investigation.
//
// ⛔ Every other instrument needs the thing that goes missing:
//
//	room health <dir>     you name the dir            ⇒ forgotten leg never named
//	a watchdog spec       the leg is in the spec      ⇒ arming and covering are
//	                                                    separate steps, and one
//	                                                    gets done without the other
//	the watcher's alarm   the watcher to be running   ⇒ the check lives INSIDE
//	                                                    the thing that died
//
// ⚑ So discovery is by DIRECTORY SCAN, never a manifest. A manifest is a list
// someone maintains and the failure mode here IS the unmaintained list.
const rosterUsage = `<dir>...`

func newRoomRosterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roster " + rosterUsage,
		Short: "What watcher legs exist under these directories, and is each still polling?",
		Long: `Scan directories for watcher state and report every leg found.

'room health' asks about a state dir you NAME. This finds the ones you would not
think to name — which is the only kind that goes unnoticed, because a leg nobody
remembers is a leg nobody checks.

Each leg's threshold is DERIVED from its own observed poll interval rather than
declared. That is the property that matters: a forgotten leg has no declaration
left, but its state dir still records how often it used to poll, so "was polling
every 60s, has not polled in 13h" is computable from disk alone.

A leg is reported STALE when its silence exceeds 10x its own observed interval —
far outside any failure-and-retry envelope, so it means stopped rather than
struggling. The age and interval are both printed; a leg at 3x is visible in the
table before it trips.

Exit: 0 when every leg is alive or deliberately retired, 3 when any is not.

⚠️ It reports. It does not restart anything. Legs are retired on purpose all the
time, and an auto-restart would fight every deliberate retirement.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runRoomRoster,
	}
	// ⚠️ The multiple is a flag because 10x is a judgement, not a measurement.
	// A leg polling every 30m is silent for 5 hours before it trips, which is
	// right for "definitely stopped" and wrong for "tell me early".
	cmd.Flags().Int("stale-multiple", 10,
		"Report STALE when silence exceeds this many times the leg's own observed interval.")
	cmd.Flags().Bool("all", false,
		"Include retired legs in the table (they are counted either way, never a fault).")
	return cmd
}

// rosterLeg is one discovered leg. Interval is OBSERVED, never configured —
// nothing on disk records what a leg was told to do, only what it did.
type rosterLeg struct {
	Name            string `json:"name"`
	StateDir        string `json:"state_dir"`
	State           string `json:"state"`
	AgeSeconds      int64  `json:"age_seconds,omitempty"`
	IntervalSeconds int64  `json:"interval_seconds,omitempty"`
	// Ratio is age/interval — the number that says "stopped" versus "slow".
	// Reported even when it does not trip, so a leg drifting toward the
	// threshold is visible before it crosses it.
	Ratio  float64 `json:"ratio,omitempty"`
	Detail string  `json:"detail,omitempty"`

	// What this leg watches, stamped at arm time. Empty on a leg armed before
	// the stamp existed — which is why it prints as a dash rather than as an
	// assertion that the leg watches nothing.
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
	BaseURL    string `json:"base_url,omitempty"`
	// Scope is the filter set the leg was armed with (#406). Empty on a leg
	// armed before it was recorded — printed as nothing rather than as "no
	// filters", which is a different and far more alarming claim.
	Scope string `json:"scope,omitempty"`
}

func runRoomRoster(cmd *cobra.Command, args []string) error {
	mult, _ := cmd.Flags().GetInt("stale-multiple")
	if mult < 2 {
		return fmt.Errorf("--stale-multiple must be at least 2: %d would report a leg "+
			"stale during its own normal poll gap", mult)
	}
	showAll, _ := cmd.Flags().GetBool("all")

	var legs []rosterLeg
	for _, root := range args {
		found, err := scanRoster(root, mult)
		if err != nil {
			return err
		}
		legs = append(legs, found...)
	}
	sort.Slice(legs, func(i, j int) bool { return legs[i].Name < legs[j].Name })

	if isJSON(cmd) {
		return printJSON(struct {
			PfwVersion string      `json:"pfw_version"`
			Legs       []rosterLeg `json:"legs"`
		}{Version, legs})
	}

	if len(legs) == 0 {
		// ⚠️ NOT an error and NOT "all clear". Nothing of ours is here, which
		// is what a wrong path looks like — say which, rather than printing an
		// empty table that reads as a healthy fleet.
		fmt.Printf("no watcher state found under %v — check the path; an empty roster is not a clean one\n", args)
		return nil
	}

	bad, unstamped := 0, 0
	// ⚠️ "AGE/IVL", not "RATIO" (@Sten). The bare word invited a comparison
	// between legs — 0.3x against 0.8x reads as a ranking when both are fine —
	// and said nothing about which direction is bad. Naming the two columns it
	// divides makes it self-describing, and the trip point goes in the header
	// so the reader has the threshold in front of them rather than in the docs.
	fmt.Printf("%-28s %10s %10s %8s  %s\n", "LEG", "LAST POLL", "INTERVAL", "AGE/IVL", "STATE")
	fmt.Printf("%-28s %10s %10s %8s  (stale above %dx)\n", "", "", "", "", mult)
	for _, l := range legs {
		if l.State == watcher.LiveRetired && !showAll {
			continue
		}
		if l.State != watcher.LiveAlive && l.State != watcher.LiveRetired {
			bad++
		}
		iv, ratio := "—", "—"
		if l.IntervalSeconds > 0 {
			iv = (time.Duration(l.IntervalSeconds) * time.Second).String()
			if l.Ratio > 0 {
				ratio = fmt.Sprintf("%.1fx", l.Ratio)
			}
		}
		age := "—"
		if l.AgeSeconds > 0 {
			age = (time.Duration(l.AgeSeconds) * time.Second).Round(time.Second).String()
		}
		fmt.Printf("%-28s %10s %10s %8s  %s\n", trunc(l.Name, 28), age, iv, ratio, l.State)
		// ⚑ WHAT it watches, on its own line and only when known. "Is this leg
		// polling" was never the whole question — @Sten had a live PROD watcher
		// for 25.5 hours and no artifact could say which room or which
		// environment, so "should it exist" was unanswerable from disk.
		if l.EntityID != "" {
			fmt.Printf("%-28s %s %s %s\n", "", l.EntityType, l.EntityID, l.BaseURL)
			// ⚑ SAY "PENDING", NEVER NOTHING (@Aldric). A room line with no scope
			// line means the leg was armed by a binary that predates the scope
			// record — which IS distinguishable, because a current binary always
			// writes the second line even when every filter is empty.
			//
			// ⛔ Silence made "armed before we recorded this" look identical to
			// "this leg has no filters", and the window is one full interval per
			// leg: a 5m leg closes it in 5 minutes, a 30m prod leg in 30. The rows
			// most likely to sit in it are the long-interval prod ones — the same
			// inversion as the subject stamp itself (#380).
			renderScope(os.Stdout, l)
		} else {
			unstamped++
		}
		// ⚠️ FIRST LINE ONLY. A retired leg's note is a paragraph — deliberately,
		// because it records what that leg covered — and printing it whole turned
		// a four-row table into forty lines. The full text is in `-o json` and in
		// `room health <dir>`, both of which are asked one leg at a time.
		if l.Detail != "" && l.State != watcher.LiveAlive {
			fmt.Printf("%-28s %s\n", "", trunc(firstLine(l.Detail), 96))
		}
	}
	// 🛑 SAY THAT THE INVENTORY IS INCOMPLETE (@Sten). A leg records its subject
	// at ARM, so adoption is inverted against the use case: 60s demo legs appear
	// within a minute of publish and 30m PROD legs — the rows an authorisation
	// review actually needs — can be half an hour behind.
	//
	// ⛔ Without this line the first person auditing prod reads a short list as a
	// complete one. That is the same silence-reads-as-clean shape the whole
	// command exists to remove, reintroduced by the feature meant to serve it.
	if unstamped > 0 {
		fmt.Printf("\n⚠️  %d leg(s) have not recorded what they watch. A leg stamps it at its\n"+
			"    next ARM, so a 30m leg can be up to 30m behind a publish. This list is\n"+
			"    NOT a complete inventory until that column is filled.\n", unstamped)
	}
	if bad > 0 {
		os.Exit(3)
	}
	return nil
}

// scanRoster walks one directory looking for watcher state.
//
// 🛑 IDENTIFY A LEG BY ITS CONTENTS, NEVER BY ITS DIRECTORY NAME (@Aldric).
// The first version matched the literal name `state`, and on a node with two
// live legs it returned an EMPTY TABLE and rc=0 — because the cutover recipe
// this repo documents tells people to arm a NEW state dir so the old one keeps
// running, and they had called theirs `state-v2`.
//
// ⛔ That is the fatal polarity for this command. A false POSITIVE is noisy and
// self-correcting; a false NEGATIVE from a tool whose entire purpose is "finds
// the legs you would not think to name" hands back a clean bill for exactly the
// case it exists to catch — and rc=0 on an empty table is indistinguishable
// from rc=0 on a healthy node.
//
// ⚑ A `.cursor` or a `health-poll` is what MAKES a directory a leg. The name is
// a convention, and conventions are the thing that drifts.
func scanRoster(root string, mult int) ([]rosterLeg, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", root, err)
	}
	if leg, ok := legAt(root, filepath.Base(root), mult); ok {
		return []rosterLeg{leg}, nil
	}

	var out []rosterLeg
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() || path == root {
			return nil //nolint:nilerr // an unreadable subtree is skipped, not fatal
		}
		// ⚠️ BOUNDED. A cron directory is shallow; an unbounded walk over a home
		// directory is a different command with a different cost.
		if rel, _ := filepath.Rel(root, path); strings.Count(rel, string(os.PathSeparator)) >= 3 {
			return filepath.SkipDir
		}
		if !holdsWatcherState(path) {
			return nil
		}
		leg, ok := legAt(path, legName(root, path), mult)
		if ok {
			out = append(out, leg)
		}
		// Never descend into a leg: its telemetry and lock files are not legs.
		return filepath.SkipDir
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", root, err)
	}
	return out, nil
}

// holdsWatcherState is the discovery predicate: does this directory contain the
// files a watcher leaves behind, whatever it is called?
func holdsWatcherState(dir string) bool {
	for _, f := range []string{".cursor", "health-poll", "health-arm"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}

// legName labels a leg by its path below the root, so two state dirs under one
// leg directory — which is what a cutover leaves behind — are distinguishable.
// The conventional `<leg>/state` keeps reading as `<leg>`.
func legName(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Base(path)
	}
	if base := filepath.Base(rel); base == "state" {
		if parent := filepath.Dir(rel); parent != "." {
			return parent
		}
	}
	return rel
}

// legAt reports whether a directory holds watcher state, and if so how it is
// doing. The threshold is derived from the leg's own observed interval.
func legAt(stateDir, name string, mult int) (rosterLeg, bool) {
	probe, err := watcher.CheckLiveness(stateDir, time.Hour)
	if err != nil || probe == nil {
		return rosterLeg{}, false
	}
	// 🛑 REJECT `unknown` OUTRIGHT, and do NOT gate on NoStateFound. That field
	// distinguishes "this dir exists and holds nothing of ours" from "this dir
	// cannot be read" — and a path that does not exist at all is the SECOND
	// case, so NoStateFound is FALSE for it.
	//
	// ⛔ Gating on NoStateFound therefore admitted every non-existent candidate
	// path as a leg. Measured: `roster` listed `watchdog-state` — a watchdog
	// directory with no watcher state anywhere in it — and exited 3 on it. A
	// discovery tool whose first output is a false alarm is worse than no
	// discovery tool, because the next real one is read as more of the same.
	//
	// ⚑ A leg with real state on disk never reads `unknown`; that word means
	// exactly "I found nothing here", by either route.
	if probe.State == watcher.LiveUnknown {
		return rosterLeg{}, false
	}

	leg := rosterLeg{
		Name: name, StateDir: stateDir,
		AgeSeconds: probe.AgeSeconds, IntervalSeconds: probe.IntervalSeconds,
		State: probe.State, Detail: probe.Detail,
	}
	leg.EntityType, leg.EntityID, leg.BaseURL, leg.Scope = watcher.Subject(stateDir)

	// ⛔ A leg with no observed interval cannot be judged, and saying so beats
	// picking a default. It has armed once, or just re-exec'd (#401) — either
	// way the number this check needs does not exist yet, and inventing one
	// would produce a verdict with nothing behind it.
	if probe.IntervalSeconds <= 0 {
		if probe.State == watcher.LiveAlive || probe.State == watcher.LiveRetired {
			return leg, true
		}
		// ⛔ APPEND, never replace. The first version assigned over Detail and
		// threw away CheckLiveness's actual diagnosis — so a leg silent for 67
		// hours reported only "no observed interval", which is a caveat about
		// the RATIO and says nothing about the leg. The measurement is the
		// finding; the caveat is a bound on one column of it.
		if leg.Detail != "" {
			leg.Detail += " "
		}
		leg.Detail += "(no observed interval — the ratio column cannot be computed)"
		return leg, true
	}

	iv := time.Duration(probe.IntervalSeconds) * time.Second
	leg.Ratio = float64(probe.AgeSeconds) / float64(probe.IntervalSeconds)

	// Re-ask with the DERIVED threshold. CheckLiveness already distinguishes
	// stale from failing from never; all this supplies is the number it could
	// not know.
	l, err := watcher.CheckLiveness(stateDir, time.Duration(mult)*iv)
	if err == nil && l != nil {
		leg.State, leg.Detail = l.State, l.Detail
	}
	return leg, true
}

// renderScope writes a leg's filter set, or says the record is pending.
// Extracted so the pending case is testable without capturing stdout.
func renderScope(w io.Writer, l rosterLeg) {
	if l.Scope != "" {
		fmt.Fprintf(w, "%-28s %s\n", "", l.Scope)
		return
	}
	fmt.Fprintf(w, "%-28s scope: pending — armed before scope was recorded; "+
		"adopts at its next arm\n", "")
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// firstLine keeps a table a table. Retired notes are paragraphs by design.
func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	return s
}
