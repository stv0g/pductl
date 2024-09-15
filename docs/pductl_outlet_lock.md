## pductl outlet lock

Lock or unlock an outlet

```
pductl outlet lock OUTLET STATE [flags]
```

### Options

```
  -h, --help   help for lock
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

* [pductl outlet](pductl_outlet.md)	 - Control outlets

