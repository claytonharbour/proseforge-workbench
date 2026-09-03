package watcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// fakeRoom is a room that can also fail on demand — the failure arms matter more
// than the happy path here.
type fakeRoom struct {
	msgs         []Message
	newest       string
	email        string
	readErr      error
	reads        int
	ignoreFilter bool
	lastMatch    []string
	lastFrom     []string
	whoErr       error
	lastExclude  []string
}

func (f *fakeRoom) ReadSince(_ context.Context, _, _, since string, limit int,
	match, from, excludeFrom []string) ([]Message, string, error) {
	f.reads++
	f.lastMatch, f.lastFrom, f.lastExclude = match, from, excludeFrom
	if f.readErr != nil {
		return nil, "", f.readErr
	}
	var out []Message
	seen := since == ""
	for _, m := range f.msgs {
		if !seen {
			if m.ID == since {
				seen = true
			}
			continue
		}
		out = append(out, m)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	// The scan boundary is the furthest id EXAMINED, which is what the room API
	// returns and what a server-side filter makes differ from the last message.
	last := ""
	if len(out) > 0 {
		last = out[len(out)-1].ID
	}
	// Mirror the real server (proseforge#1096): match on Content+Target, from on
	// Agent, everything OR'd. ignoreFilter simulates a pre-#1096 backend, which
	// drops unknown query params and returns everything — a silent fail-open.
	if (len(match) > 0 || len(from) > 0) && !f.ignoreFilter {
		kept := out[:0]
		for _, m := range out {
			hit := false
			for _, p := range match {
				if regexp.MustCompile("(?i)" + p).MatchString(m.Content + " " + m.Target) {
					hit = true
				}
			}
			for _, p := range from {
				if regexp.MustCompile("(?i)" + p).MatchString(m.Agent) {
					hit = true
				}
			}
			if hit {
				kept = append(kept, m)
			}
		}
		out = kept
	}
	return out, last, nil
}

func (f *fakeRoom) Newest(context.Context, string, string) (string, error) {
	return f.newest, nil
}

func (f *fakeRoom) WhoAmI(context.Context) (string, error) {
	if f.whoErr != nil {
		return "", f.whoErr
	}
	return f.email, nil
}

func msgs(n int) []Message {
	out := make([]Message, n)
	for i := range out {
		out[i] = Message{
			ID:        fmt.Sprintf("%d-0", 100+i),
			Agent:     "Clayton",
			Content:   fmt.Sprintf("message %d", i),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}
	return out
}

func opts(t *testing.T) (WatchOptions, string, string) {
	t.Helper()
	dir := t.TempDir()
	q := filepath.Join(dir, "inbox.jsonl")
	s := filepath.Join(dir, "state")
	return WatchOptions{
		EntityType: "story", EntityID: "e1",
		QueuePath: q, StateDir: s,
		ExpectEmail: "agent@example.com",
	}, q, s
}

func lines(t *testing.T, path string) int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatal(err)
	}
	n := 0
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
}

// A first run baselines and queues NOTHING — and must leave a marker saying so,
// or a new watcher is indistinguishable from one whose queue was lost (#311).
func TestFirstRunBaselinesAndLeavesAMarker(t *testing.T) {
	o, q, s := opts(t)
	room := &fakeRoom{msgs: msgs(5), newest: "104-0", email: o.ExpectEmail}

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Baselined || res.Queued != 0 {
		t.Fatalf("first run: baselined=%v queued=%d, want true/0", res.Baselined, res.Queued)
	}
	if lines(t, q) != 0 {
		t.Error("baseline queued prior history")
	}
	marker := readTrimmed(filepath.Join(s, fileBaseline))
	if !strings.Contains(marker, "104-0") {
		t.Errorf("baseline marker must record the id it started from: %q", marker)
	}
	if readTrimmed(filepath.Join(s, fileCursor)) != "104-0" {
		t.Error("cursor was not set to newest")
	}
}

