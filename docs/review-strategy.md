# Review Strategy Guide

How to approach a story review — for AI agents and human reviewers alike.

## The Golden Rule

**Fix structural issues before polishing sentences.** A beautifully worded paragraph in a duplicated section is still a duplicated section. Quality scores are dominated by structural dimensions (continuity 30%, progression 25%), not surface quality.

## Phase 1: Diagnose

Before making any changes, understand the story's problems.

```bash
# Check quality scores and issues
pfw story quality <story-id>

# Read the full story
pfw story export <story-id> --format json
```

Or via MCP:
```
story_quality {story_id: "..."}
story_export {story_id: "...", format: "json"}
```

**Verify before acting:** The quality checker may produce false positives —
especially gender-flip detection in multi-character scenes, dialogue flagged as
gibberish, and format artifacts that don't exist in the actual section content.
After reading the full story, cross-check each flagged issue against the exported
text. If the text is correct and the checker is wrong, note the mismatch in your
summary — do not "fix" text that isn't broken.

**What to look for:**
- Which dimensions score lowest?
- What are the critical/major issues?
- Which sections are flagged most often?
- Are there patterns (e.g., multiple sections with the same problem)?

**Prioritize by impact:**
1. Critical issues in high-weight dimensions (continuity, progression)
2. Major issues affecting multiple sections
3. Tone/voice consistency issues
4. Sentence-level polish

**Track quality across passes:** Snapshot scores before and after each review cycle:
```bash
pfw story quality <story-id> -o json > before.json
# ... incorporate feedback ...
pfw story assess <story-id> --force
pfw story quality <story-id> -o json > after.json
diff before.json after.json
```

## Phase 2: Structural Fixes (Section Rewrites)

For issues scoring 0-3, sentence-level replacements won't help. You need to rewrite sections.

### When to rewrite a section

| Problem | Signal | Action |
|---------|--------|--------|
| Duplicate openings | Two sections start with similar phrases | Rewrite one section's opening |
| Cross repetition | 15%+ content overlap between sections | Rewrite the later section to advance the narrative |
| Bloat | Section exceeds 3,000 words with repetition | Cut to 1,500-2,500 words, remove repeated phrases |
| POV shift | Section switches person (e.g., third → second) | Rewrite entire section in consistent POV |
| Gender flip | Pronouns inconsistent with character | Full pronoun pass on the section |
| Story restart | Section opens with "In the beginning..." | Rewrite opening to continue from previous section |

### How to rewrite

```bash
# Read the section (currently via story export, extract the section you need)
pfw story export <story-id> --format json

# Write the rewritten content and pipe it in
pfw feedback section update <story-id> <review-id> <section-id> --stdin < rewritten.txt
```

Or via MCP:
```
feedback_section_update {
  story_id: "...",
  review_id: "...",
  section_id: "...",
  content: "Full rewritten section content..."
}
```

### Rewriting guidelines

- **Preserve the author's voice.** You're fixing structure, not rewriting their style.
- **Keep the narrative beats.** The same events should happen — just expressed more cleanly.
- **Cut aggressively.** If a section has 36 repeated phrases, cutting 60% is appropriate.
- **Fix one problem at a time** per section if multiple issues exist. Makes the diff easier to review.

## Phase 3: Surface Polish (Replacements)

After structural fixes, address sentence-level issues.

```bash
echo '{"sectionId":"...","type":"replacement","text":"original phrase","suggested":"improved phrase","rationale":"why"}' \
  | pfw feedback item add <story-id> <review-id> --stdin
```

**Good targets for replacement:**
- Mixed metaphors ("a needle pulling through a tapestry")
- Purple prose (overwrought phrasing that obscures meaning)
- Repetitive sentence openings ("He walked... He saw... He felt...")
- Weak transitions between paragraphs
- Passive voice where active would be stronger

## Phase 4: Commentary

Even if you fix everything mechanically, the author needs to understand what happened and why.

### Feedback item types for commentary

- **`opportunity`** — "This section has potential but the POV shifts weaken it"
- **`strength`** — "The dialogue in this section is natural and engaging" (positive feedback matters)
- **`context`** — "Character 'Future' is male in section 3 but referenced as female in section 4"
- **`suggestion`** — "Consider splitting this 6,000-word section into two at the scene break"

### Review summary

Before submitting, add a `context` item summarizing the review:

```
Review Summary:

Targeted dimensions: Continuity (1.0→?), Progression (1.0→?), Tone (1.0→?)

Actions taken:
- Rewrote sections 4, 5, 8 to eliminate duplicate openings
- Cut section 8 from 6,445 to 2,100 words
- Fixed POV in section 8 (second person → third person)
- Fixed gender pronouns in sections 4, 7, 10
- 8 sentence-level replacements for prose quality

Remaining issues:
- Section 12 has a "story restart" pattern that needs narrative restructuring
- Section 9 ending is abrupt — consider extending

Replacements: 8 | Section rewrites: 4 | Commentary: 6
```

## Phase 5: Submit

```bash
pfw feedback submit <review-id>
```

The author will see:
- Section rewrites as diffs (before/after comparison)
- Replacement suggestions they can accept/reject individually
- Your commentary explaining what you found and what to do about it

## Common Mistakes

1. **Only doing replacements.** If the quality score is below 5, replacements won't move the needle. Start with section rewrites.
2. **Rewriting the author's voice.** Fix structure and consistency, not style preference. "Tighter prose" is subjective — "fixing a POV shift" is objective.
3. **Not explaining why.** Every section rewrite should have an accompanying `opportunity` or `context` item explaining what was wrong.
4. **Ignoring strengths.** Noting what works well helps the author understand what to preserve during future edits.
5. **Replacing too much of a section.** When you rewrite a section, preserve as much of the author's original text as possible. Fix the specific problem — don't rewrite the whole section from scratch. If you're changing more than 30-40% of the text, you're ghostwriting, not editing. Use targeted replacements for sentence-level fixes.
