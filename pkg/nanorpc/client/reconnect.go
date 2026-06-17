package client

import (
	"context"
	"net"

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
	c.LogDebug(conn.RemoteAddr(), nil, "connected")

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
		c.LogDebug(conn.RemoteAddr(), nil, "disconnected")
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
	c.LogError(addr, err, nil, "error")

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

func (c *Client) endSession(cs *Session) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cs == nil {
		// Already disconnected: keep the live open channel so that
		// waiters holding it are not orphaned by a spurious reset.
		return
	}

	if cs != nil && c.cs != cs {
		// A newer session already replaced this one; leave it and its
		// readiness channel intact. Unreachable while reconnect
		// serialises the lifecycle, kept as defence in depth. A nil cs
		// means "end unconditionally", the shape the tests drive.
		return
	}

	c.cs = nil
	c.connected = make(chan struct{})
}

func (c *Client) setSession(cs *Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case cs == nil:
		return ErrNoSession
	case c.cs != nil:
		return ErrSessionAttached
	default:
		c.cs = cs
		close(c.connected)
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

// Shutdown gracefully stops the [Client]: it initiates shutdown and waits
// until the session workers drain, or ctx times out. It is the clean-stop
// counterpart to [Client.Connect] — the reconnect loop halts and no new
// session is dialled. After a caller-initiated Shutdown, [Client.Err]
// reports nil; a non-nil result means the drain overran ctx. Shutdown is
// promoted from the embedded [reconnect.WorkGroup]; this wrapper documents
// that contract for the client's lifecycle surface.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.WorkGroup.Shutdown(ctx)
}

// Connected returns a channel that is closed while the [Client] holds an
// active session. The channel is replaced after each disconnect, so it
// signals only the next readiness edge — callers waiting across a
// reconnect cycle must fetch a fresh channel via Connected.
func (c *Client) Connected() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.connected
}

// IsConnected reports whether the [Client] currently holds an active
// session. It is a point-in-time snapshot; the state can change between
// the call returning and the next operation.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cs != nil
}

// WaitConnected blocks until the [Client] is connected or ctx is done.
// It returns nil on connect, or ctx.Err() if ctx fires first. The wait
// is bounded by the caller's ctx — pass a deadline if you want to limit
// how long callers will tolerate reconnection.
func (c *Client) WaitConnected(ctx context.Context) error {
	ch := c.Connected()

	// Prefer a ready connection over an already-cancelled ctx: a lone
	// select with both cases ready would choose pseudo-randomly, so an
	// already-connected client could spuriously return ctx.Err().
	select {
	case <-ch:
		return nil
	default:
	}

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
