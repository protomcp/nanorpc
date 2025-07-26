// Package client implements a reconnecting NanoRPC client.
package client

import (
	"context"
	"sync"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/net/reconnect"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/common"
)

// Client is a reconnecting NanoRPC client.
type Client struct {
	reconnect.WorkGroup

	rc           *reconnect.Client
	cs           *Session
	reqCounter   *RequestCounter
	hc           *nanorpc.HashCache
	getPathOneOf func(string) nanorpc.PathOneOf
	logger       slog.Logger

	callOnConnect    func(context.Context, reconnect.WorkGroup) error
	callOnDisconnect func(context.Context) error
	callOnError      func(context.Context, error) error

	idleReadTimeout time.Duration
	mu              sync.Mutex
	queueSize       uint
}

func (c *Client) getOnConnect() func(context.Context, reconnect.WorkGroup) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.callOnConnect
}

func (c *Client) getOnDisconnect() func(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.callOnDisconnect
}
func (c *Client) getOnError() func(context.Context, error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.callOnError
}

// New creates a new [Client] using given [Config].
// If Config.HashCache is nil, the global package-level hashCache will be used.
func (cfg *Config) New() (*Client, error) {
	var c = new(Client)

	ro, err := cfg.Export()
	if err != nil {
		return nil, err
	}

	rc, err := reconnect.New(ro, c.preInit)
	if err != nil {
		return nil, err
	}

	if err := c.init(cfg, rc); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) init(cfg *Config, rc *reconnect.Client) error {
	reqCounter, err := NewRequestCounter()
	if err != nil {
		return core.Wrap(err, "RequestCounter")
	}

	c.WorkGroup = rc
	c.rc = rc

	c.queueSize = cfg.QueueSize
	c.reqCounter = reqCounter
	c.idleReadTimeout = cfg.IdleTimeout

	c.hc = cfg.getHashCache()
	c.getPathOneOf = cfg.newGetPathOneOf(c.hc)

	c.callOnConnect = cfg.OnConnect
	c.callOnDisconnect = cfg.OnDisconnect
	c.callOnError = cfg.OnError

	// Set logger from config, add component field if provided
	c.logger = cfg.Logger
	if c.logger != nil {
		c.logger = c.logger.WithField(common.FieldComponent, common.ComponentClient)
	}

	return nil
}

// NewClient creates a new [Client] with default options.
// Uses the global package-level hashCache for path hashing.
func NewClient(ctx context.Context, address string) (*Client, error) {
	cfg := Config{
		Context: ctx,
		Remote:  address,
	}

	return cfg.New()
}

// RequestCallback handles a response to a request
type RequestCallback func(context.Context, int32, *nanorpc.NanoRPCResponse) error