func TestSecondRunQueuesOnlyWhatIsNew(t *testing.T) {
	o, q, s := opts(t)
	all := msgs(6)
	room := &fakeRoom{msgs: all[:3], newest: "102-0", email: o.ExpectEmail}

	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all // three more arrive

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Queued != 3 {
		t.Fatalf("queued %d, want 3", res.Queued)
	}
	if lines(t, q) != 3 {
		t.Errorf("queue has %d lines, want 3", lines(t, q))
	}
	if readTrimmed(filepath.Join(s, fileCursor)) != "105-0" {
		t.Error("cursor did not advance to the last queued id")
	}
}

// 🛑 A failed read must not advance the cursor and must not stamp health. Advance
// on failure and the room considers messages delivered that nothing holds — the
// one failure in this system with no recovery path.
func TestFailedReadAdvancesNothing(t *testing.T) {
	o, _, s := opts(t)
	room := &fakeRoom{msgs: msgs(3), newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	before := readTrimmed(filepath.Join(s, fileCursor))
	stampBefore := readTrimmed(filepath.Join(s, fileHealthOK))

	room.readErr = errors.New("connection refused")
	if _, err := Watch(context.Background(), room, o); err == nil {
		t.Fatal("a failed read must return an error")
	}
	if got := readTrimmed(filepath.Join(s, fileCursor)); got != before {
		t.Errorf("cursor moved on a failed read: %q -> %q", before, got)
	}
	if got := readTrimmed(filepath.Join(s, fileHealthOK)); got != stampBefore {
		t.Error("a failed read stamped health OK — it would look like a quiet tick")
	}
}

// Replaying the same batch must be safe: a crash between the append and the
// cursor write redelivers, and the retry cannot duplicate.
func TestRedeliveredBatchDoesNotDuplicate(t *testing.T) {
	o, q, s := opts(t)
	all := msgs(4)
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	// Simulate the crash: rewind the cursor, so the same batch comes again.
	if err := writeFileAtomic(filepath.Join(s, fileCursor), "100-0"); err != nil {
		t.Fatal(err)
	}
	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Queued != 0 || res.Duplicates != 3 {
		t.Errorf("replay queued=%d dupes=%d, want 0/3", res.Queued, res.Duplicates)
	}
	if lines(t, q) != 3 {
		t.Errorf("queue has %d lines after replay, want 3", lines(t, q))
	}
}

// Shared mode: whoami is blind to the collision, so a handle is mandatory.
func TestSharedModeRefusesWithoutAHandle(t *testing.T) {
	o, _, _ := opts(t)
	o.ExpectEmail, o.Shared = "", true
	room := &fakeRoom{newest: "1-0"}

	_, err := Watch(context.Background(), room, o)
	if !errors.Is(err, ErrIdentityGuard) {
		t.Fatalf("got %v, want ErrIdentityGuard", err)
	}
}

func TestIndividualModeRefusesOnTheWrongAccount(t *testing.T) {
	o, _, _ := opts(t)
	room := &fakeRoom{newest: "1-0", email: "someone-else@example.com"}

	_, err := Watch(context.Background(), room, o)
	if !errors.Is(err, ErrIdentityGuard) {
		t.Fatalf("got %v, want ErrIdentityGuard", err)
	}
	if !strings.Contains(err.Error(), "someone-else@example.com") {
		t.Errorf("the error must name who it actually is: %v", err)
	}
}

// Neither mode is optional — they check different things.
func TestNoIdentityModeRefuses(t *testing.T) {
	o, _, _ := opts(t)
	o.ExpectEmail = ""
	room := &fakeRoom{newest: "1-0"}

	if _, err := Watch(context.Background(), room, o); !errors.Is(err, ErrIdentityGuard) {
		t.Fatalf("got %v, want ErrIdentityGuard", err)
	}
}

func TestConcurrentInstanceIsLockedOut(t *testing.T) {
	o, _, s := opts(t)
	room := &fakeRoom{msgs: msgs(2), newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s, fileLock), []byte("99999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Watch(context.Background(), room, o); !errors.Is(err, ErrLocked) {
		t.Fatalf("got %v, want ErrLocked", err)
	}
}

// A process killed mid-poll must not stop all future polling — that is the
// instrument disabling the thing it measures.
func TestStaleLockIsReclaimed(t *testing.T) {
	o, _, s := opts(t)
	room := &fakeRoom{msgs: msgs(2), newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}

	lock := filepath.Join(s, fileLock)
	if err := os.WriteFile(lock, []byte("99999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * staleLockAge)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatalf("a stale lock wedged the watcher: %v", err)
	}
}

// 🛑 The cursor follows the SCAN BOUNDARY, never the last kept message.
//
// With a server-side filter the two differ by construction. Advance to the last
// survivor and every dropped message behind it is re-read forever; a tick that
// filtered out its whole batch would never advance at all. Verified live: an
// unfiltered read and a server-filtered read of the same window landed on the
// identical cursor while fetching 19 and 8 messages respectively.
func TestCursorFollowsScanBoundaryNotLastKept(t *testing.T) {
	o, _, s := opts(t)
	all := msgs(6)
	all[5].Content = "@Tate only the last one matches"
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	o.Match = []string{"@Tate"}
	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(room.lastMatch) != 1 || room.lastMatch[0] != "@Tate" {
		t.Errorf("--match should be pushed server-side, server got %v", room.lastMatch)
	}
	if res.Kept != 1 {
		t.Fatalf("kept %d, want 1", res.Kept)
	}
	if got := readTrimmed(filepath.Join(s, fileCursor)); got != "105-0" {
		t.Errorf("cursor = %s, want the scan boundary 105-0", got)
	}
}

// A batch filtered down to nothing must still advance, or a busy room whose
// traffic never matches wedges the watcher permanently.
func TestFullyFilteredBatchStillAdvances(t *testing.T) {
	o, q, s := opts(t)
	all := msgs(4)
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	o.Match = []string{"nothing-matches-this"}
	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kept != 0 || res.Queued != 0 {
		t.Fatalf("kept %d queued %d, want 0/0", res.Kept, res.Queued)
	}
	if got := readTrimmed(filepath.Join(s, fileCursor)); got == "100-0" {
		t.Fatal("cursor did not advance past a fully-filtered batch — the watcher is wedged")
	}
	if lines(t, q) != 0 {
		t.Error("filtered-out messages were queued")
	}
}

// 🛑 BOTH filters now go server-side, and the OR still holds.
//
// This test previously asserted the opposite: before proseforge#1096 the server
// had no sender filter and its match was DESTRUCTIVE, so a --from had to force
// everything local — pushing match down would have discarded sender-only hits
// before the client could OR them back. Gordon made both repeatable and OR'd
// them server-side, so that constraint is gone.
//
// Verified live: adding --from widened the result from 8 to 13, and all 5
// extras were from Clayton with no @Tate mention at all.
func TestBothFiltersGoServerSideAndTheOrSurvives(t *testing.T) {
	o, _, _ := opts(t)
	all := msgs(4)
	// msgs() makes everything "Clayton", which would make --from match all of
	// them and hide whether the OR did any work. Give the batch a realistic
	// mix: two irrelevant, one sender-only, one that matches nothing.
	all[1].Agent, all[2].Agent = "Vance", "Aldric"
	all[3].Agent = "Clayton Harbour" // sender-only: no @Tate anywhere in it
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	o.Match = []string{"@Tate"}
	o.From = []string{"Clayton"}

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(room.lastMatch) != 1 || len(room.lastFrom) != 1 {
		t.Errorf("both filters must reach the server: match=%v from=%v",
			room.lastMatch, room.lastFrom)
	}
	if res.Kept != 1 {
		t.Fatalf("kept %d, want the 1 sender-only match", res.Kept)
	}

	// Prove the OR did the work: without --from, that message is invisible.
	o2, _, _ := opts(t)
	o2.Match = []string{"@Tate"}
	room2 := &fakeRoom{msgs: all[:1], newest: "100-0", email: o2.ExpectEmail}
	if _, err := Watch(context.Background(), room2, o2); err != nil {
		t.Fatal(err)
	}
	room2.msgs = all
	res2, err := Watch(context.Background(), room2, o2)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Kept != 0 {
		t.Errorf("--match alone kept %d; the sender-only message should be invisible to it", res2.Kept)
	}
}

// ⚑ The local pass is a SAFETY NET against an older backend, not redundancy.
// A server without proseforge#1096 ignores unknown query params and returns
// EVERYTHING — a silent fail-open. ServerMissed counts what the net caught.
func TestLocalNetCatchesABackendThatIgnoresTheFilter(t *testing.T) {
	o, _, _ := opts(t)
	all := msgs(4)
	all[1].Content = "@Tate this one is mine"
	all[1].Agent, all[2].Agent, all[3].Agent = "Sten", "Vance", "Aldric"
	// ignoreFilter makes the fake behave like a pre-#1096 backend.
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail, ignoreFilter: true}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	o.Match = []string{"@Tate"}
	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kept != 1 {
		t.Fatalf("kept %d, want 1 — the local net must catch what the server ignored", res.Kept)
	}
	if res.ServerMissed != 2 {
		t.Errorf("ServerMissed = %d, want 2 — a silent fail-open must be COUNTED, not just fixed",
			res.ServerMissed)
	}
}

// 🛑 A watcher must never wake you with your own post (@Sten, #328).
//
// You sign with your handle and address others by theirs, so your own messages
// match your own gate. Measured: 8 of my 64 posts matched my own filter, and
// for anyone gating on a bare handle rather than "@handle" it is every post.
func TestOwnPostsAreSkipped(t *testing.T) {
	o, _, _ := opts(t)
	all := msgs(4)
	all[1].Agent, all[2].Agent = "Sten", "Gordon"
	all[3].Agent = "Agent" // derived from agent@example.com
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.SelfSkipped != 1 {
		t.Errorf("SelfSkipped = %d, want 1", res.SelfSkipped)
	}
	if res.Kept != 2 {
		t.Fatalf("kept %d, want 2 (Sten + Gordon, not myself)", res.Kept)
	}
}

// Self-exclusion must survive a SERVER-side match, or the fix silently stops
// applying for exactly the configuration that uses it most.
func TestOwnPostsSkippedEvenWithServerSideMatch(t *testing.T) {
	o, _, _ := opts(t)
	all := msgs(3)
	all[1].Agent = "Sten"
	all[2].Agent, all[2].Content = "Agent", "@agent quoting myself"
	all[1].Content = "@agent asking you something"
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all

	o.Match = []string{"@agent"} // goes server-side
	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(room.lastMatch) == 0 {
		t.Fatal("this arm needs the server-side path to be exercised")
	}
	if res.SelfSkipped != 1 || res.Kept != 1 {
		t.Errorf("self_skipped=%d kept=%d, want 1/1 — exclusion must apply after a server match",
			res.SelfSkipped, res.Kept)
	}
}

// The escape hatch has to work, or debugging your own posts becomes impossible.
func TestIncludeSelfOptsBackIn(t *testing.T) {
	o, _, _ := opts(t)
	all := msgs(2)
	all[1].Agent = "Tate"
	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all
	o.IncludeSelf = true

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.SelfSkipped != 0 || res.Kept != 1 {
		t.Errorf("self_skipped=%d kept=%d, want 0/1", res.SelfSkipped, res.Kept)
	}
}

// 🛑 Three conditions, three sentinels. They collapsed into one until #335, and
// because every recipe treats "refused to start" as benign, a total outage was
// classified as normal — the loud-failure alerts built to catch it could never
// fire (@Sten, verified against a closed port).
//
// This test FAILS if any two collapse, which is the property that was missing.
func TestThreeFailuresStayDistinct(t *testing.T) {
	unreachableErr := &url.Error{
		Op: "Get", URL: "http://localhost:9999/api/v1/users/me",
		Err: &net.OpError{Op: "dial", Err: errors.New("connection refused")},
	}

	cases := []struct {
		name     string
		setup    func(*WatchOptions, *fakeRoom, string)
		sentinel error
	}{
		{"backend unreachable", func(o *WatchOptions, r *fakeRoom, _ string) {
			r.whoErr = unreachableErr
		}, ErrUnreachable},

		{"identity mismatch", func(o *WatchOptions, r *fakeRoom, _ string) {
			r.email = "someone-else@example.com"
		}, ErrIdentityGuard},

		{"lock held", func(o *WatchOptions, r *fakeRoom, state string) {
			_ = os.MkdirAll(state, 0o755)
			_ = os.WriteFile(filepath.Join(state, fileLock), []byte("99999\n"), 0o644)
		}, ErrLocked},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o, _, state := opts(t)
			room := &fakeRoom{msgs: msgs(2), newest: "100-0", email: o.ExpectEmail}
			tc.setup(&o, room, state)

			_, err := Watch(context.Background(), room, o)
			if err == nil {
				t.Fatal("expected a failure")
			}
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("got %v, want %v", err, tc.sentinel)
			}
			// The collapse this test exists to prevent: no condition may match
			// more than one sentinel.
			for _, other := range []error{ErrUnreachable, ErrIdentityGuard, ErrLocked} {
				if other != tc.sentinel && errors.Is(err, other) {
					t.Errorf("%v ALSO matches %v — the two have collapsed", err, other)
				}
			}
		})
	}
}

