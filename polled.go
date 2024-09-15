// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"errors"
	"log/slog"
	"time"
)

var ErrNotPolledYet = errors.New("status has not been polled yet")

type PolledPDU struct {
	PDU

	pollInterval time.Duration
	username     string
	password     string
	lastStatus   *Status
	stop         chan any
	trigger      chan any
	onStatus     func(*Status)
}

func NewPolledPDU(p PDU, interval time.Duration, username, password string, onStatus func(*Status)) *PolledPDU {
	pp := &PolledPDU{
		PDU: p,

		pollInterval: interval,
		username:     username,
		password:     password,
		stop:         make(chan any),
		trigger:      make(chan any, 16),
		onStatus:     onStatus,
	}

	go pp.loop()

	return pp
}

func (p *PolledPDU) Close() error {
	close(p.stop)

	return p.PDU.Close()
}

func (p *PolledPDU) SwitchOutlet(id string, state bool) (err error) {
	if err := p.PDU.SwitchOutlet(id, state); err != nil {
		return err
	}

	p.trigger <- nil

	return nil
}

func (p *PolledPDU) LockOutlet(id string, state bool) (err error) {
	if err := p.PDU.LockOutlet(id, state); err != nil {
		return err
	}

	p.trigger <- nil

	return nil
}

func (p *PolledPDU) RebootOutlet(id string) error {
	if err := p.PDU.RebootOutlet(id); err != nil {
		return err
	}

	p.trigger <- nil

	return nil
}

func (p *PolledPDU) ClearMaximumCurrents() error {
	if err := p.PDU.ClearMaximumCurrents(); err != nil {
		return err
	}

	p.trigger <- nil

	return nil
}

func (p *PolledPDU) Status(detailed bool) (*Status, error) {
	if p.lastStatus == nil {
		return nil, ErrNotPolledYet
	}

	if detailed {
		return p.lastStatus, nil
	} else {
		sts := *p.lastStatus
		sts.Outlets = nil

		return &sts, nil
	}
}

func (p *PolledPDU) Temperature() (float64, error) {
	if p.lastStatus == nil {
		return -1, ErrNotPolledYet
	}

	return float64(p.lastStatus.Temperature), nil
}

func (p *PolledPDU) loop() {
	tmr := time.NewTicker(p.pollInterval)

	if pp, ok := p.PDU.(LoginPDU); ok {
		if err := pp.Login(p.username, p.password); err != nil {
			slog.Error("Failed to login", slog.Any("error", err))
			return
		}

		defer pp.Logout()
	}

	for {
		newSts, err := p.PDU.Status(true)
		if err != nil {
			slog.Error("Failed to get status", slog.Any("error", err))
			continue
		}

		p.lastStatus = newSts

		if p.onStatus != nil {
			p.onStatus(p.lastStatus)
		}

		// Wait for next tick
		select {
		case <-p.stop:
			break
		case <-tmr.C:
		case <-p.trigger:
		}
	}
}
