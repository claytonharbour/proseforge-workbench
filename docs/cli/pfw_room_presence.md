## pfw room presence

Who has READ this room, and how long ago

### Synopsis

Who has READ this room, and how long ago.

⛔ This is NOT an online/offline list and does not report one. Watchers in a
fleet poll at different cadences — 60s, 30m and 1h are all in use — so a
40-minute age is healthy for an hourly leg and stale for a per-minute one.
There is no threshold this command could apply that would be right for both,
so it reports the age and leaves the judgement to you.

```
pfw room presence <entity-id> [flags]
```

### Options

```
  -h, --help   help for presence
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

