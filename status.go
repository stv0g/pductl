// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

type BreakerStatus struct {
	Name string `json:"name"`
	ID   int    `json:"id"`

	TrueRMSCurrent float64 `json:"true_rms_current"`
	PeakRMSCurrent float64 `json:"peak_rms_current"`
}

type GroupStatus struct {
	Name string `json:"name"`

	ID        int `json:"id"`
	BreakerID int `json:"breaker_id"`

	TrueRMSCurrent float64 `json:"true_rms_current"`
	PeakRMSCurrent float64 `json:"peak_rms_current"`
	TrueRMSVoltage float64 `json:"true_rms_voltage"`
	AveragePower   float64 `json:"avg_power"`
	VoltAmps       float64 `json:"va"`
}

type OutletStatus struct {
	Name      string `json:"name"`
	ID        int    `json:"id"`
	GroupID   int    `json:"group_id"`
	BreakerID int    `json:"breaker_id"`

	State  bool `json:"state"`
	Locked bool `json:"locked"`

	TrueRMSCurrent float64 `json:"true_rms_current"`
	PeakRMSCurrent float64 `json:"peak_rms_current"`
	TrueRMSVoltage float64 `json:"true_rms_voltage"`
	AveragePower   float64 `json:"avg_power"`
	VoltAmps       float64 `json:"va"`
}

type Status struct {
	Temperature float64         `json:"temp"`
	TotalKWh    int64           `json:"kwh"`
	Switches    []bool          `json:"switches"`
	Breakers    []BreakerStatus `json:"breakers"`
	Groups      []GroupStatus   `json:"groups"`
	Outlets     []OutletStatus  `json:"outlets"`
}
