## pfw completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	pfw completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
pfw completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
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

