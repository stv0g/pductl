// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"

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

	detailed = false

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pductl",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
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

	statusCmd = &cobra.Command{
		Use:                "status",
		Short:              "Show PDU status",
		RunE:               status,
		Args:               cobra.MaximumNArgs(1),
		PersistentPreRunE:  preRun,
		PersistentPostRunE: postRun,
	}

	statusBreakerCmd = &cobra.Command{
		Use:     "breakers",
		Aliases: []string{"breaker", "brk"},
		Short:   "Show PDU breaker status",
		RunE:    status,
	}

	statusGroupCmd = &cobra.Command{
		Use:     "groups",
		Aliases: []string{"group", "grp"},
		Short:   "Show PDU group status",
		RunE:    status,
	}

	statusOutletsCmd = &cobra.Command{
		Use:     "outlets",
		Aliases: []string{"outlet"},
		Short:   "Show PDU outlet status",
		RunE:    status,
	}

	userCmd = &cobra.Command{
		Use:                "user",
		Short:              "Manage users",
		PersistentPreRunE:  preRun,
		PersistentPostRunE: postRun,
	}

	whoAmICmd = &cobra.Command{
		Use:                "whoami",
		Short:              "Displays the current user name",
		RunE:               whoAmI,
		PersistentPreRunE:  preRun,
		PersistentPostRunE: postRun,
	}

	tempCmd = &cobra.Command{
		Use:                "temperature",
		Aliases:            []string{"temp"},
		Short:              "Read current temperature",
		RunE:               temp,
		PersistentPreRunE:  preRun,
		PersistentPostRunE: postRun,
	}

	clearCmd = &cobra.Command{
		Use:                "clear",
		Short:              "Reset the maximum detected current",
		RunE:               clearMaximumCurrent,
		PersistentPreRunE:  preRun,
		PersistentPostRunE: postRun,
	}

	outletCmd = &cobra.Command{
		Use:                "outlet",
		Short:              "Control outlets",
		PersistentPreRunE:  preRun,
		PersistentPostRunE: postRun,
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
	rootCmd.AddCommand(statusCmd, tempCmd, clearCmd, outletCmd, userCmd, genDocs)
	userCmd.AddCommand(whoAmICmd)
	statusCmd.AddCommand(statusBreakerCmd, statusGroupCmd, statusOutletsCmd)
	outletCmd.AddCommand(outletLockCmd, outletRebootCmd, outletSwitchCmd, outletStatusCmd)

	pf := rootCmd.PersistentFlags()
	pf.String("config", "", "Path to YAML-formatted configuration file")
	pf.String("address", "tcp://10.208.1.1:4141", "Address for PDU communication")
	pf.String("format", "pretty-rounded", "Output format")
	pf.String("username", "admin", "Username")
	pf.String("password", "admin", "password")
	pf.String("tls-cacert", "", "Certificate Authority to validate client certificates against")
	pf.String("tls-cert", "", "Server certificate")
	pf.String("tls-key", "", "Server key")
	pf.Bool("tls-insecure", false, "Skip verification of server certificate")

	pf = statusCmd.PersistentFlags()
	pf.BoolVar(&detailed, "detailed", false, "Show detailed status")
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
	if cfg.TLS.Cert == "" || cfg.TLS.Key == "" {
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
		q, err := baytech.NewPDU(cfg.Address)
		if err != nil {
			return nil, err
		}

		if err := q.Login(cfg.Username, cfg.Password); err != nil {
			return nil, fmt.Errorf("failed to login to PDU: %w", err)
		}

		p = q
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

func status(cmd *cobra.Command, args []string) error {
	if cmd.Use == "outlets" {
		detailed = true
	}

	sts, err := p.Status(detailed)
	if err != nil {
		return fmt.Errorf("Failed to get status: %w", err)
	}

	switch cmd.Use {
	case "status":
		sts.Print(os.Stdout, cfg.Format)
	case "outlets":
		sts.PrintOutlets(os.Stdout, cfg.Format)
	case "groups":
		sts.PrintGroups(os.Stdout, cfg.Format)
	case "breakers":
		sts.PrintBreakers(os.Stdout, cfg.Format)
	}

	return nil
}

func whoAmI(_ *cobra.Command, _ []string) error {
	user, err := p.WhoAmI()
	if err != nil {
		return fmt.Errorf("Failed to send command: %w", err)
	}

	fmt.Print(user)

	return nil
}

func temp(_ *cobra.Command, _ []string) error {
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
	arg := args[0]
	sts, err := p.Status(true)
	if err != nil {
		return err
	}

	id := -1
	if i, err := strconv.ParseInt(arg, 0, 64); err == nil {
		id = int(i)
	}

	idx := -1
	for i, o := range sts.Outlets {
		if o.Name == arg {
			idx = i
		}

		if id >= 0 && o.ID == id {
			idx = i
		}
	}

	if idx < 0 {
		return pdu.ErrInvalidOutletID
	}

	sts.Outlets[idx].Print(os.Stdout, cfg.Format)

	return nil
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
