## pfw notifications list

List your notifications, newest first

### Synopsis

List your notifications, newest first.

⚠️ A notification is not a complete record of what happened to you. Only the
events the server chose to raise appear here — a room reaction, for instance,
raises nothing at all, so an empty feed does not mean nobody responded.

```
pfw notifications list [flags]
```

### Options

```
  -h, --help         help for list
      --limit int    Maximum notifications to return (1-100) (default 25)
      --offset int   Pagination offset
      --unread       Show only unread notifications
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

* [pfw notifications](pfw_notifications.md)	 - What the server has told you about — mentions, invitations, review requests

