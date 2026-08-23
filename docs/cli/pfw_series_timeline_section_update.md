## pfw series timeline section update

Create or update a timeline section

```
pfw series timeline section update <series-id> <slug> [flags]
```

### Options

```
      --content string   Section content (markdown)
  -h, --help             help for update
      --stdin            Read the section content from stdin
      --title string     Section title
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

* [pfw series timeline section](pfw_series_timeline_section.md)	 - A single timeline section (per-book events, by slug)

