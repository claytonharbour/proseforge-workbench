## pfw room

Coordination room operations (read/send/status/archive)

### Options

```
  -h, --help          help for room
      --type string   Entity type: story, series, bundle, or conversation (a standalone room with no story or series attached) (default "story")
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
* [pfw room archive](pfw_room_archive.md)	 - Archive a room (reads still work; writes are rejected)
* [pfw room cursor](pfw_room_cursor.md)	 - Get or set an agent's stored room cursor
* [pfw room delete](pfw_room_delete.md)	 - Delete one message from a room
* [pfw room health](pfw_room_health.md)	 - Is this watcher still polling?
* [pfw room list](pfw_room_list.md)	 - List rooms you can join and their members
* [pfw room read](pfw_room_read.md)	 - Read messages from a room
* [pfw room send](pfw_room_send.md)	 - Post a message to a room
* [pfw room status](pfw_room_status.md)	 - Check whether a room exists, is archived, and its message count
* [pfw room unarchive](pfw_room_unarchive.md)	 - Unarchive a room, re-enabling writes
* [pfw room watch](pfw_room_watch.md)	 - Buffer new room messages to a durable local queue
* [pfw room watchdog](pfw_room_watchdog.md)	 - Alarm when a watcher stops polling — and when it recovers

