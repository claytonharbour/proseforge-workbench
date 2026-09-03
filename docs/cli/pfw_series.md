## pfw series

Series operations

### Synopsis

Series operations.

Granting and revoking access to a series is handled by the contributor command
with the --series flag, not by a subcommand here:

  pfw contributor grant  <series-id> <email> --series --capability room:enter
  pfw contributor list   <series-id>         --series
  pfw contributor revoke <series-id> <email> --series --capability room:enter

Both parties must already be accepted friends, or the grant returns 400.
Creating a series requires a subscription; a bench account gets 403 tier_required.

### Options

```
  -h, --help   help for series
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
* [pfw series archive](pfw_series_archive.md)	 - Archive a series
* [pfw series character](pfw_series_character.md)	 - Series characters
* [pfw series create](pfw_series_create.md)	 - Create a new series
* [pfw series get](pfw_series_get.md)	 - Get a series' details
* [pfw series list](pfw_series_list.md)	 - List your series
* [pfw series plan](pfw_series_plan.md)	 - Seed a Story Forge planning session for the next book in a series
* [pfw series stories](pfw_series_stories.md)	 - List stories in a series
* [pfw series stories-add](pfw_series_stories-add.md)	 - Link an existing story to a series
* [pfw series stories-remove](pfw_series_stories-remove.md)	 - Unlink a story from a series
* [pfw series stories-reorder](pfw_series_stories-reorder.md)	 - Set the order of stories in a series
* [pfw series timeline](pfw_series_timeline.md)	 - Series canon timeline
* [pfw series update](pfw_series_update.md)	 - Update a series' metadata
* [pfw series world](pfw_series_world.md)	 - Series world doc (direction, doctrine, canon)

