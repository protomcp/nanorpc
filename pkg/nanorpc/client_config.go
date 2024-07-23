package nanorpc

import (
	"context"
	"net"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/net/reconnect"
)

// ClientConfig describes how the [Client] will operate
type ClientConfig struct {
	Context context.Context
	Logger  slog.Logger
	Remote  string

	QueueSize uint

	// OnConnect is called when the connection is established and workers spawned.
	OnConnect func(context.Context, reconnect.WorkGroup) error
	// OnDisconnect is called after closing the connection, purging callbacks and
	// can be used to prevent further connection retries.
	OnDisconnect func(context.Context) error
	// OnError is called after all errors and gives us the opportunity to
	// decide how the error should be treated by the reconnection logic.
	OnError func(context.Context, error) error
}

// SetDefaults fills gaps in [ClientConfig]
func (cfg *ClientConfig) SetDefaults() error {
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
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
		Logger:  cfg.Logger,
		Remote:  cfg.Remote,
	}

	return out, nil
}
