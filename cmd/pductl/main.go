// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	pdu "github.com/stv0g/pductl"
	"github.com/stv0g/pductl/baytech"
	"github.com/stv0g/pductl/client"
	"github.com/stv0g/pductl/internal/api"
)

var (
	p pdu.PDU

	cfg *pdu.Config

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pductl",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
	}

	genDocs = &cobra.Command{
		Use:    "docs",
		Short:  "Generate docs",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll("./docs", 0o755); err != nil {
				return err
			}

			return doc.GenMarkdownTree(rootCmd, "./docs")
		},
	}

	getStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show PDU status",
		RunE:  getStatus,
	}

	userCmd = &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	whoAmICmd = &cobra.Command{
		Use:   "whoami",
		Short: "Displays the current user name",
		RunE:  whoAmI,
	}

	readTempCmd = &cobra.Command{
		Use:   "temp",
		Short: "Read current temperature",
		RunE:  readTemp,
	}

	clearMaximumCurrentCmd = &cobra.Command{
		Use:   "clear",
		Short: "Reset the maximum detected current",
		RunE:  clearMaximumCurrent,
	}

	outletCmd = &cobra.Command{
		Use:   "outlet",
		Short: "Control outlets",
	}

	outletRebootCmd = &cobra.Command{
		Use:   "reboot OUTLET",
		Short: "Reboot an outlet",
		RunE:  outletReboot,
		Args:  cobra.ExactArgs(1),
	}

	outletSwitchCmd = &cobra.Command{
		Use:   "switch OUTLET STATE",
		Short: "Switch an outlet on/off",
		RunE:  outletSwitch,
		Args:  cobra.ExactArgs(2),
	}

	outletLockCmd = &cobra.Command{
		Use:   "lock OUTLET STATE",
		Short: "Lock or unlock an outlet",
		RunE:  outletLock,
		Args:  cobra.ExactArgs(2),
	}

	outletStatusCmd = &cobra.Command{
		Use:   "status OUTLET",
		Short: "Get status of outlet",
		RunE:  outletStatus,
		Args:  cobra.ExactArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(getStatusCmd, readTempCmd, clearMaximumCurrentCmd, outletCmd, userCmd, genDocs)
	userCmd.AddCommand(whoAmICmd)
	outletCmd.AddCommand(outletLockCmd, outletRebootCmd, outletSwitchCmd, outletStatusCmd)

	pf := rootCmd.PersistentFlags()
	pf.String("config", "", "Path to YAML-formatted configuration file")
	pf.String("address", "tcp://10.208.1.1:4141", "Address for PDU communication")
	pf.String("username", "admin", "Username")
	pf.String("password", "admin", "password")
	pf.Duration("ttl", -1, "Caching time-to-live. 0 disables caching")
	pf.String("tls-cacert", "", "Certificate Authority to validate client certificates against")
	pf.String("tls-cert", "", "Server certificate")
	pf.String("tls-key", "", "Server key")
	pf.Bool("tls-insecure", false, "Skip verification of server certificate")

	rootCmd.PersistentPreRunE = preRun
	rootCmd.PersistentPostRunE = postRun
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	flags := rootCmd.PersistentFlags()
	if cfg, err = pdu.ParseConfig(flags); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if p, err = newPDU(cfg); err != nil {
		return fmt.Errorf("failed to setup PDU: %w", err)
	}

	return err
}

func postRun(cmd *cobra.Command, args []string) error {
	if err := p.Close(); err != nil {
		return fmt.Errorf("Failed to close PDU: %w", err)
	}

	return nil
}

func newHTTPClient(cfg *pdu.Config) (c *http.Client, err error) {
	if cfg.TLS == nil {
		return &http.Client{}, nil
	}

	var clientCerts []tls.Certificate
	if clientCert, err := tls.LoadX509KeyPair(cfg.TLS.Cert, cfg.TLS.Key); err != nil {
		return nil, fmt.Errorf("Error loading certificate and key file: %v", err)
	} else {
		clientCerts = append(clientCerts, clientCert)
	}

	// Configure the client to trust TLS server certs issued by a CA.
	var certPool *x509.CertPool
	if cfg.TLS.CACert == "" {
		if certPool, err = x509.SystemCertPool(); err != nil {
			return nil, fmt.Errorf("failed to create system certificate pool: %w", err)
		}
	} else {
		certPool = x509.NewCertPool()
		if caCertPEM, err := os.ReadFile(cfg.TLS.CACert); err != nil {
			return nil, fmt.Errorf("failed to read CA cerfificate: %w", err)
		} else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, fmt.Errorf("invalid cert in CA PEM")
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				Certificates:       clientCerts,
				InsecureSkipVerify: cfg.TLS.Insecure,
			},
		},
	}, err
}

func newPDU(cfg *pdu.Config) (p pdu.PDU, err error) {
	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		c, err := newHTTPClient(cfg)
		if err != nil {
			return nil, err
		}

		if p, err = client.NewPDU(cfg.Address, api.WithHTTPClient(c)); err != nil {
			return nil, err
		}

	default:
		if p, err = baytech.NewPDU(cfg.Address, cfg.Username, cfg.Password); err != nil {
			return nil, err
		}
	}

	p = &pdu.Cached{
		PDU: p,
		TTL: cfg.TTL,
	}

	return p, err
}

func parseState(s string) (state bool, err error) {
	switch s {
	case "off", "false", "0":
		state = false

	case "on", "true", "1":
		state = true

	default:
		return false, fmt.Errorf("failed to parse outlet state: %w", err)
	}

	return state, nil
}

func getStatus(_ *cobra.Command, _ []string) error {
	sts, err := p.Status()
	if err != nil {
		return fmt.Errorf("Failed to get status: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "   ")

	return enc.Encode(sts)
}

func whoAmI(_ *cobra.Command, _ []string) error {
	user, err := p.WhoAmI()
	if err != nil {
		return fmt.Errorf("Failed to send command: %w", err)
	}

	fmt.Print(user)

	return nil
}

func readTemp(_ *cobra.Command, _ []string) error {
	temp, err := p.Temperature()
	if err != nil {
		return fmt.Errorf("Failed to send command: %w", err)
	}

	fmt.Print(temp)

	return nil
}

func clearMaximumCurrent(_ *cobra.Command, _ []string) error {
	if err := p.ClearMaximumCurrents(); err != nil {
		return fmt.Errorf("Failed to clear maximum current: %w", err)
	}

	return nil
}

func outletReboot(_ *cobra.Command, args []string) error {
	id := args[0]
	if err := p.RebootOutlet(id); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func outletSwitch(_ *cobra.Command, args []string) error {
	id := args[0]
	state, err := parseState(args[1])
	if err != nil {
		return err
	}

	if err := p.SwitchOutlet(id, state); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func outletLock(_ *cobra.Command, args []string) error {
	id := args[0]
	state, err := parseState(args[1])
	if err != nil {
		return err
	}

	if err := p.LockOutlet(id, state); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func outletStatus(_ *cobra.Command, args []string) error {
	id := args[0]
	sts, err := p.StatusOutlet(id)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "   ")

	return enc.Encode(sts)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
