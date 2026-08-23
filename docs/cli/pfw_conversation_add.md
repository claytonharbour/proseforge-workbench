## pfw conversation add

Add a friend, by principal ID

### Synopsis

Names the person by PRINCIPAL ID — the id from `pfw friends list`, not an
email address. The API does not disclose friends' addresses, so an id is the
only name you can supply.

Creator only, and only people you are already friends with: a refusal here is
usually about the friendship rather than the capability.

```
pfw conversation add <conversation-id> [flags]
```

### Options

```
      --capability string   room:post (read+write) or room:enter (read only) (default "room:post")
  -h, --help                help for add
      --principal string    Principal ID of the person to add (required)
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

