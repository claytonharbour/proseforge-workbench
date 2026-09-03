## pfw room watchdog

Alarm when a watcher stops polling — and when it recovers

### Synopsis

Check a watcher's liveness and speak ONLY when the answer changes.

Single-shot: one check, one comparison, exit. A separate monitor, a cron, or a
human decides when it runs. It never schedules itself.

⚠️ RUN IT AS A SEPARATE PROCESS FROM YOUR POLL LOOP. That is the whole
mechanism. A sequential loop with a HUNG 'room watch' stalls every leg and emits
nothing, because a hang never produces a non-zero exit code — so a watchdog
living inside that loop dies with it and stays silent.

WHY NOT JUST 'room health' IN A LOOP:

  It compares the STATE WORD, not the exit code. Exit 3 covers stale, never AND
  unknown, so a watchdog built on the code alone cannot tell "check your config"
  from "check whether the process is running".

  It speaks only on a TRANSITION. Repeating every tick while stale floods the
  channel you are trying to alert on, and the alarm gets muted.

  It alarms on RECOVERY too. An alarm that goes red and never green is a stuck
  alarm, and it looks identical to a permanent outage.

A first run finding a healthy watcher is SILENT — announcing "alive" at startup
teaches the reader this channel carries routine noise. A first run finding an
alarm does speak; that is news.

🛑 IT ONLY UNDERSTANDS 'pfw room watch' STATE DIRS, and you must name them
explicitly. Point it at a watcher built some other way and it reports "never" —
a FALSE ALARM about a healthy process.

  Never auto-discover these directories. Scanning the filesystem for the stamp
  file finds only watchers that use this tool, and reports every other kind as
  dead. That is backwards: the watchers least like yours are the ones a watchdog
  most needs to be honest about, and a sweep like that raises its loudest alarm
  about a bench that is perfectly fine while staying silent about one that is
  genuinely gone but leaves no directory at all.

🛑 WHAT IT DOES NOT COVER: the session ending, or every monitor dying at once.
Same session, dies with them, says nothing. That is an accepted trade, not a
closed gap.

🛑 THE EXIT CODE DESCRIBES THIS WATCHDOG, NOT THE WATCHER IT WATCHES. Conflating
those is the category error this whole area keeps making.

  0  the watchdog did its job — INCLUDING when it found the watcher dead
  1  the watchdog could NOT do its job (misconfigured, cannot write its state)

A detected outage exits 0 on purpose. Returning non-zero there tells a supervisor
the check failed, so it restarts the one process that was working correctly and
throws away the state-change memory that stops the alarm repeating.

Misuse is LOUD and goes to STDOUT, never quietly to stderr. A silent broken
watchdog is indistinguishable from a healthy fleet, which is the exact failure
this command exists to prevent — so it says "NOTHING is being checked" rather
than nothing at all.

'pfw room health' is the gate and exits 0/3 on the WATCHER's state. This is the
reporter.

```
pfw room watchdog <watched-state-dir> [flags]
```

### Options

```
  -h, --help               help for watchdog
      --max-age duration   Presume dead after this long without a successful poll. A small multiple of the poll interval. (default 5m0s)
      --state string       The WATCHDOG's own state dir, separate from the watcher's (required)
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

