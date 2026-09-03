## pfw story credits estimate

Estimate the credit cost of an operation

```
pfw story credits estimate [flags]
```

### Options

```
  -h, --help               help for estimate
      --images             Include image generation (generate)
      --operation string   Operation: narrate, generate, rewrite, image, avatar, patch, insights
      --sections int       Number of sections (narrate/generate/rewrite)
      --segments int       Number of segments (patch)
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

* [pfw story credits](pfw_story_credits.md)	 - Credit balance, cost estimates, and transaction history

