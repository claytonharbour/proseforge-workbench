## pfw listing create

Add a store listing to a story

```
pfw listing create <story-id> [flags]
```

### Options

```
      --format string        Format: ebook, audiobook, paperback
  -h, --help                 help for create
      --listing-url string   Store listing URL
      --status string        Status: live, preorder, draft, pulled
      --store string         Store: amazon, google_play, apple_books, audible, kobo, other
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

* [pfw listing](pfw_listing.md)	 - External store listings (Amazon, Apple Books, etc.) for a story

