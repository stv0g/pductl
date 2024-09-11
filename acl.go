// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

type AccessControlEntry struct {
	Name string `mapstructure:"name"`

	Status  bool `mapstructure:"status"`
	Clear   bool `mapstructure:"clear"`
	Outlets []struct {
		ID     string   `mapstructure:"id"`
		Access []string `mapstructure:"access"`
	} `mapstructure:"outlets"`
}
