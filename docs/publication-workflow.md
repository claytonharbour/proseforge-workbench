# Publication Workflow — From ProseForge to Storefront

> Vision document for automating the last mile between "story complete on
> ProseForge" and "book live on KDP / Google Play / other storefronts."
>
> Ticket: forge/proseforge-workbench#162

## The Gap

ProseForge handles the creative pipeline well: writing, reviewing, images,
narration, quality scoring. But the moment a book is ready to publish, the
workflow drops to manual copy-paste across multiple storefronts, each with
its own taxonomy, field limits, and formatting expectations.

The gap is not large per book, but it multiplies across a 12-book series
and across platforms. A single Corbin book currently requires:

| Step | Platform | Manual Work |
|------|----------|-------------|
| Description | KDP | Write ~250 words, paste into description field |
| Description | Google Play | Same text, different formatting rules |
| Categories | KDP | Pick 2 BISAC paths from a search UI |
| Categories | Google Play | Pick from Google's taxonomy (different from BISAC) |
| Keywords | KDP | Choose 7 keyword phrases (not visible to readers, drives discoverability) |
| Keywords | Google Play | Not applicable (Google uses categories only) |
| Cover | KDP | Upload JPG/TIFF, 2560×1600 recommended |
| Cover | Google Play | Upload PNG/JPG, different aspect requirements |
| Pricing | Both | Set price, territories, DRM |
| Series info | Both | Link to series, set book number |

Multiply by 12 books × 2+ platforms = ~240 manual steps.

## What ProseForge Already Knows

Most of this metadata is derivable from data ProseForge already holds:

| Field | Source |
|-------|--------|
| Description | story_meta (premise, characters, plot) + story content |
| BISAC categories | Genre + subgenre signals from content |
| Keywords | Themes, character types, setting, tropes from content |
| Cover | Already generated (image_generate) |
| Series info | Already in series_stories_list |
| Book number | Already in series ordering |

The missing piece is a **publication prep layer** that transforms creative
data into storefront-ready metadata.

## BISAC Categories

BISAC (Book Industry Standards and Communications) is the standard taxonomy
for book categorization. KDP's category selector accepts BISAC path strings
directly — if the tool outputs exact strings, the user can paste them into
the search field with zero translation.

Format: `FICTION / Mystery & Detective / Police Procedural`

Key insight from the current manual workflow: **ChatGPT outputs BISAC strings
that paste directly into the KDP UI.** This is the target format for any
automation we build.

### Corbin Series Categories (reference)

**Primary (all books):**
`FICTION / Mystery & Detective / Police Procedural`

**Secondary (rotate based on book theme):**
- `FICTION / Thrillers / Crime` — Books 1, 2, 3, 5
- `FICTION / Mystery & Detective / Hard-Boiled` — Books 1, 3, 6
- `FICTION / Thrillers / Suspense` — Books 4, 7, 8
- `FICTION / Mystery & Detective / Collections & Anthologies` — if bundled

### Google Play Categories

Google uses a different taxonomy. Approximate mappings:

| BISAC | Google Play |
|-------|-------------|
| FICTION / Mystery & Detective / Police Procedural | Fiction > Mystery, Thriller & Suspense > Mystery > Police Procedurals |
| FICTION / Thrillers / Crime | Fiction > Mystery, Thriller & Suspense > Thrillers > Crime |
| FICTION / Mystery & Detective / Hard-Boiled | Fiction > Mystery, Thriller & Suspense > Mystery > Hard-Boiled |

Google's separator is ` > ` not ` / `. Google's path includes
"Mystery, Thriller & Suspense" as a grouping level that BISAC doesn't have.

## KDP Keywords

KDP allows 7 keyword phrases per book. These are not visible to readers —
they function like search metadata. Best practices:

**Do:**
- Use multi-word phrases (2-4 words), not single words
- Include subgenre terms readers actually search for
- Include setting/mood descriptors ("Rust Belt", "gritty", "noir")
- Include trope terms ("wrongful conviction", "cold case", "corrupt cops")
- Vary 3-4 keywords per book while keeping 3 consistent across the series

**Don't:**
- Repeat words already in title/subtitle (KDP indexes those separately)
- Use author names as keywords (against ToS)
- Use generic terms ("book", "fiction", "novel")
- Duplicate the same keyword in different forms

### Series-Consistent Keywords (use on all Corbin books):
1. `police procedural series`
2. `detective thriller series`
3. `Rust Belt crime fiction`

### Per-Book Keywords (vary by theme):

