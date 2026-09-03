// Package watcher runs the durable, model-free half of a room watcher.
//
// This file is the notify seam: the ONE place in the product where a runtime is
// allowed to exist (forge/proseforge-workbench#312). Everything about how a
// worker is invoked and authenticated lives behind an external command, so the
// product never names `claude`, `codex`, or any other binary. A public BYOAI tool
// cannot shell out to a subscription product, and a watcher that requires one is
// unusable for a stranger.
package watcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Outcome classifies what happened when we tried to wake somebody.
type Outcome string

const (
	// OutcomeOK — the handoff was accepted by the adapter.
	//
	// It does NOT mean the work is done, and it must never acknowledge inbox
	// items by itself (Angel, #307). Only the consumer's own explicit ack
	// retires an item; otherwise a notify that succeeds while the consumer dies
	// immediately afterwards would silently retire messages nobody handled.
	OutcomeOK Outcome = "ok"

	// OutcomeCannotWake — there was nothing to wake. No credential, no runtime,
	// no adapter. The bench is DEAF.
	OutcomeCannotWake Outcome = "cannot_wake"

	// OutcomeFailed — something tried and broke.
	OutcomeFailed Outcome = "failed"
)

// Exit codes the adapter contract reserves. Ported verbatim from the shell
// adapters rather than reinvented — two of them already run against this.
const (
	exitOK          = 0
	exitCannotWake  = 10
	exitFailedFloor = 1
)

// NoPost is the sentinel an adapter prints when the worker decided not to say
// anything. A wake that produces silence is a legitimate, successful outcome —
// the alternative is a bench that posts to prove it is awake.
const NoPost = "NO_POST"

// Notification is one wake attempt.
//
// The adapter receives the handoff TWO WAYS, and may use either: file paths on
// argv ($1 batch, $2 prompt, $3 active-work), and a JSON payload on stdin. Shell
// adapters take argv and ignore stdin; a Python one reads stdin and keeps argv
// for its own flags. Offering both is what lets one seam serve both without
// either side rewriting.
type Notification struct {
	// Command is the adapter to run. Absent means buffering-only, which is a
	// fully supported mode, not a degraded one.
	Command string

	BatchPath      string // $1 — the room messages, JSON
	PromptPath     string // $2 — per-BENCH prompt. Product passes it, never reads it.
	ActiveWorkPath string // $3 — standing work, or "" if none

	// WakeID identifies this attempt. Rides in the environment rather than on
	// argv so the three-argument contract the existing adapters implement keeps
	// working unchanged.
	WakeID string

	// Env is extra environment for the adapter (TOKEN_FILE, WORKER_MODEL, and
	// whatever else a bench's profile sets). Opaque to the product.
	Env []string

	// Messages is the batch, inlined into the stdin payload.
	//
	// ⚑ Adapters get the batch BOTH WAYS: file paths on argv, and a JSON
	// payload on stdin. That is not indecision — the two runtimes genuinely
	// want different things, and supporting both costs one goroutine.
	//
	// The shell adapters read $1 and ignore stdin. Angel's Codex adapter reads
	// stdin and takes its own config on argv, so argv-only made it exit 2 with
	// "unrecognized arguments" the first time the two halves met. An adapter
	// that ignores stdin is unaffected: the pipe closes when it exits, and the
	// copy's EPIPE is discarded.
	Messages json.RawMessage

	// ActiveWork is the standing-work record, inlined into the same payload so
	// a stdin-style adapter does not have to go read the file itself.
	ActiveWork json.RawMessage

	// Timeout bounds the wake. A hung worker must not wedge the next scheduled
	// tick — the watcher would stop polling entirely, which is the instrument
	// disabling the thing it measures.
	Timeout time.Duration
}

// payload is the stdin document. Field names match what #312 originally
// specified, which is what the Codex adapter was written against.
type payload struct {
	WakeID         string          `json:"wake_id"`
	BatchPath      string          `json:"batch_path"`
	PromptPath     string          `json:"prompt_path,omitempty"`
	ActiveWorkPath string          `json:"active_work_path,omitempty"`
	Messages       json.RawMessage `json:"messages,omitempty"`
	ActiveWork     json.RawMessage `json:"active_work,omitempty"`
}

// Result is what a wake attempt produced.
type Result struct {
	Outcome  Outcome
	ExitCode int
	Reply    string // adapter stdout, trimmed
	Stderr   string
	Duration time.Duration

	// TimedOut distinguishes a worker that ran too long from one that failed
	// fast. Both are OutcomeFailed; only one means "raise the timeout".
	TimedOut bool
}

