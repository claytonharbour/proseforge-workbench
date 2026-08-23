## pfw listing update

Update a store listing

### Synopsis

Update a store listing. Only the flags you pass are changed.

```
pfw listing update <listing-id> [flags]
```

### Options

```
      --format string        New format
  -h, --help                 help for update
      --listing-url string   New store listing URL
      --status string        New status
      --store string         New store
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

* [pfw listing](pfw_listing.md)	 - External store listings (Amazon, Apple Books, etc.) for a story

