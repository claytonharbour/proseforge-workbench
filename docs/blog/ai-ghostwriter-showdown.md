# AI Ghostwriter Showdown: Claude vs Codex

*We gave two AI models the same job — rewrite a story for audiobook. They took very different paths to very different strengths.*

---

We've been building tools that let AI agents write and review stories on [ProseForge](https://app.proseforge.com). Every story on the platform goes through our [Re-Forge rewrite system](https://claytonharbour.com/blog/re-forge-ai-story-rewrite-system) — an automated rewrite pass that feeds each section back through a larger model with full story context and machine-detected quality issues, then lets the author cherry-pick changes from the diff. That system improved average scores by 35% across 35 controlled experiments.

Re-Forge rewrites whole sections, but it's guided by what the checker finds and constrained to preserve the original structure. We wanted to know what happens when you throw away the guardrails entirely and hand the whole story to a different AI with one instruction: make it good enough to listen to on a walk.

No guardrails. No "preserve the author's voice." Full ghostwrite.

Here's what happened.

## The Contenders

**Claude** got [Rust & Bone](https://app.proseforge.com/@clayton/rust-bone/read) — a mystery about industrial decay and a killer with a grudge. Seven sections. The original had been through Re-Forge but was still flat and repetitive — the premise had teeth but the prose didn't: someone is murdering the executives who destroyed a working-class neighborhood, leaving their bodies arranged like sculptures in the ruins of the factories they closed.

**Codex** got [The Last Oracle of Elyria](https://app.proseforge.com/@clayton/the-last-oracle-of-elyria/read) — an epic fantasy about prophecy, broken seals, and a world held together by ancient sacrifice. Five sections. Also Re-Forged, still generic chosen-one fare with wiki-style exposition. The bones were there, but the flesh was cardboard.

Both stories were rewritten end-to-end through our workbench MCP tools. Export, rewrite, write back, publish, narrate. The same pipeline, the same constraints, different models.

## What Claude Built

Rust & Bone came back as a tight, propulsive thriller told entirely through Detective Miles Corbin — a man who's seen too many bodies and is starting to sympathize with the people leaving them.

Claude's instinct was compression. 9,700 words across seven sections, averaging 1,400 words each. No scene wasted. The story opens with "The body was wrong" and never lets up. Every section ends on a hook. Every character earns their page time in a handful of lines.

The real achievement is the villain. Silas Croft is a 74-year-old former factory worker whose wife died of cancer from contaminated groundwater. He's been killing the executives who covered it up, leaving corroded machine gears in their hands like rusted rosaries. He's sympathetic without the story excusing him. When Corbin finally confronts him in the basement of the factory where it all started, the dialogue is the best writing in either story:

> "Nothing brings her back. But every system has a ledger, Detective. Credits and debits. For forty years, the debits have piled up and no one's paid."

And later, in jail:

> "I was strong enough to do something much easier. I was strong enough to be angry."

The ending doesn't resolve. The city is still broken. The groundwater is still poisoned. Corbin drives toward the next case. For a walk-and-listen audiobook, that lingering unease is perfect.

## What Codex Built

The Last Oracle of Elyria came back as something considerably more ambitious — a full novella at 13,200 words with a magic system, a morally complex antagonist, and set pieces that read like cinema.

Codex's instinct was depth. Where Claude compressed, Codex expanded. The generic "elements" became five named currents — Root, Tide, Ember, Gale, and Veil — felt through the body rather than described from a distance. "Root clenched in her knees. Tide surged into her throat. Ember flashed in her blood." That's world-building you can feel in your chest.

The real achievement is the antagonist's argument. Captain Kael isn't evil. He's a man whose family drowned when the Oracle council decided his city was an acceptable sacrifice for the greater good. His fury is earned, and his logic is dangerous because it's partly right:

> "I want no more children promised to altars before they know their own names. I want no councils choosing which towns drown so their histories remain tidy. I want no sacred machine deciding what grief is affordable."

The Namarre undercroft fight is the set piece of the entire experiment — multiple combatants, a vision of the drowning of Vallorne mid-battle, a beloved character dragged under by shadow-arms, and a blind archivist who saves everyone by smashing a vial of memory onto the floor. It's a lot. It works.

The ending rejects both "preserve the old sacred machine" and "burn it all down," landing on a third path that fits the protagonist's growth. Aria doesn't become a queen or a savior. She becomes a listener. That's harder to pull off than it sounds.

## The Verdict

Here's the honest comparison:

| | Claude (Rust & Bone) | Codex (Last Oracle) |
|---|---|---|
| **Word count** | 9,700 | 13,200 |
| **Sections** | 7 | 5 |
| **Avg section length** | 1,400 | 2,600 |
| **Characters** | 5 | 9 |
| **Dialogue** | Natural, snappy | Formal, weighted |
| **Pacing** | Relentless | Expansive |
| **Best moment** | Croft's surrender | Namarre undercroft |
| **Weakness** | Sections run short | Dense for audio |

**Claude wrote a better audiobook.** The rhythm is tuned for the ear. The character count is manageable — you never lose track of who's talking. The dialogue sounds spoken, not performed. The pacing never lets you drift. If you're walking and listening, Rust & Bone is the one that keeps your feet moving.

**Codex wrote a better story.** The world is richer. The moral stakes cut deeper. The ambition is higher. But it asks more of the listener — longer passages of world description, more names to track, a more formal cadence that works beautifully on the page and requires more attention through earbuds.

Neither model played it safe. Claude leaned into noir restraint. Codex leaned into epic scale. Both delivered something genuinely worth reading — which, given that the source material was flat AI-generated prose, is the real story here.

## Quality Scores

We ran both stories through ProseForge's quality checker — a code-based assessment that scores manuscripts across five weighted dimensions — **before and after the ghostwrite**. The version history API lets us score the original AI-generated prose and the rewritten version at the exact commit SHAs where each existed.

### Before: The Original AI-Generated Prose

| Dimension | Weight | Rust & Bone (original) | Last Oracle (original) |
|-----------|--------|----------------------|----------------------|
| Continuity | 30% | **10**/10 | 6.5/10 |
| Progression | 25% | 7/10 | 6/10 |
| Coherence | 20% | 7/10 | **9**/10 |
| Tone | 15% | 1/10 | 1/10 |
| **Overall** | | **7.00** (3.5 stars) | **6.00** (3.0 stars) |

The tone scores tell the real story. Both originals scored **1 out of 10** — the checker flagged massive cross-section repetition. Phrases like "ready to face whatever challenges lay ahead," "heart pounded in her chest," and "the road ahead was fraught with" appeared verbatim across three or more sections. That's the signature of AI-generated prose that hasn't been edited: each section sounds fine in isolation, but stitch them together and the machine starts repeating itself.

The Oracle also had a gender flip (Lyrien switches from male to female mid-story), story restarts ("in the heart of" reopening the narrative), and sections that end mid-sentence. The bones were there, but the flesh was cardboard.

### After: The Ghostwrite

| Dimension | Weight | Claude (Rust & Bone) | Codex (Last Oracle) |
|-----------|--------|---------------------|---------------------|
| Continuity | 30% | **10**/10 | **9**/10 |
| Progression | 25% | **10**/10 | **10**/10 |
| Coherence | 20% | 4.5/10 | 1/10 |
| Tone | 15% | **9**/10 | **9.5**/10 |
| **Overall** | | **8.61** (4.3 stars) | **7.58** (3.8 stars) |

### The Delta

| Dimension | Rust & Bone | Last Oracle |
|-----------|-------------|-------------|
| Continuity | 10 → 10 | 6.5 → **9** (+2.5) |
| Progression | 7 → **10** (+3) | 6 → **10** (+4) |
| Coherence | 7 → 4.5 (-2.5) | 9 → 1 (-8) |
| Tone | 1 → **9** (+8) | 1 → **9.5** (+8.5) |
| **Overall** | 7.00 → **8.61** (+1.61) | 6.00 → **7.58** (+1.58) |

The dimensions that matter for audiobook listening — progression, continuity, and tone — jumped dramatically. Both models eliminated the cross-section repetition that plagued the originals. Progression went from restart-prone to flawless. Continuity held or improved.

The coherence drop is the checker, not the prose. Both models write short, punchy paragraphs with sentence fragments for dramatic effect — noir restraint in Claude's case, cinematic pacing in Codex's. The checker sees "180 short paragraphs" and flags it. It also flags fantasy terminology and evidence lists as "word-salad." We've filed upstream tickets to teach the checker about literary style.

The real takeaway: **tone is where the ghostwrite earns its keep.** Taking a story from 1/10 to 9/10 on tone means eliminating the uncanny repetition that makes AI prose sound like AI prose. That's the difference between "this was clearly generated" and "this was written."

## What This Means

The ghostwriter workflow works. Export a story, hand it to an AI with the right prompt, write it back, publish, narrate. End to end, each rewrite took about fifteen minutes of model time. The results are publishable.

We've [written before](https://claytonharbour.com/blog/orchestration-beats-model-swaps) that orchestration beats model swaps — that pipeline design, state management, and feedback loops matter more than which model you pick. That's still true for **generation**, where you're writing section-by-section and the system needs to prevent character drift and maintain continuity across a long pipeline.

But ghostwriting is a different job. The model isn't generating from prompts through a multi-step pipeline — it's reading a complete story and rewriting it in one pass per section. The orchestration is identical (export, rewrite, write back). What differs is the creative instinct the model brings to the rewrite. Claude compressed. Codex expanded. Same tools, same constraints, completely different stories.

**For generation, invest in orchestration. For ghostwriting, the model is the voice.** The BYOAI thesis accounts for both: the platform provides the structure, the user picks the AI that fits the job.

We're still building. The tools are [open source](https://github.com/claytonharbour/proseforge-workbench). The stories are live. And somewhere in the margins, a turkey watches with faint judgment.

---

*Read the stories:*
- *[Rust & Bone](https://app.proseforge.com/@clayton/rust-bone/read) — a thriller about industrial decay, written by Claude*
- *[The Last Oracle of Elyria](https://app.proseforge.com/@clayton/the-last-oracle-of-elyria/read) — an epic fantasy about prophecy and sacrifice, written by Codex*
- *[The Forge of Forgotten Scrolls](https://app.proseforge.com/@clayton/the-forge-of-forgotten-scrolls/read) — the story behind the tools, featuring an inexplicable turkey*
