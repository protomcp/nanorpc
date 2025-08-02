package client

import (
	"context"
	"io/fs"
	"net"

	"darvaza.org/core"
	"darvaza.org/x/net/reconnect"
)

// preInit adjusts the [reconnect.Config] to use
// [Client]'s callbacks.
func (c *Client) preInit(cfg *reconnect.Config) error {
	cfg.OnConnect = c.onReconnectConnect
	cfg.OnSession = c.onReconnectSession
	cfg.OnDisconnect = c.onReconnectDisconnect
	cfg.OnError = c.onReconnectError
	return nil
}

//
// callbacks
//

func (c *Client) onReconnectConnect(ctx context.Context, conn net.Conn) error {
	c.LogDebug(conn.RemoteAddr(), "connected")

	cs := newClientSession(ctx, c, c.queueSize, conn)
	return c.setSession(cs)
}

func (c *Client) onReconnectSession(ctx context.Context) error {
	cs, err := c.getSession()
	if err != nil {
		return err
	}

	defer c.endSession(cs)

	if err := cs.Spawn(); err != nil {
		return err
	}

	if fn := c.getOnConnect(); fn != nil {
		if err := fn(ctx, cs); err != nil {
			return err
		}
	}

	return cs.Wait()
}

func (c *Client) onReconnectDisconnect(ctx context.Context, conn net.Conn) error {
	fn := c.getOnDisconnect()

	if fn == nil {
		c.LogDebug(conn.RemoteAddr(), "disconnected")
	}

	if cs, _ := c.getSession(); cs != nil {
		_ = cs.Close()
	}

	if fn != nil {
		return fn(ctx)
	}

	return nil
}

func (c *Client) onReconnectError(ctx context.Context, conn net.Conn, err error) error {
	var addr net.Addr

	if fn := c.getOnError(); fn != nil {
		return fn(ctx, err)
	}

	if conn != nil {
		// conn is nil when connection failed
		addr = conn.RemoteAddr()
	}
	c.LogError(addr, err, "error")

	return err
}

//
// session hooks
//

func (c *Client) getSession() (*Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cs != nil {
		return c.cs, nil
	}

	return nil, reconnect.ErrNotConnected
}

func (c *Client) endSession(*Session) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cs = nil
}

func (c *Client) setSession(cs *Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case cs == nil:
		return core.QuietWrap(fs.ErrInvalid, "missing session")
	case c.cs != nil:
		return core.QuietWrap(fs.ErrInvalid, "session already attached")
	default:
		c.cs = cs
		return nil
	}
}

//
// pass through
//

// Connect initiates the nanorpc reconnecting connection.
func (c *Client) Connect() error {
	return c.rc.Connect()
}
