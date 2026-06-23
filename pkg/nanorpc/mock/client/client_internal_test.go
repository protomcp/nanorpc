package client

import (
	"net"
	"testing"
	"time"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// listen opens a loopback listener and registers its closure.
func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	t.Cleanup(func() { _ = ln.Close() })
	return ln
}

// TestClient_New_dialFailure covers New's branch where the dial fails: it
// reports a fatal failure rather than returning a broken client.
func TestClient_New_dialFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	core.AssertMustNoError(t, err, "listen")
	addr := ln.Addr().String()
	// Close the listener so the address refuses connections.
	core.AssertMustNoError(t, ln.Close(), "close listener")

	mt := &core.MockT{}
	ok := mt.Run("dial failure", func(tt core.T) { New(tt, addr) })
	core.AssertFalse(t, ok, "New must fail to dial")
	core.AssertTrue(t, mt.HasErrors(), "fatal recorded")
}

// TestClient_Recv_closed exercises Send on the request path, then covers
// Recv's branch where the server hangs up before a response arrives.
func TestClient_Recv_closed(t *testing.T) {
	ln := listen(t)

	mt := &core.MockT{}
	cli := New(mt, ln.Addr().String())

	srvConn, err := ln.Accept()
	core.AssertMustNoError(t, err, "accept")

	cli.Send(&nanorpc.NanoRPCRequest{
		RequestId:   1,
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	})
	core.AssertMustNoError(t, srvConn.Close(), "close server side")

	ok := mt.Run("recv closed", func(core.T) { cli.Recv() })
	core.AssertFalse(t, ok, "Recv must fail when the server hangs up")
	core.AssertTrue(t, mt.HasErrors(), "fatal recorded")
	core.AssertNoError(t, cli.Close(), "close client")
}

// TestClient_Recv_timeout covers Recv's branch where no response arrives
// within the timeout. defaultTimeout is shrunk so the wait is brief.
func TestClient_Recv_timeout(t *testing.T) {
	old := defaultTimeout
	defaultTimeout = 10 * time.Millisecond
	defer func() { defaultTimeout = old }()

	ln := listen(t)

	mt := &core.MockT{}
	cli := New(mt, ln.Addr().String())

	srvConn, err := ln.Accept()
	core.AssertMustNoError(t, err, "accept")
	defer func() { _ = srvConn.Close() }()

	ok := mt.Run("recv timeout", func(core.T) { cli.Recv() })
	core.AssertFalse(t, ok, "Recv must time out")
	core.AssertTrue(t, mt.HasErrors(), "fatal recorded")
	core.AssertNoError(t, cli.Close(), "close client")
}
