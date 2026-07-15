// Package server provides a mock NanoRPC server for tests: a real client
// dials it (or it serves a supplied connection), it decodes the requests the
// client sends, and the test scripts the responses sent back.
package server

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/workgroup"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/mock/wire"
)

// acceptBacklog bounds the buffered channel of accepted connections awaiting
// an [Server.Accept] call.
const acceptBacklog = 4

// defaultTimeout bounds the blocking test helpers. It is a var so white-box
// tests can shrink it to exercise the timeout paths without a real wait.
var defaultTimeout = 2 * time.Second

// Server is a listening mock NanoRPC server. It accepts connections on a
// loopback address so a real client can dial [Server.Addr]; each accepted
// connection is surfaced as a [Conn] by [Server.Accept].
type Server struct {
	t        core.T
	ln       net.Listener
	closeErr error
	wg       *workgroup.Group
	conns    chan *Conn
	open     []*Conn

	mu        sync.Mutex
	closeOnce sync.Once
	closed    bool
}

// New starts a [Server] listening on a loopback port and registers its
// shutdown with t.Cleanup. It fails the test if the listener cannot be
// opened.
func New(t core.T) *Server {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")

	s := &Server{
		t:     t,
		ln:    ln,
		wg:    workgroup.New(context.Background()),
		conns: make(chan *Conn, acceptBacklog),
	}
	core.AssertMustNoError(t, s.wg.GoCatch(s.acceptLoop, nil), "accept loop")
	registerCleanup(t, func() { core.AssertNoError(t, s.Close(), "mock server") })

	return s
}

// Addr returns the address the server is listening on, suitable for a
// client's Remote configuration.
func (s *Server) Addr() string {
	return s.ln.Addr().String()
}

// Accept returns the next client connection, failing the test if none
// arrives within the timeout.
func (s *Server) Accept() *Conn {
	s.t.Helper()
	select {
	case c := <-s.conns:
		return c
	case <-time.After(defaultTimeout):
		s.t.Fatal("timed out waiting for a client connection")
		return nil
	}
}

// Close stops accepting, closes the listener and every connection it handed
// out, and waits for the accept loop to unwind. The resource cleanup runs
// exactly once via closeOnce and never gates on winning wg.Cancel, so a
// workgroup already cancelled by an accept fault still releases its listener
// and tracked connections. It is safe to call more than once.
func (s *Server) Close() error {
	s.wg.Cancel(nil)

	s.closeOnce.Do(s.cleanup)

	return new(errors.CompoundError).
		AppendError(s.wg.Wait(), s.closeErr).AsError()
}

// cleanup closes the listener and every handed-out connection, collecting
// their errors into s.closeErr. It runs once, under closeOnce.
func (s *Server) cleanup() {
	lnErr := s.ln.Close()

	s.mu.Lock()
	s.closed = true
	open := s.open
	s.open = nil
	s.mu.Unlock()

	// Close the handed-out connections concurrently and collect every
	// framing fault, so a client that sent a corrupt frame fails its test
	// rather than slipping through a discarded error.
	errs := new(errors.CompoundError)
	var wg sync.WaitGroup
	for _, c := range open {
		wg.Go(func() {
			_ = errs.AppendError(c.Close())
		})
	}
	wg.Wait()

	s.closeErr = errs.AppendError(lnErr).AsError()
}

// acceptLoop accepts connections until the workgroup is cancelled or the
// listener is closed. A failure after cancellation is the shutdown the test
// initiated and is reported as nil.
func (s *Server) acceptLoop(ctx context.Context) error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return wire.StopErr(ctx, err)
		}
		if !s.deliver(ctx, conn) {
			return nil
		}
	}
}

// deliver wraps conn as a tracked [Conn] and hands it to Accept, reporting
// false when the server is closing or the workgroup is cancelled before the
// hand-off completes.
func (s *Server) deliver(ctx context.Context, conn net.Conn) bool {
	c := newConn(s.t, conn)
	if !s.track(c) {
		// The server is closing and will not drain this conn; close it
		// here rather than leak its reader, and stop the accept loop.
		_ = c.Close()
		return false
	}
	select {
	case s.conns <- c:
		return true
	case <-ctx.Done():
		// Hand-off abandoned; track succeeded, so c is in the set Close
		// snapshots and Close owns its shutdown.
		return false
	}
}

// track records a connection so Close can shut it down, reporting false once
// the server is closing so the caller disposes of the connection itself.
func (s *Server) track(c *Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.open = append(s.open, c)
	return true
}

// Conn is one accepted client connection: it decodes the requests the client
// sends and scripts the responses sent back.
type Conn struct {
	t    core.T
	peer *wire.Peer[*nanorpc.NanoRPCResponse, *nanorpc.NanoRPCRequest]
}

// Serve wraps an existing connection as a server-side [Conn] without a
// listener, for tests that supply their own transport such as a net.Pipe.
// Its shutdown is registered with t.Cleanup.
func Serve(t core.T, conn net.Conn) *Conn {
	t.Helper()
	c := newConn(t, conn)
	registerCleanup(t, func() {
		core.AssertNoError(t, c.Close(), "mock server connection")
	})
	return c
}

func newConn(t core.T, conn net.Conn) *Conn {
	return &Conn{
		t: t,
		peer: wire.New(wire.Config[*nanorpc.NanoRPCResponse, *nanorpc.NanoRPCRequest]{
			Conn:      conn,
			Split:     nanorpc.Split,
			Encode:    encodeResponse,
			Decode:    decodeRequest,
			QueueSize: wire.DefaultQueueSize,
		}),
	}
}

// Recv returns the next request the client sent, failing the test if the
// connection closes or no request arrives within the timeout.
func (c *Conn) Recv() *nanorpc.NanoRPCRequest {
	c.t.Helper()
	select {
	case req, ok := <-c.peer.Recv():
		if !ok {
			c.t.Fatal("connection closed before a request arrived")
			return nil
		}
		return req
	case <-time.After(defaultTimeout):
		c.t.Fatal("timed out waiting for a request from the client")
		return nil
	}
}

// Reply sends a response to the client.
func (c *Conn) Reply(res *nanorpc.NanoRPCResponse) {
	c.t.Helper()
	core.AssertMustNoError(c.t, c.peer.Send(res), "send response")
}

// Close stops the connection's reader and closes the underlying transport.
func (c *Conn) Close() error {
	return c.peer.Close()
}

// registerCleanup runs fn at test end when t supports cleanup. A *testing.T
// does; a core.MockT (used in white-box failure tests) does not, leaving the
// test to close explicitly.
func registerCleanup(t core.T, fn func()) {
	if tc, ok := t.(interface{ Cleanup(func()) }); ok {
		tc.Cleanup(fn)
	}
}

func encodeResponse(w io.Writer, res *nanorpc.NanoRPCResponse) error {
	_, err := nanorpc.EncodeResponseTo(w, res, nil)
	return err
}

func decodeRequest(data []byte) (*nanorpc.NanoRPCRequest, error) {
	req, _, err := nanorpc.DecodeRequest(data)
	return req, err
}
