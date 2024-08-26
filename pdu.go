// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"errors"
)

const (
	All = 0 // Use id == 0 to control all outlets at once

	// For Baytech MMP-14
	NumOutlets  = 20
	NumBreakers = 3
	NumGroups   = 4
	NumSwitches = 2

	promptReady    = "MMP-14>"
	promptPassword = "Enter Password: "
	promptUsername = "Enter user name: "
)

var (
	ErrDecode          = errors.New("failed to decode")
	ErrInvalidOutletID = errors.New("invalid outlet ID")
)

type PDU interface {
	Close() error
	SwitchOutlet(id int, state bool) (err error)
	LockOutlet(id int, state bool) (err error)
	RebootOutlet(id int) error
	Status() (*Status, error)
	StatusOutlets() ([]OutletStatus, error)
	ClearMaximumCurrents() error
	Temperature() (float64, error)
	Logout() error
	WhoAmI() (string, error)
}

func NewPDU(address string) (p *PDU, err error) {
	return p, nil
}
