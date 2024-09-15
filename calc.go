// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

func CalcEnergy(prevSts, newSts *Status) {
	deltaT := float32(newSts.Timestamp.Sub(prevSts.Timestamp).Hours())

	for i := range newSts.Groups {
		prevGroup := prevSts.Groups[i]
		newGroup := newSts.Groups[i]

		prevPower := prevGroup.TrueRMSCurrent * prevGroup.TrueRMSVoltage // W
		newPower := newGroup.TrueRMSCurrent * newGroup.TrueRMSVoltage    // W

		energy := 1e-3 * deltaT * (prevPower + newPower) / 2 // kWh

		newSts.Groups[i].Energy = prevGroup.Energy + energy
	}

	for i := range newSts.Outlets {
		prevOutlet := prevSts.Outlets[i]
		newOutlet := newSts.Outlets[i]

		prevPower := prevOutlet.TrueRMSCurrent * prevOutlet.TrueRMSVoltage // W
		newPower := newOutlet.TrueRMSCurrent * newOutlet.TrueRMSVoltage    // W

		energy := 1e-3 * deltaT * (prevPower + newPower) / 2 // kWh

		newSts.Outlets[i].Energy = prevOutlet.Energy + energy
	}
}
