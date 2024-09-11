// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"fmt"
	"regexp"
)

type AccessControlEntry struct {
	Name       string   `mapstructure:"name"`
	Operations []string `mapstructure:"operations"`
	Outlets    []struct {
		ID         string   `mapstructure:"id"`
		Operations []string `mapstructure:"operations"`

		regexID *regexp.Regexp
	} `mapstructure:"outlets"`

	regexName *regexp.Regexp
}

type AccessControlList []AccessControlEntry

func (a AccessControlList) Init() (err error) {
	for i := range a {
		e := &a[i]

		if e.regexName, err = regexp.Compile(e.Name); err != nil {
			return fmt.Errorf("invalid ACL name expression: %s: %w", e.Name, err)
		}

		for j := range e.Outlets {
			o := &e.Outlets[j]

			if o.regexID, err = regexp.Compile(o.ID); err != nil {
				return fmt.Errorf("invalid outlet ID expression: %s: %w", o.ID, err)
			}
		}
	}

	return nil
}

func (a AccessControlList) Check(commonName, operationID, outletID string) bool {
	for _, e := range a {
		if !e.regexName.MatchString(commonName) {
			continue
		}

		for _, op := range e.Operations {
			if op == operationID {
				return true
			}
		}

		for _, o := range e.Outlets {
			if !o.regexID.MatchString(outletID) {
				continue
			}

			for _, op := range o.Operations {
				if op == operationID {
					return true
				}
			}
		}
	}

	return false
}
