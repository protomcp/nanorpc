package nanorpc

import (
	"sync"

	"darvaza.org/x/net/reconnect"
)

// Client is a reconnecting NanoRPC client.
type Client struct {
	reconnect.WorkGroup

	mu sync.Mutex
	rc *reconnect.Client
	cs *ClientSession

	queueSize uint
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
	c.WorkGroup = rc
	c.rc = rc
	c.queueSize = cfg.QueueSize
	return nil
}
