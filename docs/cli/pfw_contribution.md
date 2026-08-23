## pfw contribution

Work on a story you don't own, and review work on one you do

### Synopsis

Contributions are branch-based edits to someone else's story.

A grant (see 'pfw contributor') lets you edit; your writes land on your own
branch, never the owner's trunk. Mark it ready when it is worth looking at, and
the owner reviews, suggests, and merges.

### Options

```
  -h, --help   help for contribution
```

### Options inherited from parent commands

```
      --credentials-file string   Path to a credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read per invocation, so a rotated key applies immediately
      --debug                     Enable debug logging
  -o, --output string             Output format: table, json, brief (default "table")
      --token string              API token (env: PROSEFORGE_TOKEN). Accepts a quoted env reference, e.g. --token '${PROSEFORGE_TOKEN}', which keeps the key out of argv
      --url string                API base URL (env: PROSEFORGE_URL). Accepts a quoted env reference, e.g. --url '${PROSEFORGE_URL}'
```

### SEE ALSO

* [pfw](pfw.md)	 - ProseForge Workbench — CLI for AI-assisted story review
* [pfw contribution accounting](pfw_contribution_accounting.md)	 - Show contribution totals for a story, or a whole series
* [pfw contribution diff](pfw_contribution_diff.md)	 - Show a contribution's changes against trunk
* [pfw contribution discard](pfw_contribution_discard.md)	 - Throw away an unmerged contribution branch
* [pfw contribution list](pfw_contribution_list.md)	 - List contributions on a story you own
* [pfw contribution merge](pfw_contribution_merge.md)	 - Merge a contribution into your story
* [pfw contribution mine](pfw_contribution_mine.md)	 - Show your own contribution to a story
* [pfw contribution ready](pfw_contribution_ready.md)	 - Submit a contribution for review
* [pfw contribution ready-for-owner](pfw_contribution_ready-for-owner.md)	 - Tell the owner a delegated review is finished
* [pfw contribution request-changes](pfw_contribution_request-changes.md)	 - Ask the contributor to revise
* [pfw contribution resolve](pfw_contribution_resolve.md)	 - Accept or reject one suggestion
* [pfw contribution suggest](pfw_contribution_suggest.md)	 - Suggest a section rewrite on a contribution
* [pfw contribution suggestions](pfw_contribution_suggestions.md)	 - List suggestions raised against a contribution
* [pfw contribution sync](pfw_contribution_sync.md)	 - Merge trunk into a contribution branch

