## pfw room watch

Buffer new room messages to a durable local queue

### Synopsis

Read a room since this watcher's own cursor and append what is new.

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
recipe that treated 10 as benign silently ignored outages.

```
pfw room watch <entity-id> [flags]
```

### Options

```
      --alarm-after int            Alarm on the Nth CONSECUTIVE failing tick. 1 catches everything; raise it on a noisy environment where short bounces are routine — but a high value makes the alarm itself hard to ever observe. (default 1)
      --exclude-from stringArray   VETO a sender: drop their messages even if --match/--from kept them. REPEATABLE. Mutes volume WITHOUT narrowing your gate.
      --expect-email string        Refuse to poll unless the token authenticates as this address
      --from stringArray           Also keep messages FROM a sender matching this regex. REPEATABLE, OR'd with every --match.
      --from-start                 FIRST RUN ONLY: take the whole room instead of baselining. Bounded by --limit per tick.
      --handle string              This watcher's handle (required with --shared)
  -h, --help                       help for watch
      --include-quoted             Also queue messages whose only --match hit is inside quoted code. Off by default: a routing pattern shown as code is the SUBJECT of a message, not its address (#417)
      --include-self               Queue your own posts too (default: skipped — a watcher should not wake you with your own words)
      --interval duration          Sleep between ticks WHEN HEALTHY (e.g. 30m). Refused without --loop. A FAILING tick costs interval + ~30s, and the leg then looks dead until it recovers — so ONE failed poll shows as ~2 intervals since success. Size --max-age at 2x this or more.
      --limit int                  Max messages to SCAN per poll, not to keep. A filter can keep 0 of a full scan and still advance. (default 60)
      --loop                       Poll repeatedly instead of exiting after one tick. Requires --interval.
      --match stringArray          Keep messages whose content/target match this RE2 regex (case-insensitive). REPEATABLE.
      --no-reexec                  Pin the binary image: do NOT re-exec when the executable changes on disk. Default is to pick up rebuilds at the next tick, matching the shell loops --loop replaced.
      --queue string               JSONL queue to append to (required)
      --shared                     Shared-token mode: skip the whoami assertion, require --handle
      --since string               FIRST RUN ONLY: start from this MESSAGE ID (<millis>-<seq>) instead of baselining. Not a timestamp — a timestamp is refused (#357). Ignored once a cursor exists.
      --state string               State dir: cursor, lock, health stamps, baseline marker (required)
      --telemetry string           Where --loop writes per-tick telemetry (default: <state>/telemetry.log). Rotated at 5 MiB. Failures still go to stderr.
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

