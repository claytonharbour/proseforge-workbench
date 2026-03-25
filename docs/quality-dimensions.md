# Quality Dimensions

ProseForge assesses story quality across five dimensions. Each dimension is scored 0-10 and weighted to produce an overall score.

## Dimension Overview

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| Continuity | 30% | Narrative thread consistency across sections |
| Progression | 25% | Story advancement without restarts or stalls |
| Coherence | 20% | Logical sense and formatting quality |
| Tone | 15% | Consistent writing style and voice |
| Source Integration | 10% | Natural incorporation of source material |

**Overall score** = weighted average of dimension scores (0-10)
**Star rating** = overall / 2 (0-5 stars)

## Continuity (30%)

The highest-weighted dimension. Measures whether the story reads as one continuous narrative rather than a collection of disconnected sections.

### Issue Types

| Issue | Severity | What it means |
|-------|----------|--------------|
| Duplicate Opening | major | Two sections start with the same or very similar phrases |
| Cross Repetition | critical | Significant content overlap between sections (15%+ of sentences shared) |
| Gender Flip | major | A character's gender/pronouns change between sections |

### How to Fix

**Duplicate openings:** Read the openings of flagged sections side by side. Rewrite one to use a completely different entry point — different setting, different character focus, different action.

**Cross repetition:** This usually means an AI-generated story repeated itself. Read both sections, identify what's unique to each, and rewrite the later section to contain only new narrative content. It's okay to cut 50%+ of a section if it's duplicated material.

**Gender flips:** Establish the character's correct gender from context, then rewrite the affected section with consistent pronouns throughout. Don't just find-and-replace pronouns — read the full section to catch all references.

## Progression (25%)

Measures whether the story moves forward. AI-generated stories often loop back to the beginning or stall in repetitive description.

### Issue Types

| Issue | Severity | What it means |
|-------|----------|--------------|
| Story Restart | critical | A section opens as if the story is starting over ("In the beginning...", "Once upon a time...") |
| Abrupt Ending | minor | A section ends mid-thought or without resolution |
| Circular Plot | major | Events repeat without advancement |

### How to Fix

**Story restarts:** The section opening should continue from where the previous section left off. Read the ending of the prior section, then rewrite the flagged section's opening to pick up that thread.

**Abrupt endings:** Extend the section's conclusion with a transition that connects to the next section's events. Or add a paragraph of resolution before the scene break.

**Circular plot:** Identify what's repeated and what's new. Rewrite to keep only the new material, optionally adding a brief reference to past events rather than replaying them.

## Coherence (20%)

Measures whether the story makes logical sense and is well-formatted.

### Issue Types

| Issue | Severity | What it means |
|-------|----------|--------------|
| Logic Gap | major | Events don't follow logically (character is in two places at once, unexplained time jumps) |
| Formatting | minor | Structural formatting issues (broken paragraphs, inconsistent section headers) |

### How to Fix

**Logic gaps:** Add bridging sentences or short paragraphs that explain transitions. "Two days later..." or "Having left the temple..." can bridge a gap without requiring major rewrites.

**Formatting:** Usually sentence-level replacements are sufficient. Fix broken paragraph breaks, standardize how sections are titled.

## Tone (15%)

Measures whether the writing maintains a consistent voice and perspective.

### Issue Types

| Issue | Severity | What it means |
|-------|----------|--------------|
| POV Shift | critical | The narrative perspective changes (third person → second person, past tense → present) |
| Register Shift | major | The formality level changes dramatically (literary prose → casual dialogue-style narration) |
| Tense Shift | major | Verb tense changes within a section |

### How to Fix

**POV shifts:** This almost always requires a full section rewrite. You can't fix a section written in second person ("you walk into the room") with find-and-replace — the entire sentence structure changes when converting to third person ("he walked into the room"). Rewrite the section in the story's dominant POV.

**Register/tense shifts:** Sometimes replacements work for isolated shifts. For pervasive shifts, a section rewrite is more reliable.

## Source Integration (10%)

Measures how naturally source/reference material is incorporated. Currently weighted at 10% (or 0% for stories without sources).

## Interpreting Scores for Review Priority

| Score Range | Meaning | Review Approach |
|-------------|---------|----------------|
| 0-2 | Critical structural issues | Heavy section rewrites. Focus on continuity and progression. |
| 3-5 | Moderate issues | Mix of section rewrites and replacements. All dimensions need attention. |
| 6-8 | Surface issues | Primarily sentence-level polish. Structure is sound. |
| 9-10 | Minor tweaks | Light touch. Mostly suggestions and positive feedback. |
