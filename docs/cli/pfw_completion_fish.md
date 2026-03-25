## pfw completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	pfw completion fish | source

To load completions for every new session, execute once:

	pfw completion fish > ~/.config/fish/completions/pfw.fish

You will need to start a new shell for this setup to take effect.


```
pfw completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
  -o, --output string   Output format: table, json, brief (default "table")
      --token string    API token (env: PROSEFORGE_TOKEN)
      --url string      API base URL (env: PROSEFORGE_URL)
```

### SEE ALSO

* [pfw completion](pfw_completion.md)	 - Generate the autocompletion script for the specified shell

