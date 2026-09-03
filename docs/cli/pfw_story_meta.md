## pfw story meta

Story planning data (premise, characters, plot outline)

### Options

```
  -h, --help   help for meta
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

* [pfw story](pfw_story.md)	 - Story operations
* [pfw story meta acknowledge](pfw_story_meta_acknowledge.md)	 - Acknowledge meta staleness (clear the stale flag)
* [pfw story meta get](pfw_story_meta_get.md)	 - Read all story planning data (story / characters / plot)
* [pfw story meta stale](pfw_story_meta_stale.md)	 - List sections affected by meta changes since last generation
* [pfw story meta upsert](pfw_story_meta_upsert.md)	 - Write a story planning document (story | characters | plot)

