## pfw active-work claim

Take the lease, or find out who holds it

```
pfw active-work claim <path> [flags]
```

### Options

```
  -h, --help            help for claim
      --holder string   Who is claiming (required)
      --ttl duration    How long the claim lasts (default 15m0s)
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

* [pfw active-work](pfw_active-work.md)	 - Standing task that survives a session boundary

