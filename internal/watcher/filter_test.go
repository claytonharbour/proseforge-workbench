package watcher

import (
	"regexp"
	"testing"
)

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

// #417 (@Wayland): a message ABOUT routing patterns matched the patterns it
// quoted, waking every bench those patterns name. Measured on their node:
// @Gordon's 1787663670329-0 contained "@Team" and "@All" ONLY inside the fenced
// pattern `@Gordon|@Team|@All|@Leads`, and "@Wayland" zero times — yet it woke
// @Wayland's leg.
//
// ⚑ The topic that REQUIRES spelling your filter out is exactly the topic that
// broadcasts to everyone the filter names, which makes it structural.
func TestMatchIgnoresRoutingPatternsQuotedAsCode(t *testing.T) {
	f := Filter{Match: mustRes(t, `@Wayland|@Team|@All`)}

	noise := Message{
		Agent:   "Gordon",
		Content: "Here is my gate, for the record:\n```\nmatch=@Gordon|@Team|@All|@Leads\n```\nNothing else to report.",
	}
	if f.Allows(noise) {
		t.Errorf("#417: woke on a message whose only match tokens are inside a fenced " +
			"routing pattern — the message names neither @Wayland nor any real address")
	}

	// 🛑 CONTROL, and it is the one @Wayland flagged as the dangerous direction:
	// a message genuinely addressed to a group AND quoting a pattern must STILL
	// wake. Silencing this would fail toward MISSED messages, worse than noise.
	addressed := Message{
		Agent:   "Gordon",
		Content: "@Team please check your gates. Mine is:\n```\nmatch=@Gordon|@Team|@All\n```",
	}
	if !f.Allows(addressed) {
		t.Errorf("#417 fix over-reached: silenced a message actually addressed to @Team " +
			"in prose. Failing toward missed messages is worse than the noise it fixes")
	}

	// Control: inline backticks, same rule.
	inline := Message{Agent: "Aldric", Content: "my gate is `@Aldric|@Team|@All` and nothing else"}
	if f.Allows(inline) {
		t.Errorf("#417: woke on a routing pattern in inline backticks")
	}

	// Control: ordinary prose mention is untouched.
	plain := Message{Agent: "Sten", Content: "@Wayland can you check the prod leg?"}
	if !f.Allows(plain) {
		t.Errorf("regression: an ordinary prose mention must still match")
	}
}

func mustRes(t *testing.T, pats ...string) []*regexp.Regexp {
	t.Helper()
	var out []*regexp.Regexp
	for _, p := range pats {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			t.Fatalf("compile %q: %v", p, err)
		}
		out = append(out, re)
	}
	return out
}

// 🛑 The unterminated-delimiter rule, which is the safety property of the #417
// fix rather than a detail. A stray backtick must NOT swallow the rest of the
// message: that would convert a noise bug into a silent-miss bug, which is
// strictly worse. Every elision requires a matched pair.
func TestUnterminatedCodeDelimiterStillWakesYou(t *testing.T) {
	f := Filter{Match: mustRes(t, `@Wayland|@Team|@All`)}

	for name, content := range map[string]string{
		"stray inline backtick": "the flag is `--match and @Wayland you should see this",
		"unclosed fence":        "```\nmatch=@Gordon|@Leads\n\n@Team this is a real address",
		"odd backtick count":    "`a` `b @Team",
	} {
		if !f.Allows(Message{Agent: "Gordon", Content: content}) {
			t.Errorf("%s: an unbalanced delimiter silenced a real address — "+
				"the fix must fail toward WAKING, never toward missing: %q", name, content)
		}
	}
}

// Control on the control: a *balanced* pair really does elide, so the test above
// is not passing merely because elision never happens.
func TestBalancedCodeReallyElides(t *testing.T) {
	f := Filter{Match: mustRes(t, `@Team`)}
	if f.Allows(Message{Agent: "Gordon", Content: "see `@Team` here"}) {
		t.Error("balanced inline code did not elide — the suppression arm is inert, " +
			"so the unterminated-delimiter test proves nothing")
	}
}

// AllowsRaw must mirror the SERVER's semantics (no #417 elision), because the
// watcher uses the difference between the two to tell an expected divergence
// apart from a genuine backend fail-open.
//
// ⚠️ SCOPE, stated because I got this wrong: this test covers the FILTER only.
// It does NOT guard the counter split in watch.go — I wrote it believing it did,
// then deliberately merged the counters and watched it stay GREEN. The guard
// that actually fires is TestQuotedOnlyDropsAreNotCountedAsServerFailOpen in
// watch_test.go. A test whose comment overstates its reach is worse than no
// test, because it stops anyone writing the real one.
func TestQuotedCodeDropIsNotAServerFailOpen(t *testing.T) {
	f := Filter{Match: mustRes(t, `@Wayland|@Team`)}

	quoted := Message{Agent: "Gordon", Content: "my gate:\n```\n@Gordon|@Team\n```"}
	if f.Allows(quoted) {
		t.Fatal("setup: expected the quoted-pattern message to be dropped locally")
	}
	if !f.AllowsRaw(quoted) {
		t.Error("AllowsRaw must mirror the SERVER (no elision) — if it drops this too, " +
			"the drop gets miscounted as a backend fail-open and warns falsely")
	}

	// A message the server really should have dropped: no match anywhere, quoted
	// or otherwise. This one IS a genuine fail-open and must stay countable.
	unrelated := Message{Agent: "Gordon", Content: "nothing here concerns anyone"}
	if f.AllowsRaw(unrelated) {
		t.Error("AllowsRaw kept a message with no match at all — ServerMissed would " +
			"then never fire and a real stale backend would go unreported")
	}
}
