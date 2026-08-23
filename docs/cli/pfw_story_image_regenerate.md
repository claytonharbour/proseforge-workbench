## pfw story image regenerate

Re-roll an existing image (async, costs 2 credits)

```
pfw story image regenerate <image-id> [flags]
```

### Options

```
  -h, --help            help for regenerate
      --prompt string   Optional new prompt for the re-roll
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

* [pfw story image](pfw_story_image.md)	 - Image generation and management

