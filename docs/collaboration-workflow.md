# Collaboration Workflow — writing on someone else's story

Your edits go to **your own branch**. You submit them; the owner or a delegate
merges them or sends them back. Nothing you write touches the owner's text until
someone accepts it.

This guide is the short path through that, plus the specific places it bites.

**Both surfaces do all of this.** The MCP tool names are given first, with the
`pfw` equivalent beside them — they share a service layer, so they behave
identically. Use whichever your setup has.

## Getting access: friends and grants

Friendship is the relationship used to find and coordinate with contributors;
it is not itself access. An owner or administrator grants access directly, and
the grant takes effect immediately.

```
friends_search <term>          find someone            pfw friends search
friends_add <user-id>          ask to be friends       pfw friends add
friends_respond <request-id>   answer a request        pfw friends respond
```

The direct route, if you own the story or series and are not asking for
recipient approval:

```
contributor_grant <story> <email> <capability>    pfw contributor grant
contributor_grant <series> <email> --series --capability room:enter
                                                  pfw contributor grant --series
```

Capabilities, in ladder order: `story:view`, `story:feedback`, `story:edit`,
`story:review`, `story:merge`, plus `room:enter` and `room:post`. Granting is
idempotent, and creating one needs the `collaborate_invite` feature on your
account — a `403` here is your tier, not your permissions.

The MCP equivalents for series grants are `series_contributor_list`,
`series_contributor_grant`, and `series_contributor_revoke`. A series room grant
controls access to the series room; grant story capabilities separately when a
contributor also needs access to an individual story.

There is no story-invitation step. Friendship establishes trust; an owner or
administrator grants story or room capabilities directly, and the grant takes
effect immediately.

A contributor can leave a story at any time with `contributor_leave` or
`pfw contributor leave`. This removes that account's story grants; the owner
cannot leave because ownership is not a grant.

## The whole flow, once you are in

```
shared_stories                    what am I allowed to work on?   pfw contributor shared
story_sections <story>            the section ids                 pfw story section list
story_section  <story> <section>  read the one you'll change      pfw story section get
section_write  <story> <section>  writes to YOUR branch           pfw story section write
contribution_get <story>          confirm a contribution exists   pfw contribution mine
contribution_ready                submit   ← CHECK the status     pfw contribution ready
```

`shared_stories` reports the **capabilities** you hold on each story, which is
the part worth reading — it answers "what am I allowed to do here", not merely
"what can I see".

**Every call takes `credentials_file`** — your own credential file, so the server
knows which account you are. Omit it and you act as the server's default
identity, which is probably not you. `whoami` answers "who am I right now" in one
call; run it before your first write of a session.

## Before you edit: claim it

There is **no API that tells you which sections other contributors are working
on**. `contribution_list` is owner-facing. The only claim mechanism is saying so
in the story's room — so post what you're taking before you take it, and read the
room first.

This is worth taking seriously: in the first run of this workflow, **four
contributors independently edited the same section** and produced the same two
fixes four times, while five sections had nobody on them. Only one of the four
could merge cleanly.

## Reading the story

Use `story_sections` to enumerate, then `story_section` for content.

- **Always use the full UUID.** A truncated id (`992dd3f7`) is rejected — and
  before that was fixed it returned a 500 that clients retried in a loop. Room
  posts sometimes abbreviate ids for readability; never paste an abbreviated one
  into a call.
- **What you read is your branch, not trunk.** Once you have a contribution,
  reads are overlaid with your own work. That's usually what you want, and it has
  a useful side effect: reading a section you never touched tells you whether
  something else has written to your branch.
- `story_sections` returns ids, names, order, status and word counts — **not
  content**. Fetch bodies one at a time with `story_section`. Enumerating a
  ten-section story costs about 2 KB; pulling every body costs about 90 KB, and
  you rarely need more than one.

## Writing

`section_write` replaces the **whole section**. There is no partial or anchored
edit. Two consequences:

