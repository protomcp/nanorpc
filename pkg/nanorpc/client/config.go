package client

import (
	"context"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/config"
	"darvaza.org/x/net/reconnect"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// hashCache is the default hash cache for the client package.
// It is used when Config.HashCache is nil and provides automatic
// FNV-1a path hashing with collision detection.
var hashCache = new(nanorpc.HashCache)

// Config describes how the [Client] will operate
type Config struct {
	Context         context.Context
	Logger          slog.Logger
	WaitReconnect   reconnect.Waiter
	HashCache       *nanorpc.HashCache
	OnConnect       func(context.Context, reconnect.WorkGroup) error
	OnDisconnect    func(context.Context) error
	OnError         func(context.Context, error) error
	Remote          string
	DialTimeout     time.Duration `default:"2s"`
	ReadTimeout     time.Duration `default:"2s"`
	IdleTimeout     time.Duration `default:"10s"`
	WriteTimeout    time.Duration `default:"2s"`
	ReconnectDelay  time.Duration `default:"5s"`
	KeepAlive       time.Duration `default:"5s"`
	QueueSize       uint
	AlwaysHashPaths bool
}

// SetDefaults fills gaps in [Config].
// If HashCache is nil, assigns the global package-level hashCache.
func (cfg *Config) SetDefaults() error {
	if err := config.Set(cfg); err != nil {
		return err
	}

	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	if cfg.HashCache == nil {
		// use global cache
		cfg.HashCache = hashCache
	}

	if cfg.WaitReconnect == nil {
		cfg.WaitReconnect = reconnect.NewConstantWaiter(cfg.ReconnectDelay)
	}

	return nil
}

// Export generates a [reconnect.Config]
func (cfg *Config) Export() (*reconnect.Config, error) {
	// Validate remote address using reconnect package which supports both TCP and Unix sockets
	if err := reconnect.ValidateRemote(cfg.Remote); err != nil {
		return nil, core.Wrap(err, "Remote")
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

func (cfg *Config) getHashCache() *nanorpc.HashCache {
	if hc := cfg.HashCache; hc != nil {
		// use given HashCache
		return hc
	}

	// use global cache
	return hashCache
}

func (cfg *Config) newGetPathOneOf(hc *nanorpc.HashCache) func(string) nanorpc.PathOneOf {
	if cfg.AlwaysHashPaths {
		// use path_hash
		if hc == nil {
			hc = cfg.getHashCache()
		}

		return func(path string) nanorpc.PathOneOf {
			hash, err := hc.Hash(path)
			if err != nil {
				cfg.logErrorf(err, "Falling back to string path to maintain compatibility")
				// Fall back to string path on hash collision
				return nanorpc.GetPathOneOfString(path)
			}

			return nanorpc.GetPathOneOfHash(hash)
		}
	}

	// use string
	return func(path string) nanorpc.PathOneOf {
		return nanorpc.GetPathOneOfString(path)
	}
}

func (cfg *Config) logErrorf(err error, msg string, args ...any) {
	if cfg != nil && cfg.Logger != nil {
		logger := cfg.Logger.Error()
		logger = logger.WithField(slog.ErrorFieldName, err)
		logger.Printf(msg, args...)
	}
}
