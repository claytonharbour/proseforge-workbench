## pfw active-work set

Create or update the record

```
pfw active-work set <path> [flags]
```

### Options

```
      --blocker stringArray    Something blocking progress (repeatable)
      --clear-blockers         Drop existing blockers before applying --blocker
      --clear-evidence         Drop existing evidence before applying --evidence
      --evidence stringArray   Pointer to evidence (repeatable)
  -h, --help                   help for set
      --holder string          Write through a lease you hold
      --next string            The next CONCRETE action — a thing someone can do, not a status
      --objective string       What is being accomplished
      --owner string           Bench that owns this work
      --ticket string          Ticket URL this work is authorized by
      --version int            Expected current version; refuses if it has moved (-1 skips the check) (default -1)
      --work-id string         Identifier for this piece of work
```

### Options inherited from parent commands

```
      --credentials-file string   Path to a credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read per invocation, so a rotated key applies immediately
      --debug                     Enable debug logging
  -o, --output string             Output format: table, json, brief (default "table")
      --token string              API token (env: PROSEFORGE_TOKEN). Accepts a quoted env reference, e.g. --token '${PROSEFORGE_TOKEN}', which keeps the key out of argv
      --url string                API base URL (env: PROSEFORGE_URL). Accepts a quoted env reference, e.g. --url '${PROSEFORGE_URL}'
```

### SEE ALSO

* [pfw active-work](pfw_active-work.md)	 - Standing task that survives a session boundary

