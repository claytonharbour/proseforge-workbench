## pfw room stats

Per-author message and character volume for a room

### Synopsis

Group a room's messages by author and report count, characters and average length.

Computed entirely client-side from `room read`, so it needs no server support and counts your OWN messages — which `room watch` skips by default, and which is exactly the half you cannot otherwise see.

Characters are RUNES, not bytes: these rooms are full of box-drawing and emoji, and a byte count ranks formatting style rather than volume.

```
pfw room stats <entity-id> [flags]
```

### Options

```
      --by-day         Also break the totals down by day
      --from string    Only count messages from this sender
  -h, --help           help for stats
      --limit int      Max messages to SCAN (0 = server default). This bounds the window the numbers describe.
      --order string   Sort order to SCAN in: desc (newest first, the default here) or asc (default "desc")
      --since string   Only count messages after this message ID
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

