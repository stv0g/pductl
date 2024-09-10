// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package pductl -generate models,client,std-http,strict-server,skip-prune -o api.gen.go openapi.yaml
