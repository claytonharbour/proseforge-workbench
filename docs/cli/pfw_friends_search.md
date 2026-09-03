## pfw friends search

Find people you could add as a friend

### Synopsis

Search for people to add as a friend.

Match is against name and email. Omitting the term currently returns the whole
directory, so a term is required here.

```
pfw friends search <term> [flags]
```

### Options

```
  -h, --help   help for search
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

* [pfw friends](pfw_friends.md)	 - Your friends directory — the people you can invite to collaborate

