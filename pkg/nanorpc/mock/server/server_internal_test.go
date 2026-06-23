package server

import (
	"net"
	"testing"
	"time"

	"darvaza.org/core"
)

// TestConn_Recv_closed covers the branch where the connection closes before a
// request arrives: Recv reports it as a fatal failure.
func TestConn_Recv_closed(t *testing.T) {
	mt := &core.MockT{}
	a, b := net.Pipe()
	c := newConn(mt, b)

	// Close the peer end so the reader stops and Recv's channel closes.
	core.AssertMustNoError(t, a.Close(), "close peer end")

	ok := mt.Run("recv on closed conn", func(core.T) { c.Recv() })
	core.AssertFalse(t, ok, "Recv must fail when the conn closes")
	core.AssertTrue(t, mt.HasErrors(), "fatal recorded")
	core.AssertNoError(t, c.Close(), "close conn")
}

// TestConn_Recv_timeout covers the branch where no request arrives within the
// timeout. defaultTimeout is shrunk so the wait is brief.
func TestConn_Recv_timeout(t *testing.T) {
	old := defaultTimeout
	defaultTimeout = 10 * time.Millisecond
	defer func() { defaultTimeout = old }()

	mt := &core.MockT{}
	a, b := net.Pipe()
	defer func() { _ = a.Close() }()
	c := newConn(mt, b)

	ok := mt.Run("recv timeout", func(core.T) { c.Recv() })
	core.AssertFalse(t, ok, "Recv must time out")
	core.AssertTrue(t, mt.HasErrors(), "fatal recorded")
	core.AssertNoError(t, c.Close(), "close conn")
}

// TestServer_Accept_timeout covers Accept's branch where no client connects
// within the timeout.
func TestServer_Accept_timeout(t *testing.T) {
	old := defaultTimeout
	defaultTimeout = 10 * time.Millisecond
	defer func() { defaultTimeout = old }()

	mt := &core.MockT{}
	srv := New(mt)

	ok := mt.Run("accept timeout", func(core.T) { srv.Accept() })
	core.AssertFalse(t, ok, "Accept must time out")
	core.AssertTrue(t, mt.HasErrors(), "fatal recorded")
	core.AssertNoError(t, srv.Close(), "close server")
}

// TestServer_deliverAfterClose pins the shutdown-leak fix: once the server is
// closing, deliver must refuse a freshly accepted connection and close it
// rather than leak its reader.
func TestServer_deliverAfterClose(t *testing.T) {
	srv := New(t)
	core.AssertMustNoError(t, srv.Close(), "close server")

	a, b := net.Pipe()
	defer func() { _ = a.Close() }()

	delivered := srv.deliver(srv.wg.Context(), b)
	core.AssertFalse(t, delivered, "deliver refused after close")

	// b is closed by deliver, so a write to it now fails.
	_, err := b.Write([]byte("x"))
	core.AssertError(t, err, "conn closed by deliver")
}
