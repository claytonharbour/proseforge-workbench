package watcher

import "testing"

// 🛑 The test #336 exists for: exclusion is a VETO, not a third OR'd term.
//
// If ExcludeFrom were OR'd with Match/From, a message from an excluded sender
// would be KEPT rather than dropped — the exact inversion, and it would look
// like a working filter in every test that only checks positive matches.
func TestExcludeFromIsAVetoNotAnOrTerm(t *testing.T) {
	f, err := compileFilter([]string{"@Tate"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := compileExcludeFrom([]string{"Tate"})
	if err != nil {
		t.Fatal(err)
	}
	f.ExcludeFrom = ex

	// My own post that names me: the match HITS and the veto must still win.
	mine := Message{Agent: "Tate", Content: "@Tate reminder to self"}
	if f.Allows(mine) {
		t.Error("veto lost to a positive match — exclusion is being treated as an OR term")
	}

	// Someone else naming me: kept.
	theirs := Message{Agent: "Sten", Content: "@Tate look at this"}
	if !f.Allows(theirs) {
		t.Error("dropped a message that matched and was not vetoed")
	}
}

// A filter with ONLY vetoes keeps everything else — it must not be treated as
// Empty (which would keep the vetoed sender too).
func TestVetoOnlyFilterIsNotEmpty(t *testing.T) {
	var f Filter
	ex, err := compileExcludeFrom([]string{"noisybot"})
	if err != nil {
		t.Fatal(err)
	}
	f.ExcludeFrom = ex

	if f.Empty() {
		t.Fatal("a veto-only filter reported Empty — the vetoed sender would be kept")
	}
	if f.Allows(Message{Agent: "noisybot", Content: "spam"}) {
		t.Error("vetoed sender was allowed")
	}
	if !f.Allows(Message{Agent: "Sten", Content: "anything"}) {
		t.Error("veto-only filter dropped an unrelated sender")
	}
}

// Case-insensitive by default, consistent with --match/--from.
func TestExcludeFromIsCaseInsensitiveByDefault(t *testing.T) {
	ex, err := compileExcludeFrom([]string{"tate"})
	if err != nil {
		t.Fatal(err)
	}
	f := Filter{ExcludeFrom: ex}
	if f.Allows(Message{Agent: "Tate", Content: "x"}) {
		t.Error("lowercase pattern did not veto a capitalised sender")
	}
}
