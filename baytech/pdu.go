// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package baytech

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	pdu "github.com/stv0g/pductl"
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

	reTemperature   = regexp.MustCompile(`(?m)^Int\. Temp:\s*([0-9\.]+)\s*F`)
	reWhoami        = regexp.MustCompile(`(?m)^Current User:\s*([A-Za-z0-9-]+)\s*$`)
	reStatusKWh     = regexp.MustCompile(`(?m)^Total kW-h: (\d+)`)
	reStatusSwitch  = regexp.MustCompile(`(?m)^Switch 1: (Open|Closed) 2: (Open|Closed)`)
	reStatusBreaker = regexp.MustCompile(`(?m)^\|\s*(CKT[1-2]|Input [A-Z]|Circuit M[1-4])\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*$`)
	reStatusGroup   = regexp.MustCompile(`(?m)^\|\s*(CKT[1-2]|Input [A-Z]|Circuit M[1-4])\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*([0-9\.]+)\s+Volts\s*\|\s*([0-9\.]+)\s+Watts\s*\|\s*([0-9\.]+)\s+VA\s*\|`)
	reOStatusOutlet = regexp.MustCompile(`(?m)^\|\s*([A-Za-z0-9- ]+?)\s*\|\s*([0-9\.]+)\s+A\s*\|\s*([0-9\.]+)\s+A\s*\|\s*([0-9\.]+)\s+V\s*\|\s*([0-9\.]+)\s+W\s*\|\s*([0-9\.]+)\s+VA\s*\|\s*(On|Off)\s*?(Locked|)\s*\|`)
)

type PDU struct {
	Username string
	Password string

	conn    net.Conn
	timeout time.Duration
}

func NewPDU(address string, username, password string) (p *PDU, err error) {
	p = &PDU{
		Username: username,
		Password: password,

		timeout: 300 * time.Millisecond,
	}

	if p.conn, err = net.Dial("tcp", address); err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	return p, nil
}

func (p *PDU) Close() error {
	return p.conn.Close()
}

func (p *PDU) SwitchOutlet(id int, state bool) (err error) {
	if id < 0 || id > NumOutlets {
		return ErrInvalidOutletID
	}

	if state {
		_, err = p.execute("On %d", id)
	} else {
		_, err = p.execute("Off %d", id)
	}
	return err
}

func (p *PDU) LockOutlet(id int, state bool) (err error) {
	if id < 0 || id > NumOutlets {
		return ErrInvalidOutletID
	}

	if state {
		_, err = p.execute("Lock %d", id)
	} else {
		_, err = p.execute("Unlock %d", id)
	}
	return err
}

func (p *PDU) RebootOutlet(id int) error {
	if id < 0 || id > NumOutlets {
		return ErrInvalidOutletID
	}

	_, err := p.execute("Reboot %d", id)
	return err
}

func (p *PDU) Status() (*pdu.Status, error) {
	sts := &pdu.Status{}

	out, err := p.execute("Status")
	if err != nil {
		return sts, err
	}

	// Total KWh
	m := reStatusKWh.FindStringSubmatch(out)
	if m == nil {
		return sts, fmt.Errorf("%w: total Kwh", ErrDecode)
	}

	if sts.TotalKWh, err = strconv.ParseInt(m[1], 10, 64); err != nil {
		return sts, fmt.Errorf("%w: total KWh %w", ErrDecode, err)
	}

	// Temperature
	m = reTemperature.FindStringSubmatch(out)
	if m == nil {
		return sts, fmt.Errorf("%w temp", ErrDecode)
	}

	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return sts, fmt.Errorf("%w temp: %s", ErrDecode, err)
	}

	sts.Temperature = float64(f-32) * 5 / 9

	// Switches
	m = reStatusSwitch.FindStringSubmatch(out)
	if m == nil {
		return sts, fmt.Errorf("%w: switches", ErrDecode)
	}

	for _, sw := range m[1:] {
		sts.Switches = append(sts.Switches, sw == "Closed")
	}

	n := reStatusBreaker.FindAllStringSubmatch(out, -1)
	if n == nil {
		return sts, fmt.Errorf("%w: breakers", ErrDecode)
	}

	for i, b := range n {
		breaker := pdu.BreakerStatus{
			Name: b[1],
			ID:   i,
		}

		if breaker.TrueRMSCurrent, err = strconv.ParseFloat(b[2], 64); err != nil {
			return sts, fmt.Errorf("%w breaker current: %w", ErrDecode, err)
		}

		if breaker.PeakRMSCurrent, err = strconv.ParseFloat(b[3], 64); err != nil {
			return sts, fmt.Errorf("%w breaker current: %w", ErrDecode, err)
		}

		sts.Breakers = append(sts.Breakers, breaker)
	}

	n = reStatusGroup.FindAllStringSubmatch(out, -1)
	if n == nil {
		return sts, fmt.Errorf("%w: groups", ErrDecode)
	}

	for i, g := range n {
		group := pdu.GroupStatus{
			Name: g[1],
			ID:   i + 1,
		}

		switch group.ID {
		case 1, 2:
			group.BreakerID = 1
		case 3, 4:
			group.BreakerID = 2
		}

		if group.TrueRMSCurrent, err = strconv.ParseFloat(g[2], 64); err != nil {
			return sts, fmt.Errorf("%w group current: %w", ErrDecode, err)
		}

		if group.PeakRMSCurrent, err = strconv.ParseFloat(g[3], 64); err != nil {
			return sts, fmt.Errorf("%w group peak current: %w", ErrDecode, err)
		}

		if group.TrueRMSVoltage, err = strconv.ParseFloat(g[4], 64); err != nil {
			return sts, fmt.Errorf("%w group voltage: %w", ErrDecode, err)
		}

		if group.AveragePower, err = strconv.ParseFloat(g[5], 64); err != nil {
			return sts, fmt.Errorf("%w group average power: %w", ErrDecode, err)
		}

		if group.VoltAmps, err = strconv.ParseFloat(g[6], 64); err != nil {
			return sts, fmt.Errorf("%w group VA: %w", ErrDecode, err)
		}

		sts.Groups = append(sts.Groups, group)
	}

	if sts.Outlets, err = p.StatusOutlets(); err != nil {
		return sts, err
	}

	return sts, nil
}

