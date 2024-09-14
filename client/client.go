// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"

	pdu "github.com/stv0g/pductl"
	"github.com/stv0g/pductl/internal/api"
)

var _ pdu.PDU = (*Client)(nil)

type Client struct {
	client *api.ClientWithResponses
	ctx    context.Context
}

func NewPDU(address string, opts ...api.ClientOption) (c *Client, err error) {
	c = &Client{}

	if c.client, err = api.NewClientWithResponses(address+"/api/v1", opts...); err != nil {
		return nil, err
	}

	c.ctx = context.Background()

	return c, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) SwitchOutlet(id string, state bool) (err error) {
	r, err := c.client.SwitchOutletWithResponse(c.ctx, id, state)
	if err != nil {
		return err
	} else if p := r.JSON400; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON401; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON404; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return errors.New(p.Error)
	}

	return nil
}

func (c *Client) LockOutlet(id string, state bool) (err error) {
	r, err := c.client.LockOutletWithResponse(c.ctx, id, state)
	if err != nil {
		return err
	} else if p := r.JSON400; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON401; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON404; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return errors.New(p.Error)
	}

	return nil
}

func (c *Client) RebootOutlet(id string) error {
	r, err := c.client.RebootOutletWithResponse(c.ctx, id)
	if err != nil {
		return err
	} else if p := r.JSON400; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON401; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON404; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return errors.New(p.Error)
	}

	return nil
}

func (c *Client) Status(detailed bool) (*pdu.Status, error) {
	r, err := c.client.StatusWithResponse(c.ctx, &api.StatusParams{
		Detailed: &detailed,
	})
	if err != nil {
		return nil, err
	} else if p := r.JSON401; p != nil {
		return nil, errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return nil, errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return nil, errors.New(p.Error)
	}

	return r.JSON200, nil
}

func (c *Client) ClearMaximumCurrents() error {
	r, err := c.client.ClearMaximumCurrentsWithResponse(c.ctx)
	if err != nil {
		return err
	} else if p := r.JSON400; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON401; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return errors.New(p.Error)
	}

	return nil
}

func (c *Client) Temperature() (float64, error) {
	r, err := c.client.TemperatureWithResponse(c.ctx)
	if err != nil {
		return -1, err
	} else if p := r.JSON401; p != nil {
		return -1, errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return -1, errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return -1, errors.New(p.Error)
	}

	return float64(r.JSON200.Temperature), nil
}

func (c *Client) WhoAmI() (string, error) {
	r, err := c.client.WhoAmIWithResponse(c.ctx)
	if err != nil {
		return "", err
	} else if p := r.JSON401; p != nil {
		return "", errors.New(p.Error)
	} else if p := r.JSON403; p != nil {
		return "", errors.New(p.Error)
	} else if p := r.JSON500; p != nil {
		return "", errors.New(p.Error)
	}

	return r.JSON200.Username, nil
}
