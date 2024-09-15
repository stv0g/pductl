## pdud completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(pdud completion bash)

To load completions for every new session, execute once:

#### Linux:

	pdud completion bash > /etc/bash_completion.d/pdud

#### macOS:

	pdud completion bash > $(brew --prefix)/etc/bash_completion.d/pdud

You will need to start a new shell for this setup to take effect.


```
pdud completion bash
```

### Options

```
  -h, --help              help for bash
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

