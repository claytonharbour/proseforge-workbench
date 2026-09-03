## pfw story narration patch

Patch multiple segments/sections with voice changes and rebuild

### Synopsis

Batch re-voice segments and sections in one operation. Rebuilds audiobook once when done.
Use --segment sectionID:segmentID:voice (repeatable) and --section sectionID:voice (repeatable).

```
pfw story narration patch <story-id> [flags]
```

### Options

```
  -h, --help                  help for patch
      --section stringArray   Section to patch: sectionID:voice (repeatable)
      --segment stringArray   Segment to patch: sectionID:segmentID:voice (repeatable)
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

