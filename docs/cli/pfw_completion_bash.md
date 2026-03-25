## pfw completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(pfw completion bash)

To load completions for every new session, execute once:

#### Linux:

	pfw completion bash > /etc/bash_completion.d/pfw

#### macOS:

	pfw completion bash > $(brew --prefix)/etc/bash_completion.d/pfw

You will need to start a new shell for this setup to take effect.


```
pfw completion bash
```

### Options

```
  -h, --help              help for bash
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

