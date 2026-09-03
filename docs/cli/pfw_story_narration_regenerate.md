## pfw story narration regenerate

Regenerate narration for a specific section

```
pfw story narration regenerate <story-id> <section-id> [flags]
```

### Options

```
      --force          Regenerate even if content hasn't changed
  -h, --help           help for regenerate
      --voice string   Voice override for this section (e.g., Puck, af_sarah)
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

* [pfw story narration](pfw_story_narration.md)	 - Narration management

