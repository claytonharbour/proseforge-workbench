## pfw completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(pfw completion zsh)

To load completions for every new session, execute once:

#### Linux:

	pfw completion zsh > "${fpath[1]}/_pfw"

#### macOS:

	pfw completion zsh > $(brew --prefix)/share/zsh/site-functions/_pfw

You will need to start a new shell for this setup to take effect.


```
pfw completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
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

