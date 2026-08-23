## pfw active-work

Standing task that survives a session boundary

### Synopsis

Read and write the standing-work record a watcher hands to a worker.

A turn-based runtime has no background task to resume: a session handed only a
room batch will do the room batch and whatever it was halfway through simply
stops. This record is what makes a fresh session as good as a resumed one.

Keep it small. It should POINT AT evidence — a ticket, a commit, a path — never
contain it.

### Options

```
  -h, --help   help for active-work
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
* [pfw active-work claim](pfw_active-work_claim.md)	 - Take the lease, or find out who holds it
* [pfw active-work release](pfw_active-work_release.md)	 - Drop a lease you hold
* [pfw active-work set](pfw_active-work_set.md)	 - Create or update the record
* [pfw active-work show](pfw_active-work_show.md)	 - Print the current record