// ⚠️ The specific regression: a transport failure must not surface as an
// identity problem. whoami is simply the first call that touches the wire, so
// an outage used to be reported as a config error — which is how it exited 10
// and got treated as benign.
func TestTransportFailureIsNotAnIdentityProblem(t *testing.T) {
	o, _, _ := opts(t)
	room := &fakeRoom{
		msgs: msgs(2), newest: "100-0", email: o.ExpectEmail,
		whoErr: &url.Error{Op: "Get", URL: "http://x", Err: errors.New("connection refused")},
	}

	_, err := Watch(context.Background(), room, o)
	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v, want ErrUnreachable", err)
	}
	if errors.Is(err, ErrIdentityGuard) {
		t.Fatal("an outage is being reported as an identity guard — this is #335")
	}
}

// A failed READ must be classified too, not just a failed whoami. The read is
// where an outage lands once identity is cached or already verified.
func TestUnreachableOnTheReadIsAlsoClassified(t *testing.T) {
	o, _, _ := opts(t)
	room := &fakeRoom{msgs: msgs(2), newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.readErr = &url.Error{Op: "Get", URL: "http://x", Err: errors.New("connection refused")}

	_, err := Watch(context.Background(), room, o)
	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v, want ErrUnreachable on a failed read", err)
	}
}

