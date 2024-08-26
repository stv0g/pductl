// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"time"
)

type Cached struct {
	PDU
	TTL time.Duration

	lastUpdate time.Time
	lastStatus *Status
}

func (p *Cached) Status() (*Status, error) {
	now := time.Now()

	if p.lastUpdate.Add(p.TTL).Before(now) {
		sts, err := p.PDU.Status()
		if err != nil {
			return nil, err
		}

		p.lastStatus = sts
		p.lastUpdate = now
	}

	return p.lastStatus, nil
}