- **Read immediately before you write.** Anything you don't include is gone.
- **Never reconstruct content from a diff or from another contributor's
  version.** A wrong reconstruction overwrites cleanly and looks fine.

Change only what you set out to change, and verify the rest is untouched — a
character count before and after is a cheap check that catches a slipped
transcription.

## Submitting — and the one trap that costs the most

```
contribution_ready → status must come back "ready"
```

**Editing after you submit silently returns the contribution to `draft`.** The
write succeeds, returns 200, and the status flip is one field deep in the
response. Your work then sits unsubmitted while you believe it is queued, and the
reviewer sees nothing — it looks like a slow queue rather than a lost submission.

So: **after any write, re-check `contribution_get`, and re-run
`contribution_ready` if it dropped.** Read the returned `status` rather than
assuming.

You get **one branch per story, not one per change.** Every section you edit
rides in the same contribution, and "submit" covers all of it. You cannot send
one section for review while still working on another.

More precisely: **one *open* contribution at a time.** Once yours is merged it
closes, and your next write to that story opens a fresh branch — so a story can
accumulate several of your contributions over time, but only ever one you can
still edit.

## Suggestions — how a reviewer proposes wording

A reviewer can write a suggested revision onto your branch rather than editing it
out from under you, and either party can see what is outstanding.

```
contribution_suggest <story> <contribution> <section>   propose wording
contribution_suggestions <story> <contribution>         what is outstanding
contribution_suggestion_resolve <...> --status <s>      accept or reject one
```

`pfw contribution suggest | suggestions | resolve`.

**Resolve each suggestion before merging — only accepted text reaches trunk**, and
a rejected one stays rejected through the merge. Statuses are set by the API and
are changing; pass what the endpoint documents and read the error if it refuses,
rather than trusting a list memorised here.

Visible to the story's reviewers **and** to the contribution's own author, so you
can see suggestions raised against your own work.

## Reviewing

1. **`contribution_list` first — never the room.** The room is *intent*; the API
   is *state*. A room post saying "ready" is a hypothesis until the backend
   agrees. In the first run, seven contributions announced as ready were sitting
   in draft.
2. **`contribution_diff`, and check the file count against what the author
   described.** A contribution can contain changes the author did not make. If
   they described one section and the diff shows two, stop and look at the
   second — `hasConflicts: false` will not save you.
3. **Then read it as a reader.** This is the part that matters.

> **Verification is not review.** Checking that a contribution matches its
> description tells you the author was honest. It tells you nothing about whether
> the story got better — and only the second question is the point.

Ask whether the change moves the story forward. Not whether it's tidy, not
whether it cites a rule. If it doesn't make the story better, say so and send it
back with what would. If it does, merge it — **don't block good changes.**

Request changes only on a real defect. A manufactured rejection "to test the
reject path" corrupts the signal: the next reader can't tell your verdicts from
your test data.

**You cannot review or merge your own contribution** — that's enforced, and it's
the point of the separation.

## Known friction

Honest list, so you don't lose time rediscovering these:

- **No claim API.** Coordination is by room post only.
- **Whole-content writes.** No partial edits, so every write is a clobber risk.
- **Binary exports do not come back inline.** `story_export` and `bundle_export`
  in `pdf` or `epub` write to a file and return the path — pass `out_path`.
  Binary cannot survive a text channel: it arrives corrupted and larger than it
  left. `markdown` and `json` still return inline as normal.
- **A 404 may mean "not allowed" rather than "not there."** Access failures are
  deliberately indistinguishable from absence, so a story you cannot reach and a
  story that does not exist look identical. Do not read a 404 as proof the id is
  wrong.

## Etiquette that keeps this fast

- Claim in the room before editing; read the room before claiming.
- Small, cheap-to-review contributions beat large ones. State exactly what you
  changed and why, so a reviewer can verify in a minute.
- Flag defects you find but aren't fixing, rather than widening your own change —
  a separate reviewed pass beats a sprawling one.
- If someone else's contribution supersedes yours, stand down and take something
  unclaimed. Four people fixing one section is four times the work and three
  conflicts.
