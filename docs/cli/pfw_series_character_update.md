## pfw series character update

Update a character

### Synopsis

Update a character. Only the flags you pass are changed.

```
pfw series character update <series-id> <slug> [flags]
```

### Options

```
  -h, --help             help for update
      --name string      New name
      --profile string   New profile (markdown)
      --role string      New role
      --status string    New status
      --stdin            Read the profile from stdin
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

* [pfw series character](pfw_series_character.md)	 - Series characters

