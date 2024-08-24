## pductl completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(pductl completion bash)

To load completions for every new session, execute once:

#### Linux:

	pductl completion bash > /etc/bash_completion.d/pductl

#### macOS:

	pductl completion bash > $(brew --prefix)/etc/bash_completion.d/pductl

You will need to start a new shell for this setup to take effect.


```
pductl completion bash
```

### Options

```
  -h, --help              help for bash
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

