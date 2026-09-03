package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

// Per-leg telemetry, written by the tool rather than plumbed by the operator
// (#382, @Clayton's question).
//
// ⚑ THE INCONSISTENCY THIS REMOVES. `pfw` already owns a directory per leg and
// writes `.cursor`, `health-poll`, `health-arm` and `running-version` into it
// without anyone's help. Tick telemetry was the ONLY output it made the caller
// route by hand — and that single gap produced, in one afternoon:
//
//	6 benches merged stderr into stdout with 2>&1
//	3 then built greps to survive the flood they had just created
//	1 of those greps dropped ~96% of message bodies ('queued=2' vs '2 queued')
//
// Worse, the correct instruction SPLIT BY SUPERVISOR — a redirect was required
// under cron and a no-op under a Claude Monitor, which captures stderr without
// notifying. Two audiences, two instructions, for a stream the tool could place
// itself. Now it does.
//
// 🛑 WHAT DOES NOT MOVE: a genuine failure still reaches stderr. A tool that
// files its own crash report where nobody is looking is the silent-death class
// this whole subsystem exists to remove. Failures are TEE'd — file for the
// record, stderr so a human sees them now.

const (
	telemetryFile = "telemetry.log"
	// telemetryMaxBytes bounds one file before it is rotated aside. @Rowan runs
	// --interval 15s — ~5,760 lines/day/leg — so a file that only grows is the
	// same problem slowed down.
	telemetryMaxBytes = 5 << 20 // 5 MiB
)

// telemetry writes a leg's diagnostic record to its own state dir.
//
// ⚠️ Every method is nil-safe. Telemetry must never be the reason a watcher
// fails to poll: an unwritable state dir degrades to stderr-only, which is
// exactly the behaviour callers had before this existed.
type telemetry struct {
	f    *os.File
	path string

	// header identifies this leg — room, interval, version. Re-emitted after
	// every rotation so a rotated file is self-describing.
	//
	// ⚑ @Clayton made this log the fleet's diagnostic channel: "ship your logs
	// in a ticket and we fix it in one place." A log that cannot say which leg,
	// which room and which build produced it is a log nobody can act on — and
	// without this the FIRST rotation silently strips exactly that, because the
	// banner is written once at arm time and never again.
	header  string
	written int64
}

// openTelemetry prepares <stateDir>/telemetry.log, rotating it if it has grown
// past the cap. A failure here is reported once and then tolerated.
func openTelemetry(stateDir, override string) *telemetry {
	path := override
	if path == "" {
		if stateDir == "" {
			return nil
		}
		path = filepath.Join(stateDir, telemetryFile)
	}
	dir := filepath.Dir(path)
	if fi, err := os.Stat(path); err == nil && fi.Size() > telemetryMaxBytes {
		// One generation back. Two is a retention policy nobody asked for.
		_ = os.Rename(path, path+".1")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "telemetry: cannot create %s (%v) — logging to stderr only\n", dir, err)
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "telemetry: cannot write %s (%v) — logging to stderr only\n", path, err)
		return nil
	}
	written := int64(0)
	if fi, err := f.Stat(); err == nil {
		written = fi.Size()
	}
	return &telemetry{f: f, path: path, written: written}
}

// setHeader records the identity line to re-emit after each rotation.
func (t *telemetry) setHeader(format string, args ...any) {
	if t == nil {
		return
	}
	t.header = fmt.Sprintf(format, args...)
}

// rotate moves the current file aside and starts a new one carrying the header.
//
// 🛑 Checked on WRITE, not only at open. The first version rotated once at arm
// time, which means a leg that never restarts never rotates — and the whole
// point of --loop is not restarting. @Rowan at --interval 15s writes ~5,760
// lines a day, so "rotates at 5 MiB" was true only for legs that were already
// being restarted for other reasons.
func (t *telemetry) rotate() {
	if t == nil || t.f == nil {
		return
	}
	_ = t.f.Close()
	_ = os.Rename(t.path, t.path+".1")
	f, err := os.OpenFile(t.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "telemetry: rotation failed for %s (%v) — telemetry stops here\n", t.path, err)
		t.f = nil
		return
	}
	t.f, t.written = f, 0
	if t.header != "" {
		t.log("%s (continued after rotation)", t.header)
	}
}

func (t *telemetry) Path() string {
	if t == nil {
		return ""
	}
	return t.path
}

// log records a line in the file ONLY. This is the per-tick path: a log must be
// complete, and a wake must be rare — different audiences, different channels.
func (t *telemetry) log(format string, args ...any) {
	// ⚠️ DEGRADE TO STDERR, never to nothing. With no state dir and no
	// --telemetry there is no file to write, and dropping the line silently
	// would make a running checker indistinguishable from a stopped one — the
	// exact failure this telemetry exists to prevent, reintroduced by the fix
	// for it. Caught by TestHealthLoopIsSilentWhileEverythingIsHealthy.
	if t == nil || t.f == nil {
		fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
		return
	}
	n, _ := fmt.Fprintf(t.f, "%s %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, args...))
	t.written += int64(n)
	if t.written > telemetryMaxBytes {
		t.rotate()
	}
}

// warn records a line AND puts it on stderr.
//
// ⚑ @Clayton: "if we get crashes at least we will have logs and can have agents
// send them to us." So a failure goes to BOTH — the file gives it a stable path
// an agent can retrieve later, stderr makes it visible now. Neither substitutes
// for the other: under a Monitor stderr lands in a harness file that is awkward
// to fetch, and in a terminal the log is not being watched.
func (t *telemetry) warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, msg)
	t.log("%s", msg)
}

// capturePanic records a crash to the file before letting it continue.
//
// 🛑 It RE-PANICS deliberately. Swallowing the panic here would turn a loud
// crash into a watcher that exited quietly, which is worse than the crash.
// The only thing added is that the stack survives in a place someone can find
// it afterwards — deferred in the caller, so it fires on the real panic path.
func (t *telemetry) capturePanic() {
	if r := recover(); r != nil {
		t.log("PANIC %v\n%s", r, debug.Stack())
		if t != nil && t.f != nil {
			_ = t.f.Sync()
		}
		panic(r)
	}
}

func (t *telemetry) Close() {
	if t == nil || t.f == nil {
		return
	}
	_ = t.f.Close()
}
