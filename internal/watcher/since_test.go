package watcher

import "testing"

// The forms that reach --since in practice. The timestamp cases are the whole
// point: they are what an agent types when the help says "or RFC3339", and they
// used to be written straight into the cursor, producing a watcher that reports
// healthy and delivers nothing, forever (#357).
func TestLooksLikeMessageID(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"1787423366237-0", true},
		{"1787423366237-12", true},
		{"0-0", true},
		{"0", true},             // "from the beginning" sentinel
		{"1787423366237", true}, // bare millis; sequence defaults to 0

		{"2026-08-22T00:00:00Z", false}, // #357, the reported form
		{"2026-08-22", false},           // date only
		{"1787423366237-", false},       // trailing dash
		{"-0", false},                   // leading dash
		{"", false},
		{"latest", false},
		{"1787423366237-0a", false},
	} {
		if got := looksLikeMessageID(tc.in); got != tc.want {
			t.Errorf("looksLikeMessageID(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
