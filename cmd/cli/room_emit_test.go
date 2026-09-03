package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

func msg(id, body string) watcher.Message {
	return watcher.Message{ID: id, Agent: "Clayton", Timestamp: "2026-08-22T12:00:00Z", Content: body}
}

func result(msgs ...watcher.Message) *watcher.WatchResult {
	return &watcher.WatchResult{Queued: len(msgs), QueuedMessages: msgs}
}

// 🛑 THE BODY, NOT A COUNT. "2 queued" still makes the reader go and look,
// which is the round trip the watcher exists to remove (@Tuner).
func TestEmitCarriesTheBodyNotJustACount(t *testing.T) {
	var out bytes.Buffer
	emitQueued(&out, result(msg("1-0", "the dev API is down, no listener")))

	got := out.String()
	if !strings.Contains(got, "the dev API is down, no listener") {
		t.Fatalf("emission does not contain the message body:\n%s", got)
	}
	if !strings.Contains(got, "Clayton") || !strings.Contains(got, "1-0") {
		t.Errorf("emission lacks sender or id, so the reader cannot find the original:\n%s", got)
	}
}

// ⚑ A SINGLE MESSAGE ARRIVES WHOLE. The wake is the expensive half; once it has
// happened, another 1500 characters is rounding error, while a truncated body
// forces a follow-up read costing a whole tool call and its output.
func TestOneLongMessageIsNotTruncatedBelowTheBudget(t *testing.T) {
	body := strings.Repeat("x", 1600) // the real case: a 1611-char message cut to 400
	var out bytes.Buffer
	emitQueued(&out, result(msg("1-0", body)))

	if !strings.Contains(out.String(), body) {
		t.Fatalf("a %d-char message was truncated; inherited `cut -c1-400` cost more than it saved", len(body))
	}
	if strings.Contains(out.String(), "withheld") {
		t.Error("claimed to withhold something from a message that fits")
	}
}

// 🛑 SILENT TRUNCATION READS AS COMPLETENESS — the same failure as a silent
// watcher in different clothes. What is held back must be NAMED, with a number.
func TestOversizeMessageDisclosesExactlyWhatItHeldBack(t *testing.T) {
	body := strings.Repeat("y", emitBodyChars+750)
	var out bytes.Buffer
	emitQueued(&out, result(msg("9-0", body)))

	got := out.String()
	if !strings.Contains(got, "withheld") {
		t.Fatalf("truncated in SILENCE — the reader cannot tell whether the rest mattered:\n%s", got)
	}
	if !strings.Contains(got, "750") {
		t.Errorf("disclosure does not say HOW MUCH was withheld (want 750):\n%s", got)
	}
	if !strings.Contains(got, "9-0") {
		t.Errorf("disclosure does not say where to get the rest:\n%s", got)
	}
}

// A backlog drain is bounded — and says how many it did not show, so the
// consumer never mistakes a capped emission for the whole room.
func TestBacklogDrainIsBoundedAndSaysWhatItWithheld(t *testing.T) {
	var msgs []watcher.Message
	for i := 0; i < 100; i++ {
		msgs = append(msgs, msg(fmt.Sprintf("%d-0", i), strings.Repeat("z", 300)))
	}
	var out bytes.Buffer
	emitQueued(&out, result(msgs...))

	got := out.String()
	if !strings.Contains(got, "NOT shown") {
		t.Fatalf("capped the emission WITHOUT saying so — a hidden backlog is indistinguishable from a quiet room:\n%s", got[:min(400, len(got))])
	}
	// ⚠️ and it must reassure that nothing was lost, or the reader reads a cap
	// as data loss and goes looking for a bug that is not there.
	if !strings.Contains(got, "nothing was dropped") {
		t.Error("cap notice does not say the messages are still in the queue")
	}
	if len(got) > emitTotalBudget*3 {
		t.Errorf("emission %d bytes — the budget did not bound it", len(got))
	}
}

// ⚑ DEGRADED (@Tuner): queued N, rendered fewer, and no budget reason. Silent
// partial delivery is the worst outcome available — the queue says it arrived
// and the reader never saw it.
func TestQueuedCountAboveRenderedCountIsLoud(t *testing.T) {
	var out bytes.Buffer
	emitQueued(&out, &watcher.WatchResult{
		Queued:         5, // the tick says five
		QueuedMessages: []watcher.Message{msg("1-0", "only one survived")},
	})

	got := out.String()
	if !strings.Contains(got, "DEGRADED") {
		t.Fatalf("queued=5 rendered=1 passed SILENTLY:\n%s", got)
	}
	if !strings.Contains(got, "5") || !strings.Contains(got, "1") {
		t.Errorf("degraded warning does not name queued vs rendered:\n%s", got)
	}
}

// Nothing queued must produce nothing at all — an emitter that speaks on empty
// ticks is a per-tick flood by another name.
func TestNothingQueuedEmitsNothing(t *testing.T) {
	var out bytes.Buffer
	emitQueued(&out, result())
	emitQueued(&out, nil)
	if out.Len() != 0 {
		t.Fatalf("emitted %q for an empty tick", out.String())
	}
}

// A multi-byte body must not be cut mid-rune: a replacement glyph reads like
// corruption rather than truncation, and sends the reader after the wrong bug.
func TestTruncationCutsOnARuneBoundary(t *testing.T) {
	body := strings.Repeat("é", emitBodyChars) // 2 bytes each
	var out bytes.Buffer
	emitQueued(&out, result(msg("1-0", body)))
	if strings.Contains(out.String(), "�") {
		t.Fatal("truncation produced a replacement character — reads as corruption, not as a cut")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
