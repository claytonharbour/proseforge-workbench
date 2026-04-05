## pfw feedback section update

Rewrite a section's content (reads from stdin with --stdin)

```
pfw feedback section update <story-id> <review-id> <section-id> [flags]
```

### Options

```
      --content string   Section content (for short content; prefer --stdin for full sections)
  -h, --help             help for update
      --stdin            Read section content from stdin
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
  -o, --output string   Output format: table, json, brief (default "table")
      --token string    API token (env: PROSEFORGE_TOKEN)
      --url string      API base URL (env: PROSEFORGE_URL)
```

### SEE ALSO

* [pfw feedback section](pfw_feedback_section.md)	 - Section-level feedback operations

