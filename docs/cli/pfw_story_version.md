## pfw story version

Version history operations

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

* [pfw story](pfw_story.md)	 - Story operations
* [pfw story version diff](pfw_story_version_diff.md)	 - Show diff between two story versions
* [pfw story version get](pfw_story_version_get.md)	 - Get story content at a specific version
* [pfw story version list](pfw_story_version_list.md)	 - List version history for a story
* [pfw story version restore](pfw_story_version_restore.md)	 - Restore a story to a previous version

