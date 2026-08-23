## pfw reviewer

Reviewer pool operations

### Options

```
  -h, --help   help for reviewer
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
* [pfw reviewer add](pfw_reviewer_add.md)	 - Add a reviewer to a story (by user ID or email)
* [pfw reviewer list](pfw_reviewer_list.md)	 - List your accepted reviewers
* [pfw reviewer request](pfw_reviewer_request.md)	 - Request someone as a reviewer
* [pfw reviewer respond](pfw_reviewer_respond.md)	 - Accept or decline a reviewer request