// A refusal from a LIVE server is not an outage. Without this, any read error
// would be called unreachable and the distinction would be worthless.
func TestServerErrorIsNotUnreachable(t *testing.T) {
	o, _, _ := opts(t)
	room := &fakeRoom{msgs: msgs(2), newest: "100-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.readErr = errors.New("room read: 500 internal server error")

	_, err := Watch(context.Background(), room, o)
	if err == nil {
		t.Fatal("expected a failure")
	}
	if errors.Is(err, ErrUnreachable) {
		t.Error("a 500 from a live server was classified as unreachable")
	}
}

// --since starts a FIRST RUN somewhere other than a baseline, so the messages
// just before you joined are delivered instead of skipped.
//
// ⚑ @Tuner: "baselining eats the config instructions addressed to you."
// @Smiley lost the Leads room's only message to this within five minutes of
// arming (#340).
func TestSinceSeedsInsteadOfBaselining(t *testing.T) {
	o, q, s := opts(t)
	all := msgs(5)
	o.Since = "101-0" // start after the second message
	room := &fakeRoom{msgs: all, newest: "104-0", email: o.ExpectEmail}

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Baselined {
		t.Fatal("--since must not baseline")
	}
	if res.SeededFrom != "101-0" {
		t.Errorf("SeededFrom = %q, want 101-0 — the log must say which path ran", res.SeededFrom)
	}
	if res.Queued != 3 {
		t.Fatalf("queued %d, want 3 (102,103,104) — history was not delivered", res.Queued)
	}
	if lines(t, q) != 3 {
		t.Errorf("queue has %d, want 3", lines(t, q))
	}
	_ = s
}

func TestFromStartTakesTheWholeRoom(t *testing.T) {
	o, q, _ := opts(t)
	o.FromStart = true
	room := &fakeRoom{msgs: msgs(5), newest: "104-0", email: o.ExpectEmail}

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Baselined || res.SeededFrom != "(start)" {
		t.Fatalf("baselined=%v seeded=%q, want false/(start)", res.Baselined, res.SeededFrom)
	}
	if lines(t, q) != 5 {
		t.Errorf("queue has %d, want all 5", lines(t, q))
	}
}

// 🛑 An EXISTING cursor always wins. A watcher's position must not be
// resettable by a stray flag on a routine tick — one bad invocation would
// otherwise silently replay or skip a day.
func TestExistingCursorIgnoresSinceAndFromStart(t *testing.T) {
	o, _, s := opts(t)
	all := msgs(6)
	room := &fakeRoom{msgs: all[:3], newest: "102-0", email: o.ExpectEmail}
	if _, err := Watch(context.Background(), room, o); err != nil { // baselines at 102-0
		t.Fatal(err)
	}
	room.msgs = all

	// Both flags set, and both must be inert now that a cursor exists.
	o.Since, o.FromStart = "100-0", true
	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.SeededFrom != "" {
		t.Errorf("SeededFrom = %q — a live watcher was re-seeded", res.SeededFrom)
	}
	if res.CursorFrom != "102-0" {
		t.Fatalf("resumed from %q, want 102-0 — --since dragged the cursor backwards", res.CursorFrom)
	}
	if res.Queued != 3 {
		t.Errorf("queued %d, want 3 — only what is genuinely new", res.Queued)
	}
	if got := readTrimmed(filepath.Join(s, fileCursor)); got == "100-0" {
		t.Fatal("the cursor file was reset by --since")
	}
}

// A dead backend must report rc=12 on the FIRST run too (#335 follow-up).
//
// 🛑 #335 wrapped unreachable on the delta path only. The baseline path — the
// very first poll a new watcher makes — returned a bare error, so a total outage
// during setup surfaced as rc=1 "failed" rather than rc=12 "backend
// unreachable". Every shared recipe treats those differently, and the first poll
// is exactly when an operator is watching.
func TestUnreachableOnFirstRunIsNotPlainFailure(t *testing.T) {
	dir := t.TempDir()
	_, err := Watch(context.Background(), unreachableReader{}, WatchOptions{
		EntityType: "story", EntityID: "e1",
		QueuePath: filepath.Join(dir, "q.jsonl"), StateDir: filepath.Join(dir, "s"),
		Shared: true, Handle: "Tate", Limit: 10,
	})
	if err == nil {
		t.Fatal("want an error from an unreachable backend")
	}
	if !errors.Is(err, ErrUnreachable) {
		t.Errorf("first-run outage must classify as ErrUnreachable (rc=12), got %v", err)
	}
}

// unreachableReader fails Newest the way a refused connection does.
type unreachableReader struct{}

func (unreachableReader) Newest(context.Context, string, string) (string, error) {
	return "", &net.OpError{Op: "dial", Err: errors.New("connect: connection refused")}
}
func (unreachableReader) ReadSince(context.Context, string, string, string, int, []string, []string, []string) ([]Message, string, error) {
	return nil, "", &net.OpError{Op: "dial", Err: errors.New("connect: connection refused")}
}
func (unreachableReader) WhoAmI(context.Context) (string, error) { return "agent@example.com", nil }

// Self-exclusion is pushed SERVER-side as well as applied locally (#336).
//
// ⚠️ The local pass must stay: an older backend ignores unknown query params and
// fails OPEN, so a server-only exclusion would silently stop working. This
// asserts the server is ASKED — not that the local net was removed.
func TestSelfExclusionIsAlsoPushedServerSide(t *testing.T) {
	dir := t.TempDir()
	room := &fakeRoom{msgs: []Message{{ID: "1", Agent: "Tate", Content: "@Tate mine"}}}
	if _, err := Watch(context.Background(), room, WatchOptions{
		EntityType: "story", EntityID: "e1",
		QueuePath: filepath.Join(dir, "q.jsonl"), StateDir: filepath.Join(dir, "s"),
		Shared: true, Handle: "Tate", Limit: 10, Since: "0",
		Match: []string{"@Tate"},
	}); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range room.lastExclude {
		if strings.Contains(e, "Tate") {
			found = true
		}
	}
	if !found {
		t.Errorf("own handle not pushed as a server-side veto; excludeFrom=%v", room.lastExclude)
	}
}

// --include-self must NOT push a self veto — otherwise opting in still filters.
func TestIncludeSelfSendsNoSelfVeto(t *testing.T) {
	dir := t.TempDir()
	room := &fakeRoom{msgs: []Message{{ID: "1", Agent: "Tate", Content: "@Tate mine"}}}
	if _, err := Watch(context.Background(), room, WatchOptions{
		EntityType: "story", EntityID: "e1",
		QueuePath: filepath.Join(dir, "q.jsonl"), StateDir: filepath.Join(dir, "s"),
		Shared: true, Handle: "Tate", Limit: 10, Since: "0",
		Match: []string{"@Tate"}, IncludeSelf: true,
	}); err != nil {
		t.Fatal(err)
	}
	for _, e := range room.lastExclude {
		if strings.Contains(e, "Tate") {
			t.Errorf("--include-self still pushed a self veto: %v", room.lastExclude)
		}
	}
}

// 🛑 THE ARM STAMP MUST SURVIVE A FAILED POLL (#367). This is the whole reason
// it is written before the identity check and before a byte goes over the wire.
//
// A watcher whose backend is refusing it leaves NO success stamp — so on the
// success stamp alone it is indistinguishable from a process that is not
// running at all. Those are opposite emergencies: one says go and look at your
// credentials, the other says go and look at whether anything is alive. My own
// rc=10 exemption made the second unreachable for a day because it could not
// tell them apart.
//
// ⚠️ Asserts on a poll that ERRORS, deliberately. Arming on the happy path is
// easy and proves nothing — the failing path is the one the state exists for.
func TestArmStampIsWrittenEvenWhenThePollFails(t *testing.T) {
	dir := t.TempDir()
	state := filepath.Join(dir, "s")

	_, err := Watch(context.Background(), unreachableReader{}, WatchOptions{
		EntityType: "story", EntityID: "e1",
		QueuePath: filepath.Join(dir, "q.jsonl"), StateDir: state,
		Shared: true, Handle: "Tate", Limit: 10,
	})
	if err == nil {
		t.Fatal("precondition: want a failed poll")
	}

	if !fileExists(state, fileHealthArm) {
		t.Fatal("a FAILED poll left no arm stamp — 'the backend is refusing me' is now indistinguishable from 'nothing is running'")
	}
	if fileExists(state, fileHealthOK) {
		t.Fatal("a failed poll wrote a SUCCESS stamp — that reports a broken watcher as healthy, the worst error this file can make")
	}

	// And the pair must read as `failing`, not as a dead loop: an attempt just
	// happened, no success ever did.
	l, _ := CheckLiveness(state, 5*time.Minute)
	if l.State != LiveNever {
		t.Errorf("State = %q, want %q — attempts recorded, no success ever", l.State, LiveNever)
	}
	if !strings.Contains(l.Detail, "attempted") {
		t.Errorf("detail %q does not tell the reader polls ARE being attempted", l.Detail)
	}
}

// 🛑 A 5xx IS an outage, and getting this wrong killed looping watchers dead.
//
// Only dial-level errors were classified unreachable, so a backend that was UP
// but returning 503 — a restart, a deploy, an overloaded API — fell through to
// the identity branch and surfaced as rc=10 "identity guard, needs a human".
// rc=10 is treated as PERMANENT and ends the loop, so a transient 503 killed the
// watcher outright and it never came back: the one event it exists to report.
//
// ⚑ Found by the #368 red-proof harness on its first real use. Both of @Vance's
// dev deploys that morning severed the connection (dial error, correctly rc=12),
// which is exactly why every live watcher survived them and nobody saw this.
// Running longer would never have found it.
func TestUnreachableClassifiesServerErrorsAsOutages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
		why  string
	}{
		{"503 service unavailable", &api.APIError{StatusCode: 503}, true,
			"a restarting backend — transient, the loop must survive it"},
		{"500 internal error", &api.APIError{StatusCode: 500}, true,
			"the server is failing, not the caller's identity"},
		{"502 bad gateway", &api.APIError{StatusCode: 502}, true,
			"proxy in front of a dead backend, same class"},
		{"401 unauthorized", &api.APIError{StatusCode: 401}, false,
			"a REAL credential problem — must stay permanent, or a dead token retries forever"},
		{"403 forbidden", &api.APIError{StatusCode: 403}, false,
			"access denied is not an outage"},
		{"404 not found", &api.APIError{StatusCode: 404}, false,
			"a missing room will still be missing next tick"},
		{"nil", nil, false, "no error is not an outage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unreachable(tt.err); got != tt.want {
				t.Errorf("unreachable = %v, want %v — %s", got, tt.want, tt.why)
			}
		})
	}
}

