## pfw story narrate

Start narration generation for a story

```
pfw story narrate <story-id> [flags]
```

### Options

```
      --confirm-replace   Required on a story with an existing complete/error narration: forwards the backend force flag (#477) to DELETE the existing track and re-render everything from scratch (full credit burn). Without it the backend 409s narration_exists. No effect on an in-flight narration (409 conflict — wait or cancel first). To change voice non-destructively, regenerate sections with --voice then 'story narration rebuild'.
      --emphasis          Pin the TTS worker's emphasis-pause prosody for this narration (#642, RunPod Kokoro only). Omit the flag to use the worker default; --emphasis=false pins it off.
  -h, --help              help for narrate
      --voice string      TTS voice name (e.g., 'Kore'). Omit for server default.
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

