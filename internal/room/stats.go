package room

import (
	"sort"
	"unicode/utf8"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// AuthorStat is one bench's volume in one room, over one window.
type AuthorStat struct {
	Agent    string `json:"agent"`
	Messages int    `json:"messages"`
	Chars    int    `json:"chars"`
	AvgChars int    `json:"avgChars"`
}

// DayStat is per-day volume across all authors, for --by-day.
type DayStat struct {
	Day      string `json:"day"`
	Messages int    `json:"messages"`
	Chars    int    `json:"chars"`
}

// RoomStats is the whole answer, window bound included.
//
// 🛑 Window IS PART OF THE RESULT, not context around it. Every number here is
// "within the messages actually read", and a room read is capped — so a total
// published without its window reads as the room's lifetime volume when it may be
// one afternoon. That mistake has been made in this room repeatedly: a 30-hour
// window compared against a 7-day one produced a "3%" that was really 6%.
type RoomStats struct {
	Authors  []AuthorStat `json:"authors"`
	Days     []DayStat    `json:"days,omitempty"`
	Messages int          `json:"messages"`
	Chars    int          `json:"chars"`

	// Earliest and Latest bound the window the numbers cover. Empty when no
	// messages were read.
	Earliest string `json:"earliest,omitempty"`
	Latest   string `json:"latest,omitempty"`
}

// Stats aggregates messages by author.
//
// ⚠️ Chars counts RUNES, not bytes. These messages are full of box-drawing and
// emoji, where a byte count roughly triples the apparent length of the benches who
// use them most — which would rank formatting style rather than volume.
//
// ⚑ Counting is done here rather than in the CLI because the MCP surface needs the
// identical answer; a second implementation is a second set of rounding decisions.
func Stats(messages []api.RoomMessage, byDay bool) *RoomStats {
	out := &RoomStats{}
	byAuthor := map[string]*AuthorStat{}
	byDate := map[string]*DayStat{}

	for _, m := range messages {
		n := utf8.RuneCountInString(m.Content)
		out.Messages++
		out.Chars += n

		a, ok := byAuthor[m.Agent]
		if !ok {
			a = &AuthorStat{Agent: m.Agent}
			byAuthor[m.Agent] = a
		}
		a.Messages++
		a.Chars += n

		if m.Timestamp != "" {
			if out.Earliest == "" || m.Timestamp < out.Earliest {
				out.Earliest = m.Timestamp
			}
			if m.Timestamp > out.Latest {
				out.Latest = m.Timestamp
			}
		}
		if byDay && len(m.Timestamp) >= 10 {
			d := m.Timestamp[:10]
			s, ok := byDate[d]
			if !ok {
				s = &DayStat{Day: d}
				byDate[d] = s
			}
			s.Messages++
			s.Chars += n
		}
	}

	for _, a := range byAuthor {
		if a.Messages > 0 {
			a.AvgChars = a.Chars / a.Messages
		}
		out.Authors = append(out.Authors, *a)
	}
	// Messages descending, then agent, so equal counts do not reorder between runs —
	// a table that shuffles on every invocation cannot be diffed against yesterday's.
	sort.Slice(out.Authors, func(i, j int) bool {
		if out.Authors[i].Messages != out.Authors[j].Messages {
			return out.Authors[i].Messages > out.Authors[j].Messages
		}
		return out.Authors[i].Agent < out.Authors[j].Agent
	})

	for _, d := range byDate {
		out.Days = append(out.Days, *d)
	}
	sort.Slice(out.Days, func(i, j int) bool { return out.Days[i].Day < out.Days[j].Day })
	return out
}
