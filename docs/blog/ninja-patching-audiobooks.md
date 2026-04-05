# Ninja-Patching Audiobooks: Fixing One Word Without Re-Recording Everything

*We built segment-level audio patching for AI-narrated audiobooks. Here's how we fixed a mispronounced word in a 7-chapter thriller without re-narrating a single chapter.*

---

There's a moment in every audiobook production where you hear it. That one word. The one the TTS engine decided to pronounce like it was reading a different language. You wince. You check the timestamp. And then you realize: to fix thirty seconds of audio, you have to re-narrate ten minutes of chapter.

We decided that was unacceptable.

## The Problem

Our TTS engine, Kokoro, is excellent. Fast, natural, expressive. But it has blind spots. "DNA" comes out as a single syllable instead of three letters. "Breathed" rhymes with "wretched" instead of "seethed." And "IIA" — Internal Affairs — gets treated like a word nobody's ever heard, because nobody has.

These aren't bugs in the traditional sense. They're pronunciation gaps in a model trained on millions of hours of natural speech, where acronyms and irregular past tenses are statistical anomalies. The model does its best. Its best is sometimes wrong.

The traditional fix is simple and expensive: re-narrate the entire chapter. For a 3,000-word section, that means re-running TTS on the whole thing, waiting for processing, and hoping the new take doesn't introduce different problems. It costs credits, it costs time, and it replaces audio that was perfectly fine.

We wanted a scalpel. We had a sledgehammer.

## The Architecture

ProseForge narration works in three layers:

```
Story → Chapters → Segments
```

A chapter is one section of your story. A segment is a chunk of roughly 3,000 characters — about 30 seconds of audio. When you narrate a story, each chapter gets split into segments, each segment gets sent to the TTS engine, and the results get stitched back into chapter audio, then assembled into an M4B audiobook.

The key insight: **segments are independent**. Each one is a separate TTS call with its own audio file. If you can identify which segment has the problem, you can re-narrate just that segment and restitch the chapter without touching anything else.

## Building the Scalpel

We added three capabilities to the [proseforge-workbench](https://github.com/claytonharbour/proseforge-workbench):

**1. Segment listing** — See exactly what text was sent to TTS for each segment, which voice narrated it, and which provider was used.

```bash
pfw story narration segments <story-id> <chapter-id>
```

**2. Segment regeneration with voice override** — Re-narrate one segment with a different voice, potentially from a different TTS provider entirely.

```bash
pfw story narration segment-regenerate <story-id> <chapter-id> <segment-id> --voice Kore
```

**3. Automatic restitching** — After the segment is regenerated, the system automatically restitches all segments into the chapter audio and reassembles the audiobook. No manual assembly step.

## Cross-Provider Patching

This is where it gets interesting. The voice override doesn't just let you pick a different voice from the same engine. It lets you switch providers mid-chapter.

Our Corbin thriller series was narrated with Kokoro's `af_sarah` voice. When Kokoro mispronounced "DNA," we patched those segments with Google Gemini's `Kore` voice. Kore handles acronyms correctly because Gemini's model was trained differently.

The result: a chapter where 29 out of 30 seconds are Kokoro, and one segment is Gemini. The voices are different enough that a careful listener might notice the switch. But "DNA" is pronounced correctly, which matters more for a crime thriller where DNA evidence is a plot point.

## The Patch

The Corbin series has four books. "DNA" appears in 13 segments across 7 chapters in two books. We mapped every occurrence:

```bash
# Find every section containing "DNA"
pfw story export <story-id> | grep -n "DNA"

# For each section, list segments and find the one with "DNA"
pfw story narration segments <story-id> <chapter-id>

# Patch it
pfw story narration segment-regenerate <story-id> <chapter-id> <segment-id> --voice Kore
```

Thirteen segments. Twenty-six credits. About fifteen minutes of processing. Zero chapters re-narrated. Zero good audio thrown away.

## What This Means

Traditional audiobook production treats chapters as atomic units. If something's wrong, you re-record the chapter. That's how human narration works — you book the studio, the narrator reads the chapter, you do it again if something's off.

AI narration doesn't have to work that way. The audio is generated from text through a deterministic pipeline. Every segment is reproducible. Every segment is replaceable. And with cross-provider patching, every segment can be narrated by a different voice if that's what the content requires.

Imagine: a thriller where the detective's internal monologue is one voice, and the interrogation scenes are another. Not because you re-narrated the chapter — because you patched individual segments with the voice that fits.

We're not there yet. The voice mismatch between providers is still noticeable. But the infrastructure is in place. Segment-level, cross-provider, automatic restitching. The scalpel exists.

## Try It

The workbench is [open source](https://github.com/claytonharbour/proseforge-workbench). The narration patching workflow is documented in the [patching guide](https://github.com/claytonharbour/proseforge-workbench/blob/main/docs/narration-patching.md). And if you want to hear the results, the Corbin series is available on [ProseForge](https://app.proseforge.com/@clayton).

The turkey in the margins approves.

---

*Read the series:*
- *[Rust & Bone](https://app.proseforge.com/@clayton/rust-bone/read) — where it started*
- *[The Gilt Edge](https://app.proseforge.com/@clayton/the-gilt-edge/read) — the follow-up*
- *[Dead Reckoning](https://app.proseforge.com/@clayton/dead-reckoning/read) — DNA enters the chat*
- *[False Positive](https://app.proseforge.com/@clayton/false-positive/read) — DNA takes over*
