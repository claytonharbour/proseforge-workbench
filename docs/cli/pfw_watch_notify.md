## pfw watch notify

Run an adapter against the wake contract and report the outcome

### Synopsis

Invoke a notify adapter and classify what came back.

This is the seam a watcher wakes a worker through, exposed on its own so an
adapter can be verified against the contract without running a whole watcher —
and without waiting for a real room message to arrive.

THE CONTRACT an adapter must satisfy:

  argv    $1 batch file   $2 prompt file   $3 active-work file ("" if none)
  stdout  the worker's reply, or the literal NO_POST
  exit    0  ok — the handoff was accepted
          10 CANNOT WAKE — no runtime, no credential, nothing to talk to
          1  woke and broke

Exit 10 must never be folded into 1. A deaf bench and a broken bench are
different problems, and only one of them announces itself: broken is loud by
nature, deaf stays silent and silence looks like calm.

Exit 0 means the handoff was accepted. It does NOT acknowledge anything — only
the consumer's own explicit ack retires a message.

```
pfw watch notify [flags]
```

### Options

```
      --active-work string   Active-work file handed to the adapter as $3
      --adapter string       Adapter command to run (empty = buffering-only)
      --batch string         Batch file handed to the adapter as $1 (required)
      --env stringArray      Extra environment for the adapter, KEY=VALUE (repeatable)
  -h, --help                 help for notify
      --prompt string        Prompt file handed to the adapter as $2
      --timeout duration     Bound on the wake (default 10m0s)
      --wake-id string       Identifier for this wake (reaches the adapter as WATCHER_WAKE_ID)
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

* [pfw watch](pfw_watch.md)	 - Room watcher: buffer messages and hand them to a worker

