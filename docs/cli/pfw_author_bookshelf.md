## pfw author bookshelf

Get an author's complete bookshelf (titles, taglines, covers, series)

### Synopsis

Get an author's complete bookshelf in a single call — stories with titles, taglines, slugs, cover URLs, section images, and series membership. Replaces the N+1 pattern of story list + per-story images.

Works unauthenticated for published stories; authenticated calls also include drafts and pitches.

```
pfw author bookshelf <handle> [flags]
```

### Options

```
  -h, --help            help for bookshelf
      --limit int       Max results (default 50, max 100)
      --offset int      Pagination offset
      --q string        Search title, tagline, or series name
      --series string   Filter by series slug
      --sort string     Sort order: series (default), title, newest, oldest
      --status string   Filter by status: draft, published, pitch
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

* [pfw author](pfw_author.md)	 - Author-facing operations