| Book | Keywords (4 per book) |
|------|----------------------|
| 1 — Rust & Bone | vigilante justice, factory town murder, environmental crime, retired detective |
| 2 — The Gilt Edge | real estate corruption, gentrification thriller, financial crime, urban development |
| 3 — Dead Reckoning | police corruption, internal affairs, blue wall of silence, evidence tampering |
| 4 — False Positive | wrongful conviction, innocence project, DNA exoneration, cold case murder |
| 5 — The Debt | organized crime network, judicial corruption, contract killer, shadow government |
| 6 — The Deserving | domestic violence mystery, community leader murder, family secrets, moral ambiguity |
| 7 — Testimony | unreliable testimony, domestic abuse case, courtroom thriller, he said she said |
| 8 — The Pathologist | forensic mystery, medical examiner thriller, evidence chain, institutional trust |
| 9 — Cold Ground | cold case investigation, buried evidence, generational crime, environmental poisoning |
| 10 — The Apprentice | new detective partner, mentor thriller, training case, police procedural |
| 11 — The Return | exonerated man returns, redemption thriller, wrongful conviction aftermath, second chance |
| 12 — Full Circle | series finale, career retrospective, systemic corruption, justice delayed |

## Proposed Tooling (Levels)

### Level 0 — Now (manual, agent-assisted)

The agent writes KDP descriptions in conversation. The user copy-pastes.
Categories and keywords are suggested by the agent using this document
as reference. Works but doesn't scale and requires the agent to have
context on every book.

### Level 1 — Workbench MCP Tools

New tools on the proseforge-workbench MCP server:

```
story_publish_prep
  story_id: string
  platform: "kdp" | "google" | "all"
  
  Returns:
    description: string (formatted for platform)
    categories: string[] (BISAC paths for KDP, Google paths for Google)
    keywords: string[] (7 phrases for KDP, omitted for Google)
    cover_spec: { url, recommended_size, current_size }
    series_info: { name, book_number, total_books }
    consistency_check: { shared_categories, shared_keywords }
```

This tool reads the story's meta, content, genre, series context, and
quality scores, then generates platform-specific metadata. The output
is copy-paste ready — BISAC strings for KDP, Google taxonomy strings
for Google Play.

Implementation: the tool calls the AI provider with a structured prompt
that includes the story summary, genre, series context, and the BISAC
taxonomy. The AI returns categories and keywords in the exact format
needed. No human interpretation required.

### Level 2 — Platform Storage

Add publication metadata fields to the ProseForge story record:

```
story.publication_metadata:
  kdp:
    description: string
    categories: string[2]
    keywords: string[7]
    asin: string (after publish)
  google:
    description: string
    categories: string[]
    google_id: string (after publish)
```

This lets the metadata persist across sessions. The "publish prep" view
shows all fields for all platforms with copy buttons. The agent can
update individual fields via `story_publish_meta_update`.

### Level 3 — Direct Upload

Google Play Books has a [partner API](https://developers.google.com/books/partner)
that accepts metadata and manuscript uploads. KDP does not have a public
API, but the metadata can be pre-staged.

For Google Play, a `story_publish_google` tool could:
1. Export the story as EPUB
2. Upload the cover at the correct dimensions
3. Set categories, description, pricing
4. Submit for review

KDP would remain manual upload but with all metadata pre-formatted
and a single "copy all" action.

## Cover Specifications by Platform

| Platform | Format | Min Size | Recommended | Aspect |
|----------|--------|----------|-------------|--------|
| KDP eBook | JPG/TIFF | 625×1000 | 2560×1600 | 1.6:1 |
| KDP Print | PDF/TIFF | varies by trim | 300 DPI | varies |
| Google Play | PNG/JPG | 640×480 | 2560×1600 | varies |
| Apple Books | PNG/JPG | 1400×1873 | 2560×3840 | ~1.5:1 |

ProseForge currently generates 1024×1024 images. For KDP covers, we'd
need to either:
- Generate at the correct aspect ratio (1.6:1 for eBook)
- Post-process the square image to add vertical space
- Use a different generation pipeline for "cover" vs "section" images

This is a separate concern from metadata but belongs in the same
publication prep workflow.

## Next Steps

1. **Immediate:** Use this document as a reference when publishing Books 3-12.
   The agent can read this doc and output correctly formatted BISAC strings
   and keyword lists without re-deriving them each time.

2. **Near-term:** Build `story_publish_prep` as a workbench MCP tool (Level 1).
   Start with KDP since that's the primary platform.

3. **Medium-term:** Add publication metadata to the ProseForge story record
   (Level 2). This removes the need for the agent to re-generate metadata
   each session.

4. **Future:** Google Play API integration (Level 3). KDP remains manual
   until/unless Amazon opens a publishing API.

## Related

- forge/proseforge-workbench#162 — Publication metadata workflow ticket
- forge/proseforge-workbench#157 — Image regeneration (completed)
- forge/proseforge-workbench#158 — Narration rebuild (in progress)
