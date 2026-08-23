## pfw

ProseForge Workbench — CLI for AI-assisted story review

### Examples

```
  # Set credentials via environment (recommended)
  export PROSEFORGE_URL=https://app.proseforge.ai
  export PROSEFORGE_TOKEN=pf_your_token
  pfw story list

  # Or pass inline
  pfw --url https://app.proseforge.ai --token pf_your_token story list

  # Review workflow
  pfw story get <story-id>
  pfw feedback list <story-id>
  pfw feedback diff <story-id> <review-id>
```

### Options

```
      --credentials-file string   Path to a credential file (key=value lines with PROSEFORGE_TOKEN, or api_key). Read per invocation, so a rotated key applies immediately
      --debug                     Enable debug logging
  -h, --help                      help for pfw
  -o, --output string             Output format: table, json, brief (default "table")
      --token string              API token (env: PROSEFORGE_TOKEN). Accepts a quoted env reference, e.g. --token '${PROSEFORGE_TOKEN}', which keeps the key out of argv
      --url string                API base URL (env: PROSEFORGE_URL). Accepts a quoted env reference, e.g. --url '${PROSEFORGE_URL}'
```

### SEE ALSO

* [pfw active-work](pfw_active-work.md)	 - Standing task that survives a session boundary
* [pfw author](pfw_author.md)	 - Author-facing operations
* [pfw bundle](pfw_bundle.md)	 - Bundle operations (package stories into EPUB/PDF/markdown/JSON)
* [pfw completion](pfw_completion.md)	 - Generate the autocompletion script for the specified shell
* [pfw contribution](pfw_contribution.md)	 - Work on a story you don't own, and review work on one you do
* [pfw contributor](pfw_contributor.md)	 - Who may work on your story, and at what level
* [pfw conversation](pfw_conversation.md)	 - Standalone rooms with no story or series attached (proseforge#821)
* [pfw feedback](pfw_feedback.md)	 - Feedback review operations
* [pfw friends](pfw_friends.md)	 - Your friends directory — the people you can invite to collaborate
* [pfw genre](pfw_genre.md)	 - Genre operations
* [pfw listing](pfw_listing.md)	 - External store listings (Amazon, Apple Books, etc.) for a story
* [pfw review](pfw_review.md)	 - Review operations
* [pfw reviewer](pfw_reviewer.md)	 - Reviewer pool operations
* [pfw room](pfw_room.md)	 - Coordination room operations (read/send/status/archive)
* [pfw series](pfw_series.md)	 - Series operations
* [pfw server](pfw_server.md)	 - Facts about the backend this CLI is pointed at
* [pfw story](pfw_story.md)	 - Story operations
* [pfw terms](pfw_terms.md)	 - Terms of service acceptance
* [pfw watch](pfw_watch.md)	 - Room watcher: buffer messages and hand them to a worker
* [pfw whoami](pfw_whoami.md)	 - Show which account this token authenticates as

