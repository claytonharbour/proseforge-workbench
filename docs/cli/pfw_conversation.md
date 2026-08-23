## pfw conversation

Standalone rooms with no story or series attached (proseforge#821)

### Synopsis

Create and manage conversations — rooms you create empty and add specific
people to, rather than inheriting whoever can see a story.

Talk in one with: pfw room send <id> --type conversation --content "..."

### Options

```
  -h, --help   help for conversation
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

* [pfw](pfw.md)	 - ProseForge Workbench — CLI for AI-assisted story review
* [pfw conversation add](pfw_conversation_add.md)	 - Add a friend, by principal ID
* [pfw conversation create](pfw_conversation_create.md)	 - Start a new, empty conversation
* [pfw conversation leave](pfw_conversation_leave.md)	 - Leave a conversation — also how you refuse an invitation
* [pfw conversation list](pfw_conversation_list.md)	 - Conversations you started or were added to
* [pfw conversation remove](pfw_conversation_remove.md)	 - Remove someone else, by principal ID

