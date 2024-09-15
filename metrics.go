// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type BreakerMetrics struct {
	TrueRMSCurrent prometheus.Gauge
	PeakRMSCurrent prometheus.Gauge
}

type GroupMetrics struct {
	TrueRMSCurrent prometheus.Gauge
	TrueRMSVoltage prometheus.Gauge
	PeakRMSCurrent prometheus.Gauge
	AveragePower   prometheus.Gauge
	Power          prometheus.Gauge
	Energy         prometheus.Counter
}

type OutletMetrics struct {
	TrueRMSCurrent prometheus.Gauge
	TrueRMSVoltage prometheus.Gauge
	PeakRMSCurrent prometheus.Gauge
	AveragePower   prometheus.Gauge
	Power          prometheus.Gauge
	State          prometheus.Gauge
	Locked         prometheus.Gauge
	Energy         prometheus.Counter
}

type Metrics struct {
	Timestamp   prometheus.Gauge
	Temperature prometheus.Gauge
	TotalEnergy prometheus.Counter

	Breakers []BreakerMetrics
	Groups   []GroupMetrics
	Outlets  []OutletMetrics
}

func (m *Metrics) Update(prevSts, newSts *Status) {
	m.Timestamp.Set(float64(newSts.Timestamp.UnixNano()) / 1e9)
	m.Temperature.Set(float64(newSts.Temperature))

	if prevSts != nil {
		m.TotalEnergy.Add(float64(newSts.TotalEnergy - prevSts.TotalEnergy))
	}

	for i := range newSts.Breakers {
		newBreaker := newSts.Breakers[i]

		m.Breakers[i].TrueRMSCurrent.Set(float64(newBreaker.TrueRMSCurrent))
		m.Breakers[i].PeakRMSCurrent.Set(float64(newBreaker.PeakRMSCurrent))
	}

	for i := range newSts.Groups {
		newGroup := newSts.Groups[i]

		m.Groups[i].TrueRMSCurrent.Set(float64(newGroup.TrueRMSCurrent))
		m.Groups[i].TrueRMSVoltage.Set(float64(newGroup.TrueRMSVoltage))
		m.Groups[i].PeakRMSCurrent.Set(float64(newGroup.PeakRMSCurrent))
		m.Groups[i].AveragePower.Set(float64(newGroup.AveragePower))
		m.Groups[i].Power.Set(float64(newGroup.Power))

		if prevSts != nil {
			prevGroup := prevSts.Groups[i]

			m.Groups[i].Energy.Add(float64(newGroup.Energy - prevGroup.Energy))
		}
	}

	for i := range newSts.Outlets {
		newOutlet := newSts.Outlets[i]

		m.Outlets[i].TrueRMSCurrent.Set(float64(newOutlet.TrueRMSCurrent))
		m.Outlets[i].TrueRMSVoltage.Set(float64(newOutlet.TrueRMSVoltage))
		m.Outlets[i].PeakRMSCurrent.Set(float64(newOutlet.PeakRMSCurrent))
		m.Outlets[i].AveragePower.Set(float64(newOutlet.AveragePower))
		m.Outlets[i].Power.Set(float64(newOutlet.Power))

		if prevSts != nil {
			prevOutlet := prevSts.Outlets[i]

			m.Outlets[i].Energy.Add(float64(newOutlet.Energy - prevOutlet.Energy))
		}

		if newOutlet.Locked {
			m.Outlets[i].Locked.Set(1)
		} else {
			m.Outlets[i].Locked.Set(0)
		}

		if newOutlet.State {
			m.Outlets[i].State.Set(1)
		} else {
			m.Outlets[i].State.Set(0)
		}
	}
}

func NewMetrics(sts *Status) *Metrics {
	m := &Metrics{
		Temperature: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "temperature",
		}),
		TotalEnergy: promauto.NewCounter(prometheus.CounterOpts{
			Name: "total_energy",
		}),
		Timestamp: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "timestamp",
		}),
	}

	for _, breaker := range sts.Breakers {
		labels := prometheus.Labels{
			"name": breaker.Name,
		}

		m.Breakers = append(m.Breakers, BreakerMetrics{
			TrueRMSCurrent: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Name:        "true_rms_current",
				ConstLabels: labels,
			}),
			PeakRMSCurrent: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Name:        "peak_rms_current",
				ConstLabels: labels,
			}),
		})
	}

	for _, group := range sts.Groups {
		labels := prometheus.Labels{
			"name":       group.Name,
			"id":         fmt.Sprint(group.ID),
			"breaker_id": fmt.Sprint(group.BreakerID),
		}

		m.Groups = append(m.Groups, GroupMetrics{
			TrueRMSCurrent: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "group",
				Name:        "true_rms_current",
				ConstLabels: labels,
			}),
			TrueRMSVoltage: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "group",
				Name:        "true_rms_voltage",
				ConstLabels: labels,
			}),
			PeakRMSCurrent: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "group",
				Name:        "peak_rms_current",
				ConstLabels: labels,
			}),
			AveragePower: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "group",
				Name:        "avg_power",
				ConstLabels: labels,
			}),
			Power: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "group",
				Name:        "power",
				ConstLabels: labels,
			}),
			Energy: promauto.NewCounter(prometheus.CounterOpts{
				Namespace:   "pdu",
				Subsystem:   "group",
				Name:        "energy",
				ConstLabels: labels,
			}),
		})
	}

	for _, outlet := range sts.Outlets {
		labels := prometheus.Labels{
			"name":       outlet.Name,
			"id":         fmt.Sprint(outlet.ID),
			"group_id":   fmt.Sprint(outlet.GroupID),
			"breaker_id": fmt.Sprint(outlet.BreakerID),
		}

		m.Outlets = append(m.Outlets, OutletMetrics{
			TrueRMSCurrent: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "true_rms_current",
				ConstLabels: labels,
			}),
			TrueRMSVoltage: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "true_rms_voltage",
				ConstLabels: labels,
			}),
			PeakRMSCurrent: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "peak_rms_current",
				ConstLabels: labels,
			}),
			AveragePower: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "avg_power",
				ConstLabels: labels,
			}),
			Power: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "power",
				ConstLabels: labels,
			}),
			Energy: promauto.NewCounter(prometheus.CounterOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "energy",
				ConstLabels: labels,
			}),
			State: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "state",
				ConstLabels: labels,
			}),
			Locked: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace:   "pdu",
				Subsystem:   "outlet",
				Name:        "locked",
				ConstLabels: labels,
			}),
		})
	}

	return m
}
