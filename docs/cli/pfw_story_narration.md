## pfw story narration

Narration management

### Options

```
  -h, --help   help for narration
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
* [pfw story narration acknowledge](pfw_story_narration_acknowledge.md)	 - Acknowledge narration staleness (clear the stale flag)
* [pfw story narration cancel](pfw_story_narration_cancel.md)	 - Cancel a specific section's narration
* [pfw story narration delete](pfw_story_narration_delete.md)	 - Delete all narration data for a story
* [pfw story narration drift](pfw_story_narration_drift.md)	 - Cheap drift check — just the drift envelope from narration status
* [pfw story narration patch](pfw_story_narration_patch.md)	 - Patch multiple segments/sections with voice changes and rebuild
* [pfw story narration rebuild](pfw_story_narration_rebuild.md)	 - Rebuild audiobook from existing per-section audio
* [pfw story narration regenerate](pfw_story_narration_regenerate.md)	 - Regenerate narration for a specific section
* [pfw story narration regenerate-stale](pfw_story_narration_regenerate-stale.md)	 - Re-narrate only the sections whose content drifted
* [pfw story narration resume](pfw_story_narration_resume.md)	 - Resume a stuck narration
* [pfw story narration retry](pfw_story_narration_retry.md)	 - Retry a failed/stuck section narration
* [pfw story narration segment-regenerate](pfw_story_narration_segment-regenerate.md)	 - Regenerate a single segment's audio
* [pfw story narration segments](pfw_story_narration_segments.md)	 - List segments for a section with text content and voice info
* [pfw story narration status](pfw_story_narration_status.md)	 - Get narration status for a story
* [pfw story narration voices](pfw_story_narration_voices.md)	 - List available TTS voices

