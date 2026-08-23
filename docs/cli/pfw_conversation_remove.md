## pfw conversation remove

Remove someone else, by principal ID

### Synopsis

Names the person by PRINCIPAL ID — from `pfw conversation list` members or
`pfw friends list`, not an email address.

Creator only. Removal revokes access; it does not delete anything they wrote.

```
pfw conversation remove <conversation-id> [flags]
```

### Options

```
  -h, --help               help for remove
      --principal string   Principal ID of the person to remove (required)
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

