// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package baytech

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"sync"

	"go.bug.st/serial"

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
	promptInvalidPassword = "Invalid user/password!"
)

var (
	ErrDecode          = errors.New("failed to decode")
	ErrLoginRequired   = errors.New("login required")
	ErrInvalidPassword = errors.New("invalid password")

	reTemperature   = regexp.MustCompile(`(?m)^Int\. Temp:\s*([0-9\.]+)\s*F`)
	reWhoami        = regexp.MustCompile(`(?m)^Current User:\s*([A-Za-z0-9-]+)\s*$`)
	reStatusKWh     = regexp.MustCompile(`(?m)^Total kW-h: (\d+)`)
	reStatusSwitch  = regexp.MustCompile(`(?m)^Switch 1: (Open|Closed) 2: (Open|Closed)`)
	reStatusBreaker = regexp.MustCompile(`(?m)^\|\s*(CKT[1-2]|Input [A-Z]|Circuit M[1-4])\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*$`)
	reStatusGroup   = regexp.MustCompile(`(?m)^\|\s*(CKT[1-2]|Input [A-Z]|Circuit M[1-4])\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*([0-9\.]+)\s+Amps\s*\|\s*([0-9\.]+)\s+Volts\s*\|\s*([0-9\.]+)\s+Watts\s*\|\s*([0-9\.]+)\s+VA\s*\|`)
	reOstatusOutlet = regexp.MustCompile(`(?m)^\|\s*([A-Za-z0-9- ]+?)\s*\|\s*([0-9\.]+)\s+A\s*\|\s*([0-9\.]+)\s+A\s*\|\s*([0-9\.]+)\s+V\s*\|\s*([0-9\.]+)\s+W\s*\|\s*([0-9\.]+)\s+VA\s*\|\s*(On|Off)\s*?(Locked|)\s*\|`)
)

type OutletID string

type PDU struct {
	conn    io.ReadWriteCloser
	timeout time.Duration
	mu sync.Mutex
	muLogin sync.Mutex
}

func NewPDU(uri string) (p *PDU, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	p = &PDU{
		timeout: 300 * time.Millisecond,
	}

	switch u.Scheme {
	case "tcp":
		if p.conn, err = net.Dial("tcp", u.Host); err != nil {
			return nil, fmt.Errorf("failed to establish TCP connection: %w", err)
		}

	case "serial", "":
		if p.conn, err = serial.Open(u.Path, &serial.Mode{
			BaudRate: 9600,
			DataBits: 8,
			StopBits: serial.OneStopBit,
			Parity:   serial.NoParity,
		}); err != nil {
			return nil, fmt.Errorf("failed to open serial port: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported PDU address: %s", uri)
	}

	return p, nil
}

func (p *PDU) Close() error {
	if err := p.Logout(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close: %w", err)
	}

	return nil
}

func (p *PDU) SwitchOutlet(idStr string, state bool) (err error) {
	id, err := p.lookupID(idStr)
	if err != nil {
		return fmt.Errorf("invalid outlet ID: %s", err)
	}

	if state {
		_, err = p.execute("On %d", id)
	} else {
		_, err = p.execute("Off %d", id)
	}
	return err
}

func (p *PDU) LockOutlet(idStr string, state bool) (err error) {
	id, err := p.lookupID(idStr)
	if err != nil {
		return fmt.Errorf("%w: %s", pdu.ErrInvalidOutletID, err)
	}

	if state {
		_, err = p.execute("Lock %d", id)
	} else {
		_, err = p.execute("Unlock %d", id)
	}
	return err
}

func (p *PDU) RebootOutlet(idStr string) error {
	id, err := p.lookupID(idStr)
	if err != nil {
		return fmt.Errorf("%w: %s", pdu.ErrInvalidOutletID, err)
	}

	_, err = p.execute("Reboot %d", id)
	return err
}

func (p *PDU) Status(detailed bool) (*pdu.Status, error) {
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

	kwh, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return sts, fmt.Errorf("%w: total KWh %w", ErrDecode, err)
	}

	sts.TotalKwh = float32(kwh)

	// Temperature
	m = reTemperature.FindStringSubmatch(out)
	if m == nil {
		return sts, fmt.Errorf("%w temp", ErrDecode)
	}

	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return sts, fmt.Errorf("%w temp: %s", ErrDecode, err)
	}

	sts.Temperature = float32(f-32) * 5 / 9

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

		cur, err := strconv.ParseFloat(b[2], 32)
		if err != nil {
			return sts, fmt.Errorf("%w breaker current: %w", ErrDecode, err)
		}

		curPeak, err := strconv.ParseFloat(b[3], 32)
		if err != nil {
			return sts, fmt.Errorf("%w breaker current: %w", ErrDecode, err)
		}

		breaker.TrueRMSCurrent = float32(cur)
		breaker.PeakRMSCurrent = float32(curPeak)

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

		cur, err := strconv.ParseFloat(g[2], 64)
		if err != nil {
			return sts, fmt.Errorf("%w group current: %w", ErrDecode, err)
		}

		curPeak, err := strconv.ParseFloat(g[3], 64)
		if err != nil {
			return sts, fmt.Errorf("%w group peak current: %w", ErrDecode, err)
		}

		volt, err := strconv.ParseFloat(g[4], 64)
		if err != nil {
			return sts, fmt.Errorf("%w group voltage: %w", ErrDecode, err)
		}

		avgPower, err := strconv.ParseFloat(g[5], 64)
		if err != nil {
			return sts, fmt.Errorf("%w group average power: %w", ErrDecode, err)
		}

		va, err := strconv.ParseFloat(g[6], 64)
		if err != nil {
			return sts, fmt.Errorf("%w group VA: %w", ErrDecode, err)
		}

		group.TrueRMSCurrent = float32(cur)
		group.PeakRMSCurrent = float32(curPeak)
		group.TrueRMSVoltage = float32(volt)
		group.AveragePower = float32(avgPower)
		group.VoltAmps = float32(va)

		sts.Groups = append(sts.Groups, group)
	}

	if detailed {
		if sts.Outlets, err = p.statusOutlets(); err != nil {
			return nil, err
		}
	}

	return sts, nil
}

