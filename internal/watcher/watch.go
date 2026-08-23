package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// The buffer stage (forge/proseforge-workbench#308).
//
// One read, one append, one cursor advance, exit. Single-shot by design: cron
// runs it on a timer, an event listener runs it when something fires, a human
// runs it by hand. The trigger decides WHEN the loop body runs and never what it
// does, which is why this needed no decision about websockets to build.
//
// Success means "room data is durably local." Never "an agent saw it." Polling
// must keep working while filtering, auth and wake are all broken — they are
// independent failures and must fail independently.

// State file names, matching what the shell watchers already write so `pfw inbox
// health` reads both without a translation layer.
const (
	fileCursor   = ".cursor"
	fileHealthOK = "health-poll"
	fileBaseline = "baselined"
	fileLock     = ".lock"
	staleLockAge = 30 * time.Minute
	defaultLimit = 60

	// ⚑ ARM stamp — written when a poll STARTS, as against fileHealthOK which
	// is written only when one SUCCEEDS (#367).
	//
	// The pair is the whole point. One stamp answers "did a poll succeed
	// recently" and cannot separate the two failures that need opposite
	// responses:
	//
	//   attempts fresh, successes stale   the loop RUNS, the backend refuses
	//   attempts stale, successes stale   nothing is polling at all
	//
	// The first is "go look at auth or the API", the second is "go look at
	// whether your process exists". Collapsed into one word, half the fleet
	// investigates the wrong thing — and I shipped a watcher for a day that
	// treated the first as benign, so its alarm for the second was unreachable.
	fileHealthArm = "health-arm"
	// fileHealthArmPrev holds the PREVIOUS arm stamp, so the observed poll
	// interval is derivable from state alone. Without it, --max-age has nothing
	// to be checked against and a guaranteed-false-alarm config stays silent.
	fileHealthArmPrev = "health-arm-prev"
	// fileVersion records the version of the binary that is ACTUALLY RUNNING
	// this leg, rewritten every tick.
	//
	// 🛑 Provenance cannot live in a stream that also carries room messages.
	// @Sten grepped his Monitor output for `pfw=` and matched FIVE different
	// values — one his, the others @Smiley's tick line quoted in a message, my
	// own prose describing the field, and @Smiley's grep pattern. `^tick=`
	// anchoring does not save you: a quoted tick line sits at column 0 too. The
	// naive check told him he was on b57afd3 while he ran 9009986.
	//
	// ⚑ A room can quote anything, including your instrument. A FILE cannot be
	// quoted into existence by a message, so provenance goes here instead.
	fileVersion = "running-version"
)

// RoomReader is the slice of the room API this needs. An interface so the buffer
// stage is testable without a server.
type RoomReader interface {
	// serverMatch is the room API's own content filter. Empty means fetch
	// everything and filter locally.
	// Returns the messages AND the scan boundary. They differ whenever the
	// server filtered: lastID is the furthest id examined, the last message is
	// only the furthest one KEPT. Advancing to the latter re-reads the gap on
	// every tick, forever.
	ReadSince(ctx context.Context, entityType, entityID, since string, limit int,
		match, from, excludeFrom []string) (msgs []Message, lastID string, err error)
	Newest(ctx context.Context, entityType, entityID string) (string, error)
	WhoAmI(ctx context.Context) (string, error)
}

