## pductl completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	pductl completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
pductl completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
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

