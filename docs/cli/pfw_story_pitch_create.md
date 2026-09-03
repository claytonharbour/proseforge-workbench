## pfw story pitch create

Create a pitch (a pre-writing story idea)

### Synopsis

Create a pitch — a pre-writing story idea with planning data but no sections. Next: story meta upsert (premise/characters/plot), then story promote.

```
pfw story pitch create [flags]
```

### Options

```
      --genre string     Genre name (e.g., "Historical Fiction")
  -h, --help             help for create
      --tagline string   Story tagline (optional)
      --title string     Story title (optional)
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

* [pfw story pitch](pfw_story_pitch.md)	 - Pre-writing story pitches (idea + planning data, no sections yet)

