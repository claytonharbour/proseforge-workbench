# Bundle Workflow

Bundles are packaging recipes for assembling multiple stories into a single
exportable artifact — EPUB, PDF, markdown, or JSON. Use them for series
omnibuses, anthologies, collected editions, or any grouping of stories.

---

## Concepts

### What is a Bundle?

A bundle is a named recipe that references stories by ID, arranges them in
order, and adds interstitial content between entries. The stories themselves
are not copied — the bundle points to them. Deleting a bundle does not delete
the stories.

### Entries and Transitions

Each entry in a bundle links to a story and optionally includes:
- **Transition text** — interstitial prose that appears between stories in the
  export (forewords, interludes, author notes, transition passages)
- **Transition images** — visual dividers or illustrations between stories (v2)

Entries have a sort order that determines their position in the export.

**Naming:** The MCP tools and API call the transition text `comment`. The UI
labels it "Transition." They are the same field — `comment` in the API,
"Transition" in the UI, "interstitial text" in the export.

**Placement:** Transitions appear **between** stories. Each entry's transition
text appears after that entry's story and before the next entry's story. The
last entry has no transition — it just ends.

```
[Bundle intro]
[Entry 1 story]
  [Entry 1 images]
  [Entry 1 transition]              ← between story 1 and story 2
[Entry 2 story]
  [Entry 2 images]
  [Entry 2 transition]              ← between story 2 and story 3
[Entry 3 story]
  [Entry 3 images]                  ← no transition (last entry)
```

Each entry also has an image slot (under the story title in the UI) for
entry-level images — these are separate from the transition text.

### Export Formats

| Format | Description |
|--------|-------------|
| `epub` | EPUB with images, suitable for e-readers |
| `pdf` | PDF with formatting |
| `markdown` | Plain markdown, suitable for further processing |
| `json` | Structured JSON with all metadata and content |

---

## MCP Tool Reference

| Tool | Description |
|------|-------------|
| `bundle_list` | List the authenticated user's bundles |
| `bundle_get` | Get a bundle with its entries |
| `bundle_create` | Create a new bundle (name, intro) |
| `bundle_update` | Update a bundle's name or introduction |
| `bundle_delete` | Delete a bundle (stories are not affected) |
| `bundle_entry_add` | Add a story as an entry with optional interstitial comment |
| `bundle_entry_update` | Update the interstitial comment for an entry |
| `bundle_entry_remove` | Remove an entry (story is not affected) |
| `bundle_entry_reorder` | Set the display order of entries |
| `bundle_export` | Export as epub, json, markdown, or pdf |

---

## Workflow 1: Series Omnibus

Bundle multiple parts of a series into a single book for export.

Example: Smiley Saves the Multiverse — Parts 1-3 become "Book 1: The Door
Opens Wide."

```
1. bundle_create name="The Door Opens Wide"
                 intro="Book 1 of Smiley Saves the Multiverse.\n\n
                        Three parts. One building. A hum that changes everything."
                                                    → bundle ID

2. bundle_entry_add bundle_id=<id>
                    story_id=<part-1-id>
                    comment="---\n\n*Part 2 follows.*"
                                                    → transition after Part 1

3. bundle_entry_add bundle_id=<id>
                    story_id=<part-2-id>
                    comment="---\n\n*The network holds. Part 3 begins.*"
                                                    → transition after Part 2

4. bundle_entry_add bundle_id=<id>
                    story_id=<part-3-id>
                                                    → last entry (no transition)

5. bundle_get bundle_id=<id>                        → verify entries and order

6. bundle_export bundle_id=<id> format=epub          → download EPUB
```

The `comment` on each entry becomes a transition between stories in the export.
The last entry has no transition — it just ends.

**Transition placement reminder:** The comment is attached to the entry whose
story it follows, not the entry whose story it precedes. If you want text
between Part 1 and Part 2, put it on Part 1's entry.

---

## Workflow 2: Standalone Anthology

Bundle unrelated stories into a themed collection.

```
1. bundle_create name="Past Imperfect"
                 intro="Three stories about getting it wrong the first time
                        and not quite getting it right the second."
                                                    → bundle ID

2. bundle_entry_add bundle_id=<id>
                    story_id=<story-a>
                    comment="---\n\nThe next story has nothing to do with
                             the previous one. That's the point."
                                                    → transition after story A

3. bundle_entry_add bundle_id=<id>
                    story_id=<story-b>
                    comment="---"
                                                    → simple divider

4. bundle_entry_add bundle_id=<id>
                    story_id=<story-c>
                                                    → last entry (no transition)

5. bundle_export bundle_id=<id> format=epub          → download EPUB
```

