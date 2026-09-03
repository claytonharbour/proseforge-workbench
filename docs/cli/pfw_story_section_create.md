## pfw story section create

Create a new section in a story

```
pfw story section create <story-id> [flags]
```

### Options

```
      --content string   Initial content (optional)
  -h, --help             help for create
      --name string      Section name (e.g., "Chapter 1")
      --order int        Position to insert at (0-indexed) (default -1)
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

