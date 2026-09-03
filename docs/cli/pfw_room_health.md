## pfw room health

Is this watcher still polling?

### Synopsis

Report whether a watcher's last poll actually succeeded, and how long ago.

A watcher that has STOPPED looks exactly like a room with nothing to report.
Every quiet tick is indistinguishable from death, so the failure is invisible
precisely while it is happening.

This asks the health-poll stamp, which 'room watch' refreshes on every
SUCCESSFUL poll and never on a failure. So its age answers "when did a poll last
actually succeed" without depending on what any exit code means — which matters,
because enumerating exit codes is a losing game and a guard built on a stale
enumeration is silently wrong.

STATES, and they are deliberately not one boolean. "never" and "stale" need
different fixes — check your config, versus check whether the process is
running — so they must not collapse into "unhealthy":
  alive    a poll succeeded within --max-age
  failing  polls ARE being attempted and none succeeds — the process is fine,
           the backend or the credentials are not. Do NOT restart it.
  stale    nothing is polling at all — THE failure case
  never    attempts recorded, no success ever
  unknown  no pfw watcher state here, or the dir cannot be read

Exit: 0 alive · 3 anything else. The code is also PRINTED on line one, because
'health | head -1' returns HEAD's status, so an rc-only caller behind a pipe
reads stale, never and unreadable alike as alive.

🛑 THAT LAST READER IS STILL UNREACHABLE and no change here can fix it — $? is
shell semantics, not ours to grant. Printing rc into the payload rescues the
caller who reads the OUTPUT; a caller who reads only $? behind a pipe still
gets 0. If you must pipe, take the real code explicitly:

  pfw room health "$STATE" --max-age 5m | head -1
  rc=${PIPESTATUS[0]}        # bash/zsh — NOT $?, which is head's

  # or simply do not pipe:
  out=$(pfw room health "$STATE" --max-age 5m); rc=$?

```
pfw room health <state-dir>[:max-age]... [flags]
```

### Options

```
      --check-cadence duration   One-shot only: evaluate this leg's threshold as if a watchdog polled it every X (e.g. 2m), and report whether a real breach could fall between samples (#400). Implied by --loop's own --interval.
  -h, --help                     help for health
      --interval duration        Sleep between checks in --loop mode (e.g. 5m). Refused without --loop.
      --loop                     Check repeatedly instead of exiting. Requires --interval.
      --max-age duration         Default: presume dead after this long without a successful poll. MUST be at least 2x your poll interval: one failed poll leaves a leg looking dead for ~2 intervals, because it sleeps a full interval before retrying. Budget 2x interval + 60s to survive ONE failed poll; measured overhead is 46-57s per failure cycle and is FIXED, not a multiple of the interval. Override PER LEG with <dir>:<duration> — a 30m poller and a 60s poller cannot share a threshold. (default 5m0s)
      --no-reexec                Do not replace this process when the binary changes. The default swap is how a fix reaches a long-lived watchdog at all.
      --state string             Remember each leg's last state HERE, so a restart does not re-alarm about an unchanged failure. Without it, state is in-memory and lost on restart.
      --telemetry string         Where --loop writes per-cycle telemetry (default: <state>/telemetry.log). Rotated at 5 MiB.
```

### Options inherited from parent commands

```
      --credentials-file string   Path to a credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read per invocation, so a rotated key applies immediately
      --debug                     Enable debug logging
  -o, --output string             Output format: table, json, brief (default "table")
      --token string              API token (env: PROSEFORGE_TOKEN). Accepts a quoted env reference, e.g. --token '${PROSEFORGE_TOKEN}', which keeps the key out of argv
      --type string               Entity type: story, series, bundle, or conversation (a standalone room with no story or series attached) (default "story")
      --url string                API base URL (env: PROSEFORGE_URL). Accepts a quoted env reference, e.g. --url '${PROSEFORGE_URL}'
```

### SEE ALSO

* [pfw room](pfw_room.md)	 - Coordination room operations (read/send/status/archive)

