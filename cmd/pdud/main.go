package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	pdu "github.com/stv0g/pductl"
)

var (
	p *pdu.PDU

	// Flags
	address  string
	username string
	password string
	listen   string
	outlet   string
	outletID int

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pdud",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
		PreRunE:           setupMetrics,
		RunE:              daemon,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			p, err = pdu.NewPDU(address, username, password)
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

	pf.StringVar(&address, "address", "10.208.1.1:4141", "Address of TCP socket for PDU communication")
	pf.StringVar(&username, "username", "admin", "Username")
	pf.StringVar(&password, "password", "admin", "password")
	pf.StringVar(&listen, "listen", ":8080", "Address for HTTP listener")
}

var opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
	Name: "myapp_processed_ops_total",
	Help: "The total number of processed events",
})

var (
	lastUpdate time.Time
	lastStatus *pdu.Status
)

func cachedStatus() (*pdu.Status, error) {
	ttl := 1 * time.Minute
	now := time.Now()

	if lastUpdate.Add(ttl).Before(now) {
		fmt.Println("Fetching status")
		s, err := p.Status()
		if err != nil {
			return nil, err
		}

		lastStatus = &s
		lastUpdate = now
		fmt.Println("Fetched status")
	}

	return lastStatus, nil
}

func withStatus(cb func(sts *pdu.Status) float64) func() float64 {
	return func() float64 {
		sts, _ := cachedStatus()
		return cb(sts)
	}
}

func withBoolStatus(cb func(sts *pdu.Status) bool) func() float64 {
	return func() float64 {
		sts, _ := cachedStatus()
		if cb(sts) {
			return 1
		} else {
			return 0
		}
	}
}

func setupMetrics(_ *cobra.Command, _ []string) error {
	sts, err := cachedStatus()
	if err != nil {
		return fmt.Errorf("failed to get PDU status")
	}

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "temperature",
	}, withStatus(func(sts *pdu.Status) float64 {
		return sts.Temperature
	}))

	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "total_kwh",
	}, withStatus(func(sts *pdu.Status) float64 {
		return float64(sts.TotalKWh)
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
			return sts.Breakers[i].TrueRMSCurrent
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Name:        "peak_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Breakers[i].PeakRMSCurrent
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
			return sts.Groups[i].TrueRMSCurrent
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "peak_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Groups[i].PeakRMSCurrent
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "avg_power",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Groups[i].AveragePower
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "va",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Groups[i].VoltAmps
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "group",
			Name:        "true_rms_voltage",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Groups[i].TrueRMSVoltage
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
			return sts.Outlets[i].TrueRMSCurrent
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "peak_rms_current",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Outlets[i].PeakRMSCurrent
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "avg_power",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Outlets[i].AveragePower
		}))

		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "pdu",
			Subsystem:   "outlet",
			Name:        "va",
			ConstLabels: labels,
		}, withStatus(func(sts *pdu.Status) float64 {
			return sts.Outlets[i].VoltAmps
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
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listen, nil)

	return nil
}

func main() {
	rootCmd.Execute()
}
