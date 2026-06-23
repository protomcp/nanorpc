// Package client provides a mock NanoRPC client for tests: it dials a server,
// sends requests, and reads the responses returned. Use it to drive a real
// server.Server over the wire.
package client

import (
	"io"
	"net"
	"time"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/mock/wire"
)

// dialTimeout bounds client.New's dial so an unreachable address fails the
// test instead of hanging it.
const dialTimeout = 2 * time.Second

// defaultTimeout bounds the blocking test helpers. It is a var so white-box
// tests can shrink it to exercise the timeout paths without a real wait.
var defaultTimeout = 2 * time.Second

// Client is a mock NanoRPC client over a single dialled connection.
type Client struct {
	t    core.T
	peer *wire.Peer[*nanorpc.NanoRPCRequest, *nanorpc.NanoRPCResponse]
}

// New dials addr and returns a mock [Client], registering its shutdown with
// t.Cleanup. It fails the test if the connection cannot be established.
func New(t core.T, addr string) *Client {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	core.AssertMustNoError(t, err, "dial")

	c := &Client{
		t: t,
		peer: wire.New(wire.Config[*nanorpc.NanoRPCRequest, *nanorpc.NanoRPCResponse]{
			Conn:      conn,
			Split:     nanorpc.Split,
			Encode:    encodeRequest,
			Decode:    decodeResponse,
			QueueSize: wire.DefaultQueueSize,
		}),
	}
	registerCleanup(t, func() { core.AssertNoError(t, c.Close(), "mock client") })

	return c
}

// Send encodes req and writes it to the server, failing the test on error.
func (c *Client) Send(req *nanorpc.NanoRPCRequest) {
	c.t.Helper()
	core.AssertMustNoError(c.t, c.peer.Send(req), "send request")
}

// Recv returns the next response from the server, failing the test if the
// connection closes or none arrives within the timeout.
func (c *Client) Recv() *nanorpc.NanoRPCResponse {
	c.t.Helper()
	select {
	case res, ok := <-c.peer.Recv():
		if !ok {
			c.t.Fatal("connection closed before a response arrived")
			return nil
		}
		return res
	case <-time.After(defaultTimeout):
		c.t.Fatal("timed out waiting for a response from the server")
		return nil
	}
}

// Close closes the connection to the server.
func (c *Client) Close() error {
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

func encodeRequest(w io.Writer, req *nanorpc.NanoRPCRequest) error {
	_, err := nanorpc.EncodeRequestTo(w, req, nil)
	return err
}

func decodeResponse(data []byte) (*nanorpc.NanoRPCResponse, error) {
	res, _, err := nanorpc.DecodeResponse(data)
	return res, err
}