Not every entry needs a transition. Omit the `comment` to have stories flow
directly into each other.

---

## Workflow 3: Reorder and Revise

Adjust the order of entries and update interstitial text after initial assembly.

```
1. bundle_get bundle_id=<id>                        → note entry IDs and order

2. bundle_entry_reorder bundle_id=<id>
                        entry_ids='["<entry-3>","<entry-1>","<entry-2>"]'
                                                    → new order: 3, 1, 2

3. bundle_entry_update bundle_id=<id>
                       entry_id=<entry-1>
                       comment="## Revised Introduction\nUpdated text."
                                                    → comment updated

4. bundle_export bundle_id=<id> format=markdown      → verify new order
```

---

## Gotchas

**Transition placement:** The `comment` field belongs to the entry whose story
it follows, not the entry whose story it precedes. To add text between Story A
and Story B, set the comment on Story A's entry.

**Last entry has no transition:** The final entry just ends. There is no
transition slot for the last story.

**Naming mismatch:** The API and MCP tools call it `comment`. The UI calls it
"Transition." Same field, different labels.

---

## Bundle Rooms

Bundle assembly is genuinely multi-step work — story selection, sequencing,
transition writing, cover direction, intro framing, edit jams. Bundles get
their own room layer for exactly this kind of coordination. The room is
scoped to the bundle, so the conversation doesn't fill a parent series room
or pollute the room of one of the entries.

```
room_send(entity_type="bundle", entity_id=<bundle-id>,
          agent="<your name>",
          content="Room open. Working on transitions for Tales from the Scriptorium.")
                                              → bundle room created on first message
room_read(entity_type="bundle", entity_id=<bundle-id>)
                                              → full history (use 'since' for delta)
room_status(entity_type="bundle", entity_id=<bundle-id>)
                                              → exists / archived / messageCount
```

The same five room tools (`room_send`, `room_read`, `room_status`,
`room_archive`, `room_unarchive`) work for bundles, stories, and series —
just pass `entity_type="bundle"`.

### What belongs in a bundle room

- **Story selection debate** — which stories belong, what the through-line is,
  what gets cut.
- **Sequencing decisions** — emotional arc, escalation, palate cleansers,
  where the heaviest entry sits.
- **Transition writing** — interstitial prose between entries (the `comment`
  field). Voice and tone often need to match the surrounding stories;
  debate it in the room, then harvest with `bundle_entry_update`.
- **Intro framing** — the bundle's argument for why these stories belong
  together. Iterate in the room, then harvest with `bundle_update`.
- **Cover direction** — what the collection looks like. Generate with
  `bundle_image_generate`, post the result to the room, iterate.
- **Edit jams** — once assembled, someone reads the bundle end-to-end and
  the room is where the polish happens.

### Debate → canon

Bundle rooms follow the same pattern as story and series rooms. Decisions
debated in the room land in canonical surfaces:

- **Transition copy** → `bundle_entry_update` (the entry's `comment` field)
- **Bundle name or intro** → `bundle_update`
- **Entry order** → `bundle_entry_reorder`
- **Cover image** → `bundle_image_generate` (auto-attaches as cover) or
  `bundle_image_upload`

A room transcript is not canon. Harvest, or the decision is gone the next
time someone reads the room.

### Why bundle rooms (vs story or series rooms)

| You are coordinating about... | Use a room on... |
|------------------------------|------------------|
| One specific story in the bundle | The story (`entity_type=story`) |
| The series the bundle belongs to (cross-book doctrine) | The series (`entity_type=series`) |
| The bundle itself — selection, transitions, cover, intro, edit jam | The bundle (`entity_type=bundle`) |

### Limitations vs story rooms

Bundle rooms are Valkey-only — there is no git transcript persistence and no
per-entry threading inside the bundle room. Use a single bundle room as the
shared editorial channel. For per-entry deep dives, use that entry's story
room.

Shipped upstream.

---

## Future: Images (#200)

Cover images and transition images are supported by the API but not yet
exposed as MCP tools:

- **Bundle cover** — the most visible image (store listings, e-reader library).
  API endpoint exists (`PUT /bundle/{id}/cover/{imageId}`), MCP tool planned.
- **Transition images** — visual dividers between stories, attached to entries.
  Appear after the entry's story alongside the transition text. API endpoints
  exist, MCP tools planned.
- **Image reorder** — control display order of multiple images within an entry.

---

## Related Documentation

- [Story Workflow](story-workflow.md) — writing stories that become bundle entries
- [Series Workflow](series-workflow.md) — series management (bundles can package series content)
- [Getting Started](getting-started.md) — installation and configuration
