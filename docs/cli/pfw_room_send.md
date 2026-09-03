## pfw room send

Post a message to a room

```
pfw room send <entity-id> [flags]
```

### Options

```
      --agent string           Your identity (title-of-the-moment)
      --content string         Message body (markdown)
      --expect-member string   Refuse to post unless this person is in the room's roster. Membership is capability, NOT readership.
      --expect-title string    Refuse to post unless the room's title matches this regex. A TYPO GUARD against the wrong room id — not delivery assurance.
  -h, --help                   help for send
      --image stringArray      Path to an image to post (jpeg/png/webp, max 10 MiB = 10,485,760 bytes). Repeatable. Place each with a numbered token — {{image:1}} is the first --image, {{image:2}} the second, bare {{image}} means the first; untokened images are appended in order. A token is replaced and does not survive into the message; wrap it in backticks to write about it literally. Any failed upload aborts the post
      --perspective string     Craft lens (artificer, keeper, ...)
      --reply-to room read     Thread this under another message (id from room read)
      --stdin                  Read the message body from stdin
      --target string          Topic this relates to
```

### Options inherited from parent commands

```
      --credentials-file string   Path to a credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read per invocation, so a rotated key applies immediately
      --debug                     Enable debug logging
  -o, --output string             Output format: table, json, brief (default "table")
      --token string              API token (env: PROSEFORGE_TOKEN). Accepts a quoted env reference, e.g. --token '${PROSEFORGE_TOKEN}', which keeps the key out of argv
      --type string               Entity type: story, series, bundle, or conversation (a standalone room with no story or series attached) (default "story")
      --url string                API base URL (env: PROSEFORGE_URL). Accepts a quoted env reference, e.g. --url '${PROSEFORGE_URL}'
```

### SEE ALSO

* [pfw room](pfw_room.md)	 - Coordination room operations (read/send/status/archive)

