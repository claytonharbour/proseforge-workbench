## pfw room roster

What watcher legs exist under these directories, and is each still polling?

### Synopsis

Scan directories for watcher state and report every leg found.

'room health' asks about a state dir you NAME. This finds the ones you would not
think to name — which is the only kind that goes unnoticed, because a leg nobody
remembers is a leg nobody checks.

Each leg's threshold is DERIVED from its own observed poll interval rather than
declared. That is the property that matters: a forgotten leg has no declaration
left, but its state dir still records how often it used to poll, so "was polling
every 60s, has not polled in 13h" is computable from disk alone.

A leg is reported STALE when its silence exceeds 10x its own observed interval —
far outside any failure-and-retry envelope, so it means stopped rather than
struggling. The age and interval are both printed; a leg at 3x is visible in the
table before it trips.

Exit: 0 when every leg is alive or deliberately retired, 3 when any is not.

⚠️ It reports. It does not restart anything. Legs are retired on purpose all the
time, and an auto-restart would fight every deliberate retirement.

```
pfw room roster <dir>... [flags]
```

### Options

```
      --all                  Include retired legs in the table (they are counted either way, never a fault).
  -h, --help                 help for roster
      --stale-multiple int   Report STALE when silence exceeds this many times the leg's own observed interval. (default 10)
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

