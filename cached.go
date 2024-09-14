// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"time"
)

type Cached struct {
	PDU
	TTL time.Duration

	LastUpdate time.Time
	LastUpdateDetailed time.Time

	LastStatus *Status
}

func (p *Cached) Status(detailed bool) (*Status, error) {
	if !p.IsValid(detailed) {
		sts, err := p.PDU.Status(detailed)
		if err != nil {
			return nil, err
		}

		p.LastStatus = sts
		p.LastUpdate = time.Now()

		if detailed {
			p.LastUpdateDetailed = time.Now()
		}
	}

	sts := *p.LastStatus
	if !detailed {
		sts.Outlets = nil
	}

	return &sts, nil
}

func (p *Cached) IsValid(detailed bool) bool {
	if detailed {
		return p.LastUpdateDetailed.Add(p.TTL).After(time.Now())
	} else {
		return p.LastUpdate.Add(p.TTL).After(time.Now())
	}
}
