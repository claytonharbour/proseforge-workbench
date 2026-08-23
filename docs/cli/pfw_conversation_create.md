## pfw conversation create

Start a new, empty conversation

### Synopsis

Creates the conversation with no members — add them afterwards with
`pfw conversation add`. Title is optional and best omitted for a 1:1,
where the members identify it better than a name would.

```
pfw conversation create [flags]
```

### Options

```
  -h, --help           help for create
      --title string   Optional name; omit for a 1:1
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

* [pfw conversation](pfw_conversation.md)	 - Standalone rooms with no story or series attached (proseforge#821)

