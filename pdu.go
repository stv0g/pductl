// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"errors"
	"time"
)

const (
	All = "all"

	DefaultTTL = 1 * time.Minute
)

var (
	ErrDecode          = errors.New("failed to decode")
	ErrInvalidOutletID = errors.New("invalid outlet ID")
)

type PDU interface {
	Close() error
	SwitchOutlet(id string, state bool) (err error)
	LockOutlet(id string, state bool) (err error)
	RebootOutlet(id string) error
	Status() (*Status, error)
	StatusOutlet(id string) (*OutletStatus, error)
	StatusOutletAll() ([]OutletStatus, error)
	ClearMaximumCurrents() error
	Temperature() (float64, error)
	WhoAmI() (string, error)
}
