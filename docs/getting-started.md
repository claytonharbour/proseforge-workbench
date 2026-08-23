# Getting Started

ProseForge Workbench (`pfw`) is a CLI and MCP server for AI-assisted story review on ProseForge. It enables BYOAI (Bring Your Own AI) — connect your preferred AI agent to review stories via the ProseForge API.

## Requirements

- Go 1.22+ (to build from source)
- A ProseForge account with Loremaster tier (for API access)
- Your API token (generate at ProseForge → Settings → API Tokens)

## Installation

```bash
git clone https://github.com/claytonharbour/proseforge-workbench.git
cd proseforge-workbench
make build
```

This produces two binaries in `build/bin/`:
- `pfw` — CLI tool
- `workbench-mcp` — MCP server for AI agents

## Configuration

Set your ProseForge API URL and token:

```bash
export PROSEFORGE_URL=https://app.proseforge.ai
export PROSEFORGE_TOKEN=pf_your_token_here
```

Or pass them per-command:

```bash
pfw --url https://app.proseforge.ai --token pf_your_token story list
```

Environment notes:

- Production base URL: `https://app.proseforge.ai`
- Use the base URL only. The client appends `/api/v1`.
- Production is sacred. Prefer read-only verification before any write.

## First Commands

```bash
# List your stories
pfw story list

# Get story details
pfw story get <story-id>

# Export a story as JSON (ideal for AI processing)
pfw story export <story-id> --format json

# Export as markdown (human-readable)
pfw story export <story-id> --format markdown

# View story quality scores
pfw story quality <story-id>
```

## Output Formats

All list/get commands support `--output` (`-o`) for different formats:

```bash
pfw story list -o json      # Structured JSON (for AI/programmatic use)
pfw story list -o table     # Formatted table (default, for humans)
```

## MCP Documentation Resources

The MCP server exposes workflow documentation as resources. Agents can read
these on demand without needing filesystem access:

- `docs://story-workflow` — story lifecycle, planning, rooms, writing
- `docs://series-workflow` — series management, world-building, characters
- `docs://review-flow` — AI-assisted story reviews
- `docs://getting-started` — this document
- `docs://quality-dimensions` — quality score explanations

To access from an MCP client, use `list_mcp_resources` to discover available
docs and `read_mcp_resource` to read one.

## Next Steps

- [Story Workflow](story-workflow.md) — Pitch-to-publish authoring guide
- [Review Flow](review-flow.md) — How to use AI-assisted story reviews
- [MCP Setup](mcp-setup.md) — Configure the MCP server for Claude Code or other AI tools
- [CLI Reference](cli/) — Full command reference
