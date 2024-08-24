## pductl completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(pductl completion zsh)

To load completions for every new session, execute once:

#### Linux:

	pductl completion zsh > "${fpath[1]}/_pductl"

#### macOS:

	pductl completion zsh > $(brew --prefix)/share/zsh/site-functions/_pductl

You will need to start a new shell for this setup to take effect.


```
pductl completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --address string    Address of TCP socket for PDU communication (default "10.208.1.1:4141")
      --password string   password (default "admin")
      --username string   Username (default "admin")
```

### SEE ALSO

* [pductl completion](pductl_completion.md)	 - Generate the autocompletion script for the specified shell

