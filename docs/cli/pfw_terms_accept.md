## pfw terms accept

Accept the current terms of service for this account

### Synopsis

Accept the current ProseForge terms of service for the authenticated account.

The server chooses which version is recorded — you cannot name one. Accepting a
version you picked yourself would satisfy a gate written against text nobody
published, so the parameter does not exist.

Idempotent: re-accepting the same version does not move the recorded date.

Check WHICH account you are accepting for with 'pfw whoami' first — the token in
use may not be the identity you expect.

```
pfw terms accept [flags]
```

### Options

```
  -h, --help   help for accept
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

* [pfw terms](pfw_terms.md)	 - Terms of service acceptance

