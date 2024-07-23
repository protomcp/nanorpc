package nanorpc

import (
	"context"
	"net"

	"darvaza.org/core"
	"darvaza.org/x/net/reconnect"
)

// ClientConfig describes how the [Client] will operate
type ClientConfig struct {
	Context context.Context
	Remote  string

	QueueSize uint
}

// SetDefaults fills gaps in [ClientConfig]
func (cfg *ClientConfig) SetDefaults() error {
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	return nil
}

// Export generates a [reconnect.Config]
func (cfg *ClientConfig) Export() (*reconnect.Config, error) {
	_, port, err := core.SplitHostPort(cfg.Remote)
	switch {
	case err != nil:
		return nil, core.Wrap(err, "Remote")
	case port == "", port == "0":
		return nil, &net.AddrError{
			Addr: cfg.Remote,
			Err:  "Remote: port not specified",
		}
	}

	if err := cfg.SetDefaults(); err != nil {
		return nil, err
	}

	out := &reconnect.Config{
		Context: cfg.Context,
		Remote:  cfg.Remote,
	}
	return out, nil
}
