## pdud completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	pdud completion fish | source

To load completions for every new session, execute once:

	pdud completion fish > ~/.config/fish/completions/pdud.fish

You will need to start a new shell for this setup to take effect.


```
pdud completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --address string      Address of TCP socket for PDU communication (default "tcp://10.208.1.1:4141")
      --config string       Path to YAML-formatted configuration file
      --listen string       Address for HTTP listener (default ":8080")
      --password string     password (default "admin")
      --tls-cacert string   Certificate Authority to validate client certificates against
      --tls-cert string     Server certificate
      --tls-insecure        Skip verification of client certificates
      --tls-key string      Server key
      --ttl duration        Caching time-to-live. 0 disables caching (default 1m0s)
      --username string     Username (default "admin")
```

### SEE ALSO

* [pdud completion](pdud_completion.md)	 - Generate the autocompletion script for the specified shell

