// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNotFound        = errors.New("failed to find outlet")
	ErrInvalidOutletID = errors.New("invalid outlet ID")
	ErrLoginRequired   = errors.New("login required")
	ErrInvalidPassword = errors.New("invalid password")
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func toKebabCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")

	return strings.ToLower(snake)
}
