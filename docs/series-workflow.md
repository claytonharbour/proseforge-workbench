# Series Forge Workflow

Series Forge is ProseForge's system for managing shared universes across multiple
stories (books). It provides persistent world-building, character management, and
canon tracking so that each new installment in a series is grounded in everything
that came before.

Series Forge is a **repository**, not a pipeline. It stores the series bible —
world, characters, timeline — and makes it available to any writing process. How
the story gets written (BYOAI direct write, Story Forge pipeline, or manual) is
up to the author.

## Upstream API Reference

The full API specification with architecture diagram, endpoint details, and git
repository layout is maintained upstream:

- **API Reference:** [SERIES_FORGE_CHAT.md](https://git.proseforge.ai/forge/proseforge/src/branch/master/docs/SERIES_FORGE_CHAT.md)
- **Story Forge Chat:** [STORY_FORGE_CHAT.md](https://git.proseforge.ai/forge/proseforge/src/branch/master/docs/STORY_FORGE_CHAT.md)

This document covers the **workbench MCP tool surface** and practical workflows.

---

## Concepts

### Series Bible

Every series has three core documents stored in a git repository:

| Document | Description | Tool |
|----------|-------------|------|
| **World** | Setting, themes, tone, rules, anti-patterns | `series_world_get` / `series_world_update` |
| **Characters** | Named characters with profiles, roles, status | `series_character_create` / `series_character_list` / `series_character_get` / `series_character_update` |
| **Timeline** | Chronological events across all books | `series_timeline_get` / `series_timeline_update` |

### Story Lifecycle

Stories follow a status lifecycle:

```
pitch → draft → published → (unpublished → draft)
```

- **Pitch** — a pre-writing idea. Has planning data (meta) but no sections.
  Created via `story_pitch_create`. Excluded from the default story list.
- **Draft** — a story being written. Sections can be created and edited.
  Pitches become drafts via `story_promote`.
- **Published** — live and readable. Published via `story_publish`.

### Story Planning Data (Meta)

Every story has three planning documents stored as markdown in git. The format
is freeform markdown — no schema enforcement — but the generation pipeline
reads specific structures. Write content that follows these conventions for
best results.

#### `story` — Premise & Overview

The high-level concept. The pipeline injects the entire file as context when
generating each section. More detail gives the AI better context.

```markdown
## Genre & Tone
Detective thriller with procedural realism, set in a dark tone

## Central Theme
The cost of seeing clearly in a system that rewards looking away

## Setting
An unnamed Rust Belt city in post-industrial decline

## Core Conflict
A detective investigating institutional corruption that operates within the law
```

#### `characters` — Character Profiles

One `## Name` header per character. The pipeline reads the whole document as
context — field names don't need to match exactly.

```markdown
## Miles Corbin
Role: Protagonist, homicide detective
Background: Twenty years on the force, humbled by a wrongful conviction
Motivation: Finding truth, even when the system resists it

## Dr. Lena Hanson
Role: Chief forensic pathologist, moral anchor
Background: Sharp, clinical, dark-humored. Two cups of coffee.
Motivation: Precision as a form of care
```

#### `plot` — Section-by-Section Outline

**Important:** Use `## Section 1`, `## Section 2` headers. The pipeline parses
these to inject per-section context during generation. Without these headers,
it falls back to the whole outline.

```markdown
## Section 1
Introduction — the case arrives. Corbin meets the client.
Tension level: Low

## Section 2
Investigation begins. First witness interview reveals the pattern.
Tension level: Rising

## Section 3
Evidence mounts. The suspect lawyers up.
Tension level: High
```

After generation, the pipeline appends `### What was written` under each
section header with a summary for continuity.

Meta works on any story regardless of status — pitches, drafts, or published.
Use it as a planning surface before writing sections.

### Writing Paths

There are three ways to write a story in a series:

1. **BYOAI Direct Write** — An external AI agent reads the series bible, then
   writes sections directly via `section_write`. The agent is responsible for
   voice consistency and continuity. This is the primary path used for the
   Corbin series (12 books).

2. **Story Forge Pipeline** — Use `series_plan` to create a Story Forge Chat
   session pre-seeded with series context. The AI interviews the author with
   full awareness of the universe, then generates the story through the
   standard pipeline (outline → approve → sections).

3. **Manual** — Read the series bible, write the story outside ProseForge,
   then upload and link it to the series.

---

## MCP Tool Reference

### Series Management

| Tool | Description |
|------|-------------|
| `series_list` | List the authenticated user's series |
| `series_create` | Create a new series |
| `series_get` | Get series details by ID |
| `series_update` | Update series metadata (name, description, genre, tone) |
| `series_archive` | Archive (soft-delete) a series |

### World & Timeline

| Tool | Description |
|------|-------------|
| `series_world_get` | Get the world overview document (markdown) |
| `series_world_update` | Update the world overview document |
| `series_timeline_get` | Get the canon timeline (unified view, all sections assembled) |
| `series_timeline_update` | Update the canon timeline (full rewrite — prefer per-section tools below) |
| `series_timeline_sections` | List timeline sections with slugs, titles, and sort order |
| `series_timeline_section_get` | Get a single timeline section by slug |
| `series_timeline_section_update` | Update a single timeline section by slug (safe for parallel writes) |

### Characters

| Tool | Description |
|------|-------------|
| `series_character_create` | Create a character (name, role, profile, status) |
| `series_character_list` | List all characters in a series |
| `series_character_get` | Get a character's profile by slug |
| `series_character_update` | Update a character's profile |
| `series_character_delete` | Delete a character |

**Role values:** `protagonist`, `antagonist`, `supporting`, `minor`
**Status values:** `active`, `retired`, `deceased`

### Story Lifecycle & Planning

| Tool | Description |
|------|-------------|
| `story_pitch_create` | Create a pitch (pre-writing idea, status "pitch") |
| `story_promote` | Promote a pitch to draft (enables sections and publishing) |
| `story_meta_upsert` | Write story planning data (story, characters, or plot) — creates if missing |
| `storyforge_meta_get` | Read story planning data (story, characters, plot) |

### Story Linking

| Tool | Description |
|------|-------------|
| `series_stories_list` | List stories linked to a series |
| `series_stories_add` | Link an existing story to a series |
| `series_stories_remove` | Remove a story from a series |

### Series Chat (World-Building Interview)

| Tool | Description |
|------|-------------|
| `series_chat_create` | Start a new world-building chat session |
| `series_chat_list` | List chat sessions for a series |
| `series_chat_get` | Get a chat session with its messages |
| `series_chat_send` | Send a message and get AI response |
| `series_chat_finalize` | Finalize a completed chat session |
| `series_chat_harvest` | Extract world/character/timeline data from a session to git |
| `series_harvest_all` | Harvest all sessions in a series |

### Series Plan (Handoff to Story Forge)

| Tool | Description |
|------|-------------|
| `series_plan` | Create a Story Forge Chat session seeded with series context (world, characters, timeline). The AI interviews the author with full series awareness. |

### Story Forge Pipeline

These tools drive the Story Forge generation pipeline after `series_plan`:

| Tool | Description |
|------|-------------|
| `storyforge_chat_create` | Start a new Story Forge Chat session |
| `storyforge_chat_list` | List Story Forge Chat sessions |
| `storyforge_chat_get` | Get a session with message history |
| `storyforge_chat_send` | Send a message and get AI response |
| `storyforge_chat_finalize` | Finalize interview and trigger generation |
| `storyforge_status` | Get generation pipeline status |
| `storyforge_meta_get` | Read story planning data (story, characters, plot) |
| `storyforge_meta_approve` | Approve outline and start section generation |
| `storyforge_meta_regenerate` | Regenerate the outline (free retry) |
| `storyforge_resume` | Resume a failed or paused generation |

---

## Workflow 1: Build a Series from Scratch

Direct construction — you provide the content for each component.

```
1. series_create                    → series ID
2. series_world_update              → set world overview (markdown)
3. series_character_create          → create characters (repeat for each)
4. series_timeline_update           → set canon timeline (markdown)
```

### World Document Template

```markdown
# The Setting

Description of the world, location, time period.

## Key Locations

- Location 1 — description
- Location 2 — description

## Recurring Institutions

- Institution 1 — role in the story
- Institution 2 — role in the story

## Themes

- Theme 1
- Theme 2

## Prose Style Rules

- POV, tense, tone
- Dialogue style
- Section length targets

## Anti-Patterns (DO NOT REPEAT)

- Pattern to avoid — reason
- Pattern to avoid — reason
```

### Character Profile Template

```markdown
Name, age, role. Physical description. Background.

Speech patterns. Key relationships. Defining moments.

"Signature line or quote."

Appears in: Book 1, Book 3.
```

---

## Workflow 2: Build via Chat Interview

Use the AI-guided world-building interview to develop the series bible
conversationally.

```
1. series_create                              → series ID
2. series_chat_create                         → session ID
3. LOOP:
   series_chat_send                           → world-building conversation
4. series_chat_finalize                       → mark complete
5. series_chat_harvest                        → extract to git (world, characters, timeline)
```

After harvest, the series bible is populated from the conversation. You can
then edit individual components with `series_world_update`,
`series_character_update`, etc.

---

## Workflow 3: Write a Story (BYOAI Direct Write)

This is the workflow used for the Corbin series. An external AI agent reads
the series bible and writes sections directly.

```
1. series_world_get                           → read world document
2. series_character_list + series_character_get → read character profiles
3. series_timeline_get                        → read canon timeline
4. story_pitch_create                         → create a pitch (or story_create for draft)
5. story_meta_upsert                          → write premise, characters, plot outline
6. story_promote                              → promote pitch to draft (if created as pitch)
7. section_create + section_write             → write each section
8. series_stories_add                         → link story to series
9. story_publish                              → publish
10. narration_start                           → start audiobook narration (optional)
```

Steps 4-6 are the **planning phase**. Pitches live in the Pitches section of
My Library, separate from drafts and published stories. Use `story_meta_upsert`
to develop the premise, characters, and plot before committing to writing.
When ready, `story_promote` transitions to draft and enables section creation.

### After Writing: Harvest Back

After a story is written, new characters and events should be harvested back
to the series bible so the next installment has the full context.

**Characters (append operation — safe for parallel writes):**
```
1. series_character_list                      → check existing characters
2. series_character_create                    → create each new character not already present
3. series_character_update                    → update existing characters with new developments
```

**Timeline (per-section — safe for parallel writes):**
```
1. series_timeline_sections                   → list sections, find your book's slug
2. series_timeline_section_update             → write your book's events by slug
```

Each book has its own timeline section with a slug (e.g., `book-3-dead-reckoning`).
Agents write only their own book's events without touching other sections. The
unified read (`series_timeline_get`) still assembles all sections into one document.

The full-rewrite `series_timeline_update` tool is still available for bulk
operations but should not be used for per-book harvest.

---

## Workflow 4: Plan a Story from Series Context

Use `series_plan` to seed a Story Forge Chat with the full series bible, then
let the platform generate the story through the standard pipeline.

```
1. series_plan                                → seeded Story Forge Chat session ID
2. LOOP:
   storyforge_chat_send                       → refine story idea with series context
3. storyforge_chat_finalize                   → trigger generation → story ID
4. storyforge_status                          → poll (metaStatus = awaiting_approval)
5. storyforge_meta_get                        → review generated outline
6. storyforge_meta_approve                    → start section generation
7. storyforge_status                          → poll (completedCount = totalSections)
8. series_stories_add                         → link to series
```

The `series_plan` tool accepts optional parameters:
- `book_number` — which installment (0 = auto-detect next)
- `include_characters` — character slugs to include (null = all)
- `notes` — author notes injected into the AI context

---

## Ghostwrite Prompt Template

For BYOAI direct writes, a prompt template ensures consistent quality and
workflow across agents. See `.plans/prompts/corbin-ghostwrite.md` for the
production template used for the Corbin series.

Key elements of a ghostwrite prompt:
- Environment verification (prod vs dev)
- Series bible setup (read world, characters, timeline)
- Story requirements (premise, length, voice, anti-patterns)
- Before writing: create a tracking ticket with outline
- After writing: link, publish, narrate, harvest characters + timeline
- Test criteria (does it feel like it belongs in the series?)

---

## Series Bible as Quality Control

The series bible serves multiple functions beyond continuity:

**Anti-contamination:** The world document provides structure, not prose. It
tells agents what the city looks like, how characters speak, and what themes
to explore — without giving them sentences to copy. This prevents voice
contamination across different AI agents writing in the same universe.

**Anti-patterns:** The world document includes an explicit list of patterns
to avoid — repeated plot devices, overused metaphors, resolved storylines.
This prevents the series from becoming repetitive as it grows.

**Character voice:** Each character profile includes speech patterns and
signature lines. Agents read these before writing and use them as constraints,
not templates. The test is whether the character sounds right, not whether
they sound the same as last time.

---

## Related Documentation

- [Upstream API Reference](https://git.proseforge.ai/forge/proseforge/src/branch/master/docs/SERIES_FORGE_CHAT.md) — full endpoint spec, architecture diagram, git layout
- [Story Forge Chat](https://git.proseforge.ai/forge/proseforge/src/branch/master/docs/STORY_FORGE_CHAT.md) — the generation pipeline that `series_plan` feeds into
- [Review Flow](review-flow.md) — how AI-assisted reviews work (separate from series)
- [MCP Setup](mcp-setup.md) — server configuration and registration
- [Getting Started](getting-started.md) — installation and first commands
