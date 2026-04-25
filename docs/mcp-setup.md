# MCP Server Setup

The ProseForge Workbench MCP server exposes review tools for AI agents over the standard MCP stdio transport.

## Registration

### Claude Code

```bash
claude mcp add proseforge-workbench \
  -e PROSEFORGE_URL=https://app.proseforge.ai \
  -e PROSEFORGE_TOKEN=pf_your_token_here \
  -- /path/to/build/bin/workbench-mcp
```

### Generic MCP Client

The server communicates via JSON-RPC over stdio. Launch with environment variables:

```bash
PROSEFORGE_URL=https://app.proseforge.ai \
PROSEFORGE_TOKEN=pf_your_token \
/path/to/workbench-mcp
```

## Configuration

| Environment Variable | Description |
|---------------------|-------------|
| `PROSEFORGE_URL` | ProseForge API base URL |
| `PROSEFORGE_TOKEN` | Default API token for authentication |

Both can be overridden per-tool-call via optional `url` and `token` parameters on any tool.

Known environment URLs:

- Production: `https://app.proseforge.ai`

Use the base URL only. Do not include `/api/v1`.

## Operational Notes

- Production is sacred. Prefer read-only verification before any write.
- For Codex, persistent MCP registrations live in `~/.codex/config.toml`.
- During active development, the current recommended persistent default is
  `dev`, with `demo` and `prod` accessed via explicit per-call `url` and
  `token` overrides.
- This is a development-time recommendation. For released tooling, the default
  may eventually shift to production once the product is ready for that posture.
- The current Codex config uses `allowed_tools = ["*"]` for the registered MCP
  servers, so the tool allowlist itself does not need to be re-approved across
  sessions.
- Important limitation: in current Codex behavior, `allowed_tools` appears to
  control tool availability, not all approval prompts. If Codex is running with
  a stricter global approval policy such as `approval_policy = "on-request"`,
  mutating MCP tools may still prompt for confirmation even when the server is
  fully allowlisted.
- In practice, this means read-only tools may run silently while write tools
  such as `feedback_item_add` or `feedback_submit` still ask for approval.
- At the time of writing, we have not identified a per-server approval override
  in the local Codex config surface; approvals appear to be governed separately
  from the MCP server's `allowed_tools` list.
- This persisted setup was verified after session reset in
  `forge/proseforge-workbench#63`.
- If your MCP client loses server approval or registration across session
  resets, treat that as an operational setup problem and re-register explicitly.
- After changing MCP registration or auth, verify behavior in a fresh session,
  because an existing session may still be attached to the old server process.

## Available Tools

### Story Operations
| Tool | Description |
|------|-------------|
| `story_list` | List stories (optional: status, limit) |
| `story_get` | Get story details including sections |
| `story_export` | Download story as json, markdown, or pdf |
| `story_quality` | Get quality assessment scores |
| `story_assess` | Trigger quality assessment |
| `story_insights` | Get combined quality and AI analysis |

### Review Operations
| Tool | Description |
|------|-------------|
| `review_list` | List pending review assignments |
| `review_accept` | Accept a review assignment |
| `review_decline` | Decline a review assignment |
| `reviewer_available` | List users available as reviewers |
| `reviewer_add` | Add a reviewer to a story (author) |

### Feedback Operations
| Tool | Description |
|------|-------------|
| `feedback_list` | List feedback reviews for a story |
| `feedback_get` | Get feedback review details |
| `feedback_suggestions` | Get suggestions for a review |
| `feedback_diff` | Get diff of suggested changes |
| `feedback_create` | Create an AI feedback review (author only) |
| `feedback_item_add` | Add a feedback item to a review |
| `feedback_submit` | Submit a review for the author |
| `feedback_incorporate` | Incorporate changes (author) |

## Per-Call Credential Overrides

Every tool accepts optional `url` and `token` parameters. This is useful for:
- Testing both author and reviewer roles from one server
- Switching environments without re-registering

```json
{
  "name": "story_list",
  "arguments": {
    "token": "pf_different_users_token",
    "status": "unpublished"
  }
}
```

If not provided, the tool uses the default credentials from environment variables.

## Example: AI Review Session

```
1. review_list                              → see pending reviews
2. review_accept {review_id: "..."}         → accept assignment
3. story_export {story_id: "...", format: "json"}  → read the story
4. feedback_item_add {story_id, review_id,  → add suggestions
     type: "replacement",
     section_id: "...",
     text: "original",
     suggested: "improved",
     rationale: "why"}
5. feedback_submit {review_id: "..."}       → submit for author
```
