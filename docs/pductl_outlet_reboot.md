## pductl outlet reboot

Reboot an outlet

```
pductl outlet reboot OUTLET [flags]
```

### Options

```
  -h, --help   help for reboot
```

### Options inherited from parent commands

```
      --address string      Address for PDU communication (default "tcp://10.208.1.1:4141")
      --config string       Path to YAML-formatted configuration file
      --password string     password (default "admin")
      --tls-cacert string   Certificate Authority to validate client certificates against
      --tls-cert string     Server certificate
      --tls-insecure        Skip verification of server certificate
      --tls-key string      Server key
      --ttl duration        Caching time-to-live. 0 disables caching (default -1ns)
      --username string     Username (default "admin")
```

### SEE ALSO

* [pductl outlet](pductl_outlet.md)	 - Control outlets

