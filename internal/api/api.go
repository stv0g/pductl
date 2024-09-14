// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package api -generate models,client,std-http,strict-server,skip-prune -o api.gen.go ../../openapi.yaml

func OutletIDFromRequest(r any) string {
	switch r := r.(type) {
	case *LockOutletRequestObject:
		return r.Id
	case *SwitchOutletRequestObject:
		return r.Id
	case *RebootOutletRequestObject:
		return r.Id
	}

	return ""
}
