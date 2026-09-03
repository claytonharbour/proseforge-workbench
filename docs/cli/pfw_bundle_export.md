## pfw bundle export

Export a bundle as epub, json, markdown, or pdf

```
pfw bundle export <bundle-id> [flags]
```

### Options

```
      --format string   Export format: epub, json, markdown, pdf
  -h, --help            help for export
      --out string      Write to this file instead of stdout (recommended for epub/pdf)
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

