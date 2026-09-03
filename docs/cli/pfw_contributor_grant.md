## pfw contributor grant

Grant one capability to an existing user on a story or series

```
pfw contributor grant <entity-id> <email> [flags]
```

### Options

```
      --capability string   Capability to grant (default "story:edit")
  -h, --help                help for grant
      --series              Treat the id as a series rather than a story
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

* [pfw contributor](pfw_contributor.md)	 - Who may work on your story, and at what level

