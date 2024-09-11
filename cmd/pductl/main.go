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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	pdu "github.com/stv0g/pductl"
	"github.com/stv0g/pductl/baytech"
	"github.com/stv0g/pductl/client"
)

var (
	p pdu.PDU

	// Flags
	address  string
	username string
	password string
	ttl      time.Duration

	tlsCA         string
	tlsKey        string
	tlsCert       string
	tlsSkipVerify bool

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pductl",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			p, err = setupPDU()
			return err
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if err := p.Close(); err != nil {
				return fmt.Errorf("Failed to close PDU: %w", err)
			}

			return nil
		},
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
	rootCmd.AddCommand(getStatusCmd)
	rootCmd.AddCommand(readTempCmd)
	rootCmd.AddCommand(clearMaximumCurrentCmd)
	rootCmd.AddCommand(outletCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(genDocs)

	userCmd.AddCommand(whoAmICmd)

	outletCmd.AddCommand(outletLockCmd)
	outletCmd.AddCommand(outletRebootCmd)
	outletCmd.AddCommand(outletSwitchCmd)
	outletCmd.AddCommand(outletStatusCmd)

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&address, "address", "tcp://10.208.1.1:4141", "Address for PDU communication")
	pf.StringVar(&username, "username", "admin", "Username")
	pf.StringVar(&password, "password", "admin", "password")
	pf.StringVar(&tlsCA, "tls-ca", "", "Certificate Authority to validate client certificates against")
	pf.StringVar(&tlsCert, "tls-cert", "", "Server certificate")
	pf.StringVar(&tlsKey, "tls-client", "", "Server key")
	pf.BoolVar(&tlsSkipVerify, "tls-skip-verify", false, "Skip verification of server certificate")
	pf.DurationVar(&ttl, "ttl", -1, "Caching time-to-live. 0 disables caching")
}

func setupHTTPClient() (c *http.Client, err error) {
	var clientCerts []tls.Certificate
	if tlsCert != "" && tlsKey != "" {
		if clientCert, err := tls.LoadX509KeyPair(tlsCert, tlsKey); err != nil {
			return nil, fmt.Errorf("Error loading certificate and key file: %v", err)
		} else {
			clientCerts = append(clientCerts, clientCert)
		}
	}

	// Configure the client to trust TLS server certs issued by a CA.
	var certPool *x509.CertPool
	if tlsCA == "" {
		if certPool, err = x509.SystemCertPool(); err != nil {
			return nil, fmt.Errorf("failed to create system certificate pool: %w", err)
		}
	} else {
		certPool = x509.NewCertPool()
		if caCertPEM, err := os.ReadFile(tlsCA); err != nil {
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
				InsecureSkipVerify: tlsSkipVerify,
			},
		},
	}, err
}

func setupPDU() (p pdu.PDU, err error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		c, err := setupHTTPClient()
		if err != nil {
			return nil, err
		}

		p, err = client.NewPDU(address, pdu.WithHTTPClient(c))

	default:
		p, err = baytech.NewPDU(address, username, password)

		if ttl < 0 {
			ttl = pdu.DefaultTTL
		}
	}

	p = &pdu.Cached{
		PDU: p,
		TTL: ttl,
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
