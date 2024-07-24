package nanorpc

import (
	"context"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/net/reconnect"
)

// Client is a reconnecting NanoRPC client.
type Client struct {
	reconnect.WorkGroup

	mu sync.Mutex
	rc *reconnect.Client
	cs *ClientSession

	queueSize  uint
	reqCounter *RequestCounter

	hc           *HashCache
	getPathOneOf func(string) isNanoRPCRequest_PathOneof

	callOnConnect    func(context.Context, reconnect.WorkGroup) error
	callOnDisconnect func(context.Context) error
	callOnError      func(context.Context, error) error
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

// New creates a new [Client] using given [ClientConfig].
func (cfg *ClientConfig) New() (*Client, error) {
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

func (c *Client) init(cfg *ClientConfig, rc *reconnect.Client) error {
	reqCounter, err := NewRequestCounter()
	if err != nil {
		return core.Wrap(err, "RequestCounter")
	}

	c.WorkGroup = rc
	c.rc = rc

	c.queueSize = cfg.QueueSize
	c.reqCounter = reqCounter

	c.hc = cfg.getHashCache()
	c.getPathOneOf = cfg.newGetPathOneOf(c.hc)

	c.callOnConnect = cfg.OnConnect
	c.callOnDisconnect = cfg.OnDisconnect
	c.callOnError = cfg.OnError

	return nil
}

// NewClient a new [Client] with default options
func NewClient(ctx context.Context, address string) (*Client, error) {
	cfg := ClientConfig{
		Context: ctx,
		Remote:  address,
	}

	return cfg.New()
}

// RequestCallback handles a response to a request
type RequestCallback func(context.Context, int32, *NanoRPCResponse) error
