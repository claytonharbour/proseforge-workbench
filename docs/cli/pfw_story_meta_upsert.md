## pfw story meta upsert

Write a story planning document (story | characters | plot)

### Synopsis

Write story planning data as markdown. --type selects which document; use --stdin for large content.

```
pfw story meta upsert <story-id> [flags]
```

### Options

```
      --content string   Planning data content (markdown)
  -h, --help             help for upsert
      --stdin            Read the content from stdin
      --type string      Document type: story, characters, or plot
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

* [pfw story meta](pfw_story_meta.md)	 - Story planning data (premise, characters, plot outline)

