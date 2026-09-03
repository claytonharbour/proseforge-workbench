## pfw story section write

Update a section's content or name

### Synopsis

Update a section's content or name. Use --stdin to read content from stdin, --content for short text, and --name to rename without rewriting.

```
pfw story section write <story-id> <section-id> [flags]
```

### Options

```
      --content string   Content to write (for short text)
  -h, --help             help for write
      --name string      New section name
      --stdin            Read content from stdin
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

* [pfw story section](pfw_story_section.md)	 - Section operations

