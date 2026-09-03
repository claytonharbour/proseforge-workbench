## pfw story list

List stories

```
pfw story list [flags]
```

### Options

```
      --audiobook       Only stories with a completed audiobook
      --cursor string   Continue from a next_cursor value
  -h, --help            help for list
      --limit int       Max results (1-100) (default 25)
      --narration       Only stories with narration
      --q string        Search story titles and text
      --search string   Alias for --q
      --sort string     Sort by date, updated time, or rating (for example date_desc)
      --status string   Filter by status (pitch, draft, queued, generating, published, failed, or all)
      --user string     Filter by user ID
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

* [pfw story](pfw_story.md)	 - Story operations

