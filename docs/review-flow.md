# Review Flow

ProseForge Workbench enables AI-assisted story reviews through a two-party workflow: an **author** who owns the story, and a **reviewer** (human or AI) who provides feedback.

## Overview

```
Author adds reviewer → Reviewer accepts → Reviewer reads & analyzes
→ Reviewer submits feedback → Author reviews suggestions → Author incorporates
```

## Author Flow

### 1. Find a reviewer

```bash
# List available reviewers (users who opted in)
pfw reviewer available
```

### 2. Add reviewer to your story

```bash
# Add by user ID (from the available list)
pfw review request <story-id> --reviewer <user-id>
```

### 3. Wait for review

Check review status:

```bash
pfw story reviewer list <story-id>
```

### 4. View feedback

Once the reviewer submits, view their suggestions:

```bash
# List feedback reviews for your story
pfw feedback list <story-id>

# View the diff of suggested changes
pfw feedback diff <story-id> <review-id>
```

### 5. Incorporate changes

```bash
# Accept all suggested changes
pfw feedback incorporate <story-id> <review-id> --all

# Or selectively accept/reject via --selections JSON
pfw feedback incorporate <story-id> <review-id> \
  --selections '{"path/to/file.md": true, "path/to/other.md": false}'
```

## Reviewer Flow (AI Agent)

This is the flow your AI agent follows when reviewing a story.

### 1. Check pending reviews

```bash
pfw review list
```

### 2. Accept a review

```bash
pfw review accept <review-id>
```

This transitions the review to `running` and grants read access to the story.

### 3. Read the story

```bash
# Structured JSON with section IDs (recommended for AI)
pfw story export <story-id> --format json

# Human-readable markdown
pfw story export <story-id> --format markdown

# Story metadata and section list
pfw story get <story-id> -o json
```

### 4. Check quality scores (optional)

```bash
pfw story quality <story-id>
pfw story insights <story-id>
```

### 5. Submit feedback items

For each suggestion, add a feedback item:

```bash
# Text replacement
echo '{
  "sectionId": "<section-id>",
  "type": "replacement",
  "text": "original text to replace",
  "suggested": "improved replacement text",
  "rationale": "why this is better"
}' | pfw feedback item add <story-id> <review-id> --stdin

# General suggestion
echo '{
  "sectionId": "<section-id>",
  "type": "suggestion",
  "text": "Consider restructuring this paragraph for better flow",
  "rationale": "The current structure buries the key information"
}' | pfw feedback item add <story-id> <review-id> --stdin

# Batch mode (one JSON object per line)
cat suggestions.jsonl | pfw feedback item add <story-id> <review-id> --stdin --batch
```

**Feedback item types:**
- `replacement` — specific text replacement (requires `text` + `suggested`)
- `suggestion` — general improvement suggestion
- `strength` — something the author does well (positive feedback)
- `opportunity` — area with potential for improvement
- `context` — contextual note (characters, plot, tone, threads)

### 6. Submit the review

```bash
pfw feedback submit <review-id>
```

This marks the review as `ready_for_review` — the author can now see and incorporate the feedback.

## Review Status Lifecycle

```
pending → running (accept) → ready_for_review (submit) → completed (incorporate)
```

- **pending** — reviewer invited, hasn't accepted
- **running** — reviewer accepted, actively reviewing (has story access)
- **ready_for_review** — reviewer submitted, author reviewing (reviewer still has access)
- **completed** — author incorporated changes (reviewer access revoked)

## Important: Do NOT use `feedback_create`

The `feedback_create` tool / `pfw feedback create` command triggers the **built-in AI Story Coach** — it is for story owners to start an automated AI review. BYOAI reviewers do **not** use this endpoint.

The BYOAI review is created automatically when the author adds you as a reviewer and you accept:
```
reviewer_add → review_accept → (review exists, git branch created) → feedback_item_add / feedback_section_update → feedback_submit
```

If `feedback_create` is called with the author token during a BYOAI review, it will trigger a separate, concurrent AI review that conflicts with the human review.

## Error Handling for AI Reviewers

If any tool call fails, **stop and report the error**. Do not work around failures using bash, curl, python, direct API calls, or any method other than the provided MCP tools. Tool failures indicate an issue that needs to be fixed, not worked around.

## Using with MCP (AI Agents)

The same flow works through MCP tools. See [MCP Setup](mcp-setup.md) for configuration.

```
review_list {}                                    → see pending reviews
review_accept {review_id}                         → accept assignment
story_export {story_id, format: "json"}           → read the story
feedback_section_update {story_id, review_id,     → rewrite sections
  section_id, content}
feedback_item_add {story_id, review_id,           → add suggestions
  type: "replacement", text, suggested, rationale}
feedback_submit {review_id}                       → submit for author
```
