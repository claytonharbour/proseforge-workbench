# Narration Patching Guide

How to fix audio issues in ProseForge audiobooks without re-narrating entire stories.

## When to Use This

- A word is mispronounced (e.g., Kokoro says "brethed" instead of "breathed")
- An acronym isn't spoken correctly (e.g., "DNA" read as a word, not spelled out)
- A segment has audio artifacts or glitches
- You want a different voice for a specific passage

## Concepts

ProseForge narration has three levels:

```
Narration (story-level)
  └── Chapters (one per section)
       └── Segments (chunks of ~3000 characters each)
```

You can fix audio at any level:
- **Segment** — re-narrate one chunk (~30 seconds of audio). Cheapest, most precise.
- **Chapter** — re-narrate an entire section. Use when segment boundaries shifted.
- **Story** — delete and re-narrate everything. Nuclear option.

## Segment Patching Workflow

### Step 1: Find the problem chapter

```bash
pfw story narration status <story-id>
```

This lists all chapters with their IDs and status. Note the chapter ID for the
section containing the problem.

### Step 2: List segments and find the bad one

```bash
pfw story narration segments <story-id> <chapter-id>
```

Returns each segment with:
- `id` — segment ID (needed for regeneration)
- `text` — the exact text that was sent to TTS
- `voice` — which voice narrated it
- `provider` — which TTS provider was used
- `content_changed` — whether the story text has changed since narration

Search the `text` field for the mispronounced word to identify which segment
needs fixing.

### Step 3: Regenerate the segment with a different voice

```bash
pfw story narration segment-regenerate <story-id> <chapter-id> <segment-id> --voice Kore
```

This:
1. Deletes the old segment audio
2. Re-narrates just that segment with the specified voice
3. Restitches all segments into the chapter audio
4. Reassembles the audiobook automatically

**Voice selection:** Use `pfw story narration voices` to list available voices.
Common choices:
- `Kore` — Gemini female (good for fixing Kokoro pronunciation issues)
- `Puck` — Gemini male
- `af_sarah` — Kokoro female (default)
- `am_adam` — Kokoro male

### Step 4: Verify

```bash
pfw story narration segments <story-id> <chapter-id>
```

The regenerated segment should now show the new voice and provider.

## MCP Tool Workflow

For AI agents using the MCP server:

```
narration_status {story_id}           → find chapter with the problem
narration_segments {story_id, chapter_id}  → find the segment with bad text
narration_segment_regenerate {story_id, chapter_id, segment_id, voice: "Kore"}
narration_status {story_id}           → poll until complete
```

## Examples

### Fix a mispronounced word

Kokoro says "brethed" instead of "breathed":

```bash
# Find which chapter has "breathed"
pfw story export <story-id> | grep -n "breathed"

# List segments for that chapter
pfw story narration segments <story-id> <chapter-id>

# Find the segment containing "breathed" and regenerate with Gemini
pfw story narration segment-regenerate <story-id> <chapter-id> <segment-id> --voice Kore
```

### Fix an acronym

Kokoro doesn't pronounce "DNA" as an acronym:

```bash
# Same flow — find the segment, regenerate with a voice that handles acronyms
pfw story narration segment-regenerate <story-id> <chapter-id> <segment-id> --voice Kore
```

## Cost

Each segment regeneration costs **2 credits** (same as any TTS operation).
Use `pfw story credits` to check your balance before and after.

## Limitations

- **Content must not have changed.** If you edited the story text after narration,
  segment boundaries may have shifted. The endpoint will reject with
  `content_changed: true`. Use chapter-level regeneration instead.
- **Voice mismatch.** Replacing one segment with a different voice will sound
  different from surrounding segments. This is a trade-off — correct pronunciation
  vs consistent voice. For critical fixes, consider regenerating the entire chapter.
- **One segment at a time.** There's no batch segment regeneration. If multiple
  segments need fixing, regenerate each one individually or use chapter-level
  regeneration.

## Other Narration Operations

| Command | Purpose |
|---------|---------|
| `story narrate <story-id>` | Start full narration |
| `story narration regenerate <story-id> <chapter-id> --force --voice Kore` | Re-narrate entire chapter |
| `story narration rebuild --chapter-announcements` | Reassemble audiobook with chapter titles |
| `story narration retry <story-id> <chapter-id>` | Retry a failed chapter |
| `story narration delete <story-id>` | Delete narration and start fresh |
| `story narration resume <story-id>` | Resume stuck narration |
| `story credits` | Check credit balance |
