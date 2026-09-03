## pfw contribution resolve

Accept or reject one suggestion

### Synopsis

Move one suggestion's status.

The status vocabulary is set by the API and is being simplified upstream, so it
is passed through rather than checked here — an unknown one comes back as an
API error naming it.

```
pfw contribution resolve <story-id> <contribution-id> <suggestion-id> [flags]
```

### Options

```
  -h, --help            help for resolve
      --status string   New status (required)
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

* [pfw contribution](pfw_contribution.md)	 - Work on a story you don't own, and review work on one you do

