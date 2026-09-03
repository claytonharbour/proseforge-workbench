## pfw reviewer add

Add a reviewer to a story (by user ID or email)

```
pfw reviewer add <story-id> [flags]
```

### Options

```
      --email string         Reviewer's email (alternative to --reviewer-id)
  -h, --help                 help for add
      --reviewer-id string   Reviewer's user ID
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

* [pfw reviewer](pfw_reviewer.md)	 - Reviewer pool operations

