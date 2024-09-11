## pdud completion

Generate the autocompletion script for the specified shell

### Synopsis

Generate the autocompletion script for pdud for the specified shell.
See each sub-command's help for details on how to use the generated script.


### Options

```
  -h, --help   help for completion
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

* [pdud](pdud.md)	 - A command line utility, REST API and Prometheus Exporter for Baytech PDUs
* [pdud completion bash](pdud_completion_bash.md)	 - Generate the autocompletion script for bash
* [pdud completion fish](pdud_completion_fish.md)	 - Generate the autocompletion script for fish
* [pdud completion powershell](pdud_completion_powershell.md)	 - Generate the autocompletion script for powershell
* [pdud completion zsh](pdud_completion_zsh.md)	 - Generate the autocompletion script for zsh

