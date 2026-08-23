package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// The emit layer for `room watch --loop` (#366).
//
// 🛑 A COUNT IS UNACTIONABLE. "2 queued" still makes the reader go and look,
// which is the exact work a watcher exists to remove (@Tuner). The bodies are
// already in hand at queue time; not printing them means every delivery costs a
// second round trip through the queue file.
//
// ⚑ THE COST MODEL, because it is the opposite of the obvious one (@Aldric,
// #333). The expensive half of a delivery is THE WAKE — a model turn against
// full context. Once that has happened, another 1500 characters is rounding
// error, while a truncated body forces a follow-up read that costs a whole
// extra tool call AND its output. I inherited `cut -c1-400` from a publisher
// whose constraint was a terminal line, not a context window, and paid it for
// weeks: a 1611-char message arrived cut to 400 and I ran a Bash command to
// read the rest, treating that as normal. It is not normal. The truncation was
// charging more than it saved.
//
// So: a BUDGET, not per-message truncation. A single message arrives whole. A
// backlog drain is bounded in total. And whatever is withheld is NAMED —
// silent truncation reads as completeness, which is the same failure as a
// silent watcher wearing different clothes.

const (
	// emitBodyChars bounds ONE message. Above this a room post is a document,
	// and the pointer to it is worth more than its first screenful.
	emitBodyChars = 2000
	// emitTotalBudget bounds one emit, so a backlog drain cannot flood the
	// consumer's context in a single wake.
	emitTotalBudget = 8000
	// emitMaxMessages bounds the COUNT as well as the bytes: sixty short
	// messages fit the byte budget and are still sixty things to read.
	emitMaxMessages = 40
)

// emitQueued writes the bodies of what a tick queued.
//
// 🛑 It writes to w, which MUST be stdout. A Monitor raises events from stdout
// only, so an arrival notice on stderr reaches nobody — that is precisely how
// my watcher went deaf while looking healthy. The parameter exists so a test
// can assert on the file descriptor rather than on "output appeared": a test
// capturing 2>&1 passes whether the line went to stdout or stderr, and stderr
// is the bug.
func emitQueued(w io.Writer, res *watcher.WatchResult) {
	if res == nil || len(res.QueuedMessages) == 0 {
		return
	}

	msgs := res.QueuedMessages
	var b strings.Builder
	spent, shown := 0, 0
	heldChars := 0

	for _, m := range msgs {
		if shown >= emitMaxMessages || spent >= emitTotalBudget {
			break
		}
		body, held := clipBody(m.Content, emitBodyChars)
		heldChars += held

		fmt.Fprintf(&b, "🔔 [%s] %s  %s\n", m.Agent, m.Timestamp, m.ID)
		if m.Target != "" {
			fmt.Fprintf(&b, "%s\n", m.Target)
		}
		fmt.Fprintf(&b, "%s\n", body)
		if held > 0 {
			// ⚠️ NEVER trail off mid-word and leave the reader unable to tell
			// whether the rest mattered. Say the size and where to get it.
			fmt.Fprintf(&b, "… +%d chars withheld — full text: grep %s in the queue\n", held, m.ID)
		}
		b.WriteString("\n")
		spent += len(body)
		shown++
	}

	// 🛑 SAY WHAT WAS WITHHELD, with the count. A drain that hides messages is
	// indistinguishable from a quiet room — the failure this whole design
	// exists to prevent, reached from the inside.
	if withheld := len(msgs) - shown; withheld > 0 {
		fmt.Fprintf(&b, "… %d further message(s) NOT shown (budget %d chars / %d messages). "+
			"They ARE in the queue — nothing was dropped.\n", withheld, emitTotalBudget, emitMaxMessages)
	}

	// ⚑ DEGRADED GUARD (@Tuner). queued N but rendered fewer than N, with no
	// budget reason, means messages went missing between the append and here.
	// Loud, because silent partial delivery is the worst outcome available:
	// the queue says it arrived and the reader never saw it.
	if len(msgs) != res.Queued {
		fmt.Fprintf(&b, "🛑 DEGRADED — queued %d but only %d message(s) available to render. "+
			"Some arrivals are NOT in this emission. Read the queue directly.\n",
			res.Queued, len(msgs))
	}

	fmt.Fprint(w, b.String())
}

// clipBody bounds one message and reports how much it held back, so the caller
// can disclose it rather than truncating in silence.
func clipBody(s string, max int) (string, int) {
	if len(s) <= max {
		return s, 0
	}
	// Cut on a rune boundary; a half-written multi-byte character renders as a
	// replacement glyph and reads like corruption rather than truncation.
	cut := max
	for cut > 0 && !isRuneStart(s[cut]) {
		cut--
	}
	return s[:cut], len(s) - cut
}

func isRuneStart(b byte) bool { return b&0xC0 != 0x80 }
