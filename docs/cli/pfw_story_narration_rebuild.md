## pfw story narration rebuild

Rebuild audiobook from existing per-section audio

```
pfw story narration rebuild <story-id> [flags]
```

### Options

```
      --emphasis                Update the stored emphasis narration option (#642). Omit the flag to leave it untouched; takes effect on the next re-TTS, not this rebuild.
  -h, --help                    help for rebuild
      --section-announcements   Insert TTS-generated section title announcements
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

