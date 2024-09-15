## pductl

A command line utility, REST API and Prometheus Exporter for Baytech PDUs

### Options

```
      --address string      Address for PDU communication (default "tcp://10.208.1.1:4141")
      --config string       Path to YAML-formatted configuration file
      --format string       Output format (default "pretty-rounded")
  -h, --help                help for pductl
      --password string     password (default "admin")
      --tls-cacert string   Certificate Authority to validate client certificates against
      --tls-cert string     Server certificate
      --tls-insecure        Skip verification of server certificate
      --tls-key string      Server key
      --username string     Username (default "admin")
```

### SEE ALSO

* [pductl clear](pductl_clear.md)	 - Reset the maximum detected current
* [pductl completion](pductl_completion.md)	 - Generate the autocompletion script for the specified shell
* [pductl outlet](pductl_outlet.md)	 - Control outlets
* [pductl status](pductl_status.md)	 - Show PDU status
* [pductl temperature](pductl_temperature.md)	 - Read current temperature
* [pductl user](pductl_user.md)	 - Manage users

