## pfw story image upload

Upload a pre-made image (BYOAI path, max 10 MiB)

```
pfw story image upload [flags]
```

### Options

```
      --file string   Path to image file (jpeg/png/webp)
  -h, --help          help for upload
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

