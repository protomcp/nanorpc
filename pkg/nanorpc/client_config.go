package nanorpc

import (
	"context"
	"net"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/config"
	"darvaza.org/x/net/reconnect"
)

// ClientConfig describes how the [Client] will operate
type ClientConfig struct {
	Context context.Context
	Logger  slog.Logger
	Remote  string

	QueueSize uint

	KeepAlive    time.Duration `default:"5s"`
	DialTimeout  time.Duration `default:"2s"`
	ReadTimeout  time.Duration `default:"2s"`
	WriteTimeout time.Duration `default:"2s"`

	ReconnectDelay time.Duration `default:"5s"`
	WaitReconnect  reconnect.Waiter

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
	if err := config.Set(cfg); err != nil {
		return err
	}

	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	if cfg.WaitReconnect == nil {
		cfg.WaitReconnect = reconnect.NewConstantWaiter(cfg.ReconnectDelay)
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

		KeepAlive:    cfg.KeepAlive,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		WaitReconnect: cfg.WaitReconnect,
	}

	return out, nil
}
