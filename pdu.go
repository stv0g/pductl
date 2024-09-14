// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"time"

	"github.com/stv0g/pductl/internal/api"
)

const (
	All = "all"

	DefaultTTL = 1 * time.Minute
)

type (
	Status        = api.Status
	BreakerStatus = api.BreakerStatus
	OutletStatus  = api.OutletStatus
	GroupStatus   = api.GroupStatus
)

type PDU interface {
	Close() error
	SwitchOutlet(id string, state bool) (err error)
	LockOutlet(id string, state bool) (err error)
	RebootOutlet(id string) error
	Status(detailed bool) (*Status, error)
	ClearMaximumCurrents() error
	Temperature() (float64, error)
	WhoAmI() (string, error)
}

type LoginPDU interface {
	Login(username, password string) error
	Logout() error
	WithLogin(username, password string, cb func()) error
}
