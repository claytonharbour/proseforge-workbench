## pfw server version

Report this CLI's build and the backend's build

### Synopsis

Report the version of this CLI binary and of the ProseForge backend it is
configured to reach, along with the resolved backend URL.

Answers "am I measuring what I think I am measuring" before you trust a result:
the URL comes from the resolved config (flag, env, or --credentials-file), not
from what you believe it to be, and the backend build comes from the server
rather than from any local assumption.

The backend version endpoint is public — this needs no admin credentials.

```
pfw server version [flags]
```

### Options

```
  -h, --help   help for version
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

* [pfw server](pfw_server.md)	 - Facts about the backend this CLI is pointed at

