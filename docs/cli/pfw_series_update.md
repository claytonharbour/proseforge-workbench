## pfw series update

Update a series' metadata

### Synopsis

Update a series' metadata. Only the flags you pass are changed.

```
pfw series update <series-id> [flags]
```

### Options

```
      --description string   New description
      --genre-id string      New genre ID
  -h, --help                 help for update
      --name string          New name
      --tone-id string       New tone ID
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