func (p *PDU) statusOutlets() ([]pdu.OutletStatus, error) {
	outlets := []pdu.OutletStatus{}

	out, err := p.execute("Ostatus")
	if err != nil {
		return nil, err
	}

	n := reOstatusOutlet.FindAllStringSubmatch(out, -1)
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

		cur, err := strconv.ParseFloat(o[2], 64)
		if err != nil {
			return nil, fmt.Errorf("%w group current: %w", ErrDecode, err)
		}

		curPeak, err := strconv.ParseFloat(o[3], 64)
		if err != nil {
			return nil, fmt.Errorf("%w group peak current: %w", ErrDecode, err)
		}

		volt, err := strconv.ParseFloat(o[4], 64)
		if err != nil {
			return nil, fmt.Errorf("%w group voltage: %w", ErrDecode, err)
		}

		avgPower, err := strconv.ParseFloat(o[5], 64)
		if err != nil {
			return nil, fmt.Errorf("%w group average power: %w", ErrDecode, err)
		}

		va, err := strconv.ParseFloat(o[6], 64)
		if err != nil {
			return nil, fmt.Errorf("%w group VA: %w", ErrDecode, err)
		}

		outlet.TrueRMSCurrent = float32(cur)
		outlet.PeakRMSCurrent = float32(curPeak)
		outlet.TrueRMSVoltage = float32(volt)
		outlet.AveragePower = float32(avgPower)
		outlet.VoltAmps = float32(va)

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

func (p *PDU) WhoAmI() (string, error) {
	out, err := p.execute("Whoami")
	if err != nil {
		return "", err
	}

	m := reWhoami.FindStringSubmatch(out)
	if m == nil {
		return "", fmt.Errorf("failed to find user name")
	}

	return strings.TrimSpace(m[1]), err
}

func (p *PDU) WithLogin(username, password string, cb func()) error {
	p.muLogin.Lock()
	defer p.muLogin.Unlock()

	if err := p.Login(username, password); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	cb()

	if err := p.Logout(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

func (p *PDU) Login(username, password string) error {
	if user, err := p.WhoAmI(); err != nil {
		if !errors.Is(err, ErrLoginRequired) {
			return err
		}
	} else if user != username {
		slog.Debug("Already logged in with wrong user. Logging out...", slog.String("username", user))
		if err := p.Logout(); err != nil {
			return err
		}
	} else {
		slog.Debug("Already logged in", slog.String("username", username))
		return nil // Already logged in
	}

	slog.Debug("Logging in", slog.String("username", username))

	str := ""
	sentUsername := false
	sentPassword := false

	if _, err := p.communicate(func(buf string) (bool, string, error) {
		str += buf

		switch {
		case strings.HasSuffix(str, promptReady):
			return true, "", nil

		case strings.HasSuffix(str, promptUsername):
			if err := p.send(username); err != nil {
				return false, "", err
			}

			sentUsername = true

		case strings.HasSuffix(str, promptPassword):
			if err := p.send(password); err != nil {
				return false, "", err
			}

			sentPassword = true

		case strings.HasSuffix(str, promptInvalidPassword):
			if sentUsername && sentPassword {
				return false, "", ErrInvalidPassword
			}
		}

		return false, "", nil
	}); err != nil {
		return err
	}

	user, err := p.WhoAmI()
	if err != nil {
		return err
	}

	slog.Debug("Logged in", slog.String("username", user))

	return err
}

func (p *PDU) Logout() error {
	_, err := p.execute("Logout")
	return err
}

func (p *PDU) send(cmd string) error {
	_, err := p.conn.Write([]byte(cmd + "\r\n"))
	return err
}

func (p *PDU) execute(cmd string, args ...any) (string, error) {
	cmd = fmt.Sprintf(cmd, args...)
	str := ""
	sent := false

	started := time.Now()

	res, err := p.communicate(func(buf string) (bool, string, error) {
		str += buf

		// There is no prompt after logout
		if sent == true && cmd == "Logout" {
			if str == cmd {
				time.Sleep(500 * time.Millisecond)
				return true, "", nil
			}

			return false, "", nil
		}

		switch {
		case strings.HasSuffix(str, promptReady):
			if sent {
				str = strings.TrimPrefix(str, cmd)
				str = strings.TrimSuffix(str, "\r\n"+promptReady)
				str = strings.TrimSpace(str)

				return true, str, nil
			}
			
			if err := p.send(cmd); err != nil {
				return false, "", err
			}

			sent = true
			str = ""

		case strings.HasSuffix(str, promptUsername), strings.HasSuffix(str, promptPassword):
			return false, "", ErrLoginRequired 
		}

		return false, "", nil
	})

	finished := time.Now()

	opts := []any{slog.String("command", cmd), slog.Duration("took", finished.Sub(started))}
	if len(args) > 0 {
		opts = append(opts, slog.Any("args", args))
	}

	if err != nil {
		opts = append(opts, slog.Any("error", err))
	} else if res != "" {
		opts = append(opts, slog.Int("#result", len(res)))
	}

	slog.Debug("Executed PDU command", opts...)

	return res, err
}

func (p *PDU) communicate(cb func(out string) (bool, string, error)) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.send(""); err != nil {
		return "", err
	}

	for {
		switch c := p.conn.(type) {
		case *net.TCPConn:
			if err := c.SetReadDeadline(time.Now().Add(p.timeout)); err != nil {
				return "", fmt.Errorf("failed to set read deadline: %w", err)
			}

		case serial.Port:
			if err := c.SetReadTimeout(p.timeout); err != nil {
				return "", fmt.Errorf("failed to set read deadline: %w", err)
			}

		default:
			return "", fmt.Errorf("unsupported connection type: %T", p.conn)
		}

		buf := make([]byte, 2048)
		n, err := p.conn.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue // Timeout
			}

			return "", err
		} else if n == 0 {
			continue // Timeout
		}

		sbuf := string(buf[:n])
		if done, out, err := cb(sbuf); err != nil {
			return "", err
		} else if done {
			return out, nil
		}
	}
}

func (p *PDU) lookupID(idStr string) (int, error) {
	if idStr == "all" {
		return 0, nil
	}

	if id, err := strconv.ParseInt(idStr, 0, 64); err == nil {
		if id < 0 || id > NumOutlets {
			return -1, pdu.ErrInvalidOutletID
		}

		return int(id), nil
	}

	return -1, fmt.Errorf("%w: %s", pdu.ErrNotFound, idStr)
}
