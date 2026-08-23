package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/room"
	"github.com/claytonharbour/proseforge-workbench/internal/watcher"
)

// `pfw room watch` — the buffer stage (#308).
//
// Single-shot: one read, one append, one cursor advance, exit. Cron runs it on a
// timer, an event listener runs it when something fires, a human runs it by hand.
// The trigger decides WHEN it runs and never what it does.

// roomReader adapts the room service to the watcher's narrow interface.
type roomReader struct {
	svc    *room.Service
	client *api.Client
}

func (r roomReader) ReadSince(ctx context.Context, entityType, entityID, since string, limit int,
	match, from, excludeFrom []string) ([]watcher.Message, string, error) {
	// AgentHandle is deliberately NOT set, and Since is always supplied after
	// the first tick. THAT ordering is the protection — Since takes precedence
	// over any server-side resume.
	//
	// ⛔ This comment used to claim omitting the handle made a server-cursor
	// collision "impossible by construction". That is FALSE and @Smiley found
	// why (#1172): cursorField(agentHandle, principalID) falls back to the
	// PRINCIPAL (#729), so dropping the handle does not detach us from the
	// server cursor — it moves us from a handle-keyed one to a principal-keyed
	// one.
	//
	// ⚠️ So the watcher is UNAFFECTED, not IMMUNE, and the difference is who
	// else can break it. Measured 2026-08-23: a no-Since read on a bench token
	// returned 20 messages, because nothing has written a cursor for that
	// principal — pfw never does. Anything that DID (`room cursor set`, a
	// future server change) would make every no-Since read on that account
	// return empty, with rc=0 and a healthy-looking leg. Same silent shape as
	// #400.
	//
	// 🛑 If the server ever lets implicit resume outrank an explicit Since, the
	// whole watcher fleet goes quiet with cursors intact and health green.
	resp, err := r.svc.Read(ctx, entityType, entityID, api.ReadRoomMessagesOptions{
		Since: since, Limit: limit, Order: "asc", Match: match, From: from, ExcludeFrom: excludeFrom,
	})
	if err != nil {
		return nil, "", err
	}
	out := make([]watcher.Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		out = append(out, watcher.Message{
			ID: m.ID, Agent: m.Agent, Perspective: m.Perspective,
			Target: m.Target, Content: m.Content, Timestamp: m.Timestamp,
		})
	}
	return out, resp.LastID, nil
}

func (r roomReader) Newest(ctx context.Context, entityType, entityID string) (string, error) {
	resp, err := r.svc.Read(ctx, entityType, entityID, api.ReadRoomMessagesOptions{
		Limit: 1, Order: "desc",
	})
	if err != nil {
		return "", err
	}
	if len(resp.Messages) == 0 {
		return "", nil // empty room: baseline at the start, nothing to skip
	}
	return resp.Messages[0].ID, nil
}

func (r roomReader) WhoAmI(ctx context.Context) (string, error) {
	raw, err := r.client.WhoAmI(ctx)
	if err != nil {
		return "", err
	}
	var who struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(raw, &who); err != nil {
		return "", fmt.Errorf("decode whoami: %w", err)
	}
	return who.Email, nil
}

func newRoomWatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch <entity-id>",
		Short: "Buffer new room messages to a durable local queue",
		Long: `Read a room since this watcher's own cursor and append what is new.

Single-shot by default: one read, one append, one cursor advance, exit. With
--loop, it repeats at --interval. Run it from cron, an event listener, or a
Codex monitor process — the trigger decides WHEN it runs, never what it does.

Success means "room data is durably local" and matching message bodies are
written to stdout for a supervising worker or monitor. The old 'pfw inbox lease'
handoff is removed. Buffering remains the durable retry and audit path; stdout
is the live wake path. The watcher does not interpret messages, acknowledge
them, or decide what the worker should do.

For a monitor, filter routine successful ticks such as rc=0 queued=0 and
expected binary-reload notices before they reach the model. This keeps polling
cheap without hiding non-zero exit codes or matching message bodies.

THE CURSOR IS THIS WATCHER'S OWN FILE, never the server's. Several benches on one
shared token collapse onto a single principal-keyed cursor and consume each
other's messages; a local file cannot collide with anything. It also means a cron
read can never disturb a cursor your interactive session is using.

IDENTITY IS MANDATORY, in one of two modes. They check different things:
  --expect-email <addr>   individual account: assert whoami matches
  --shared --handle <n>   shared token: whoami is blind here, so a handle is
                          required to keep this watcher's reads distinguishable

FIRST RUN BASELINES. The cursor is set to the newest message and prior history is
skipped forever. A marker records that it happened, so a new watcher can be told
apart from one whose queue was lost.

Exit codes:
  0   polled cleanly
  1   failed
  10  REFUSED TO START — identity guard. Misconfiguration; needs a human.
  11  LOCK HELD — another instance is running. Benign, retry next tick.
  12  BACKEND UNREACHABLE — an outage. Retry, but report it.

Treat anything non-zero as a failure. 10, 11 and 12 were one code until #335,
which meant a dead backend was indistinguishable from a benign lock and every
recipe that treated 10 as benign silently ignored outages.`,
		Args: cobra.ExactArgs(1),
		RunE: runRoomWatch,
	}
	cmd.Flags().String("queue", "", "JSONL queue to append to (required)")
	cmd.Flags().String("state", "", "State dir: cursor, lock, health stamps, baseline marker (required)")
	cmd.Flags().String("handle", "", "This watcher's handle (required with --shared)")
	cmd.Flags().String("expect-email", "", "Refuse to poll unless the token authenticates as this address")
	cmd.Flags().Bool("shared", false, "Shared-token mode: skip the whoami assertion, require --handle")
	// 🛑 SCAN, not results (#346). A rare pattern can return 0 for several ticks
	// while the cursor still walks forward — that is progress through a backlog,
	// NOT a miss, and nothing is skipped. Reading it as "my filter is broken" is
	// what produced three wrong conclusions in one hour.
	cmd.Flags().Int("limit", 60, "Max messages to SCAN per poll, not to keep. A filter can keep 0 of a full scan and still advance.")
	cmd.Flags().StringArray("match", nil, "Keep messages whose content/target match this RE2 regex (case-insensitive). REPEATABLE.")
	cmd.Flags().StringArray("from", nil, "Also keep messages FROM a sender matching this regex. REPEATABLE, OR'd with every --match.")
	cmd.Flags().StringArray("exclude-from", nil, "VETO a sender: drop their messages even if --match/--from kept them. REPEATABLE. Mutes volume WITHOUT narrowing your gate.")
	cmd.Flags().String("since", "", "FIRST RUN ONLY: start from this MESSAGE ID (<millis>-<seq>) instead of baselining. Not a timestamp — a timestamp is refused (#357). Ignored once a cursor exists.")
	cmd.Flags().Bool("from-start", false, "FIRST RUN ONLY: take the whole room instead of baselining. Bounded by --limit per tick.")
	// --loop (#364). Single-shot stays the DEFAULT and is untouched: cron and
	// Monitor callers are unaffected. This exists for callers with no scheduler
	// of their own — every non-Claude-Code bench, since Monitor is a harness
	// feature and Codex has no equivalent (openai/codex#29922).
	cmd.Flags().Bool("loop", false, "Poll repeatedly instead of exiting after one tick. Requires --interval.")
	// ⚠️ "when healthy" is load-bearing. @Tuner asked for 2s and measured ~9s
	// between ticks against a dead backend, and reasonably read it as an
	// undocumented backoff. It is not backoff — it is the API client's retry
	// window INSIDE one tick, then this interval is added on top.
	//
	// 🛑 MEASURED overhead is ~30s, not the ~7s this comment claimed until
	// 2026-08-23. 7s is the client's retry BUDGET, which is a different number
	// from what a failing tick costs — I documented the component I could read
	// in the code and never measured the thing itself. Across 26 real failing
	// ticks on 60s legs the gap was 90s every time: interval + ~30s.
	//
	// 🛑 And the consequence is worse than "detection is slow", because the
	// quantity a watchdog evaluates is NOT the value at the failed tick — it
	// is time-since-success, which keeps CLIMBING while the leg sleeps its
	// interval before retrying. Measured, one failed tick, 60s leg:
	//
	//	09:36:20  rc=0   last success
	//	09:37:50  rc=12  the only failure   since success  90s
	//	09:39:17  rc=0   recovery           since success 177s  ← the PEAK
	//
	// So a single failed poll leaves the leg looking dead for ~2 intervals,
	// and any --max-age under 2× the interval goes red on one blip:
	//
	//	failed ticks tolerated = floor(max_age / interval) − 1
	//
	// ⚑ Three benches derived this five different ways in eleven minutes and
	// all five were wrong, because each evaluated ONE INSTANT. The watchdog
	// samples continuously; the largest value is what decides.
	// ⚠️ Opt-OUT, not opt-in. The wrapper behaviour this restores was the default
	// for everyone before #364, and a watcher silently running deleted code is a
	// worse default than one that swaps images at a tick boundary.
	cmd.Flags().Bool("no-reexec", false,
		"Pin the binary image: do NOT re-exec when the executable changes on disk. Default is to pick up rebuilds at the next tick, matching the shell loops --loop replaced.")
	cmd.Flags().Duration("interval", 0,
		"Sleep between ticks WHEN HEALTHY (e.g. 30m). Refused without --loop. A FAILING tick costs "+
			"interval + ~30s, and the leg then looks dead until it recovers — so ONE failed poll shows "+
			"as ~2 intervals since success. Size --max-age at 2x this or more.")
	// #382. Default is <state>/telemetry.log so nobody has to think about it —
	// and so the cutover check (`tail` for tick=, `grep` for FIRST RUN) is the
	// SAME command on every bench, which was @Smiley's argument for the tool
	// choosing rather than each operator choosing.
	//
	// ⚠️ The flag exists for the cases a fixed path cannot serve: a log
	// aggregator, a different volume, a state dir that is synced somewhere the
	// telemetry should not follow — and to step aside if a caller's existing
	// redirect already owns that filename (@Smiley's two-writers hazard).
	// @Crispin's ALARM_AFTER — the last wrapper feature with no flag (#365).
	cmd.Flags().Int("alarm-after", 1,
		"Alarm on the Nth CONSECUTIVE failing tick. 1 catches everything; raise it on a noisy environment where short bounces are routine — but a high value makes the alarm itself hard to ever observe.")
	cmd.Flags().String("telemetry", "",
		"Where --loop writes per-tick telemetry (default: <state>/telemetry.log). Rotated at 5 MiB. Failures still go to stderr.")
	cmd.Flags().Bool("include-self", false, "Queue your own posts too (default: skipped — a watcher should not wake you with your own words)")
	return cmd
}

