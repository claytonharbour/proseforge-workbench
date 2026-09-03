package room

import "testing"

import "github.com/claytonharbour/proseforge-workbench/internal/api"

func msg(agent, content, ts string) api.RoomMessage {
	return api.RoomMessage{Agent: agent, Content: content, Timestamp: ts}
}

// 🛑 CHARS MUST BE RUNES, NOT BYTES (#458).
//
// These rooms are written in box-drawing and emoji. "⛔" is 3 bytes and one rune;
// a byte count inflates the benches who format heavily by ~3x and would rank
// PRESENTATION STYLE while claiming to rank volume — a wrong answer that looks
// entirely plausible, which is the kind this fleet keeps publishing.
func TestStatsCountsRunesNotBytes(t *testing.T) {
	// 10 runes, 26 bytes.
	const heavy = "⛔⚠️✅⇒⚑🛑📌"
	got := Stats([]api.RoomMessage{msg("Tate", heavy, "2026-08-29T10:00:00Z")}, false)

	if len(heavy) == got.Chars {
		t.Fatalf("Chars = %d = len(bytes); it is counting bytes, so heavy formatters "+
			"rank above heavy writers", got.Chars)
	}
	if want := []rune(heavy); got.Chars != len(want) {
		t.Errorf("Chars = %d, want %d runes", got.Chars, len(want))
	}
}

func TestStatsGroupsRanksAndBoundsTheWindow(t *testing.T) {
	in := []api.RoomMessage{
		msg("Sten", "aaaaaaaa", "2026-08-29T10:00:00Z"),
		msg("Tate", "bb", "2026-08-28T09:00:00Z"),
		msg("Sten", "cc", "2026-08-29T11:00:00Z"),
		msg("Tate", "dd", "2026-08-30T08:00:00Z"),
		msg("Tate", "ee", "2026-08-30T09:00:00Z"),
	}
	got := Stats(in, true)

	if got.Messages != 5 || got.Chars != 16 {
		t.Errorf("totals = %d msgs / %d chars, want 5 / 16", got.Messages, got.Chars)
	}
	// Tate has MORE MESSAGES (3 v 2); Sten has MORE CHARACTERS (10 v 6). Ranking is
	// by messages, so
	// Tate leads — asserted because "who talks most" and "who writes most" are
	// different questions and the table must not silently answer the other one.
	if got.Authors[0].Agent != "Tate" || got.Authors[0].Messages != 3 {
		t.Errorf("rank[0] = %+v, want Tate with 3", got.Authors[0])
	}
	if got.Authors[1].Agent != "Sten" || got.Authors[1].Chars != 10 {
		t.Errorf("rank[1] = %+v, want Sten with 10 chars", got.Authors[1])
	}
	if got.Authors[0].AvgChars != 2 {
		t.Errorf("Tate avg = %d, want 2", got.Authors[0].AvgChars)
	}

	// ⛔ THE WINDOW IS PART OF THE ANSWER. A total without its bound reads as the
	// room's lifetime volume; this fleet has already published one number that was
	// wrong by 2x for exactly that reason.
	if got.Earliest != "2026-08-28T09:00:00Z" || got.Latest != "2026-08-30T09:00:00Z" {
		t.Errorf("window = %s .. %s, want the true min and max", got.Earliest, got.Latest)
	}

	if len(got.Days) != 3 {
		t.Fatalf("days = %d, want 3", len(got.Days))
	}
	if got.Days[0].Day != "2026-08-28" || got.Days[2].Day != "2026-08-30" {
		t.Errorf("days not in ascending order: %+v", got.Days)
	}
	if got.Days[2].Messages != 2 {
		t.Errorf("2026-08-30 = %d messages, want 2", got.Days[2].Messages)
	}
}

// A table that reorders between identical runs cannot be diffed against yesterday's.
func TestStatsTiesSortDeterministically(t *testing.T) {
	in := []api.RoomMessage{msg("Zed", "x", "t"), msg("Ada", "y", "t"), msg("Mid", "z", "t")}
	first := Stats(in, false)
	for i := 0; i < 20; i++ {
		got := Stats(in, false)
		for j := range got.Authors {
			if got.Authors[j].Agent != first.Authors[j].Agent {
				t.Fatalf("run %d reordered equal counts: %v vs %v", i, got.Authors, first.Authors)
			}
		}
	}
	if first.Authors[0].Agent != "Ada" {
		t.Errorf("equal counts should break ties by agent name, got %v", first.Authors)
	}
}

func TestStatsHandlesAnEmptyWindow(t *testing.T) {
	got := Stats(nil, true)
	if got.Messages != 0 || got.Chars != 0 || len(got.Authors) != 0 {
		t.Errorf("empty input produced %+v", got)
	}
	// Empty, not a fabricated zero-time bound — "no data" must not render as a window.
	if got.Earliest != "" || got.Latest != "" {
		t.Errorf("empty window reported bounds %q..%q", got.Earliest, got.Latest)
	}
}
