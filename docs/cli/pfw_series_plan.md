## pfw series plan

Seed a Story Forge planning session for the next book in a series

```
pfw series plan <series-id> [flags]
```

### Options

```
      --book-number int      Override book number (omit to auto-detect next) (default -1)
      --characters strings   Character slugs to include (default: all)
  -h, --help                 help for plan
      --notes string         Author notes injected into AI context
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

* [pfw series](pfw_series.md)	 - Series operations

