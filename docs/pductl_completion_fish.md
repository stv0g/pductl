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
      --address string      Address for PDU communication (default "tcp://10.208.1.1:4141")
      --config string       Path to YAML-formatted configuration file
      --format string       Output format (default "pretty-rounded")
      --password string     password (default "admin")
      --tls-cacert string   Certificate Authority to validate client certificates against
      --tls-cert string     Server certificate
      --tls-insecure        Skip verification of server certificate
      --tls-key string      Server key
      --username string     Username (default "admin")
```

### SEE ALSO

* [pductl completion](pductl_completion.md)	 - Generate the autocompletion script for the specified shell

