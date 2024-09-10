// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pductl

import (
	"context"
	"errors"
)

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	PDU
}

// Get status of PDU
// (GET /status)
func (p *Server) Status(ctx context.Context, request StatusRequestObject) (StatusResponseObject, error) {
	sts, err := p.PDU.Status()
	if err != nil {
		return Status500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return Status200JSONResponse(*sts), nil
}

// Get status of PDU outlets
// (GET /status/outlets)
func (p *Server) StatusOutletAll(ctx context.Context, request StatusOutletAllRequestObject) (StatusOutletAllResponseObject, error) {
	sts, err := p.PDU.StatusOutletAll()
	if err != nil {
		return StatusOutletAll500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return StatusOutletAll200JSONResponse(sts), nil
}

// Get temperature of PDU
// (GET /temperature)
func (s *Server) Temperature(ctx context.Context, request TemperatureRequestObject) (TemperatureResponseObject, error) {
	t, err := s.PDU.Temperature()
	if err != nil {
		return Temperature500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return Temperature200JSONResponse{
		Temperature: float32(t),
	}, nil
}

// Get current user
// (GET /whoami)
func (s *Server) WhoAmI(ctx context.Context, request WhoAmIRequestObject) (WhoAmIResponseObject, error) {
	u, err := s.PDU.WhoAmI()
	if err != nil {
		return WhoAmI500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return WhoAmI200JSONResponse{
		Username: u,
	}, nil
}

// Clear peak RMS current
// (POST /clear)
func (s *Server) ClearMaximumCurrents(ctx context.Context, request ClearMaximumCurrentsRequestObject) (ClearMaximumCurrentsResponseObject, error) {
	if err := s.PDU.ClearMaximumCurrents(); err != nil {
		return &ClearMaximumCurrents500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return ClearMaximumCurrents200Response{}, nil
}

// Get status of PDU outlet
// (GET /outlet/{id}/status)
func (p *Server) StatusOutlet(ctx context.Context, request StatusOutletRequestObject) (StatusOutletResponseObject, error) {
	sts, err := p.PDU.StatusOutlet(request.Id)
	if err != nil {
		return StatusOutlet500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return StatusOutlet200JSONResponse(*sts), nil
}

// Switch lock state of outlet
// (POST /outlet/{id}/lock)
func (s *Server) LockOutlet(ctx context.Context, request LockOutletRequestObject) (LockOutletResponseObject, error) {
	if err := s.PDU.LockOutlet(request.Id, *request.Body); err != nil {
		if errors.Is(err, ErrNotFound) {
			return &LockOutlet404JSONResponse{
				NotFoundJSONResponse: NotFoundJSONResponse{
					Error: err.Error(),
				},
			}, nil
		}

		return &LockOutlet500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	if request.Body == nil {
		return &LockOutlet400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Error: "Missing request body",
			},
		}, nil
	}

	if err := s.PDU.LockOutlet(request.Id, *request.Body); err != nil {
		return &LockOutlet500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return LockOutlet200Response{}, nil
}

// Reboot the outlet
// (POST /outlet/{id}/reboot)
func (s *Server) RebootOutlet(ctx context.Context, request RebootOutletRequestObject) (RebootOutletResponseObject, error) {
	if err := s.PDU.RebootOutlet(request.Id); err != nil {
		return &RebootOutlet500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return RebootOutlet200Response{}, nil
}

// Switch state of outlet
// (POST /outlet/{id}/state)
func (s *Server) SwitchOutlet(ctx context.Context, request SwitchOutletRequestObject) (SwitchOutletResponseObject, error) {
	if request.Body == nil {
		return &SwitchOutlet400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse{
				Error: "Missing request body",
			},
		}, nil
	}

	if err := s.PDU.SwitchOutlet(request.Id, *request.Body); err != nil {
		return &SwitchOutlet500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Error: err.Error(),
			},
		}, nil
	}

	return SwitchOutlet200Response{}, nil
}
