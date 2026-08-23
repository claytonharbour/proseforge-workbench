package watcher

import "regexp"

// Filter is the watcher's copy of the inbox gate predicate (#322, #324).
//
// Duplicated rather than imported because internal/inbox is the CONSUMER side
// and internal/watcher is the PRODUCER side; making the producer depend on the
// consumer to decide what to store inverts the layering. The semantics must
// stay identical, so both are covered by the same acceptance arms.
//
// 🛑 Match and From are OR'd, never AND'd. A gate asks "is this FOR me, or FROM
// someone I cannot afford to miss" — two independent sufficient reasons.
//
// 🛑 ExcludeFrom is a VETO and is NOT part of that OR. It runs first and beats a
// positive match, which is the only way to express "messages naming me, but not
// the ones I wrote" (#336). Treating it as a third OR'd term would invert it:
// every message from an excluded sender would be KEPT.
type Filter struct {
	Match       []*regexp.Regexp // content + target
	From        []*regexp.Regexp // sender handle ONLY, never content
	ExcludeFrom []*regexp.Regexp // VETO on sender; beats any match/from hit
}

// Empty reports whether the gate keeps everything. ⚠️ A filter with ONLY vetoes
// is not empty — it drops those senders and keeps the rest.
func (f Filter) Empty() bool {
	return len(f.Match) == 0 && len(f.From) == 0 && len(f.ExcludeFrom) == 0
}

// Allows keeps a message if ANY match OR ANY from hits — the same rule the
// server applies (proseforge#1096), so the local net and the server agree.
func (f Filter) Allows(m Message) bool {
	if f.Empty() {
		return true
	}
	// Veto first: a match here drops the message no matter what else hits.
	for _, re := range f.ExcludeFrom {
		if re.MatchString(m.Agent) {
			return false
		}
	}
	// Vetoes alone (no positive terms) mean "keep everything except those".
	if len(f.Match) == 0 && len(f.From) == 0 {
		return true
	}
	for _, re := range f.Match {
		if re.MatchString(m.Content + " " + m.Target) {
			return true
		}
	}
	for _, re := range f.From {
		if re.MatchString(m.Agent) {
			return true
		}
	}
	return false
}
