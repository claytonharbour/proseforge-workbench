## pfw admin demo-banner update

Update the runtime demo banner state

```
pfw admin demo-banner update [flags]
```

### Options

```
      --enabled             true to show the banner; false to suppress it temporarily
      --expires-at string   UTC RFC3339 expiry within the next 24 hours
  -h, --help                help for update
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

* [pfw admin demo-banner](pfw_admin_demo-banner.md)	 - Read or update the runtime demo banner

