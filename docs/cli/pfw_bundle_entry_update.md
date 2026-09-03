## pfw bundle entry update

Update an entry's transition text

```
pfw bundle entry update <bundle-id> <entry-id> [flags]
```

### Options

```
      --comment string   New transition text (markdown)
  -h, --help             help for update
      --stdin            Read the transition text from stdin
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

* [pfw bundle entry](pfw_bundle_entry.md)	 - Manage bundle entries