// Message is a room message as queued. Mirrors internal/inbox.Message; kept
// separate so the buffer stage does not depend on the consumer.
type Message struct {
	ID          string `json:"id"`
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`
	Timestamp   string `json:"timestamp"`
}

// WatchOptions configures one poll.
type WatchOptions struct {
	EntityType string
	EntityID   string

	QueuePath string // JSONL queue to append to
	StateDir  string // cursor, lock, health stamps, baseline marker

	// Identity. Exactly one mode must be chosen — they check different things
	// and neither is optional.
	Handle      string // required in shared mode
	ExpectEmail string // individual mode: assert whoami == this
	Shared      bool

	Limit int

	// Since starts a FIRST RUN at this message id or RFC3339 timestamp instead
	// of baselining. FromStart takes the whole room. Both are ignored once a
	// cursor exists — see the first-run block in Watch.
	Since     string
	FromStart bool

	// Match filters on content+target, From on the sender's handle. Both are
	// REPEATABLE and everything is OR'd: a message is kept if ANY match OR ANY
	// from hits (proseforge#1096).
	//
	// These are raw patterns, pushed to the server AND re-applied locally.
	// See the note on the local pass in Watch.
	Match []string
	From  []string

	// ExcludeFrom vetoes senders. NOT a third OR'd term — it runs first and
	// beats a positive match, which is the only way to say "messages naming me,
	// but not the ones I wrote" (#336). Repeatable, case-insensitive.
	//
	// ⚑ The capability this adds is muting a loud sender WITHOUT narrowing your
	// gate — previously the only way to cut volume was to drop a --match and
	// lose real messages with it.
	ExcludeFrom []string

	// IncludeSelf disables self-exclusion. Off by default: a watcher must never
	// wake you with your own post (@Sten, #328).
	//
	// You sign with your handle and you address others by theirs, so your own
	// messages match your own gate. Measured on my queue: 8 of 64 of my posts
	// matched my own filter — and for anyone gating on a bare handle rather
	// than "@handle" it is every single post they make.
	//
	// ⚑ It is an EXCLUSION, applied after --match/--from and regardless of
	// whether the match ran server-side. That keeps the OR intact: exclusions
	// compose with a union, additional inclusion criteria would not.
	IncludeSelf bool

	// Version is the running binary's version string, recorded into the state
	// dir every tick so provenance never has to be parsed out of a stream that
	// also carries room messages (@Sten).
	Version string
}

// selfHandle is the handle this watcher posts under, used only to skip its own
// messages. Prefers --handle; otherwise the local part of --expect-email, since
// agent@example.com posts as "agent".
func (o WatchOptions) selfHandle() string {
	if o.Handle != "" {
		return o.Handle
	}
	if i := strings.IndexByte(o.ExpectEmail, '@'); i > 0 {
		return o.ExpectEmail[:i]
	}
	return ""
}

// compileFilter builds the local safety net.
//
// ⚑ Both patterns now go SERVER-side (proseforge#1096 made them repeatable and
// OR'd them there), which is the bandwidth win the upstream ticket asked for.
// Before that the server had no sender filter and its match was destructive, so
// a --from forced everything local.
//
// 🛑 The local pass stays anyway, and is not redundant. An OLDER backend
// ignores unknown query parameters and returns EVERYTHING — so pushing the
// filter down to a server that does not implement it yet fails open, silently,
// and the queue fills with traffic the operator asked not to see. dev has
// #1096; prod may not. Re-applying locally costs one regex per message and
// turns a silent fail-open into a no-op.
// compileExcludeFrom builds the veto list. Separate from compileFilter because
// the two compose differently: these DROP, those KEEP (#336).
func compileExcludeFrom(patterns []string) ([]*regexp.Regexp, error) {
	var out []*regexp.Regexp
	for _, p := range patterns {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			return nil, fmt.Errorf("invalid --exclude-from %q: %w", p, err)
		}
		out = append(out, re)
	}
	return out, nil
}

func compileFilter(match, from []string) (Filter, error) {
	var f Filter
	for _, p := range match {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			return f, fmt.Errorf("invalid --match %q: %w", p, err)
		}
		f.Match = append(f.Match, re)
	}
	for _, p := range from {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			return f, fmt.Errorf("invalid --from %q: %w", p, err)
		}
		f.From = append(f.From, re)
	}
	return f, nil
}

// WatchResult reports what one poll did.
type WatchResult struct {
	Baselined bool `json:"baselined"`

	// Fetched is what the read returned; Kept is what survived filtering.
	// Reported separately because silent discarding is how you discover a
	// wrong filter months later.
	Fetched     int `json:"fetched"`
	Kept        int `json:"kept"`
	SelfSkipped int `json:"self_skipped,omitempty"`

	// SeededFrom records that a first run started somewhere other than a
	// baseline, so the log says which of the two happened.
	SeededFrom   string `json:"seeded_from,omitempty"`
	ServerMissed int    `json:"server_missed,omitempty"`
	ServerFilter bool   `json:"server_filtered,omitempty"`

	Queued     int `json:"queued"`
	Duplicates int `json:"duplicates"`

	// QueuedMessages is what was appended this tick, in order.
	//
	// 🛑 Queued is a COUNT and a count is unactionable (@Tuner: "NEW 2
	// message(s)" still makes the reader go and look, which is the thing a
	// watcher exists to avoid). Carrying the bodies means the emit layer never
	// has to re-read the queue file to say what arrived.
	QueuedMessages []Message `json:"queued_messages,omitempty"`
	CursorFrom     string    `json:"cursor_from,omitempty"`
	CursorTo       string    `json:"cursor_to,omitempty"`
}

// ErrIdentityGuard means the watcher refused to poll. Distinct so a caller can
// exit 10 for it — a misconfigured watcher is deaf, not broken.
var ErrIdentityGuard = errors.New("identity guard")

// ErrConfigRefusal means the watcher refused to poll because its configuration
// or on-disk state cannot work — not because identity failed. Separate sentinel,
// SAME exit code (10): callers already treat 10 as "permanent, do not retry,
// escalate", which is exactly right here, but a caller matching on
// ErrIdentityGuard must not be told a credential problem it does not have.
var ErrConfigRefusal = errors.New("config refusal")

// ErrLocked means another instance holds the lock.
var ErrLocked = errors.New("another watcher instance is running")

// ErrUnreachable means the backend could not be reached at all.
//
// 🛑 Distinct from ErrIdentityGuard, and that distinction is the whole point of
// this sentinel. A network failure used to surface AS an identity guard —
// whoami is simply the first call that touches the wire — so a total outage
// exited 10, which every shared recipe treats as benign. An outage was being
// classified as normal, and the loud-failure alerts built to catch it could
// never fire (@Sten, #335).
var ErrUnreachable = errors.New("backend unreachable")

// unreachable reports whether err is a transport failure rather than a refusal.
//
// A non-2xx response is NOT a url.Error — the http client only produces one when
// the request never completed. So this separates "the server said no" from "there
// was no server", which is exactly the line that was blurred.
func unreachable(err error) bool {
	if err == nil {
		return false
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	// 🛑 A 5xx IS an outage. Only dial-level failures were counted here, so a
	// backend that was UP but returning 503 — a restart, a deploy, an overloaded
	// API — fell through to the identity branch and surfaced as rc=10, "identity
	// guard, needs a human".
	//
	// ⛔ That is the worst possible misclassification for a LOOPING watcher: 10 is
	// treated as permanent and ENDS the loop, so a transient 503 killed the
	// watcher outright and it never came back. A watcher must survive exactly the
	// event it exists to report.
	//
	// ⚑ Found by the #368 red-proof harness on its first real use, against a
	// fault-injecting backend — never by running longer. @Vance's two dev deploys
	// this morning happened to sever the connection (dial error, correctly rc=12)
	// rather than return 503, which is why every live watcher survived them and
	// nobody saw this.
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode >= 500 {
		return true
	}
	return false
}

// Watch performs one poll.
func Watch(ctx context.Context, r RoomReader, opts WatchOptions) (*WatchResult, error) {
	if opts.EntityID == "" {
		return nil, errors.New("no entity id")
	}
	if opts.QueuePath == "" || opts.StateDir == "" {
		return nil, errors.New("both --queue and --state are required")
	}
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}
	if err := os.MkdirAll(opts.StateDir, 0o755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}

	// 🛑 ARM BEFORE ANYTHING CAN FAIL — before the identity check, before the
	// lock, before a byte goes over the network (#367, @Crispin).
	//
	// His first version stamped inside the loop body, so nothing existed until
	// after the first sleep and a freshly-armed watcher was indistinguishable
	// from one that had never started — recreating the exact ambiguity the
	// stamp was added to remove. He found it by noticing the FILE WAS ABSENT,
	// not by reading the code, which is why absence-then-presence is a test
	// below and not a comment here.
	//
	// Placed here on purpose: an auth failure or an unreachable backend must
	// still leave evidence that something TRIED. That evidence is the only
	// thing separating "your token is dead" from "your process is dead".
	touchArm(opts.StateDir)
	StampVersion(opts.StateDir, opts.Version)

	if err := checkIdentity(ctx, r, opts); err != nil {
		return nil, err
	}

	release, err := acquireLock(opts.StateDir)
	if err != nil {
		return nil, err
	}
	defer release()

	cursorPath := filepath.Join(opts.StateDir, fileCursor)
	cursor := readTrimmed(cursorPath)

	// ── First run ───────────────────────────────────────────────────────────
	//
	// ⚠️ A baseline SKIPS ALL PRIOR HISTORY. That is the right default for a
	// genuinely fresh watcher and the WRONG one for the commonest real case —
	// joining a room that already has messages in it.
	//
	// ⚑ @Tuner's line, and it is why --since exists: baselining eats the config
	// instructions addressed to you. The messages immediately before you join
	// are disproportionately ABOUT you joining — the welcome, the handle, the
	// setup. @Smiley lost the Leads room's only message to this within five
	// minutes of arming (#340).
	//
	// 🛑 These apply on the FIRST RUN ONLY. An existing cursor always wins: a
	// watcher's position must not be resettable by a stray flag on a routine
	// tick, or one bad invocation silently replays or skips a day.
	seeded := ""
	// 🛑 REJECT a --since that is not a message id (#357).
	//
	// The server's `since` is message-id based. An RFC3339 timestamp is written
	// to the cursor, announced as "SEEDED … prior history WILL be delivered",
	// and then matches nothing forever: the watcher reports rc=0 and `alive`
	// while delivering NOTHING, permanently. @Smiley found it; the help text
	// claimed timestamps were supported and they never were.
	//
	// ⚠️ This is on the migration path as of today — @Crispin's safe cutover is
	// "new state dir + --since <messageId> seeded from your old cursor", so every
	// bench moving off a wrapper is one keystroke from a silently deaf watcher.
	// Refusing costs a retry; accepting costs a bench that never hears again.
	// 🛑 An ALREADY-POISONED cursor (#357). The guard below catches a bad --since
	// at seed time, but every watcher seeded before this fix has the timestamp
	// sitting in its cursor file RIGHT NOW, and nothing about a later tick
	// re-examines it: it polls, matches nothing, advances nothing, exits 0.
	// @Smiley's census found two on this box, one deaf ~12h.
	//
	// So refuse to poll a cursor we can prove the API cannot consume. Exiting 10
	// every tick is loud and permanent — which is correct, because the condition
	// IS permanent until a human deletes the file. Silence was the whole defect.
	if cursor != "" && !looksLikeMessageID(cursor) {
		return nil, fmt.Errorf(
			"%w: cursor %q in %s is not a message id — this watcher is DEAF and has been\n"+
				"    since it was seeded: the API cannot consume that value, so it matches\n"+
				"    nothing, advances nothing, and reports a clean poll forever (#357).\n"+
				"    Fix: delete the cursor file to re-baseline, or write a real message id\n"+
				"    into it (`pfw room read <entity> --order desc --limit 1` gives you one).",
			ErrConfigRefusal, cursor, cursorPath)
	}

	if cursor == "" && opts.Since != "" && !looksLikeMessageID(opts.Since) {
		return nil, fmt.Errorf(
			"%w: --since %q is not a message id: expected <millis>-<seq>, e.g. 1787423366237-0.\n"+
				"    A timestamp is accepted by the flag but matches nothing — the watcher would\n"+
				"    report healthy and deliver nothing, forever (#357).\n"+
				"    Take the id from your old state dir's cursor file, or from `pfw room read --order desc`.",
			ErrConfigRefusal, opts.Since)
	}

	if cursor == "" && opts.Since != "" {
		if err := writeFileAtomic(cursorPath, opts.Since); err != nil {
			return nil, err
		}
		cursor, seeded = opts.Since, opts.Since
	}
	if cursor == "" && opts.FromStart {
		// Leave the cursor empty and fall through to an ordinary poll — an
		// empty `since` asks the room for everything it has. --limit bounds
		// each tick, so a large room drains over several polls rather than one.
		seeded = "(start)"
	}

	if cursor == "" && seeded == "" {
		newest, err := r.Newest(ctx, opts.EntityType, opts.EntityID)
		if err != nil {
			// 🛑 #335 classified unreachable on the DELTA path and missed this
			// one. A dead backend during the FIRST run reported rc=1 ("failed")
			// instead of rc=12 ("backend unreachable") — so the very first poll
			// of a new watcher, the moment an operator is most likely to be
			// watching, gave the least useful answer. Found because @Tuner
			// reported an rc he could not explain.
			if unreachable(err) {
				return nil, fmt.Errorf("%w: baseline read: %v", ErrUnreachable, err)
			}
			return nil, fmt.Errorf("baseline read: %w", err)
		}
		if err := writeFileAtomic(cursorPath, newest); err != nil {
			return nil, err
		}
		// The marker is what lets `inbox health` tell a new watcher apart from
		// one whose queue was lost. Without it both look like "cursor, no
		// queue" and the pole has to shrug (#311, #316).
		stamp := time.Now().UTC().Format(time.RFC3339) + " " + newest + "\n"
		if err := writeFileAtomic(filepath.Join(opts.StateDir, fileBaseline), stamp); err != nil {
			return nil, err
		}
		_ = touchStamp(opts.StateDir, fileHealthOK)
		return &WatchResult{Baselined: true, CursorTo: newest}, nil
	}

	// ── Ordinary poll ───────────────────────────────────────────────────────
	local, err := compileFilter(opts.Match, opts.From)
	if err != nil {
		return nil, err
	}
	// Local veto too — an older backend ignores unknown query params, so the
	// server-side exclusion fails OPEN. Same reasoning as #324.
	if local.ExcludeFrom, err = compileExcludeFrom(opts.ExcludeFrom); err != nil {
		return nil, err
	}
	// ⚑ Push self-exclusion server-side too when we know our own handle (#336).
	// The local pass below STAYS — an older backend ignores unknown query params
	// and would fail open. Belt and braces: this saves fetching our own posts on
	// a current backend, and changes nothing on a stale one.
	serverExclude := opts.ExcludeFrom
	if self := opts.selfHandle(); self != "" && !opts.IncludeSelf {
		serverExclude = append(append([]string{}, serverExclude...), regexp.QuoteMeta(self))
	}
	msgs, lastID, err := r.ReadSince(ctx, opts.EntityType, opts.EntityID, cursor, opts.Limit,
		opts.Match, opts.From, serverExclude)
	if err != nil {
		// The cursor is NOT advanced and no stamp is written. A failed read must
		// look like a failed read, not a quiet tick.
		if unreachable(err) {
			return nil, fmt.Errorf("%w: room read: %v", ErrUnreachable, err)
		}
		return nil, fmt.Errorf("room read: %w", err)
	}

	res := &WatchResult{CursorFrom: cursor, CursorTo: cursor,
		Fetched: len(msgs), ServerFilter: !local.Empty(), SeededFrom: seeded}

	// ⚠️ The cursor advances past DROPPED messages too — the server keeps
	// lastID as the scan boundary and so do we. That is what makes a filter
	// non-retroactive: change it later and the skipped messages are only
	// recoverable by re-querying the room, which is the system of record.
	// 🛑 The scan boundary is what the cursor follows — never the last KEPT
	// message.
	//
	// With a server-side filter they are different by construction: the server
	// drops non-matching messages but keeps lastID at the furthest id it
	// examined. Advance to the last surviving message instead and every dropped
	// message behind it is re-read on the next tick, forever. A tick that
	// filtered out its entire batch would never advance at all.
	scanBoundary := lastID
	if scanBoundary == "" && len(msgs) > 0 {
		scanBoundary = msgs[len(msgs)-1].ID
	}
	if scanBoundary == "" {
		_ = touchStamp(opts.StateDir, fileHealthOK)
		return res, nil // nothing read at all
	}

	if !local.Empty() {
		kept := msgs[:0]
		for _, m := range msgs {
			if local.Allows(m) {
				kept = append(kept, m)
			}
		}
		// Anything dropped here is traffic the SERVER should have filtered.
		// Non-zero means the backend ignored the parameters — an old build —
		// and the local net just caught a silent fail-open.
		res.ServerMissed = len(msgs) - len(kept)
		msgs = kept
	}

	// Drop our own posts. Applied last and unconditionally, so it holds whether
	// the match ran here or on the server.
	if self := opts.selfHandle(); self != "" && !opts.IncludeSelf {
		kept := msgs[:0]
		for _, m := range msgs {
			if !strings.EqualFold(strings.TrimSpace(m.Agent), self) {
				kept = append(kept, m)
			}
		}
		res.SelfSkipped = len(msgs) - len(kept)
		msgs = kept
	}
	res.Kept = len(msgs)

	if len(msgs) > 0 {
		queuedMsgs, dupes, err := appendQueue(opts.QueuePath, msgs)
		if err != nil {
			return nil, err
		}
		res.Queued, res.Duplicates = len(queuedMsgs), dupes
		res.QueuedMessages = queuedMsgs
	}

	// Cursor advances ONLY after the batch is durably on disk. Advance first
	// and a crash loses the batch while the room considers it delivered — the
	// one failure with no recovery path. A fully-filtered batch has nothing to
	// lose, so advancing past it is safe and necessary.
	if err := writeFileAtomic(cursorPath, scanBoundary); err != nil {
		return nil, err
	}
	res.CursorTo = scanBoundary

	_ = touchStamp(opts.StateDir, fileHealthOK)
	return res, nil
}

// checkIdentity enforces that this watcher's reads are distinguishable from
// every other watcher's. The two modes check different things, so neither is
// redundant: individual mode proves it by account, shared mode by handle.
func checkIdentity(ctx context.Context, r RoomReader, opts WatchOptions) error {
	if opts.Shared {
		// whoami is BLIND to the shared-token failure: every bench authenticates
		// as the same principal, so the assertion would pass for all of them
		// while they collide on one cursor (#263). A handle is the only thing
		// that separates them.
		if strings.TrimSpace(opts.Handle) == "" {
			return fmt.Errorf("%w: --shared requires a non-empty --handle, "+
				"or this watcher's reads cannot be told apart from any other bench's", ErrIdentityGuard)
		}
		return nil
	}
	if opts.ExpectEmail == "" {
		return fmt.Errorf("%w: choose --expect-email <addr> (individual account) or --shared --handle <name>",
			ErrIdentityGuard)
	}
	got, err := r.WhoAmI(ctx)
	if err != nil {
		// ⚠️ Transport failure is NOT an identity problem. Reporting it as one
		// is how an outage got classified as a config error and exited 10.
		if unreachable(err) {
			return fmt.Errorf("%w: %v", ErrUnreachable, err)
		}
		return fmt.Errorf("%w: cannot verify identity: %v", ErrIdentityGuard, err)
	}
	if !strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(opts.ExpectEmail)) {
		return fmt.Errorf("%w: token authenticates as %q, expected %q — refusing to poll",
			ErrIdentityGuard, got, opts.ExpectEmail)
	}
	return nil
}

// appendQueue appends messages not already present, by id.
//
// Dedupe is against the whole queue because a re-delivered batch must be safe:
// a crash between the append and the cursor write replays the same messages, and
// that retry has to be idempotent.
// ⚑ It returns WHAT it queued, not just how many (#366). The emit layer used
// to recover the bodies by tailing the queue file — `tail -n "$n"` — which
// re-parses lines that were already in hand and makes a malformed line a
// silent message loss. Handing the messages back removes that class rather
// than guarding it.
func appendQueue(path string, msgs []Message) (queuedMsgs []Message, dupes int, err error) {
	seen, err := existingIDs(path)
	if err != nil {
		return nil, 0, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, 0, fmt.Errorf("create queue dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, 0, fmt.Errorf("open queue: %w", err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, m := range msgs {
		if m.ID == "" || seen[m.ID] {
			dupes++
			continue
		}
		if err := enc.Encode(m); err != nil {
			return queuedMsgs, dupes, fmt.Errorf("append queue: %w", err)
		}
		seen[m.ID] = true
		queuedMsgs = append(queuedMsgs, m)
	}
	// fsync before the caller advances the cursor. Without it the cursor can
	// outrun data still sitting in the page cache.
	if err := f.Sync(); err != nil {
		return queuedMsgs, dupes, fmt.Errorf("sync queue: %w", err)
	}
	return queuedMsgs, dupes, nil
}

func existingIDs(path string) (map[string]bool, error) {
	seen := map[string]bool{}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return seen, nil
		}
		return nil, fmt.Errorf("read queue: %w", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m struct {
			ID string `json:"id"`
		}
		if json.Unmarshal([]byte(line), &m) == nil && m.ID != "" {
			seen[m.ID] = true
		}
	}
	return seen, nil
}

// acquireLock stops two instances interleaving appends and cursor writes.
//
// A stale lock is reclaimed rather than wedging the watcher forever: a process
// killed mid-poll would otherwise stop all future polling, which is the
// instrument disabling the thing it measures.
func acquireLock(stateDir string) (func(), error) {
	path := filepath.Join(stateDir, fileLock)
	if fi, err := os.Stat(path); err == nil {
		if time.Since(fi.ModTime()) > staleLockAge {
			_ = os.Remove(path)
		} else {
			return nil, fmt.Errorf("%w (lock held since %s)", ErrLocked,
				fi.ModTime().UTC().Format(time.RFC3339))
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("take lock: %w", err)
	}
	_, _ = f.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	_ = f.Close()
	return func() { _ = os.Remove(path) }, nil
}

func readTrimmed(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func touchStamp(dir, name string) error {
	return writeFileAtomic(filepath.Join(dir, name), time.Now().UTC().Format(time.RFC3339)+"\n")
}

// touchArm records that a poll STARTED, rotating the previous arm stamp aside
// first so the observed interval between polls stays derivable (#367).
//
// ⚠️ Errors are deliberately swallowed. This is diagnostic state, and a watcher
// that refuses to poll because it could not write a diagnostic file has turned
// an observability feature into an outage. The absence of the stamp is itself
// reported by `room health`, so a failure here is visible rather than silent.
// StampVersion records which binary is running this leg. Called every tick, so
// it survives a re-exec (#371) and answers "what is ACTUALLY running" without
// parsing any stream.
func StampVersion(dir, version string) {
	if dir == "" || version == "" {
		return
	}
	_ = writeFileAtomic(filepath.Join(dir, fileVersion), version+"\n")
}

func touchArm(dir string) {
	if prev := readTrimmed(filepath.Join(dir, fileHealthArm)); prev != "" {
		_ = writeFileAtomic(filepath.Join(dir, fileHealthArmPrev), prev+"\n")
	}
	_ = touchStamp(dir, fileHealthArm)
}

func writeFileAtomic(path, body string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return fmt.Errorf("stage %s: %w", filepath.Base(path), err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString(body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", filepath.Base(path), err)
	}
	return os.Rename(tmp.Name(), path)
}

// looksLikeMessageID reports whether s has the room stream's id shape:
// <millis>-<seq>, e.g. 1787423366237-0.
//
// Deliberately shape-only — it does not check the id EXISTS. A wrong-but-valid
// id seeds from the wrong point, which is visible in the first tick's queued
// count; a wrong-SHAPED id is silently permanent, which is the one worth
// refusing.
func looksLikeMessageID(s string) bool {
	// One or two all-digit parts: "0" and "<millis>" are valid (the sequence
	// defaults to 0), and "<millis>-<seq>" is the full form. Two dashes is a
	// date — "2026-08-22" — which is exactly what we are here to refuse.
	parts := strings.Split(s, "-")
	if len(parts) == 0 || len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}