func (p *PDU) StatusOutlets() ([]pdu.OutletStatus, error) {
	outlets := []pdu.OutletStatus{}

	out, err := p.execute("OStatus")
	if err != nil {
		return nil, err
	}

	n := reOStatusOutlet.FindAllStringSubmatch(out, -1)
	if n == nil {
		return nil, fmt.Errorf("%w: groups", ErrDecode)
	}

	for i, o := range n {
		outlet := pdu.OutletStatus{
			Name:   o[1],
			ID:     i + 1,
			State:  o[7] == "On",
			Locked: o[8] == "Locked",
		}

		switch {
		case 1 <= outlet.ID && outlet.ID <= 5:
			outlet.BreakerID = 1
			outlet.GroupID = 1
		case 6 <= outlet.ID && outlet.ID <= 10:
			outlet.BreakerID = 1
			outlet.GroupID = 2
		case 11 <= outlet.ID && outlet.ID <= 15:
			outlet.BreakerID = 2
			outlet.GroupID = 3
		case 16 <= outlet.ID && outlet.ID <= 20:
			outlet.BreakerID = 2
			outlet.GroupID = 4
		}

		if outlet.TrueRMSCurrent, err = strconv.ParseFloat(o[2], 64); err != nil {
			return nil, fmt.Errorf("%w group current: %w", ErrDecode, err)
		}

		if outlet.PeakRMSCurrent, err = strconv.ParseFloat(o[3], 64); err != nil {
			return nil, fmt.Errorf("%w group peak current: %w", ErrDecode, err)
		}

		if outlet.TrueRMSVoltage, err = strconv.ParseFloat(o[4], 64); err != nil {
			return nil, fmt.Errorf("%w group voltage: %w", ErrDecode, err)
		}

		if outlet.AveragePower, err = strconv.ParseFloat(o[5], 64); err != nil {
			return nil, fmt.Errorf("%w group average power: %w", ErrDecode, err)
		}

		if outlet.VoltAmps, err = strconv.ParseFloat(o[6], 64); err != nil {
			return nil, fmt.Errorf("%w group VA: %w", ErrDecode, err)
		}

		outlets = append(outlets, outlet)
	}

	return outlets, nil
}

func (p *PDU) ClearMaximumCurrents() error {
	_, err := p.execute("Clear")
	return err
}

func (p *PDU) Temperature() (float64, error) {
	out, err := p.execute("Temp")
	if err != nil {
		return 0, err
	}

	m := reTemperature.FindStringSubmatch(out)
	if m == nil {
		return -1, fmt.Errorf("failed to find temp")
	}

	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return -1, fmt.Errorf("failed to find temp: %s", err)
	}

	c := float64(f-32) * 5 / 9

	return c, nil
}

func (p *PDU) Logout() error {
	_, err := p.execute("Logout")
	return err
}

func (p *PDU) WhoAmI() (string, error) {
	out, err := p.execute("Whoami")

	m := reWhoami.FindStringSubmatch(out)
	if m == nil {
		return "", fmt.Errorf("failed to find user name")
	}

	return strings.TrimSpace(m[1]), err
}

func (p *PDU) send(cmd string) error {
	_, err := p.conn.Write([]byte(cmd + "\r\n"))
	return err
}

func (p *PDU) execute(cmd string, args ...any) (string, error) {
	fCmd := fmt.Sprintf(cmd, args...)
	fBuf := []byte{}
	sBuf := ""

	commandSend := false

out:
	for {
		rBuf := make([]byte, 1024)

		if err := p.conn.SetReadDeadline(time.Now().Add(p.timeout)); err != nil {
			return "", fmt.Errorf("failed to set read deadline: %w", err)
		}

		n, err := p.conn.Read(rBuf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				if err := p.send(""); err != nil {
					return "", err
				}
				continue
			}

			return "", err
		} else if n == 0 {
			continue
		}

		fBuf = append(fBuf, rBuf[:n]...)
		sBuf = string(fBuf)

		switch {
		case strings.HasSuffix(sBuf, promptReady):
			if commandSend {
				break out
			}

			fBuf = nil
			err = p.send(fCmd)
			commandSend = true

		case strings.HasSuffix(sBuf, promptUsername):
			err = p.send(p.Username)

		case strings.HasSuffix(sBuf, promptPassword):
			err = p.send(p.Password)
		}

		if err != nil {
			return "", err
		}
	}

	res := strings.TrimPrefix(sBuf, fCmd)
	res = strings.TrimSuffix(res, "\r\n"+promptReady)
	res = strings.TrimSpace(res)

	return res, nil
}
