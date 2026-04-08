## pfw feedback item add

Add a feedback item (reads JSON from stdin with --stdin)

```
pfw feedback item add <story-id> <review-id> [flags]
```

### Options

```
      --batch              Read multiple items, one JSON object per line
  -h, --help               help for add
      --rationale string   Why this improves the writing
      --section string     Section ID
      --stdin              Read feedback item JSON from stdin
      --suggested string   Suggested replacement text
      --text string        Original text (replacement) or feedback text
      --type string        Item type: replacement, strength, opportunity, suggestion, context
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
  -o, --output string   Output format: table, json, brief (default "table")
      --token string    API token (env: PROSEFORGE_TOKEN)
      --url string      API base URL (env: PROSEFORGE_URL)
```

### SEE ALSO

* [pfw feedback item](pfw_feedback_item.md)	 - Feedback item operations

