// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	pdu "github.com/stv0g/pductl"
	"github.com/stv0g/pductl/baytech"
)

var (
	p *baytech.PDU

	// Flags
	address  string
	username string
	password string

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pductl",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			p, err = baytech.NewPDU(address, username, password)
			return err
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if err := p.Close(); err != nil {
				return fmt.Errorf("Failed to close PDU: %w", err)
			}

			return nil
		},
	}

	genDocs = &cobra.Command{
		Use:    "docs",
		Short:  "Generate docs",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll("./docs", 0o755); err != nil {
				return err
			}

			return doc.GenMarkdownTree(rootCmd, "./docs")
		},
	}

	getStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show PDU status",
		RunE:  getStatus,
	}

	userCmd = &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	whoAmICmd = &cobra.Command{
		Use:   "whoami",
		Short: "Displays the current user name",
		RunE:  whoami,
	}

	readTempCmd = &cobra.Command{
		Use:   "temp",
		Short: "Read current temperature",
		RunE:  readTemp,
	}

	clearMaximumCurrentCmd = &cobra.Command{
		Use:   "clear",
		Short: "Reset the maximum detected current",
		RunE:  clearMaximumCurrent,
	}

	outletCmd = &cobra.Command{
		Use:   "outlet",
		Short: "Control outlets",
	}

	outletRebootCmd = &cobra.Command{
		Use:   "reboot OUTLET",
		Short: "Reboot an outlet",
		RunE:  outletReboot,
		Args:  cobra.ExactArgs(1),
	}

	outletSwitchCmd = &cobra.Command{
		Use:   "switch OUTLET STATE",
		Short: "Switch an outlet on/off",
		RunE:  outletSwitch,
		Args:  cobra.ExactArgs(2),
	}

	outletLockCmd = &cobra.Command{
		Use:   "lock OUTLET STATE",
		Short: "Lock or unlock an outlet",
		RunE:  outletLock,
		Args:  cobra.ExactArgs(2),
	}

	outletStatusCmd = &cobra.Command{
		Use:   "status OUTLET",
		Short: "Get status of outlet",
		RunE:  outletStatus,
		Args:  cobra.ExactArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(getStatusCmd)
	rootCmd.AddCommand(readTempCmd)
	rootCmd.AddCommand(clearMaximumCurrentCmd)
	rootCmd.AddCommand(outletCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(genDocs)

	userCmd.AddCommand(whoAmICmd)

	outletCmd.AddCommand(outletLockCmd)
	outletCmd.AddCommand(outletRebootCmd)
	outletCmd.AddCommand(outletSwitchCmd)
	outletCmd.AddCommand(outletStatusCmd)

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&address, "address", "10.208.1.1:4141", "Address of TCP socket for PDU communication")
	pf.StringVar(&username, "username", "admin", "Username")
	pf.StringVar(&password, "password", "admin", "password")
}

func getOutletID(arg string) (int, error) {
	if arg == "all" {
		return pdu.All, nil
	}

	if id, err := strconv.ParseInt(arg, 0, 64); err == nil {
		return int(id), nil
	}

	// Attempt to lookup outlet by name
	outlets, err := p.StatusOutlets()
	if err != nil {
		return -1, fmt.Errorf("failed to get outlets from PDU: %w", err)
	}

	for i, outlet := range outlets {
		if outlet.Name == arg {
			return i + 1, nil
		}
	}

	return -1, fmt.Errorf("failed to find outlet %s", arg)
}

func getOutletIDandState(arg1, arg2 string) (id int, state bool, err error) {
	if id, err = getOutletID(arg1); err != nil {
		return -1, false, err
	}

	switch arg2 {
	case "off", "false", "0":
		state = false
	case "on", "true", "1":
		state = true
	default:
		return -1, false, fmt.Errorf("failed to parse outlet state: %w", err)
	}

	return id, state, nil
}

func getStatus(_ *cobra.Command, _ []string) error {
	sts, err := p.Status()
	if err != nil {
		return fmt.Errorf("Failed to get status: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "   ")

	return enc.Encode(sts)
}

func whoami(_ *cobra.Command, _ []string) error {
	user, err := p.WhoAmI()
	if err != nil {
		return fmt.Errorf("Failed to send command: %w", err)
	}

	fmt.Print(user)

	return nil
}

func readTemp(_ *cobra.Command, _ []string) error {
	temp, err := p.Temperature()
	if err != nil {
		return fmt.Errorf("Failed to send command: %w", err)
	}

	fmt.Print(temp)

	return nil
}

func clearMaximumCurrent(_ *cobra.Command, _ []string) error {
	if err := p.ClearMaximumCurrents(); err != nil {
		return fmt.Errorf("Failed to clear maximum current: %w", err)
	}

	return nil
}

func outletReboot(_ *cobra.Command, args []string) error {
	id, err := getOutletID(args[0])
	if err != nil {
		return err
	}

	if err := p.RebootOutlet(id); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func outletSwitch(_ *cobra.Command, args []string) error {
	id, state, err := getOutletIDandState(args[0], args[1])
	if err != nil {
		return err
	}

	if err := p.SwitchOutlet(id, state); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func outletLock(_ *cobra.Command, args []string) error {
	id, state, err := getOutletIDandState(args[0], args[1])
	if err != nil {
		return err
	}

	if err := p.LockOutlet(id, state); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func outletStatus(_ *cobra.Command, args []string) error {
	id, err := getOutletID(args[0])
	if err != nil {
		return err
	}

	outlets, err := p.StatusOutlets()
	if err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	var out any
	if id == pdu.All {
		out = outlets
	} else if id < 1 || id > len(outlets) {
		return fmt.Errorf("invalid outlet number: %d", id)
	} else {
		out = outlets[id-1]
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "   ")

	return enc.Encode(out)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
