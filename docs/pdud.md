## pdud

A command line utility, REST API and Prometheus Exporter for Baytech PDUs

```
pdud [flags]
```

### Options

```
      --address string           Address of TCP socket for PDU communication (default "tcp://10.208.1.1:4141")
      --config string            Path to YAML-formatted configuration file
  -h, --help                     help for pdud
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

