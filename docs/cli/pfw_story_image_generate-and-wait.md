## pfw story image generate-and-wait

Generate an AI image and block until it completes (costs 2 credits)

```
pfw story image generate-and-wait [flags]
```

### Options

```
      --height int           Image height in pixels
  -h, --help                 help for generate-and-wait
      --prompt string        Prompt describing the desired image
      --section-id string    Section ID for section-specific imagery
      --story-id string      Story ID for context-aware generation
      --template-id string   Image template ID
      --timeout duration     Max time to wait for completion (default 1m0s)
      --width int            Image width in pixels
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

