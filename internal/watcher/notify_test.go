package watcher

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixture writes a batch and prompt and returns a Notification pointing at the
// fake adapter — the positive control that makes these arms runnable with no
// credential and no model.
func fixture(t *testing.T) Notification {
	t.Helper()
	dir := t.TempDir()
	batch := filepath.Join(dir, "batch.json")
	prompt := filepath.Join(dir, "prompt.md")
	for _, f := range []struct{ path, body string }{
		{batch, `[{"id":"1-0","agent":"Clayton","content":"@Tate look at this"}]`},
		{prompt, "you are a bench"},
	} {
		if err := os.WriteFile(f.path, []byte(f.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	adapter, err := filepath.Abs("testdata/adapter.fake.sh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(adapter, 0o755); err != nil {
		t.Fatal(err)
	}
	return Notification{
		Command:    adapter,
		BatchPath:  batch,
		PromptPath: prompt,
		WakeID:     "wake-test-1",
		Timeout:    30 * time.Second,
	}
}

func notify(t *testing.T, n Notification) *Result {
	t.Helper()
	res, err := Notify(context.Background(), n)
	if err != nil {
		t.Fatalf("Notify returned an error for an adapter outcome: %v", err)
	}
	return res
}

// 🛑 The arm this whole seam exists for. A deaf bench and a broken bench must be
// distinguishable — collapsing them is how a missing token looks healthy for days.
func TestCannotWakeIsDistinctFromFailed(t *testing.T) {
	n := fixture(t)

	n.Env = []string{"FAKE_EXIT=10"}
	if got := notify(t, n); got.Outcome != OutcomeCannotWake {
		t.Errorf("exit 10: got %q, want %q", got.Outcome, OutcomeCannotWake)
	}

	n.Env = []string{"FAKE_EXIT=1"}
	if got := notify(t, n); got.Outcome != OutcomeFailed {
		t.Errorf("exit 1: got %q, want %q", got.Outcome, OutcomeFailed)
	}

	// Any other non-zero code is a failure, not a silent third category.
	n.Env = []string{"FAKE_EXIT=42"}
	if got := notify(t, n); got.Outcome != OutcomeFailed || got.ExitCode != 42 {
		t.Errorf("exit 42: got %q/%d, want failed/42", got.Outcome, got.ExitCode)
	}
}

// A config mistake must not take out the buffer stage. The watcher keeps polling
// while the operator fixes their adapter path.
func TestMissingAdapterIsCannotWakeNotACrash(t *testing.T) {
	n := fixture(t)
	n.Command = filepath.Join(t.TempDir(), "does-not-exist.sh")

	got := notify(t, n)
	if got.Outcome != OutcomeCannotWake {
		t.Fatalf("got %q, want %q", got.Outcome, OutcomeCannotWake)
	}
	if !strings.Contains(got.Stderr, "not runnable") {
		t.Errorf("stderr should say why: %q", got.Stderr)
	}
}

// A non-executable file is the same class of mistake as a missing one — a chmod
// people forget constantly.
func TestNonExecutableAdapterIsCannotWake(t *testing.T) {
	n := fixture(t)
	p := filepath.Join(t.TempDir(), "adapter.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	n.Command = p
	if got := notify(t, n); got.Outcome != OutcomeCannotWake {
		t.Errorf("got %q, want %q", got.Outcome, OutcomeCannotWake)
	}
}

// Buffering-only is a supported mode, not a degraded one — but it must report
// honestly rather than look like a successful wake.
func TestNoAdapterConfiguredIsBufferingOnly(t *testing.T) {
	n := fixture(t)
	n.Command = ""

	got := notify(t, n)
	if got.Outcome != OutcomeCannotWake {
		t.Fatalf("got %q, want %q", got.Outcome, OutcomeCannotWake)
	}
	if got.ShouldPost() {
		t.Error("buffering-only must never produce a post")
	}
}

func TestSuccessfulWakeCarriesTheReply(t *testing.T) {
	n := fixture(t)
	n.Env = []string{"FAKE_REPLY=I read it and filed #999"}

	got := notify(t, n)
	if got.Outcome != OutcomeOK {
		t.Fatalf("got %q, want ok (stderr: %q)", got.Outcome, got.Stderr)
	}
	if !got.ShouldPost() {
		t.Error("a real reply should be posted")
	}
	if got.Reply != "I read it and filed #999" {
		t.Errorf("reply = %q", got.Reply)
	}
}

// Silence is a legitimate successful outcome. The alternative is a bench that
// posts to prove it is awake, which is the room noise this system exists to cut.
func TestNoPostIsSuccessWithoutAPost(t *testing.T) {
	got := notify(t, fixture(t)) // fake defaults to NO_POST

	if got.Outcome != OutcomeOK {
		t.Fatalf("got %q, want ok", got.Outcome)
	}
	if got.ShouldPost() {
		t.Error("NO_POST must not produce a room post")
	}
}

// Argument order is a real hazard: swap batch and prompt and a worker wakes with
// the wrong input while every exit code still says success.
func TestAdapterReceivesArgvInContractOrder(t *testing.T) {
	n := fixture(t)
	n.ActiveWorkPath = n.PromptPath // any readable third arg

	got := notify(t, n)
	if got.Outcome != OutcomeOK {
		t.Fatalf("got %q (stderr %q)", got.Outcome, got.Stderr)
	}
	if !strings.Contains(got.Stderr, "batch=batch.json") ||
		!strings.Contains(got.Stderr, "prompt=prompt.md") {
		t.Errorf("adapter saw the wrong argv: %q", got.Stderr)
	}
	// The wake id must reach the adapter, or an ack cannot be tied to its wake.
	if !strings.Contains(got.Stderr, "wake=wake-test-1") {
		t.Errorf("WATCHER_WAKE_ID did not reach the adapter: %q", got.Stderr)
	}
}

// A hung worker must not wedge the next tick — a watcher that stops polling is
// the instrument disabling the thing it measures.
func TestHungAdapterTimesOutAndSaysSo(t *testing.T) {
	n := fixture(t)
	n.Env = []string{"FAKE_SLEEP=30"}
	n.Timeout = 300 * time.Millisecond

	start := time.Now()
	got := notify(t, n)
	elapsed := time.Since(start)

	if got.Outcome != OutcomeFailed {
		t.Fatalf("got %q, want failed", got.Outcome)
	}
	if !got.TimedOut {
		t.Error("TimedOut must be set, or 'raise the timeout' is indistinguishable from a crash")
	}
	if !strings.Contains(got.Stderr, "timed out") {
		t.Errorf("stderr should name the timeout: %q", got.Stderr)
	}

	// 🛑 Assert the BOUND, not just the label.
	//
	// The first version of this test checked only TimedOut and passed while
	// taking the worker's full 10s — the classification was honest and the
	// timeout did nothing. Without a wall-clock assertion the guarantee this
	// whole test claims to prove is unobservable.
	if elapsed > 10*time.Second {
		t.Errorf("timeout did not bound the tick: %s elapsed for a %s timeout "+
			"(the adapter's grandchildren are holding the pipes open)", elapsed, n.Timeout)
	}
}

// Some runtimes print fatal errors — including auth failures — to STDOUT.
// Discarding stdout on failure is how an outage stays invisible.
func TestStdoutIsCapturedEvenOnFailure(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "noisy.sh")
	script := "#!/bin/sh\necho 'Invalid API key'\nexit 1\n"
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	n := fixture(t)
	n.Command = p

	got := notify(t, n)
	if got.Outcome != OutcomeFailed {
		t.Fatalf("got %q, want failed", got.Outcome)
	}
	if !strings.Contains(got.Reply, "Invalid API key") {
		t.Errorf("stdout was discarded on failure: %q", got.Reply)
	}
}

// Two adapter styles, one seam. The shell adapters take argv and ignore stdin;
// Angel's Codex adapter reads stdin and keeps argv for its own flags. Sending
// only argv made it exit 2 with "unrecognized arguments" the first time the two
// halves met — so the payload must arrive both ways.
func TestStdinPayloadReachesAdaptersThatWantIt(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "stdin-reader.sh")
	// Echoes what it got on stdin, ignoring argv entirely.
	if err := os.WriteFile(p, []byte("#!/bin/sh\ncat\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	n := fixture(t)
	n.Command = p
	n.Messages = json.RawMessage(`[{"id":"7-0","content":"hello"}]`)
	n.ActiveWork = json.RawMessage(`{"objective":"port the seam"}`)

	got := notify(t, n)
	if got.Outcome != OutcomeOK {
		t.Fatalf("got %q (stderr %q)", got.Outcome, got.Stderr)
	}

	var doc struct {
		WakeID     string          `json:"wake_id"`
		BatchPath  string          `json:"batch_path"`
		Messages   json.RawMessage `json:"messages"`
		ActiveWork json.RawMessage `json:"active_work"`
	}
	if err := json.Unmarshal([]byte(got.Reply), &doc); err != nil {
		t.Fatalf("stdin was not valid handoff JSON: %v\ngot: %s", err, got.Reply)
	}
	if doc.WakeID != "wake-test-1" {
		t.Errorf("wake_id = %q", doc.WakeID)
	}
	if doc.BatchPath != n.BatchPath {
		t.Errorf("batch_path = %q, want %q", doc.BatchPath, n.BatchPath)
	}
	if !strings.Contains(string(doc.Messages), "hello") {
		t.Errorf("messages did not survive: %s", doc.Messages)
	}
	if !strings.Contains(string(doc.ActiveWork), "port the seam") {
		t.Errorf("active_work did not survive: %s", doc.ActiveWork)
	}
}

// An argv-style adapter that never reads stdin must be unaffected — no hang, no
// EPIPE surfacing as a failed wake.
func TestAdapterIgnoringStdinIsUnaffected(t *testing.T) {
	n := fixture(t)
	n.Messages = json.RawMessage(`[{"id":"1-0","content":"` + strings.Repeat("x", 200000) + `"}]`)

	got := notify(t, n)
	if got.Outcome != OutcomeOK {
		t.Fatalf("a large unread stdin payload broke the wake: %q (%q)", got.Outcome, got.Stderr)
	}
}
