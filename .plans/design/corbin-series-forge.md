# Corbin Series — Series Forge Integration

## Overview

The **Corbin** series is a 12-book detective thriller used as both a creative
project and a validation vehicle for the **Series Forge** API integration into
the proseforge-workbench MCP layer.

### Goals

1. Write and publish all 12 books with audiobook narration
2. Implement Series Forge MCP tools in the workbench (Phase 2 of #96)
3. Populate a machine-readable series bible via the Series Forge API
4. Run a multi-agent authorship experiment: Claude writes some books, Codex
   writes others — compare voice consistency using the `StorySeed` mechanism
5. Validate the upstream Series Forge API from a real external consumer

### Key Constraint

The primary consumer of the workbench is **another AI agent**. The series bible,
character data, and world context must be structured for machine consumption, not
human reading. The `series_plan` handoff path seeds a Story Forge Chat session
with structured context — this is the mechanism that enables multi-agent
authorship without voice contamination.

---

## Series Forge API Mapping

The upstream Series Forge API (documented in `docs/SERIES_FORGE_CHAT.md` in the
proseforge repo, tracked in workbench #96 Phase 2) provides:

### Series CRUD
| Endpoint | Tool | Purpose |
|----------|------|---------|
| `POST /series` | `series_create` | Create the Corbin series entity |
| `GET /series/{id}` | `series_get` | Retrieve series metadata |
| `PATCH /series/{id}` | `series_update` | Update series metadata |
| `GET /series` | `series_list` | List series |

### World & Timeline (Git-backed markdown)
| Endpoint | Tool | Purpose |
|----------|------|---------|
| `GET /series/{id}/world` | `series_world_get` | Read world document |
| `PUT /series/{id}/world` | `series_world_update` | Write world document |
| `GET /series/{id}/timeline` | `series_timeline_get` | Read timeline |
| `PUT /series/{id}/timeline` | `series_timeline_update` | Write timeline |

The **world document** holds: the city, institutions, neighborhoods, themes,
prose style rules, and anti-patterns. The **timeline** holds: chronological
events across all books.

### Characters (Git-persisted)
| Endpoint | Tool | Purpose |
|----------|------|---------|
| `POST /series/{id}/characters` | `series_character_create` | Add character |
| `GET /series/{id}/characters` | `series_character_list` | List all characters |
| `GET /series/{id}/characters/{cid}` | `series_character_get` | Get character detail |
| `PATCH /series/{id}/characters/{cid}` | `series_character_update` | Update character |
| `DELETE /series/{id}/characters/{cid}` | `series_character_delete` | Remove character |

Each character entry includes: name, role, description, relationships, voice
notes, appearance history (which books they appear in).

### Story Linking
| Endpoint | Tool | Purpose |
|----------|------|---------|
| `GET /series/{id}/stories` | `series_stories_list` | List linked stories |
| `POST /series/{id}/stories` | `series_stories_add` | Link story to series |
| `DELETE /series/{id}/stories/{sid}` | `series_stories_remove` | Unlink story |

### Generation Handoff (Critical Path)
| Endpoint | Tool | Purpose |
|----------|------|---------|
| `POST /series/{id}/stories/generate` | `series_generate` | Direct generation — series meta becomes story meta, enters pipeline at `awaiting_approval` |
| `POST /series/{id}/stories/plan` | `series_plan` | Creates a Story Forge Chat session pre-seeded with `StorySeed` (world, canon, prior installments, genre, tone, book number) |

### Series Chat (World-building interviews)
| Endpoint | Tool | Purpose |
|----------|------|---------|
| `POST /series/{id}/chat` | `series_chat_create` | Start world-building chat |
| `POST /series/{id}/chat/{cid}/messages` | `series_chat_send` | Send message |
| `GET /series/{id}/chat/{cid}` | `series_chat_get` | Get session |
| `POST /series/{id}/chat/{cid}/finalize` | `series_chat_finalize` | Finalize session |
| `POST /series/{id}/chat/{cid}/harvest` | `series_chat_harvest` | Extract world/character/timeline data from chat |

---

## Workflow Design

### What Series Forge Is (and Isn't)

**Series Forge is a repository**, not a pipeline enforcer. It holds the series
bible — world, characters, timeline, linked stories. Any agent or process can
read from it. It does NOT replace direct writes.

**Writing a book still has two paths:**

1. **BYOAI direct write** (primary) — An agent reads series context from Series
   Forge, then writes sections directly via `section_write`. This is the
   proseforge-workbench model. Series Forge gives the agent better context;
   the writing workflow stays the same.

2. **Story Forge pipeline** (optional) — `series_plan` seeds a Story Forge Chat
   session with series context, and the platform's built-in AI generates the
   story through the chat → meta → sections pipeline. This is an alternative
   path, not a replacement.

Both paths are valid. The workbench doesn't force you through the pipeline.

### Three Test Cases for Multi-Agent Experiment

| Test | Agent | Path | Notes |
|------|-------|------|-------|
| 1 | **Claude** | BYOAI direct write | Reads series bible via Series Forge tools, writes sections via `section_write`. Same as Books 1–4 but with structured series context. |
| 2 | **Codex** | BYOAI direct write | Same as Claude — reads series bible, writes sections directly. Tests whether a different agent can match voice using the same structured context. |
| 3 | **ProseForge AI** (magnum + gemini) | Story Forge pipeline | BYOAI agent takes the role of the Story Forge Chat, kicks off a series write using series metadata. Tests the platform's own generation pipeline seeded with series context via `series_plan`. |

Test 3 is the most interesting from an API validation perspective — it exercises
the `series_plan` → `StorySeed` → chat → generation path end-to-end.

### Anti-Contamination Strategy

The risk: an agent that reads prior books too closely will unconsciously remix
scenes, dialogue patterns, and plot structures. The mitigation:

1. **Series Forge provides structure, not prose** — character descriptions, not
   character dialogue. World rules, not world passages. Plot summaries, not
   scene-by-scene recaps.
2. **Anti-patterns list** — explicit list of things NOT to repeat, stored in the
   world document.
3. **Voice guide** — prose style rules that describe the target without
   demonstrating it.

### Development Philosophy

> "The ink won't stick" — *The Forge of Forgotten Scrolls*

Move slowly. Prefer upstream API fixes over workbench workarounds. If the Series
Forge API doesn't support what we need, file an upstream ticket and wait for the
right fix rather than building a hack that we'll have to tear out later. The
review/feedback workflow taught us this lesson already — every workaround became
technical debt that had to be unwound when the upstream API caught up.

---

## Series Bible Structure

### World Document (series_world_update)

```markdown
# The City

An unnamed Rust Belt city in post-industrial decline. Three distinct zones:

- **Industrial East Side**: Abandoned factories, contaminated groundwater,
  working-class neighborhoods (Greystone, Stonebridge). The old city.
- **Innovation District**: Glass towers, construction cranes, gentrification.
  The new city. Built on the bones of the old.
- **Midtown/Downtown**: The institutions — precinct, courthouse, morgue, city
  hall. Where the systems operate.

The city is never named. It is every city.

## Recurring Institutions
- The police department (reformed after Book 3, still imperfect)
- The DA's office (Margaret Liu)
- The morgue (Hanson's domain)
- Greystone Community Center (Ruth Bledsoe)
- The Innocence Collective (introduced Book 4)

## Themes
- The powerful consuming the vulnerable
- Institutions protecting themselves
- The cost of seeing clearly
- Justice as aspiration, not guarantee
- The gap between the system's promises and its delivery

## Prose Style Rules
- Third-person close, locked to Corbin's POV
- Short punchy sentences in tense moments
- Longer flowing prose for atmosphere and setting
- Dialogue-heavy — characters have distinct speech patterns
- No purple prose, no overwriting
- Written for the ear (audiobook-first)
- Sections: 1,500–2,500 words each
- Every section ends with a hook

## Anti-Patterns (DO NOT REPEAT)
- No more staged suicides (used in Books 2 and 3)
- No more confrontations in abandoned industrial buildings (used in Books 1, 2)
- No more corroded gears as symbols
- No more "Corbin pours bourbon and doesn't drink it" (used enough)
- No more Alderman Phelps (resolved in Books 2–3)
- Avoid the word "machinery" as metaphor for systems (overused)
- Don't have Corbin drive through the city reflecting on landmarks at the end of
  every book (used in Books 1–4 — find a different closing rhythm)
```

### Characters (series_character_create)

**Miles Corbin** — Detective, homicide. Late 40s. Twenty+ years on the force.
Sharp, tired, increasingly disillusioned. Defined by his ability to see what
others miss — and, after Book 4, by the knowledge that this ability is fallible.
Declined the captain's job twice. Drinks bourbon occasionally but has been
pouring it down the sink since Book 1. No family mentioned — the job IS his
life. Speech: direct, spare, occasionally sardonic.

**Dr. Lena Hanson** — Chief forensic pathologist. Corbin's constant. Sharp
cheekbones, sharper eyes. Dry, clinical, dark-humored. Expresses care through
precision and reliability, not words. Always has two cups of coffee (one for
Corbin). The moral anchor of the series. Speech: clipped, precise, occasionally
warm when it matters.

**Carmen Reyes** — Stonebridge resident. 70s. Dominican. Lead plaintiff in the
class-action lawsuit. Patient, wise, refuses empty promises. "Don't make
promises. Just do what you can." Represents the people the city forgets.

**Ruth Bledsoe** — Runs the Greystone Community Center. 70s. Gravel voice,
steel backbone. Reads Silas Croft's letters at community meetings. Represents
community memory.

**Martin Dahl** — Professional fixer. Former military police/combat medic.
Average in every physical dimension — designed to be forgettable. Calm,
articulate, operates on cost-benefit analysis rather than morality. "I'm a
pragmatist." Killed Elena Marsh, testified against Kessler, served 3 years.
Returns in Book 5.

**DA Margaret Liu** — District Attorney. Tough, principled, wears glasses she
sometimes wants to throw. "Everything by the book." Corbin's legal ally.

**Darnell Watts** — Wrongfully convicted in Book 4. Calm, dignified, patient.
Spent 15 years knowing the truth. "Anger is expensive in here." Now free,
becoming an advocate.

**Captain Torres** — New captain (post-Briggs). Transferred from another
district. Implementing reforms. Not yet a fully developed character — room for
growth in later books.

### Timeline (series_timeline_update)

| Event | Book | Impact |
|-------|------|--------|
| Millhaven Chemical contaminates Greystone groundwater | Backstory | Sets up the city's original sin |
| Silas Croft fired for whistleblowing | Backstory | Croft's origin |
| Margaret Croft dies of lung cancer | Backstory (2002) | Croft's motivation |
| Sarah Greenfield murdered, Darnell Watts wrongfully convicted | 15 years before Book 4 | Corbin's foundational mistake |
| Four factory-connected executives murdered by Croft | Book 1 | Corbin's first case in the series |
| Croft arrested at Millhaven Plant 3 | Book 1 | |
| Elena Marsh murdered (staged suicide) | Book 2 | |
| Kessler arrested for $14M fraud, Dahl gets 3-year deal | Book 2 | |
| Phelps resigns | Book 2 | |
| Silas Croft dies in holding cell | Book 2–3 gap | |
| Detective Marco Ruiz murdered (staged suicide) | Book 3 | |
| Halloran, Beck, Cross arrested by FBI | Book 3 | |
| Briggs retires, Torres becomes captain | Book 3 | |
| Croft's gear buried with Margaret | Book 3 | |
| Watts exonerated, Craig Linden arrested | Book 4 | |
| Corbin's case review — no other wrongful convictions found | Book 4 | |
| Stonebridge affordable housing construction begins | Book 4 era | |
| Dahl released from federal prison | Book 5 | |

---

## Environment Progression

All Series Forge integration work follows the standard environment pipeline:

| Environment | URL | Purpose | Data |
|-------------|-----|---------|------|
| **dev** | `.env.dev` | Build + iterate on MCP tools, test API surface | Throwaway series/stories for validation |
| **demo** | `.env.demo` | Validate end-to-end workflows, multi-agent experiment | Full Corbin series dry-run, voice comparison |
| **prod** | `.env.prod` | Final publication + narration | Canonical Corbin books |

### Rules

- **Dev**: Break things freely. Create test series, populate junk characters,
  hammer the API. No expectation of data quality.
- **Demo**: Treat as dress rehearsal. Populate the real series bible, run the
  real Codex experiment, generate real audiobooks. Fix issues before prod.
- **Prod**: Final pass only. Books 1–4 are already here (direct writes). Books
  5–12 go to prod only after the workflow is validated on demo.

### What this means for tool implementation

The MCP tools already support `--url` and `--token` overrides. The Series Forge
tools must follow the same pattern — every tool accepts optional `url` and
`token` parameters, defaulting to `PROSEFORGE_URL` and `PROSEFORGE_TOKEN` from
the process environment. No environment should be hardcoded.

### What this means for the series bible

The series bible (world, characters, timeline) will be populated three times:
1. **Dev** — rough draft, testing API calls
2. **Demo** — refined version, used for the multi-agent experiment
3. **Prod** — final version, canonical

The content is the same; the environment determines where it lives. The bible
content in this design doc is the source of truth — push it to whichever
environment you're targeting.

---

## Implementation Plan

### Phase 1: Series Forge MCP Tools (dependency: #96 Phase 2)

Implement the 27 Series Forge tools in the workbench following the existing
architecture pattern:

1. `internal/api/series.go` — API wrapper methods
2. `internal/series/service.go` — Service layer
3. `cmd/mcp/tools_series.go` — MCP tool registration
4. `cmd/cli/cmd_series.go` — CLI commands (for testing)

**Priority tools** (needed for the Corbin experiment):
- `series_create`, `series_get`, `series_update`
- `series_world_get`, `series_world_update`
- `series_character_create`, `series_character_list`, `series_character_get`
- `series_timeline_get`, `series_timeline_update`
- `series_stories_add`, `series_stories_list`
- `series_generate`, `series_plan`

### Phase 2: Populate Corbin Series

Using the implemented tools:
1. Create the Corbin series entity
2. Write the world document (from bible above)
3. Create character entries for all recurring characters
4. Build the timeline
5. Link Books 1–4 as existing stories
6. Test `series_plan` handoff with a Book 5 outline

### Phase 3: Multi-Agent Experiment

1. Assign books to agents (Claude vs Codex)
2. Claude books: direct write via `section_write` (current workflow)
3. Codex books: `series_plan` → StorySeed-seeded chat → generation pipeline
4. Compare voice consistency across agents
5. Document findings

### Phase 4: Upstream Feedback

Based on real usage, file upstream tickets for:
- StorySeed format adjustments
- Missing API capabilities
- Context window optimization for series with many prior installments
- Anti-pattern enforcement in the generation pipeline

---

## Book Inventory

### Completed
| # | Title | Story ID | Sections | Ticket |
|---|-------|----------|----------|--------|
| 1 | Rust & Bone | `2d3fed2a-1ffb-4a43-a171-fd5e26cedfbb` | 7 | #94 |
| 2 | The Gilt Edge | `23a91358-6b71-418f-ab3f-78afe66140b2` | 8 | #101 |
| 3 | Dead Reckoning | `98b26784-11be-4b79-ac22-851fd4f3f32e` | 7 | #104 |
| 4 | False Positive | `7a6d3f12-fce9-4c26-afcb-68de77059ba9` | 7 | #105 |

### Planned
| # | Title | Core Question | Agent |
|---|-------|--------------|-------|
| 5 | The Debt | Dahl returns — can Corbin work with Elena's killer? | TBD |
| 6 | The Deserving | A monster is murdered — is justice conditional? | TBD |
| 7 | Testimony | Zero physical evidence — can Corbin trust others' eyes? | TBD |
| 8 | The Pathologist | Hanson's story — investigate the one person you trust | TBD |
| 9 | Cold Ground | Decades-old cold case — detective as archaeologist | TBD |
| 10 | The Apprentice | Training a young detective — confronting your own reflection | TBD |
| 11 | The Return | Darnell Watts brings Corbin a case — guilt becomes partnership | TBD |
| 12 | Full Circle | The last case — back to the ruins, back to the beginning | TBD |

Agent assignment (Claude vs Codex) to be determined after Series Forge tools
are implemented and the StorySeed mechanism is validated.

---

## Agent Prompt Template (BYOAI Direct Write)

For books assigned to Codex or another external agent, the ticket should include
a **generation prompt** that the agent can use. The agent reads the series bible
from Series Forge tools first, then writes sections via `section_write`.

```
You are writing Book [N] of the Corbin detective series.

BEFORE WRITING:
1. Call series_world_get to read the world document (city, themes, prose rules)
2. Call series_character_list + series_character_get for recurring characters
3. Call series_timeline_get for the event history
4. Review the anti-patterns list in the world document

THIS BOOK:
- Title: [title]
- Core question: [one sentence]
- Plot outline: [section-by-section outline from ticket]
- New characters: [any book-specific characters not in the series bible]
- Emotional arc: [how Corbin changes across this book]

WRITING:
- Use story_create to create the story
- Use section_create for each section
- Use section_write to write each section (1,500–2,500 words)
- Every section ends with a hook
- Write for audiobook narration
- Use story_publish when complete
- Use narration_start to begin audiobook generation

VOICE RULES:
- Match the series prose style from the world document
- Do NOT read prior books' prose — use only the structured series context
- Third-person close, locked to Corbin's POV

ANTI-CONTAMINATION:
- Check the anti-patterns list in the world document
- Do not reuse plot structures, confrontation settings, or symbolic objects
  from prior installments
- This book must stand alone as a complete story
```

## Story Forge Pipeline Prompt (Test Case 3)

For testing the platform's own generation pipeline:

```
1. Call series_plan with the book outline — this creates a StorySeed-seeded
   Story Forge Chat session
2. The platform AI (magnum + gemini) receives series context automatically
3. Guide the chat to refine the story outline
4. Finalize → generation pipeline → meta review → approve → section generation
5. Publish + narrate
```

This path exercises the upstream `series_plan` → `StorySeed` handoff and
validates whether the platform's built-in AI produces voice-consistent output
from structured series context alone.
