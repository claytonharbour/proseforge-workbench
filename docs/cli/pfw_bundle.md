## pfw bundle

Bundle operations (package stories into EPUB/PDF/markdown/JSON)

### Options

```
  -h, --help   help for bundle
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
* [pfw bundle create](pfw_bundle_create.md)	 - Create a new bundle
* [pfw bundle delete](pfw_bundle_delete.md)	 - Delete a bundle (the stories are not affected)
* [pfw bundle entry](pfw_bundle_entry.md)	 - Manage bundle entries
* [pfw bundle export](pfw_bundle_export.md)	 - Export a bundle as epub, json, markdown, or pdf
* [pfw bundle get](pfw_bundle_get.md)	 - Get a bundle with its entries
* [pfw bundle list](pfw_bundle_list.md)	 - List your bundles
* [pfw bundle update](pfw_bundle_update.md)	 - Update a bundle's name or intro

