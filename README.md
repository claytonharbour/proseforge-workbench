# ProseForge Workbench

CLI and MCP server for AI-assisted story review on [ProseForge](https://app.proseforge.ai).

**BYOAI** (Bring Your Own AI) — connect any AI agent to review and improve stories through structured feedback, quality scoring, and inline suggestions.

## Install

### From release (recommended)

Download the latest binaries from [Releases](https://github.com/claytonharbour/proseforge-workbench/releases).

### From source

```bash
go install github.com/claytonharbour/proseforge-workbench/cmd/cli@latest
go install github.com/claytonharbour/proseforge-workbench/cmd/mcp@latest
```

### Build locally

```bash
git clone https://github.com/claytonharbour/proseforge-workbench.git
cd proseforge-workbench
make build
```

## Configuration

Set your ProseForge API credentials:

```bash
export PROSEFORGE_URL=https://app.proseforge.ai
export PROSEFORGE_TOKEN=pf_your_token_here
```

Get your API token from your ProseForge account settings.

## CLI Quick Start

```bash
# List your stories
pfw story list

# Export a story as JSON
pfw story export <story-id> --format json

# View quality scores
pfw story quality <story-id>

# List pending reviews
pfw review list

# Accept a review assignment
pfw review accept <review-id>

# Submit feedback
pfw feedback item add <story-id> <review-id> --type replacement --stdin < fix.json
pfw feedback submit <review-id>
```

## MCP Server

The workbench includes an MCP server with 19+ tools for AI agents. Configure it in your AI tool's MCP settings:

```json
{
  "mcpServers": {
    "proseforge-workbench": {
      "command": "/path/to/pfw-mcp",
      "env": {
        "PROSEFORGE_URL": "https://app.proseforge.ai",
        "PROSEFORGE_TOKEN": "pf_your_token_here"
      }
    }
  }
}
```

The MCP server provides tools for story reading, review management, feedback submission, quality assessment, and more — all over stdio JSON-RPC.

## Documentation

- [Getting Started](docs/getting-started.md) — Installation, configuration, first commands
- [Review Flow](docs/review-flow.md) — Complete AI-assisted review workflow
- [Review Strategy](docs/review-strategy.md) — How to approach a story review
- [Quality Dimensions](docs/quality-dimensions.md) — Understanding quality scores
- [MCP Setup](docs/mcp-setup.md) — Detailed MCP server configuration
- [CLI Reference](docs/cli/) — Auto-generated command reference

## Features

- **CLI (`pfw`)** — Story management, review workflow, feedback submission
- **MCP Server (`pfw-mcp`)** — 19+ tools for AI agents over stdio transport
- **BYOAI** — Works with any AI agent (Claude, GPT, etc.)
- **Pipe-friendly** — JSON output, stdin support for large content
- **Quality insights** — Automated quality scores and AI-powered analysis

## How It Works

```
Author publishes story on ProseForge
    → AI reviewer reads story via CLI or MCP
    → AI reviews and submits suggestions
    → Author accepts/rejects feedback
    → Better story
```

## License

See [LICENSE](LICENSE) for details.
