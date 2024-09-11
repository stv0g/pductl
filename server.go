// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"github.com/stv0g/pductl/internal/api"
)

var _ api.StrictServerInterface = (*Server)(nil)

var (
	ErrMissingClientCert = errors.New("missing client certificate")
	ErrAccessDenied      = errors.New("ACL denies access")
)

type Server struct {
	PDU
	ACL AccessControlList
}

func Handler(mux *http.ServeMux, p PDU, acl AccessControlList) http.Handler {
	svr := &Server{
		PDU: p,
		ACL: acl,
	}

	mwLog := func(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
			response, err = f(ctx, w, r, request)

			slog.Debug("API Request", slog.String("operation", toKebabCase(operationID)), slog.Any("request", request), slog.Any("response", response), slog.Any("error", err))

			return response, err
		}
	}

	mwAuth := func(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
		if len(acl) == 0 {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
			if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 || len(r.TLS.VerifiedChains[0]) == 0 {
				return nil, ErrMissingClientCert
			}

			commonName := r.TLS.VerifiedChains[0][0].Subject.CommonName
			operationID = toKebabCase(operationID)
			outletID := api.OutletIDFromRequest(request)

			if !acl.Check(commonName, operationID, outletID) {
				return nil, ErrAccessDenied
			}

			return f(ctx, w, r, request)
		}
	}

	si := api.NewStrictHandlerWithOptions(svr, []nethttp.StrictHTTPMiddlewareFunc{mwAuth, mwLog}, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  errorHandlerFunc,
		ResponseErrorHandlerFunc: errorHandlerFunc,
	})

	return api.HandlerWithOptions(si, api.StdHTTPServerOptions{
		BaseURL:          "/api/v1",
		BaseRouter:       mux,
		ErrorHandlerFunc: errorHandlerFunc,
	})
}

func errorHandlerFuncFor(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		switch {
		case errors.Is(err, ErrAccessDenied):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, ErrMissingClientCert):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(api.ApiResponse{
			Error: err.Error(),
		})
	}
}

func errorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	errorHandlerFuncFor(err)(w, r)
}

// Get status of PDU
// (GET /status)
func (p *Server) Status(ctx context.Context, request api.StatusRequestObject) (api.StatusResponseObject, error) {
	sts, err := p.PDU.Status()
	if err != nil {
		return api.Status500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.Status200JSONResponse(*sts), nil
}

// Get status of PDU outlets
// (GET /status/outlets)
func (p *Server) StatusOutletAll(ctx context.Context, request api.StatusOutletAllRequestObject) (api.StatusOutletAllResponseObject, error) {
	sts, err := p.PDU.StatusOutletAll()
	if err != nil {
		return api.StatusOutletAll500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.StatusOutletAll200JSONResponse(sts), nil
}

// Get temperature of PDU
// (GET /temperature)
func (s *Server) Temperature(ctx context.Context, request api.TemperatureRequestObject) (api.TemperatureResponseObject, error) {
	t, err := s.PDU.Temperature()
	if err != nil {
		return api.Temperature500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.Temperature200JSONResponse{
		Temperature: float32(t),
	}, nil
}

// Get current user
// (GET /whoami)
func (s *Server) WhoAmI(ctx context.Context, request api.WhoAmIRequestObject) (api.WhoAmIResponseObject, error) {
	u, err := s.PDU.WhoAmI()
	if err != nil {
		return api.WhoAmI500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.WhoAmI200JSONResponse{
		Username: u,
	}, nil
}

// Clear peak RMS current
// (POST /clear)
func (s *Server) ClearMaximumCurrents(ctx context.Context, request api.ClearMaximumCurrentsRequestObject) (api.ClearMaximumCurrentsResponseObject, error) {
	if err := s.PDU.ClearMaximumCurrents(); err != nil {
		return &api.ClearMaximumCurrents500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.ClearMaximumCurrents200Response{}, nil
}

// Get status of PDU outlet
// (GET /outlet/{id}/status)
func (p *Server) StatusOutlet(ctx context.Context, request api.StatusOutletRequestObject) (api.StatusOutletResponseObject, error) {
	sts, err := p.PDU.StatusOutlet(request.Id)
	if err != nil {
		return api.StatusOutlet500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.StatusOutlet200JSONResponse(*sts), nil
}

// Switch lock state of outlet
// (POST /outlet/{id}/lock)
func (s *Server) LockOutlet(ctx context.Context, request api.LockOutletRequestObject) (api.LockOutletResponseObject, error) {
	if err := s.PDU.LockOutlet(request.Id, *request.Body); err != nil {
		if errors.Is(err, ErrNotFound) {
			return &api.LockOutlet404JSONResponse{
				NotFoundJSONResponse: api.NotFoundJSONResponse{
					Error: err.Error(),
				},
			}, nil
		}

		return &api.LockOutlet500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	if request.Body == nil {
		return &api.LockOutlet400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error: "Missing request body",
			},
		}, nil
	}

	if err := s.PDU.LockOutlet(request.Id, *request.Body); err != nil {
		return &api.LockOutlet500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.LockOutlet200Response{}, nil
}

// Reboot the outlet
// (POST /outlet/{id}/reboot)
func (s *Server) RebootOutlet(ctx context.Context, request api.RebootOutletRequestObject) (api.RebootOutletResponseObject, error) {
	if err := s.PDU.RebootOutlet(request.Id); err != nil {
		return &api.RebootOutlet500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.RebootOutlet200Response{}, nil
}

// Switch state of outlet
// (POST /outlet/{id}/state)
func (s *Server) SwitchOutlet(ctx context.Context, request api.SwitchOutletRequestObject) (api.SwitchOutletResponseObject, error) {
	if request.Body == nil {
		return &api.SwitchOutlet400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error: "Missing request body",
			},
		}, nil
	}

	if err := s.PDU.SwitchOutlet(request.Id, *request.Body); err != nil {
		return &api.SwitchOutlet500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return api.SwitchOutlet200Response{}, nil
}
