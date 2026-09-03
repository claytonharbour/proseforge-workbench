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
  └── Sections (one per story section)
       └── Segments (chunks of ~3000 characters each)
```

You can fix audio at any level:
- **Segment** — re-narrate one chunk (~30 seconds of audio). Cheapest, most precise.
- **Section** — re-narrate an entire section. Use when segment boundaries shifted.
- **Story** — delete and re-narrate everything. Nuclear option.

> The assembled M4B audiobook itself contains chapter markers (one per section) so listeners see "Chapter N" while playing. That's audiobook-format terminology and is correct. The narration pipeline operates on **sections** — the same units you write to with `section_write`.

## Before You Start: Check the State

If the story has had heavy edits since narration was created — sections inserted, renamed, deleted, or reordered — the narration state may be inconsistent in ways that quietly break patch attempts. Always run a status check first:

```bash
pfw story narration status <story-id>
```

Or via MCP: `narration_status {story_id}`.

Look for:

- **Section count mismatch.** Compare `total_sections` in the narration response to the number of sections in the story (`pfw story get <story-id>` or `story_get`). If the narration count is lower, a section was inserted into the story but never narrated.
- **`section_number` gaps.** The narration `sections[]` array reports each section's `section_number` (1-indexed against current story sort order). A gap like 1–10, then 12 with no 11 means the section now at sortOrder 10 has no audio row. Fix: `narration_regenerate` on that section to first-narrate it, then `narration_rebuild`.
- **`is_stale: true`** on individual sections. The prose was updated after the section was last narrated; the audio is out of sync. That section needs re-narration before any segment-level patch on it will land correctly.
- **`status: "error"`** with an `error_message`. Read the message. A `signal: killed` in the ffmpeg output is the M4B assembler being OOM-killed (infra-side; segment patches won't help). Other errors may indicate transient TTS failures (retryable) or per-section content failures (operator action needed).

Patch attempts on a structurally-broken or stale narration silently no-op or fail in confusing ways. A 30-second state check before the first regenerate saves an hour of debugging "why did my fix not stick?"

## Segment Patching Workflow

### Step 1: Find the problem section

```bash
pfw story narration status <story-id>
```

This lists all sections with their IDs and status. Note the section ID for the
section containing the problem.

### Step 2: List segments and find the bad one

```bash
pfw story narration segments <story-id> <section-id>
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
pfw story narration segment-regenerate <story-id> <section-id> <segment-id> --voice Kore
```

This:
1. Deletes the old segment audio
2. Re-narrates just that segment with the specified voice
3. Restitches all segments into the section audio
4. Reassembles the audiobook automatically

**Voice selection:** Use `pfw story narration voices` to list available voices.
Common choices:
- `Kore` — Gemini female (good for fixing Kokoro pronunciation issues)
- `Puck` — Gemini male
- `af_sarah` — Kokoro female (default)
- `am_adam` — Kokoro male

### Step 4: Verify

```bash
pfw story narration segments <story-id> <section-id>
```

The regenerated segment should now show the new voice and provider.

## MCP Tool Workflow

For AI agents using the MCP server:

```
narration_status {story_id}                              → find section with the problem
narration_segments {story_id, section_id}                → find the segment with bad text
narration_segment_regenerate {story_id, section_id, segment_id, voice: "Kore"}
narration_status {story_id}                              → poll until complete
```

## Examples

### Fix a mispronounced word

Kokoro says "brethed" instead of "breathed":

```bash
# Find which section has "breathed"
pfw story export <story-id> | grep -n "breathed"

# List segments for that section
pfw story narration segments <story-id> <section-id>

# Find the segment containing "breathed" and regenerate with Gemini
pfw story narration segment-regenerate <story-id> <section-id> <segment-id> --voice Kore
```

### Fix an acronym

Kokoro doesn't pronounce "DNA" as an acronym:

```bash
# Same flow — find the segment, regenerate with a voice that handles acronyms
pfw story narration segment-regenerate <story-id> <section-id> <segment-id> --voice Kore
```

## Cost

Each segment regeneration costs **2 credits** (same as any TTS operation).
Use `pfw story credits` to check your balance before and after.

## Limitations

- **Content must not have changed.** If you edited the story text after narration,
  segment boundaries may have shifted. The endpoint will reject with
  `content_changed: true`. Use section-level regeneration instead.
- **Voice mismatch.** Replacing one segment with a different voice will sound
  different from surrounding segments. This is a trade-off — correct pronunciation
  vs consistent voice. For critical fixes, consider regenerating the entire section.
- **One segment at a time.** There's no batch segment regeneration. If multiple
  segments need fixing, regenerate each one individually or use section-level
  regeneration.

## Other Narration Operations

| Command | Purpose |
|---------|---------|
| `story narrate <story-id>` | Start full narration |
| `story narration regenerate <story-id> <section-id> --force --voice Kore` | Re-narrate entire section |
| `story narration rebuild --section-announcements` | Reassemble audiobook with section titles |
| `story narration retry <story-id> <section-id>` | Retry a failed section |
| `story narration cancel <story-id> <section-id>` | Cancel an in-flight section narration |
| `story narration delete <story-id>` | Delete narration and start fresh |
| `story narration resume <story-id>` | Resume stuck narration |
| `story credits` | Check credit balance |