// The 4xx rows above are the load-bearing half. If 5xx-is-an-outage were widened
// to all-errors-are-outages, a dead token would retry forever and never surface
// as the misconfiguration it is — trading a watcher that dies on a blip for one
// that never tells you your credentials expired.
func TestUnreachableStillTreatsClientErrorsAsPermanent(t *testing.T) {
	for _, code := range []int{400, 401, 403, 404, 409, 422} {
		if unreachable(&api.APIError{StatusCode: code}) {
			t.Errorf("HTTP %d classified as an outage; it is permanent and must not be retried forever", code)
		}
	}
}

// 🛑 The regression this fix nearly shipped, tested at the ACCOUNTING path
// rather than at the Filter — because my first attempt at guarding it was a
// Filter-level test that stayed GREEN when I deliberately merged the counters.
// A check that cannot fail is not a check; this one was caught only by
// red-proofing it.
//
// ServerMissed means "the backend ignored our filter" and prints a loud warning
// that the build is stale. The #417 elision drops messages the server correctly
// returned, so counting those as ServerMissed would make every filter discussion
// in the room accuse the backend of being old, on every leg in the fleet.
func TestQuotedOnlyDropsAreNotCountedAsServerFailOpen(t *testing.T) {
	o, _, _ := opts(t)
	all := msgs(3)
	all[0].Content = "@Tate can you look at this"                  // real address → kept
	all[1].Content = "my gate:\n```\nmatch=@Tate|@Team\n```\ndone" // #417 → QuotedOnly
	all[2].Content = "entirely unrelated chatter"                  // → ServerMissed

	room := &fakeRoom{msgs: all[:1], newest: "100-0", email: o.ExpectEmail, ignoreFilter: true}
	if _, err := Watch(context.Background(), room, o); err != nil {
		t.Fatal(err)
	}
	room.msgs = all
	o.Match = []string{"@Tate"}

	res, err := Watch(context.Background(), room, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.QuotedOnly != 1 {
		t.Errorf("QuotedOnly = %d, want 1 — the #417 drop must be counted as its own "+
			"reason, and reported so the suppression is visible", res.QuotedOnly)
	}
	if res.ServerMissed != 1 {
		t.Errorf("ServerMissed = %d, want 1 — merging the two reasons makes every "+
			"quoted routing pattern accuse the backend of being stale (got QuotedOnly=%d)",
			res.ServerMissed, res.QuotedOnly)
	}
}
