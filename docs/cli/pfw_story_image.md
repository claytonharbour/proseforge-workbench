## pfw story image

Image generation and management

### Options

```
  -h, --help   help for image
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
* [pfw story image attach](pfw_story_image_attach.md)	 - Attach an image to a story
* [pfw story image cover](pfw_story_image_cover.md)	 - Set an attached image as the story cover
* [pfw story image generate](pfw_story_image_generate.md)	 - Generate an AI image (async, costs 2 credits)
* [pfw story image generate-and-wait](pfw_story_image_generate-and-wait.md)	 - Generate an AI image and block until it completes (costs 2 credits)
* [pfw story image get](pfw_story_image_get.md)	 - Get image details and generation status
* [pfw story image list](pfw_story_image_list.md)	 - List user's image library
* [pfw story image regenerate](pfw_story_image_regenerate.md)	 - Re-roll an existing image (async, costs 2 credits)
* [pfw story image story-images](pfw_story_image_story-images.md)	 - List images attached to a story
* [pfw story image upload](pfw_story_image_upload.md)	 - Upload a pre-made image (BYOAI path, max 10 MiB)

