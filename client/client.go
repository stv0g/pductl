// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	pdu "github.com/stv0g/pductl"
)

var _ pdu.PDU = (*Client)(nil)

type Client struct {
	client *pdu.ClientWithResponses
	ctx    context.Context
}

func handleResponse(r *http.Response) error {
	if r.StatusCode == 200 {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var ap pdu.ApiResponse
	if err := json.Unmarshal(body, &ap); err != nil {
		return fmt.Errorf("failed to ")
	}

	return errors.New(ap.Error)
}

func NewPDU(address string) (c *Client, err error) {
	c = &Client{}

	if c.client, err = pdu.NewClientWithResponses(address + "/api/v1"); err != nil {
		return nil, err
	}

	c.ctx = context.Background()

	return c, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) SwitchOutlet(id string, state bool) (err error) {
	resp, err := c.client.SwitchOutletWithResponse(c.ctx, id, state)
	if err != nil {
		return err
	} else if r := resp.JSON500; r != nil {
		return errors.New(r.Error)
	}

	return nil
}

func (c *Client) LockOutlet(id string, state bool) (err error) {
	r, err := c.client.LockOutlet(c.ctx, id, state)
	if err != nil {
		return err
	}

	return handleResponse(r)
}

func (c *Client) RebootOutlet(id string) error {
	r, err := c.client.RebootOutlet(c.ctx, id)
	if err != nil {
		return err
	}

	return handleResponse(r)
}

func (c *Client) Status() (*pdu.Status, error) {
	r, err := c.client.StatusWithResponse(c.ctx)
	if err != nil {
		return nil, err
	} else if r.StatusCode() != 200 {
		return nil, handleResponse(r.HTTPResponse)
	}

	return r.JSON200, nil
}

func (c *Client) StatusOutlet(id string) (*pdu.OutletStatus, error) {
	r, err := c.client.StatusOutletWithResponse(c.ctx, id)
	if err != nil {
		return nil, err
	} else if r.StatusCode() != 200 {
		return nil, handleResponse(r.HTTPResponse)
	}

	return r.JSON200, nil
}

func (c *Client) StatusOutletAll() ([]pdu.OutletStatus, error) {
	r, err := c.client.StatusOutletAllWithResponse(c.ctx)
	if err != nil {
		return nil, err
	} else if r.StatusCode() != 200 {
		return nil, handleResponse(r.HTTPResponse)
	}

	return *r.JSON200, nil
}

func (c *Client) ClearMaximumCurrents() error {
	r, err := c.client.ClearMaximumCurrents(c.ctx)
	if err != nil {
		return err
	}

	return handleResponse(r)
}

func (c *Client) Temperature() (float64, error) {
	r, err := c.client.TemperatureWithResponse(c.ctx)
	if err != nil {
		return -1, err
	} else if r.StatusCode() != 200 {
		return -1, handleResponse(r.HTTPResponse)
	}

	return float64(r.JSON200.Temperature), nil
}

func (c *Client) WhoAmI() (string, error) {
	r, err := c.client.WhoAmIWithResponse(c.ctx)
	if err != nil {
		return "", err
	} else if r.StatusCode() != 200 {
		return "", handleResponse(r.HTTPResponse)
	}

	return r.JSON200.Username, nil
}
