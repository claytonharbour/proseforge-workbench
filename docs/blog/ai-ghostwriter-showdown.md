# AI Ghostwriter Showdown: Claude vs Codex

*We gave two AI models the same job — rewrite a story for audiobook. They took very different paths to very different strengths.*

---

We've been building tools that let AI agents write and review stories on [ProseForge](https://app.proseforge.com). Last week we decided to stress-test the ghostwriter workflow: take two unpublished stories, hand each to a different AI model, and see what comes back.

The rules were simple. Read the story. Rewrite every section. Make it good enough to listen to on a walk. Publish it. Start the audiobook.

No guardrails. No "preserve the author's voice." Full ghostwrite.

Here's what happened.

## The Contenders

**Claude** got [Rust & Bone](https://app.proseforge.com/@clayton/rust-bone/read) — a mystery about industrial decay and a killer with a grudge. Seven sections. The original was flat and repetitive, but the premise had teeth: someone is murdering the executives who destroyed a working-class neighborhood, leaving their bodies arranged like sculptures in the ruins of the factories they closed.

**Codex** got [The Last Oracle of Elyria](https://app.proseforge.com/@clayton/the-last-oracle-of-elyria/read) — an epic fantasy about prophecy, broken seals, and a world held together by ancient sacrifice. Five sections. The original was generic chosen-one fare with wiki-style exposition. The bones were there, but the flesh was cardboard.

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

## What This Means

The ghostwriter workflow works. Export a story, hand it to an AI with the right prompt, write it back, publish, narrate. End to end, each rewrite took about fifteen minutes of model time. The results are publishable.

But the comparison reveals something more interesting: **the model you choose shapes the story as much as the prompt does.** Same genre guidance. Same structural constraints. Same tools. Completely different creative instincts.

That's the BYOAI thesis in action. The platform provides the structure. The AI provides the voice. The writer picks which voice fits the story they're trying to tell.

We're still building. The tools are [open source](https://github.com/claytonharbour/proseforge-workbench). The stories are live. And somewhere in the margins, a turkey watches with faint judgment.

---

*Read the stories:*
- *[Rust & Bone](https://app.proseforge.com/@clayton/rust-bone/read) — a thriller about industrial decay, written by Claude*
- *[The Last Oracle of Elyria](https://app.proseforge.com/@clayton/the-last-oracle-of-elyria/read) — an epic fantasy about prophecy and sacrifice, written by Codex*
- *[The Forge of Forgotten Scrolls](https://app.proseforge.com/@clayton/the-forge-of-forgotten-scrolls/read) — the story behind the tools, featuring an inexplicable turkey*
