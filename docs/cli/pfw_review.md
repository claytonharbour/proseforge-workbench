## pfw review

Review operations

### Options

```
  -h, --help   help for review
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
* [pfw review accept](pfw_review_accept.md)	 - Accept a review assignment
* [pfw review active](pfw_review_active.md)	 - Show the active (running) review for a story
* [pfw review approve](pfw_review_approve.md)	 - Approve a story after review
* [pfw review decline](pfw_review_decline.md)	 - Decline a review assignment
* [pfw review list](pfw_review_list.md)	 - List pending reviews assigned to you
* [pfw review reject](pfw_review_reject.md)	 - Reject a story after review
* [pfw review request](pfw_review_request.md)	 - Add a reviewer to a story