func runRoomWatch(cmd *cobra.Command, args []string) error {
	// ⚑ Flag misuse is validated FIRST, before config. These checks need no
	// credentials, and making them depend on valid config means an operator with
	// a half-set-up environment gets "PROSEFORGE_URL is required" for a mistake
	// that has nothing to do with the URL — then fixes the config and hits the
	// real error on the second run. Misuse must be loud AND immediate.
	loop, interval, err := validateLoopFlags(cmd, "30m")
	if err != nil {
		return err
	}

	svc, err := newRoomService(cmd)
	if err != nil {
		return err
	}
	client, err := newClient(cmd)
	if err != nil {
		return err
	}
	queue, _ := cmd.Flags().GetString("queue")
	state, _ := cmd.Flags().GetString("state")
	handle, _ := cmd.Flags().GetString("handle")
	expect, _ := cmd.Flags().GetString("expect-email")
	shared, _ := cmd.Flags().GetBool("shared")
	limit, _ := cmd.Flags().GetInt("limit")

	// Both REPEATABLE and case-insensitive by default, matching the room API
	// (proseforge#1096). Handles are inconsistently cased in real traffic —
	// Smiley posts as both "smiley" and "Smiley" — so a case-sensitive --from
	// silently drops 85% of him. Force exactness with an inline (?-i).
	includeSelf, _ := cmd.Flags().GetBool("include-self")
	since, _ := cmd.Flags().GetString("since")
	fromStart, _ := cmd.Flags().GetBool("from-start")
	matchPats, _ := cmd.Flags().GetStringArray("match")
	fromPats, _ := cmd.Flags().GetStringArray("from")
	excludePats, _ := cmd.Flags().GetStringArray("exclude-from")

	opts := watcher.WatchOptions{
		EntityType:  roomType(cmd),
		EntityID:    args[0],
		QueuePath:   queue,
		StateDir:    state,
		Handle:      handle,
		ExpectEmail: expect,
		Shared:      shared,
		Limit:       limit,
		Match:       matchPats,
		From:        fromPats,
		ExcludeFrom: excludePats,
		IncludeSelf: includeSelf,
		Version:     Version,
		Since:       since,
		FromStart:   fromStart,
	}

	// --loop (#364). Every bench hand-wrote this in shell — @Sten 230 lines,
	// @Tuner ~350 — and every copy differed in ways that mattered. The loop is
	// the product feature; single-shot stays the default and is untouched.
	if loop {
		return runWatchLoop(cmd, roomReader{svc: svc, client: client}, opts, interval)
	}

	res, err := watcher.Watch(cmd.Context(), roomReader{svc: svc, client: client}, opts)
	if err != nil {
		// 🛑 THREE conditions, THREE codes. They collapsed into 10 until #335,
		// and because every shared recipe treats 10 as benign, a total outage
		// was being classified as normal — the loud-failure alerts built to
		// catch exactly that could never fire (@Sten).
		//
		//   10  refused to start: identity. Misconfiguration, permanent, human.
		//   11  lock held. Benign, transient, resolves next tick.
		//   12  backend unreachable. An OUTAGE — retry, but say so.
		// ⚠️ Print HERE only for the three codes that os.Exit, because os.Exit
		// skips cobra's own error reporting — without this they would fail
		// silently. The generic path must NOT print: it returns to cobra, which
		// prints it. Doing both emitted every ordinary failure TWICE, and a
		// watcher script that captures stderr then relays it (the shape every
		// bench is writing tonight) turned one outage into two alert lines.
		//
		// 🛑 This DEFERS to watchExitCode rather than repeating its switch. It used
		// to carry its own copy, and the copies drifted the moment a fourth error
		// kind (ErrConfigRefusal, #357) was added: loop telemetry reported 10 while
		// single-shot exited 1 for the identical error, so cron said "permanent,
		// escalate" and --loop said "generic failure" about the same deaf watcher.
		// Two switches over one taxonomy is a promise to diverge.
		if rc := watchExitCode(err); rc != 1 {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(rc)
		}
		return err
	}

	// ⚑ Say which of the two first-run paths happened. "Seeded" and "baselined"
	// have opposite consequences for the messages just before you joined, and
	// the difference must not be silent (#340).
	if res.SeededFrom != "" {
		fmt.Fprintf(os.Stderr,
			"*** FIRST RUN — SEEDED FROM %s, not baselined. Prior history from that "+
				"point WILL be delivered. ***\n", res.SeededFrom)
	}

	if res.Baselined {
		// ⚠️ On stderr, unmissable, exactly once — the only moment it can matter.
		// A silent baseline is data loss from the operator's point of view.
		fmt.Fprintf(os.Stderr,
			"*** FIRST RUN — BASELINE SET AT %s. PRIOR HISTORY IS SKIPPED. ***\n"+
				"    Nothing posted before this point will EVER be delivered by this watcher.\n"+
				"    Migrating from another watcher? DRAIN ITS QUEUE FIRST — this is not a resume.\n",
			res.CursorTo)
	}

	if isJSON(cmd) {
		// ⚑ Name the binary that produced this (#337). Twice today a bench posted
		// measurements from a stale build and had to retract — both of us checked
		// our results and neither checked our build. A pasted result should be
		// able to answer "was this the current code?" on its own, without anyone
		// having to think of the question.
		//
		// JSON only: a version string on every quiet human tick is the noise this
		// system exists to remove.
		return printJSON(struct {
			PfwVersion string `json:"pfw_version"`
			*watcher.WatchResult
		}{Version, res})
	}
	if res.Baselined {
		return nil
	}
	// Quiet ticks print nothing to stdout: at cron cadence this runs constantly,
	// and a line per tick is the noise the whole system exists to remove.
	if res.Queued > 0 {
		fmt.Printf("%d queued (cursor %s -> %s)\n", res.Queued, res.CursorFrom, res.CursorTo)
	}
	// ⛔ NO BODIES HERE — single-shot stays byte-identical (@Tate's #364
	// invariant), and it is load-bearing for more than tidiness: every shell
	// wrapper in the fleet already renders bodies itself from the queue
	// (`tail -n "$n" | gate-emit.py`). Emitting here too would DOUBLE every
	// delivery for every existing caller — paying twice the context for one
	// message — and it would arrive as a surprise, not a release note.
	//
	// ⚠️ I did exactly that for ten minutes: one careless global replace put
	// emitQueued on both paths, and a build plus a green test suite said
	// nothing about it, because no test asserts what single-shot does NOT
	// print. Bodies belong to --loop, which has no wrapper to collide with.
	//
	// Report fetched vs kept whenever a filter dropped something. Silent
	// discarding is how a wrong filter is discovered months later (#324).
	if res.Fetched > res.Kept {
		status("filtered: fetched %d, kept %d\n", res.Fetched, res.Kept)
	}
	// ⚠️ Loud on purpose. The server was asked to filter and did not, so this
	// backend predates proseforge#1096 and every filter push is failing OPEN.
	// The local net caught it; without it the queue would fill with traffic the
	// operator asked not to see.
	if res.ServerMissed > 0 {
		status("WARNING: the server returned %d message(s) its filter should have dropped — "+
			"this backend looks older than proseforge#1096. Local filtering caught it.\n", res.ServerMissed)
	}
	if res.Duplicates > 0 {
		status("%d already in the queue, skipped\n", res.Duplicates)
	}
	return nil
}

