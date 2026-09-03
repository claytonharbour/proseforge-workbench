package main

import (
	"fmt"
	"os"
	"syscall"
)

// checkOutputStreams warns at ARM TIME when this process's stdout and stderr
// have been wired in a way the room-watching recipe forbids (#409).
//
// 🛑 THE RULE EXISTED AND FIVE BENCHES BROKE IT ANYWAY. `CLAUDE.md` says
// "Never 2>&1 — it floods you and destroys the provenance line, because a room
// can quote anything including your own instrument." On 2026-08-23 @Gordon,
// @Smiley, @Tuner, @Wayland, @Sten and I all armed legs that merged the two,
// and it surfaced only after a disk incident and an hour of `lsof`. @Sten had
// the file loaded in context while producing the error it documents.
//
// ⚑ A document cannot enforce a shell redirect. The process CAN see how it was
// launched: fstat both descriptors and compare device+inode.
//
// ⚠️ WARN, never refuse. The redirect is the operator's choice and there are
// legitimate reasons to want it; the failure here was invisibility, not the
// decision. A watcher that refuses to start over an output stream would be a
// worse tool than one that mentions it once.
func checkOutputStreams(warn func(string, ...any)) {
	out, errOut, ok := streamIdentity()
	if !ok {
		return // cannot stat — say nothing rather than guess
	}

	// ⚠️ THIS TEXT WAS WRONG THREE TIMES, each time about a different thing.
	//
	// v1 "Drop the 2>&1."         @Crispin: silences startup failures on a Monitor leg.
	// v2 "it is a two-sided trade" @Sten/@Wayland: the trade does not exist. Measured,
	//                              --loop, streams separate:
	//                                STDOUT "WATCH FAILING" 1  ·  STDERR 0
	//                              (single-shot: 0 and 0 — no alarm state outside the loop)
	// v3 "you wrote 2>&1"          ⛔ MEASURED FALSE for anyone who wrote no redirect.
	//
	// 🛑 v3 is the one that matters. SOME launchers hand the child one file for
	// both descriptors. Measured here, this harness's background mode, no
	// redirect anywhere in the command:
	//
	//	no redirect       fd1 node=111059837   fd2 node=111059837   ⛔ MERGED BY THE LAUNCHER
	//	2>separate.err    fd1 node=111059837   fd2 node=111059840   ✅ escapable
	//
	// ⛔ AND v4: "a detached/background launcher merges by default" was itself an
	// overclaim — n=1 about peers. @Gordon measured his Monitor legs unmerged with
	// no redirect, and @Tuner confirmed "you wrote 2>&1" was simply TRUE of him.
	// I generalised from ONE launcher (mine) to a class, while correcting an
	// overclaim. My own leg is the written-redirect case, not the launcher case.
	//
	// ⛔ AND v5, @Crispin's: v4 answered "fstat cannot branch" by prescribing the
	// SUPERSET remedy — a new unrotated file — to everyone, including the benches
	// for whom deleting the redirect is the whole fix. We filed three tickets
	// about unbounded log files tonight (#408 among them) and the recommended fix
	// created another one per leg. @Gordon's line applies to it exactly: "tiny
	// today is a value, not a bound."
	//
	// ⚑ THE DESIGN ERROR: the PROCESS cannot distinguish the two causes, but the
	// OPERATOR always can — they know what they typed. A check that cannot branch
	// should hand the branch to the reader, not prescribe the union of both arms.
	// @Sten's wording, adopted here.
	//
	// ⚑ @Aldric's reconciliation of which path each stream carries:
	//	--loop alarms, re-shouts, bodies  → STDOUT   the merge adds nothing
	//	startup / baseline-read failure   → STDERR   the merge is what shows it
	if out == errOut {
		warn("⛔ stdout and stderr are the SAME destination, so a room quoting " +
			"text that looks like this watcher's own output (rc=, tick=, " +
			"telemetry) is indistinguishable from the real thing. TWO CAUSES, and " +
			"this check cannot tell them apart — but YOU CAN: (1) if you wrote " +
			"2>&1, just drop it; that is the whole fix and it creates no new " +
			"file. (2) If you wrote no redirect, your launcher merged them — only " +
			"then give stderr its own path, 2>\"$D/arm.log\", a DIFFERENT file " +
			"from stdout, and note pfw does NOT rotate it. NOTE: under --loop the " +
			"merge buys nothing for alarms; WATCH FAILING, WATCH STILL FAILING and " +
			"message bodies are already on STDOUT. What it does cover is the " +
			"STARTUP path — a bad room id or bad credentials fails the baseline " +
			"read on stderr, before the loop exists. If that was your reason, read " +
			"the arm banner once instead.")
	}
	if fi, err := os.Stat(os.DevNull); err == nil {
		if id, ok2 := statID(fi); ok2 && out == id {
			warn("⚠️ stdout is /dev/null — matching message bodies and alarms are " +
				"discarded. This leg buffers to its queue and reports health, but " +
				"nothing will wake anyone.")
			return
		}
	}
	if isRegularFile(os.Stdout) {
		warn("⚠️ stdout is a regular file, which pfw does not rotate — only its own " +
			"telemetry is bounded. That file grows without limit; its content is " +
			"already durable in the queue, so truncating it loses nothing.")
	}
}

type streamID struct{ dev, ino uint64 }

func statID(fi os.FileInfo) (streamID, bool) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return streamID{}, false
	}
	return streamID{dev: uint64(st.Dev), ino: uint64(st.Ino)}, true
}

func streamIdentity() (out, errOut streamID, ok bool) {
	fo, err1 := os.Stdout.Stat()
	fe, err2 := os.Stderr.Stat()
	if err1 != nil || err2 != nil {
		return out, errOut, false
	}
	o, ok1 := statID(fo)
	e, ok2 := statID(fe)
	return o, e, ok1 && ok2
}

func isRegularFile(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode().IsRegular()
}

var _ = fmt.Sprintf
