// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-systemd/v22/activation"
	daemonx "github.com/coreos/go-systemd/v22/daemon"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	pdux "github.com/stv0g/pductl"
	"github.com/stv0g/pductl/baytech"
)

var (
	pdu     pdux.PDU
	cfg     *pdux.Config
	sts     *pdux.Status
	metrics *pdux.Metrics

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pdud",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
		RunE:              daemon,
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
)

func init() {
	pf := rootCmd.PersistentFlags()

	pf.String("config", "", "Path to YAML-formatted configuration file")
	pf.String("address", "tcp://10.208.1.1:4141", "Address of TCP socket for PDU communication")
	pf.Duration("poll-interval", 10*time.Second, "Interval between status updates")
	pf.String("username", "admin", "Username")
	pf.String("password", "admin", "password")
	pf.String("listen", ":8080", "Address for HTTP listener")
	pf.String("tls-cacert", "", "Certificate Authority to validate client certificates against")
	pf.String("tls-cert", "", "Server certificate")
	pf.String("tls-key", "", "Server key")
	pf.Bool("tls-insecure", false, "Skip verification of client certificates")

	rootCmd.PersistentPreRunE = preRun
	rootCmd.PersistentPostRunE = postRun

	rootCmd.AddCommand(genDocs)
}

func preRun(cmd *cobra.Command, args []string) (err error) {
	if cfg, err = pdux.ParseConfig(rootCmd.Flags()); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	if pdu, err = baytech.NewPDU(cfg.Address); err != nil {
		return err
	}

	pdu = pdux.NewPolledPDU(pdu, cfg.PollInterval, cfg.Username, cfg.Password, onStatus)

	return err
}

func onStatus(newSts *pdux.Status) {
	prevSts := sts

	if isFirst := prevSts == nil; isFirst {
		if cfg.Metrics {
			metrics = pdux.NewMetrics(newSts)
		}

		if _, err := daemonx.SdNotify(false, daemonx.SdNotifyReady); err != nil {
			slog.Error("Failed to notify SystemD", slog.Any("error", err))
		}
	} else {
		pdux.CalcEnergy(prevSts, newSts)
	}

	if cfg.Metrics {
		metrics.Update(prevSts, newSts)
	}

	sts = newSts
}

func postRun(cmd *cobra.Command, args []string) error {
	if err := pdu.Close(); err != nil {
		return fmt.Errorf("Failed to close PDF: %w", err)
	}

	return nil
}

func daemon(_ *cobra.Command, _ []string) error {
	r := http.NewServeMux()

	if cfg.Metrics {
		r.Handle("/metrics", promhttp.Handler())
	}

	if len(cfg.ACL) == 0 {
		slog.Warn("No ACL provided. No access control checks will be performed!")
	} else if err := cfg.ACL.Init(); err != nil {
		return fmt.Errorf("failed to initialize ACL: %w", err)
	}

	h := pdux.Handler(r, pdu, cfg)

	var tc *tls.Config
	if cfg.TLS.Cert == "" || cfg.TLS.Key == "" {
		slog.Warn("No TLS configuration provided. API will be exposed unencrypted and unauthenticated!")
	} else {
		cer, err := tls.LoadX509KeyPair(cfg.TLS.Cert, cfg.TLS.Key)
		if err != nil {
			return fmt.Errorf("failed to load server key pair: %w", err)
		}

		tc = &tls.Config{
			Certificates: []tls.Certificate{cer},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
			MinVersion: tls.VersionTLS13,
		}

		if cfg.TLS.CACert != "" {
			caContents, err := os.ReadFile(cfg.TLS.CACert)
			if err != nil {
				return fmt.Errorf("failed to read CA: %w", err)
			}

			tc.ClientCAs = x509.NewCertPool()
			tc.ClientCAs.AppendCertsFromPEM(caContents)
		}
	}

	s := &http.Server{
		Handler:   h,
		Addr:      cfg.Listen,
		TLSConfig: tc,
	}

	slog.Info("Listening", slog.String("address", cfg.Listen))

	return listenAndServe(s)
}

func listenAndServe(s *http.Server) error {
	listeners, err := activation.Listeners()
	if err != nil {
		return fmt.Errorf("failed to get listeners from systemd: %w", err)
	} else if len(listeners) > 1 {
		return fmt.Errorf("got more than one socket fds from systemd")
	}

	if len(listeners) == 1 {
		slog.Debug("Inherited socket from systemd")
	} else {
		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}

		defer ln.Close()

		listeners = append(listeners, ln)
	}

	if s.TLSConfig != nil {
		err = s.ServeTLS(listeners[0], "", "")
	} else {
		err = s.Serve(listeners[0])
	}

	return err
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