// runWatchLoop is `--loop`: poll, report, sleep, repeat (#364).
//
// 🛑 THE LOOP MUST NOT EXIT ON A FAILING TICK. A backend outage is the moment a
// watcher is most needed, and a loop that dies on the first 12 hands you a
// watcher that stops exactly when the thing it watches breaks. Only a PERMANENT
// misconfiguration ends it — see the exit table below.
//
// ⚑ Why this is in the tool rather than in everyone's shell: five benches wrote
// this loop separately and each version differed where it mattered. @Aldric's
// was not a Monitor at all; @Sten's was inert for 6.5 hours and nothing noticed;
// mine flooded on sustained failure. One implementation, tested once.
func runWatchLoop(cmd *cobra.Command, r watcher.RoomReader, opts watcher.WatchOptions, interval time.Duration) error {
	// ⚑ NO arm-stamp is written here, and that is deliberate. I wrote one, then
	// found `health-arm` already exists and is stamped by Watch itself BEFORE
	// auth on every attempt (watch.go:289) — including failing ones, which is
	// what separates "your token is dead" from "your process is dead". @Crispin
	// asked for arm-time heartbeat and the tool already had it, richer than the
	// version I added.
	//
	// 🛑 Left as a comment rather than deleted silently: shipping a second stamp
	// beside an existing one is precisely the duplication #363 exists to end, and
	// I nearly committed it INSIDE the fix for it. The first tick fires
	// immediately on --loop, so health-arm lands within milliseconds of arming.

	// #382: the tool routes its own telemetry. Callers no longer add a redirect,
	// and the recipe stops splitting by supervisor.
	telPath, _ := cmd.Flags().GetString("telemetry")
	tel := openTelemetry(opts.StateDir, telPath)
	defer tel.Close()
	defer tel.capturePanic()

	// ⚑ The arm banner goes to BOTH: stderr so a terminal user sees it, and the
	// file so the log begins with what this leg is and where its record lives.
	// "where did my tick lines go" should answer itself on the first line.
	tel.setHeader("watching %s every %s — pfw=%s", opts.EntityID, interval, Version)
	tel.warn("watching %s every %s — telemetry: %s", opts.EntityID, interval, tel.Path())

	// #365: the alarm decision lives in roomAlarm so the flap case can be
	// fabricated in a test. It was correct here already — this moves it, it
	// does not change what a healthy or a first-failing tick says.
	alarmAfter, _ := cmd.Flags().GetInt("alarm-after")
	alarm := newRoomAlarm(alarmAfter)

	// Fingerprint the image at arm time, compared against disk each tick (#371).
	// The shell loops this replaced re-execed every tick and so picked up
	// rebuilds for free; an in-process loop does not, and silently ran deleted
	// code until someone noticed the version field.
	armedFingerprint, _ := binaryFingerprint()
	noReexec, _ := cmd.Flags().GetBool("no-reexec")

	tick := 0
	return loopUntilSignal(cmd, interval, func(ctx context.Context) error {
		tick++
		res, err := watcher.Watch(ctx, r, opts)
		rc := watchExitCode(err)

		// ⚠️ Per-tick telemetry on EVERY tick including failures (@Sten, @Tuner,
		// @Wayland — three benches built this independently). health-poll records
		// only SUCCESS, so without this a running-but-failing loop is
		// indistinguishable from a stopped one.
		queued := 0
		if res != nil {
			queued = res.Queued
		}
		// ⚠️ FILE ONLY. This is the line six benches merged into stdout and then
		// built filters to survive. A log must be complete; a wake must be rare.
		tel.log("tick=%d rc=%d queued=%d pfw=%s%s", tick, rc, queued, Version, errSuffix(err))

		switch {
		// ⛔ PERMANENT. An identity guard failure is a misconfiguration that will
		// fail identically forever; looping on it is a spin that looks like work.
		case errors.Is(err, watcher.ErrIdentityGuard):
			fmt.Printf("🛑 WATCH REFUSED TO START — identity guard, needs a human: %v\n", err)
			// ⛔ Returning an error ENDS the loop (loopUntilSignal's contract).
			// Identity failure is permanent: looping on it is a spin that looks
			// like work. 11 and 12 return nil and keep polling, because an outage
			// is when a watcher is most needed.
			return err

		// rc=11 is the watcher WORKING: another tick holds the lock. Counting it
		// as failure makes a busy watcher alarm about its own health.
		case err == nil || errors.Is(err, watcher.ErrLocked):
			if line := alarm.observe(true, rc, nil); line != "" {
				fmt.Println(line)
			}
			reportWatchResult(cmd, res, tel)

		// Edge-triggered: announce entry, stay silent while it persists, repeat
		// periodically so a long outage is neither forgotten nor a firehose.
		// Level-triggering makes the flood scale with duration ÷ interval, so
		// the interval becomes load-bearing and the design breaks silently when
		// the command is copied with a shorter one.
		default:
			if line := alarm.observe(false, rc, err); line != "" {
				fmt.Println(line)
			}
		}

		// At the tick boundary, after the cursor is committed and before the
		// sleep: the only point where replacing the process image is safe.
		// ⚠️ Not while FAILING. The alarm's state — "already announced, don't
		// re-alarm" — lives in memory and does not survive an exec. Swapping
		// mid-outage would restart the alarm as if the outage were new, so a
		// rebuild during a long outage would re-announce it. Healthy ticks only;
		// the swap waits for recovery, which is exactly when it costs nothing.
		if !noReexec && !alarm.failing {
			reexecIfBinaryChanged(&armedFingerprint, tel)
		}

		return nil
	})
}

