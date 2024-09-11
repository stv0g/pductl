// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	pdu "github.com/stv0g/pductl"
	"github.com/stv0g/pductl/baytech"
)

var (
	p pdu.PDU

	// Flags
	address  string
	username string
	password string
	listen   string
	ttl      time.Duration

	tlsCA   string
	tlsKey  string
	tlsCert string

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pdud",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
		PreRunE:           setupMetrics,
		RunE:              daemon,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			q, err := baytech.NewPDU(address, username, password)
			if err != nil {
				return err
			}

			p = &pdu.Cached{
				PDU: q,
				TTL: ttl,
			}

			return err
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if err := p.Close(); err != nil {
				return fmt.Errorf("Failed to close PDF: %w", err)
			}

			return nil
		},
	}
)

func init() {
	pf := rootCmd.PersistentFlags()

	pf.StringVar(&address, "address", "tcp://10.208.1.1:4141", "Address of TCP socket for PDU communication")
	pf.StringVar(&username, "username", "admin", "Username")
	pf.StringVar(&password, "password", "admin", "password")
	pf.StringVar(&listen, "listen", ":8080", "Address for HTTP listener")
	pf.StringVar(&tlsCA, "tls-ca", "", "Certificate Authority to validate client certificates against")
	pf.StringVar(&tlsCert, "tls-cert", "", "Server certificate")
	pf.StringVar(&tlsKey, "tls-client", "", "Server key")
	pf.DurationVar(&ttl, "ttl", pdu.DefaultTTL, "Caching time-to-live. 0 disables caching")
}

func withStatus(cb func(sts *pdu.Status) float64) func() float64 {
	return func() float64 {
		sts, _ := p.Status()
		return cb(sts)
	}
}

func withBoolStatus(cb func(sts *pdu.Status) bool) func() float64 {
	return func() float64 {
		sts, _ := p.Status()
		if cb(sts) {
			return 1
		} else {
			return 0
		}
	}
}

func setupMetrics(_ *cobra.Command, _ []string) error {
	sts, err := p.Status()
	if err != nil {
		return fmt.Errorf("failed to get PDU status")
	}

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "temperature",
	}, withStatus(func(sts *pdu.Status) float64 {
		return float64(sts.Temp)
	}))

	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "total_kwh",
	}, withStatus(func(sts *pdu.Status) float64 {
		return float64(sts.TotalKwh)
	}))

	for i, breaker := range sts.Breakers {
		labels := prometheus.Labels{
			"name": breaker.Name,
		}

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Name:        "true_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Breakers[i].TrueRMSCurrent)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Name:        "peak_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Breakers[i].PeakRMSCurrent)
		}))
	}

	for i, group := range sts.Groups {
		labels := prometheus.Labels{
			"name":       group.Name,
			"id":         fmt.Sprint(group.ID),
			"breaker_id": fmt.Sprint(group.BreakerID),
		}

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "true_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Groups[i].TrueRMSCurrent)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "peak_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Groups[i].PeakRMSCurrent)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "avg_power",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Groups[i].AveragePower)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "va",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Groups[i].VoltAmps)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "true_rms_voltage",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Groups[i].TrueRMSCurrent)
		}))
	}

	for i, outlet := range sts.Outlets {
		labels := prometheus.Labels{
			"name":       outlet.Name,
			"id":         fmt.Sprint(outlet.ID),
			"group_id":   fmt.Sprint(outlet.GroupID),
			"breaker_id": fmt.Sprint(outlet.BreakerID),
		}

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "true_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Outlets[i].TrueRMSCurrent)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "peak_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Outlets[i].PeakRMSCurrent)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "avg_power",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Outlets[i].AveragePower)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "va",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return float64(sts.Outlets[i].VoltAmps)
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "state",
			ConstLabels: labels,
		}, withBoolStatus(func(sts *pdu.Status) bool {
			return sts.Outlets[i].State
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "locked",
			ConstLabels: labels,
		}, withBoolStatus(func(sts *pdu.Status) bool {
			return sts.Outlets[i].Locked
		}))
	}

	return nil
}

func daemon(_ *cobra.Command, _ []string) error {
	r := http.NewServeMux()
	si := pdu.NewStrictHandler(&pdu.Server{
		PDU: p,
	}, nil)
	h := pdu.HandlerWithOptions(si, pdu.StdHTTPServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: r,
	})

	r.Handle("/metrics", promhttp.Handler())

	var tc *tls.Config
	if tlsKey != "" && tlsCert != "" {
		cer, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
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

		if tlsCA != "" {
			caContents, err := os.ReadFile(tlsCA)
			if err != nil {
				return fmt.Errorf("failed to read CA: %w", err)
			}

			tc.ClientCAs = x509.NewCertPool()
			tc.ClientCAs.AppendCertsFromPEM(caContents)
		}
	}

	s := &http.Server{
		Handler:   h,
		Addr:      listen,
		TLSConfig: tc,
	}

	return s.ListenAndServe()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
