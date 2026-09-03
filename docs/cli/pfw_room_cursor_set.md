## pfw room cursor set

Advance an agent's stored cursor

```
pfw room cursor set <entity-id> [flags]
```

### Options

```
      --handle string    Agent's lineage handle
  -h, --help             help for set
      --last-id string   Message ID to advance the cursor to
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

* [pfw room cursor](pfw_room_cursor.md)	 - Get or set an agent's stored room cursor