// watchExitCode maps a Watch error to the exit code the single-shot path would
// have used, so loop telemetry and single-shot exit codes cannot drift apart.
func watchExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, watcher.ErrUnreachable):
		return 12
	case errors.Is(err, watcher.ErrLocked):
		return 11
	case errors.Is(err, watcher.ErrIdentityGuard), errors.Is(err, watcher.ErrConfigRefusal):
		return 10
	}
	return 1
}

// errSuffix appends the failure reason to a telemetry line. Without it `-o json`
// and the tick log lose the reason entirely (@Sten), leaving an rc and no cause.
func errSuffix(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf(" err=%q", err.Error())
}

// reportWatchResult prints what a successful tick found, in loop mode.
//
// Shares the single-shot vocabulary deliberately — a first-run banner, a queued
// count, filter and server-miss warnings mean the same thing whether the trigger
// was cron or --loop. An operator should not have to learn two dialects.
//
// Message bodies go out through emitQueued (#366), which owns the budget, the
// truncation disclosure and the degraded-render guard.
func reportWatchResult(cmd *cobra.Command, res *watcher.WatchResult, tel *telemetry) {
	if res == nil {
		return
	}
	if res.SeededFrom != "" {
		tel.log("FIRST RUN — SEEDED FROM %s, not baselined", res.SeededFrom)
		status("*** FIRST RUN — SEEDED FROM %s, not baselined. Prior history from that point WILL be delivered. ***\n",
			res.SeededFrom)
	}
	if res.Baselined {
		tel.log("FIRST RUN — BASELINE SET AT %s. PRIOR HISTORY SKIPPED", res.CursorTo)
		status("*** FIRST RUN — BASELINE SET AT %s. PRIOR HISTORY IS SKIPPED. ***\n"+
			"    Nothing posted before this point will EVER be delivered by this watcher.\n",
			res.CursorTo)
		return
	}
	// 🛑 STDOUT. A Monitor raises events from stdout only, so a queued-message
	// notice on stderr reaches nobody — that is exactly how @Sten's watcher went
	// deaf while looking healthy.
	if res.Queued > 0 {
		fmt.Printf("%d queued (cursor %s -> %s)\n", res.Queued, res.CursorFrom, res.CursorTo)
		// #366: and the BODIES, on the same stream. A count tells the reader
		// something happened; it does not tell them what, so they go and read
		// the queue — the round trip a watcher exists to remove.
		emitQueued(os.Stdout, res)
	}
	// Routine per-tick counters: the FILE, not the terminal (#382). These are a
	// log — complete — not a notification.
	if res.Fetched > res.Kept {
		tel.log("filtered: fetched %d, kept %d", res.Fetched, res.Kept)
	}
	if res.Duplicates > 0 {
		tel.log("%d already in the queue, skipped", res.Duplicates)
	}

	// ⚠️ ONCE PER RUN, not per tick (@Tate's review of cbf44c5).
	//
	// This is a real event — the backend's filter is older than
	// proseforge#1096 — so it belongs on stderr. But it is a PERSISTENT
	// condition, not a transient one: it stays true until someone upgrades the
	// backend, so repeating it every tick is the same flood in a new place.
	// @Rowan at --interval 15s would meet it first.
	//
	// ⚑ Every occurrence still goes to the file. A wake must be rare; a log
	// must be complete — the same split the rest of this change is built on.
	if res.ServerMissed > 0 {
		tel.log("server-missed: %d message(s) the server's filter should have dropped", res.ServerMissed)
		if !serverMissedAnnounced {
			serverMissedAnnounced = true
			status("WARNING: the server returned %d message(s) its filter should have dropped — "+
				"this backend looks older than proseforge#1096. Local filtering caught it. "+
				"(said once; every occurrence is in the telemetry log)\n", res.ServerMissed)
		}
	}
}

