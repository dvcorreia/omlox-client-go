// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

// Defaults
const (
	DefaultRequestTimeout = 60 * time.Second
	DefaultConnectTimeout = 60 * time.Second
)

// DefaultHttpClient uses cleanhttp, which has the same default values as net/http client, but
// does not share state with other clients (see: gh/hashicorp/go-cleanhttp)
func DefaultHttpClient() *http.Client {
	return cleanhttp.DefaultPooledClient()
}

// ClientOpt is a configuration option to initialize a client.
type ClientOpt func(*Client) error

// WithHTTPClient sets the HTTP client to use for all API requests.
func WithHTTPClient(client *http.Client) ClientOpt {
	return func(c *Client) error {
		c.client = client
		return nil
	}
}

// WithRequestTimeout, given a non-negative value, will apply the timeout to
// each request function unless an earlier deadline is passed to the request
// function through context.Context.
func WithRequestTimeout(timeout time.Duration) ClientOpt {
	return func(c *Client) error {
		if timeout < 0 {
			return fmt.Errorf("request timeout must not be negative")
		}
		c.timeout = timeout
		return nil
	}
}

// WithConnectTimeout, given a non-negative value, will apply the timeout when
// connecting to the websocket interface.
func WithConnectTimeout(timeout time.Duration) ClientOpt {
	return func(c *Client) error {
		if timeout < 0 {
			return fmt.Errorf("request timeout must not be negative")
		}
		c.connectTimeout = timeout
		return nil
	}
}

// WithRateLimiter configures how frequently requests are allowed to happen.
// If this pointer is nil, then there will be no limit set. Note that an
// empty struct rate.Limiter is equivalent to blocking all requests.
func WithRateLimiter(limiter *rate.Limiter) ClientOpt {
	return func(c *Client) error {
		c.rateLimiter = limiter
		return nil
	}
}

// pendingSubscription awaiting for subscription ID from the server
type pendingSubscription struct {
	Sid int
	Err error
}

// Client represents a client connection to a Omlox Hub.
type Client struct {
	mu sync.RWMutex

	baseURL *url.URL

	client      *http.Client
	timeout     time.Duration // request timeout (includes websocket requests)
	rateLimiter *rate.Limiter

	Trackables TrackablesAPI
	Providers  ProvidersAPI

	errg   *errgroup.Group
	cancel context.CancelFunc

	// websockets connection
	conn           *websocket.Conn
	connectTimeout time.Duration
	closed         bool

	// subscriptions
	subs map[int]*Subcription

	// chanel to receive pending subcriptions ack
	// can only be one subscription per client awaiting for subscription
	// this is a design constraint from having to wait for a subscription ID
	pending chan chan pendingSubscription
}

// NewClient returns a new Client configured by the given options.
func NewClient(baseURL string, opts ...ClientOpt) (*Client, error) {
	hubURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := Client{
		baseURL: hubURL,
		timeout: DefaultRequestTimeout,

		connectTimeout: DefaultConnectTimeout,
		closed:         true,

		pending: make(chan chan pendingSubscription, 1),
		subs:    make(map[int]*Subcription),
	}

	c.Trackables = TrackablesAPI{
		client: &c,
	}

	c.Providers = ProvidersAPI{
		client: &c,
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(&c); err != nil {
				return nil, err
			}
		}
	}

	if c.client == nil {
		c.client = DefaultHttpClient()
	}

	return &c, nil
}

// sendStructuredRequestParseResponse constructs a structured request, sends it, and parses the response
func sendStructuredRequestParseResponse[ResponseT any](
	ctx context.Context,
	client *Client,
	method string,
	path string,
	body any,
	parameters url.Values,
	headers http.Header,
) (*ResponseT, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("could not encode request body: %w", err)
	}

	return sendRequestParseResponse[ResponseT](
		ctx,
		client,
		method,
		path,
		&buf,
		parameters,
		headers,
	)
}

// sendRequestParseResponse constructs a request, sends it, and parses the response.
func sendRequestParseResponse[ResponseT any](
	ctx context.Context,
	client *Client,
	method string,
	path string,
	body io.Reader,
	parameters url.Values,
	headers http.Header,
) (*ResponseT, error) {
	// apply the client-level request timeout, if set
	if client.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.timeout)
		defer cancel()
	}

	// TODO: set User-Agent and Content-Type headers

	req, err := client.newRequest(ctx, method, path, body, parameters, headers)
	if err != nil {
		return nil, err
	}

	resp, err := client.send(ctx, req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := isResponseError(resp); err != nil {
		return nil, err
	}

	return parseResponse[ResponseT](resp.Body)
}

// sendRequestParseResponse constructs a request, sends it, and parses the response.
func sendRequestParseResponseList[ResponseT any](
	ctx context.Context,
	client *Client,
	method string,
	path string,
	body io.Reader,
	parameters url.Values,
	headers http.Header,
) ([]ResponseT, error) {
	// apply the client-level request timeout, if set
	if client.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.timeout)
		defer cancel()
	}

	// TODO: set User-Agent and Content-Type headers

	req, err := client.newRequest(ctx, method, path, body, parameters, headers)
	if err != nil {
		return nil, err
	}

	resp, err := client.send(ctx, req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := isResponseError(resp); err != nil {
		return nil, err
	}

	return parseResponseList[ResponseT](resp.Body)
}

// newRequest constructs a new request.
func (c *Client) newRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	parameters url.Values,
	headers http.Header,
) (*http.Request, error) {
	// concatenate the base address with the given path
	url := c.baseURL.JoinPath(path)

	// add query parameters (if any)
	if len(parameters) != 0 {
		url.RawQuery = parameters.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, fmt.Errorf("could not create '%s %s' request: %w", method, url.String(), err)
	}

	// populate request headers
	if headers != nil {
		req.Header = headers
	}

	return req, nil
}

// send sends the given request to Omlox.
func (c *Client) send(ctx context.Context, req *http.Request) (*http.Response, error) {
	// block on the rate limiter, if set
	if c.rateLimiter != nil {
		c.rateLimiter.Wait(ctx)
	}

	return c.client.Do(req)
}

// parseResponse fully consumes the given response body without closing it and
// parses the data into a generic Response[T] structure. If the response body
// is empty, a nil value will be returned.
func parseResponse[T any](responseBody io.Reader) (*T, error) {
	// First, read the data into a buffer. This is not super efficient but we
	// want to know if we actually have a body or not.
	var buf bytes.Buffer

	_, err := buf.ReadFrom(responseBody)
	if err != nil {
		return nil, err
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	var response T
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// parseResponseList fully consumes the given response body without closing it and
// parses the data into a generic T structure list. If the response body
// is empty, a empty T list will be returned.
func parseResponseList[T any](responseBody io.Reader) ([]T, error) {
	// First, read the data into a buffer. This is not super efficient but we
	// want to know if we actually have a body or not.
	var buf bytes.Buffer

	_, err := buf.ReadFrom(responseBody)
	if err != nil {
		return nil, err
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	var response []T
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return nil, err
	}

	return response, nil
}
