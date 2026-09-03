## pfw contributor

Who may work on your story, and at what level

### Synopsis

Grants control access to a story you own.

Capabilities: story:view, story:edit, story:review, story:merge, room:enter,
room:post. Granting is idempotent.

### Options

```
  -h, --help   help for contributor
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
* [pfw contributor grant](pfw_contributor_grant.md)	 - Grant one capability to an existing user on a story or series
* [pfw contributor leave](pfw_contributor_leave.md)	 - Leave a story by removing your grants
* [pfw contributor list](pfw_contributor_list.md)	 - List the grants on your story or series
* [pfw contributor revoke](pfw_contributor_revoke.md)	 - Revoke one capability
* [pfw contributor shared](pfw_contributor_shared.md)	 - List stories other people have shared with you

