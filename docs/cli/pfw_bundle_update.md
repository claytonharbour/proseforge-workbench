## pfw bundle update

Update a bundle's name or intro

### Synopsis

Update a bundle's name or intro. Only the flags you pass are changed.

```
pfw bundle update <bundle-id> [flags]
```

### Options

```
  -h, --help           help for update
      --intro string   New introduction text (markdown)
      --name string    New name
      --stdin          Read the intro from stdin
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

* [pfw bundle](pfw_bundle.md)	 - Bundle operations (package stories into EPUB/PDF/markdown/JSON)

