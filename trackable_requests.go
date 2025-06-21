// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// TrackablesAPI_ is a simple wrapper around the client for trackables requests.
type TrackablesAPI_ struct {
	client *Client
}

// List lists all trackables.
func (c *TrackablesAPI_) List(ctx context.Context) ([]Trackable, error) {
	requestPath := "/trackables/summary"

	return sendRequestParseResponseList[Trackable](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}

// IDs lists all trackable IDs.
func (c *TrackablesAPI_) IDs(ctx context.Context) ([]uuid.UUID, error) {
	requestPath := "/trackables"

	return sendRequestParseResponseList[uuid.UUID](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}

// Create creates a trackable.
func (c *TrackablesAPI_) Create(ctx context.Context, trackable Trackable) (*Trackable, error) {
	requestPath := "/trackables"

	return sendStructuredRequestParseResponse[Trackable](
		ctx,
		c.client,
		http.MethodPost,
		requestPath,
		trackable,
		nil, // request query parameters
		nil, // request headers
	)
}

// DeleteAll deletes all trackables.
func (c *TrackablesAPI_) DeleteAll(ctx context.Context) error {
	requestPath := "/trackables"

	_, err := sendRequestParseResponse[struct{}](
		ctx,
		c.client,
		http.MethodDelete,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)

	return err
}

// Get gets a trackable.
func (c *TrackablesAPI_) Get(ctx context.Context, id uuid.UUID) (*Trackable, error) {
	requestPath := "/trackables/" + id.String()

	return sendRequestParseResponse[Trackable](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}

// Delete deletes a trackable.
func (c *TrackablesAPI_) Delete(ctx context.Context, id uuid.UUID) error {
	requestPath := "/trackables/" + id.String()

	_, err := sendRequestParseResponse[struct{}](
		ctx,
		c.client,
		http.MethodDelete,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)

	return err
}

// GetLocation gets the last most recent location for a trackable.
// It considers all recent location updates of the trackables location providers.
func (c *TrackablesAPI_) GetLocation(ctx context.Context, id uuid.UUID) (*Location, error) {
	requestPath := "/trackables/" + id.String() + "/location"

	return sendRequestParseResponse[Location](
		ctx,
		c.client,
		http.MethodGet,
		requestPath,
		nil, // request body
		nil, // request query parameters
		nil, // request headers
	)
}
