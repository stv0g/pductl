// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	pdu "github.com/stv0g/pductl"
)

var (
	p *pdu.PDU

	// Flags
	address  string
	username string
	password string
	outlet   string
	outletID int

	// Commands
	rootCmd = &cobra.Command{
		Use:               "pductl",
		Short:             "A command line utility, REST API and Prometheus Exporter for Baytech PDUs",
		DisableAutoGenTag: true,
	}

	genDocs = &cobra.Command{
		Use:    "docs",
		Short:  "Generate docs",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			os.MkdirAll("./docs", 0o755)
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
		RunE:  rebootOutlet,
		Args:  cobra.ExactArgs(1),
	}

	outletSwitchCmd = &cobra.Command{
		Use:   "switch OUTLET STATE",
		Short: "Switch an outlet on/off",
		RunE:  switchOutlet,
		Args:  cobra.ExactArgs(2),
	}

	outletLockCmd = &cobra.Command{
		Use:   "lock OUTLET STATE",
		Short: "Lock or unlock an outlet",
		RunE:  lockOutlet,
		Args:  cobra.ExactArgs(2),
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

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		p = pdu.NewPDU(address, username, password)
	}
	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if err := p.Close(); err != nil {
			return fmt.Errorf("Failed to close PDF: %w", err)
		}

		return nil
	}

	pf := rootCmd.PersistentFlags()

	pf.StringVar(&address, "address", "10.208.1.1:4141", "Address of TCP socket for PDU communication")
	pf.StringVar(&username, "username", "admin", "Username")
	pf.StringVar(&password, "password", "admin", "password")
}

func getOutletID(arg string) (int, error) {
	if arg == "all" {
		return pdu.All, nil
	}

	id, err := strconv.ParseInt(arg, 0, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse outlet number: %w", err)
	}

	return int(id), nil
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
	enc.Encode(sts)

	return nil
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
		slog.Error("Failed to clear maximum current", err)
	}

	return nil
}

func rebootOutlet(_ *cobra.Command, args []string) error {
	id, err := getOutletID(args[0])
	if err != nil {
		return err
	}

	if err := p.RebootOutlet(id); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func switchOutlet(_ *cobra.Command, args []string) error {
	id, state, err := getOutletIDandState(args[0], args[1])
	if err != nil {
		return err
	}

	if err := p.SwitchOutlet(id, state); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func lockOutlet(_ *cobra.Command, args []string) error {
	id, state, err := getOutletIDandState(args[0], args[1])
	if err != nil {
		return err
	}

	if err := p.LockOutlet(id, state); err != nil {
		return fmt.Errorf("Failed to control outlet: %w", err)
	}

	return nil
}

func main() {
	rootCmd.Execute()
}
