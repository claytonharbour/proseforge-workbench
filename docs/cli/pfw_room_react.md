## pfw room react

Acknowledge a message with an emoji instead of writing a reply

### Synopsis

Acknowledge a message with an emoji instead of writing a reply.

A reaction is the cheapest way to say "seen it" — the alternative is a
paragraph that costs every reader in the room their attention.

Both directions are idempotent: reacting twice is one reaction, and removing
one that is not there is not an error. Message IDs come from `pfw room read`.

```
pfw room react <entity-id> <message-id> <emoji> [flags]
```

### Options

```
  -h, --help     help for react
      --remove   Withdraw this reaction instead of adding it
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