// ShouldPost reports whether the reply is worth sending to the room.
func (r *Result) ShouldPost() bool {
	return r.Outcome == OutcomeOK && r.Reply != "" && strings.TrimSpace(r.Reply) != NoPost
}

// DefaultTimeout is deliberately generous. A worker doing real work — reading a
// story, running a test — legitimately takes minutes, and a timeout that fires on
// honest work teaches operators to raise it until it never fires at all.
const DefaultTimeout = 10 * time.Minute

// Notify runs the adapter and classifies the result.
//
// It never returns an error for an adapter that merely failed — a failed wake is
// a Result, not an exception, because the caller must record it either way. An
// error here means the notification itself was malformed.
func Notify(ctx context.Context, n Notification) (*Result, error) {
	if n.BatchPath == "" {
		return nil, errors.New("notify: no batch path")
	}

	// No adapter configured is buffering-only. Report it as cannot-wake so it
	// shows up honestly in health rather than masquerading as a successful tick.
	if n.Command == "" {
		return &Result{
			Outcome:  OutcomeCannotWake,
			ExitCode: exitCannotWake,
			Stderr:   "no --notify-cmd configured: buffering only",
		}, nil
	}

	// A missing or non-executable adapter is CANNOT WAKE, not a crash. The
	// operator's watcher should keep buffering while they fix their config —
	// the buffer stage must survive every downstream failure.
	if err := executable(n.Command); err != nil {
		return &Result{
			Outcome:  OutcomeCannotWake,
			ExitCode: exitCannotWake,
			Stderr:   fmt.Sprintf("adapter not runnable: %v", err),
		}, nil
	}

	timeout := n.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, n.Command, n.BatchPath, n.PromptPath, n.ActiveWorkPath)

	// ⚠️ WaitDelay is what makes Timeout real, and its absence is silent.
	//
	// CommandContext kills only the DIRECT child. An adapter is a shell script,
	// so its own children — `sleep`, a model CLI, a curl — inherit the stdout
	// pipe and keep it open. Run() then blocks waiting for EOF on that pipe
	// regardless of the context, so a 300ms timeout against a 10s worker took
	// the full 10s and still reported "timed out" honestly. The classification
	// was right and the bound was fiction.
	//
	// WaitDelay force-closes the pipes shortly after the kill, which is what
	// actually bounds the tick.
	cmd.WaitDelay = 5 * time.Second

	cmd.Env = append(os.Environ(), n.Env...)
	if n.WakeID != "" {
		cmd.Env = append(cmd.Env, "WATCHER_WAKE_ID="+n.WakeID)
	}

	// The same handoff, offered on stdin for adapters that want it that way.
	// os/exec discards EPIPE on the stdin copy, so an adapter that never reads
	// this costs nothing and cannot hang.
	if doc, err := json.Marshal(payload{
		WakeID:         n.WakeID,
		BatchPath:      n.BatchPath,
		PromptPath:     n.PromptPath,
		ActiveWorkPath: n.ActiveWorkPath,
		Messages:       n.Messages,
		ActiveWork:     n.ActiveWork,
	}); err == nil {
		cmd.Stdin = bytes.NewReader(doc)
	}

	// stdout is captured even on failure, deliberately. Some runtimes print
	// fatal errors — including auth failures — to STDOUT, and discarding it is
	// how an outage stays invisible (Sten).
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	res := &Result{
		Reply:    strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		Duration: time.Since(start),
	}

	if runErr == nil {
		res.Outcome, res.ExitCode = OutcomeOK, exitOK
		return res, nil
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
	} else {
		res.ExitCode = exitFailedFloor
	}

	if ctx.Err() != nil {
		// Killed by our own timeout. Report it as failure with the reason
		// attached; a bare "exit -1" would look like a mystery crash.
		res.Outcome, res.TimedOut = OutcomeFailed, true
		res.Stderr = strings.TrimSpace(fmt.Sprintf("timed out after %s\n%s", timeout, res.Stderr))
		return res, nil
	}

	// 🛑 10 must not fold into 1.
	//
	// A DEAF bench and a BROKEN bench are different problems, and only one of
	// them announces itself. Broken is loud by nature; deaf stays silent, and
	// silence looks exactly like calm. Collapsing these is how a missing token
	// queues a batch, stamps a green heartbeat, and reports success for days.
	if res.ExitCode == exitCannotWake {
		res.Outcome = OutcomeCannotWake
	} else {
		res.Outcome = OutcomeFailed
	}
	return res, nil
}

func executable(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if fi.Mode()&0o111 == 0 {
		return fmt.Errorf("%s is not executable", path)
	}
	return nil
}
