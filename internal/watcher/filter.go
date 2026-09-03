package watcher

import (
	"regexp"
	"strings"
)

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
func (f Filter) Allows(m Message) bool { return f.allows(m, true) }

// AllowsRaw is Allows WITHOUT the #417 code elision — i.e. exactly the semantics
// the SERVER applies (proseforge#1096).
//
// 🛑 It exists so ServerMissed keeps meaning "the backend ignored our filter".
// Without it, every quoted routing pattern we now drop locally would be counted
// as a server fail-open and print "this backend looks older than proseforge#1096"
// — turning a noise fix into a fleet-wide false alarm about the backend.
func (f Filter) AllowsRaw(m Message) bool { return f.allows(m, false) }

func (f Filter) allows(m Message, elide bool) bool {
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
	// ⚑ #417 (@Wayland): match the message's ADDRESS, not the routing patterns it
	// QUOTES. A message explaining someone's gate contains every token that gate
	// names, so the topic that requires spelling a filter out is the topic that
	// broadcasts to everyone the filter names. Measured on their node: two
	// messages in 8 seconds woke a bench neither mentioned.
	for _, re := range f.Match {
		content := m.Content
		if elide {
			content = elideCode(content)
		}
		if re.MatchString(content + " " + m.Target) {
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

// elideCode blanks fenced blocks and inline code spans so a quoted routing
// pattern is not read as an address (#417).
//
// 🛑 AN UNTERMINATED FENCE OR BACKTICK IS LEFT INTACT, deliberately. Eliding to
// end-of-message on a stray backtick would silence a real address, and
// @Wayland's bound on this whole fix is the one that matters: failing toward
// MISSED messages is worse than the noise it removes. So every elision here
// requires a matched pair; anything unbalanced falls through and still wakes you.
//
// ⚠️ A message addressed in prose AND quoting a pattern still matches, because
// only the quoted span is blanked. That case is covered by a test, because it is
// the one that would turn this fix into a worse bug than the one it fixes.
func elideCode(s string) string {
	s = elidePairs(s, "```")
	return elidePairs(s, "`")
}

func elidePairs(s, delim string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, delim)
		if i < 0 {
			break
		}
		j := strings.Index(s[i+len(delim):], delim)
		if j < 0 {
			break // unterminated — leave the remainder INTACT, see above
		}
		b.WriteString(s[:i])
		b.WriteString(" ")
		s = s[i+len(delim)+j+len(delim):]
	}
	b.WriteString(s)
	return b.String()
}
