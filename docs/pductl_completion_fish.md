## pductl completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	pductl completion fish | source

To load completions for every new session, execute once:

	pductl completion fish > ~/.config/fish/completions/pductl.fish

You will need to start a new shell for this setup to take effect.


```
pductl completion fish [flags]
```

### Options

```
  -h, --help              help for fish
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

