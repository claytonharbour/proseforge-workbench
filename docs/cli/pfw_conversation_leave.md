## pfw conversation leave

Leave a conversation — also how you refuse an invitation

### Synopsis

Leaving and refusing an invitation are the same act at different moments,
so there is one command for both.

The creator cannot leave: they hold the conversation through ownership
rather than a grant, so leaving would revoke nothing. Archive the room instead.

```
pfw conversation leave <conversation-id> [flags]
```

### Options

```
  -h, --help   help for leave
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

