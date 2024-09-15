## pdud completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	pdud completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
pdud completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --address string           Address of TCP socket for PDU communication (default "tcp://10.208.1.1:4141")
      --config string            Path to YAML-formatted configuration file
      --listen string            Address for HTTP listener (default ":8080")
      --password string          password (default "admin")
      --poll-interval duration   Interval between status updates (default 10s)
      --tls-cacert string        Certificate Authority to validate client certificates against
      --tls-cert string          Server certificate
      --tls-insecure             Skip verification of client certificates
      --tls-key string           Server key
      --username string          Username (default "admin")
```

### SEE ALSO

* [pdud completion](pdud_completion.md)	 - Generate the autocompletion script for the specified shell