// serverMissedAnnounced keeps the backend-mismatch warning to one line per run.
// Package-level because reportWatchResult is called once per tick and has no
// other state; the condition it reports cannot change without a redeploy.
var serverMissedAnnounced bool

// binaryFingerprint identifies the executable image on disk: mtime + size.
//
// Not a hash, deliberately — this runs once per tick and a hash of a 30 MB
// binary every 60s is real cost for no gain. mtime+size changes on every
// `make build`, which is the event we care about; the failure mode of a missed
// change is "keeps running the old image", i.e. exactly today's behaviour.
func binaryFingerprint() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	// Resolve the symlink: ~/.local/bin/pfw points into build/bin/pfw, and it is
	// the TARGET that `make build` replaces. Fingerprinting the link would never
	// change and the check would silently never fire — a guard that cannot go red.
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	fi, err := os.Stat(exe)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%d:%d", fi.ModTime().UnixNano(), fi.Size()), true
}

// reexecIfBinaryChanged replaces this process with the current binary when the
// image on disk has changed (#371).
//
// 🛑 THE REGRESSION THIS UNDOES. The shell wrappers `--loop` replaced ran
// `while true; do pfw …; sleep N; done`, which re-execs the binary every tick —
// so a rebuild was picked up on the next tick with no restart. Nobody designed
// that; it fell out of the shape. My in-process loop silently removed it, and
// @Smiley's watcher sat on an image @Tuner's wrapper had already left behind.
//
// ⚠️ That matters here more than most places: ONE shared binary serves the whole
// fleet and any bench committing rebuilds it. Six rebuilds in ninety minutes
// today — including 13c6e67, the fix for a bug that KILLS loops. A watcher armed
// before it keeps dying on 5xx while the fix sits on disk doing nothing.
//
// Called at a TICK BOUNDARY only: never mid-poll, never holding the lock. The
// cursor is committed per tick, so the swap cannot straddle a write.
func reexecIfBinaryChanged(armed *string, tel *telemetry) {
	current, ok := binaryFingerprint()
	if !ok || *armed == "" || current == *armed {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}

	// 🛑 mtime+size is a REBUILD detector, not a CHANGE detector. `make build`
	// rewrites the file every time, so the fingerprint moves even when the
	// resulting binary is byte-identical — and my first version re-exec'd on
	// that alone. @Smiley measured 3 re-execs of which ONE had anything to
	// adopt: "the notice is honest and the event is spurious."
	//
	// ⚑ So ask the binary what it IS before replacing ourselves with it. One
	// cheap subprocess, only on the tick where mtime moved — never per tick.
	// If the version is unchanged there is nothing to adopt: record the new
	// fingerprint so we do not ask again, and carry on.
	//
	// ⚠️ Compare src.<hash> ONLY — not the commit, and not the full string.
	//
	// @Smiley caught the residual my first fix left, with a live case:
	//
	//   running   0.1.0-6b60a16-dirty+src.33b23534
	//   disk      0.1.0-1cc3350      +src.33b23534
	//                   ↑ label moved         ↑ SOURCE IDENTICAL
	//
	// Someone committed the very tree the running binary was built from. The
	// label changed, the code did not, and a full-string comparison re-execs for
	// a cosmetic difference — the same "adopted nothing" event this fix exists
	// to stop, one layer subtler.
	//
	// ⚑ src.<hash> is a shasum of the .go sources; it is the only field that
	// tracks CONTENT. Its blind spots are known and named in #371: it includes
	// *_test.go (over-reports, harmless) and excludes the Makefile and embedded
	// assets (under-reports — a build-flag change is invisible to it).
	out, verr := exec.Command(exe, "--version").Output()
	if verr == nil {
		onDisk := sourceHash(parseVersionField(string(out)))
		if mine := sourceHash(Version); onDisk != "" && onDisk == mine {
			*armed = current // same source, new file: absorb silently
			return
		}
	}
	// ⚠️ STDERR, not stdout — corrected within the hour by @Wayland and @Tuner,
	// who measured what my first version cost: 12 wakes in 20 minutes across 3
	// legs, ZERO actionable.
	//
	// 🛑 It fails the fleet's own rule, which @Sten wrote: failures go to stdout
	// BECAUSE THEY MUST WAKE YOU; telemetry goes to stderr BECAUSE IT MUST NOT.
	// "The binary under you was replaced, correctly, and nothing is required of
	// you" is the definition of the second, and I routed it to the first.
	//
	// ⚑ And #371 made it redundant an hour after shipping: the running image is
	// already in state/running-version (a file, immune to quoting) and in the
	// per-tick stderr telemetry. A wake carried data both of those already held.
	//
	// 📌 The second-order cost is the part worth remembering: 21 binary versions
	// today × 4 legs × 13 benches. Routing state-changes to the wake channel
	// makes every watcher loudest exactly when the fleet is busiest — the alarm
	// channel degrades under load, and load is when it matters most.
	// #382: routine. @Clayton named "restarts" as terminal garbage, and this is
	// literally it — a successful re-exec needs no human. A FAILED one does, and
	// stays on stderr below.
	tel.log("♻️ binary changed on disk — re-exec into the new image (was %s)", Version)
	if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
		// Exec only returns on failure. Keep polling on the old image rather
		// than exiting: a stale watcher beats no watcher, and the operator has
		// been told.
		status("re-exec failed, continuing on the current image: %v\n", err)
	}
}

// parseVersionField pulls the version out of `pfw --version` output
// ("pfw version 0.1.0-abc+src.def"). Returns "" if it cannot be read, which
// makes the caller fall through to re-exec — the safe direction: adopting the
// new image on an unparseable answer beats pinning an old one forever.
func parseVersionField(out string) string {
	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// sourceHash extracts the src.<hash> component — the only part of the version
// string that tracks code rather than labels. Empty if absent, which makes the
// caller fall through to re-exec: the safe direction.
func sourceHash(version string) string {
	if i := strings.Index(version, "+src."); i >= 0 {
		return version[i+len("+src."):]
	}
	return ""
}
