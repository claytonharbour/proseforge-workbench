## pfw feedback section update

Rewrite a section's content (reads from stdin with --stdin)

```
pfw feedback section update <story-id> <review-id> <section-id> [flags]
```

### Options

```
      --content string   Section content (for short content; prefer --stdin for full sections)
  -h, --help             help for update
      --stdin            Read section content from stdin
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

* [pfw feedback section](pfw_feedback_section.md)	 - Section-level feedback operations

