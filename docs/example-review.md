# Example Review: Gunthor Gets Dirty

A complete walkthrough of an AI-assisted review that improved a story's quality score from **2.89 (1.4 stars) to 8.11 (4.1 stars)**.

## The Story

"Gunthor Gets Dirty" is a 12-section fantasy story about a time traveler named Gunthor and a woman named Mrs. Murphy who discover an ancient temple pulsing with forgotten magic. The story was AI-generated and had significant structural issues.

**Story ID:** `260570eb-5f13-46bc-9580-441760a2443a`

## Before: Quality Assessment

```
pfw story quality 260570eb-5f13-46bc-9580-441760a2443a
```

| Dimension | Score | Issues |
|-----------|-------|--------|
| Continuity (30%) | 1.0/10 | 6 duplicate openings, cross repetition (critical), 3 gender flips |
| Progression (25%) | 1.0/10 | 3 story restarts, 1 abrupt ending |
| Coherence (20%) | 9.5/10 | 1 minor formatting issue |
| Tone (15%) | 1.0/10 | POV shift (second→third person), 10 tone issues |
| **Overall** | **2.89/10** | **23 issues total** |

## Diagnosis

The quality scores reveal this story's problems are **structural, not surface-level**:

1. **Duplicate content** — Sections 4/5, 6/11, and 9/10 share significant content overlap (up to 48%)
2. **Bloat** — Section 9 is 6,445 words with 14-20x repeated phrases
3. **POV shifts** — Section 9 switches from third to second person
4. **Story restarts** — Sections 7 and 12 open with "In the beginning..." patterns
5. **Gender flips** — Character pronouns change across sections 4, 7, and 10

Coherence is already high (9.5) — the logical structure is fine, the content just repeats.

**Key insight:** Sentence-level replacements won't fix these issues. We need section rewrites.

## The Review

### Phase 1: Structural Fixes (Section Rewrites)

**Section 6 "Three in a Tree"** — 48% cross-repetition with Section 5.
- Read both sections, identified the overlap
- Rewrote Section 6 with entirely new narrative: a hidden passage, a seeing-well, floating script, a dark basin
- Preserved the story's mystical tone and the Gunthor/Mrs. Murphy dynamic

**Section 9 "Sacred Pulse of Connection"** — 6,445 words, second-person POV, 14-20x phrase repetition.
- Cut from 6,445 to ~1,000 words
- Fixed POV from second person to consistent third person
- Removed repeated phrases ("the air was thick enough to chew", "tapestry of")
- Preserved the altar discovery and the turkey moment

**Section 10 "One Big Bang after Another"** — 18% cross-repetition with Section 9.
- Rewrote the opening to continue from the stairwell discovered in the rewritten Section 9
- Eliminated the repetition, added a new upper chamber discovery

**Section 7 "The Final Climb"** — Story restart pattern.
- Changed "In the beginning" → "This is the source" / "We are at the wellspring of creation"
- Eliminated the story restart trigger without changing the narrative

**Section 12 "Beneath the Pluméd Monolith"** — Weak ending.
- Completely rewritten as a proper conclusion
- Added: lattice revelation (the interconnected web of all consciousness)
- Added: farewell between Gunthor and Mrs. Murphy
- Closing line: "He stepped forward into the night, carrying the light within him."

### Phase 2: Surface Polish (Replacements)

4 sentence-level replacements targeting the "tapestry of" AI phrase pattern that appeared across multiple untouched sections.

### Phase 3: Commentary

- 3 **strength** items — positive feedback on dialogue quality, world-building, the turkey motif
- 5 **opportunity** items — areas for future improvement in untouched sections
- 2 **context** items — character reference sheet and review summary

## After: Quality Assessment

```
pfw story assess 260570eb-5f13-46bc-9580-441760a2443a --force
pfw story quality 260570eb-5f13-46bc-9580-441760a2443a
```

| Dimension | Before | After | Change |
|-----------|--------|-------|--------|
| Continuity (30%) | 1.0 | 7.0 | **+6.0** |
| Progression (25%) | 1.0 | 9.0 | **+8.0** |
| Coherence (20%) | 9.5 | 9.5 | — |
| Tone (15%) | 1.0 | 7.0 | **+6.0** |
| **Overall** | **2.89** | **8.11** | **+5.22** |

Issues reduced: **23 → 9**

## What Worked

1. **Diagnosing before editing.** Reading the quality scores first revealed that 85% of the problem was structural (continuity + progression + tone), not prose quality. This directed the effort toward section rewrites rather than sentence polish.

2. **Section rewrites for structural issues.** The 5 section rewrites addressed duplicate content, bloat, POV shifts, and story restarts — problems that can't be fixed with find-and-replace.

3. **Preserving the author's voice.** The turkey motif ("It has hair all around it like a turkey") was preserved throughout as a running element. The rewritten sections maintained the mystical-yet-grounded tone of the original.

4. **Providing a real ending.** The completely rewritten Section 12 gave the story a conclusion with emotional weight (the lattice revelation, the farewell) instead of trailing off.

## What Could Be Better

1. **Sections 1-5 and 8 are untouched.** A second pass could address residual repetitive phrasing and some remaining "tapestry of" patterns in these sections.

2. **The turkey repeats.** The motif appears in multiple sections because each original section had its own instance. Consolidating to 1-2 occurrences would make it land harder.

3. **Show, don't tell.** Many sections still describe characters' feelings ("He felt a profound connection") rather than showing the connection through action and dialogue. A Line Editor pass would address this.

## Lessons for BYOAI Users

- **Check quality scores first.** They tell you exactly where to focus.
- **Score below 3 → section rewrites.** Don't polish sentences when the structure is broken.
- **Score 5-7 → mix of rewrites and replacements.** Fix remaining structural issues, then polish.
- **Score 8+ → sentence-level polish only.** The structure is sound.
- **Always include commentary.** Your review should be readable by a human author, not just machine-actionable.
