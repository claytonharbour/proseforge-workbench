## pfw contribution merge

Merge a contribution into your story

### Synopsis

Merge a contribution into your story.

Resolve outstanding suggestions first — only accepted text reaches trunk.

PARTIAL ACCEPT: --selections takes a JSON object of file path -> true/false, so
you can take two of three files instead of all or nothing.

  pfw contribution merge <story> <contrib> \
      --selections '{"content/<sectionId>.md": true, "content/<other>.md": false}'

🛑 IT IS FAIL-CLOSED. If you supply --selections it must name EVERY changed path.
An absent path is a 400 invalid_selections, NOT an implicit accept — the opposite
of the feedback merge. List the contribution's changed paths first (contribution
diff) and name all of them.

Keys are FILE PATHS (content/<sectionId>.md), not bare section ids.

Omit --selections entirely to merge everything, which is the default and
unchanged.

```
pfw contribution merge <story-id> <contribution-id> [flags]
```

### Options

```
  -h, --help                help for merge
      --selections string   JSON object mapping each changed file path to true (accept) or false (reject). Must name EVERY changed path.
      --stdin               Read the selections JSON from stdin
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

* [pfw contribution](pfw_contribution.md)	 - Work on a story you don't own, and review work on one you do

