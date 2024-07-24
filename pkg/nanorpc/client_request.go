package nanorpc

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Ping sends a ping message to keep the connection alive.
// Ping returns false if the [Client] isn't connected.
func (c *Client) Ping() bool {
	// assemble header
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_PING,
	}

	_, err := c.enqueue(m, nil, nil)
	return err == nil
}

// Pong returns a channel that waits until a ping
// is answered.
// the channel returns nil on success or ErrPingTimeout
// if not connected or disconnected before answered.
func (c *Client) Pong() <-chan error {
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_PING,
	}

	// size 1 so we can write even if no-one is listening.
	ch := make(chan error, 1)

	// handler
	h := func(err error) {
		defer close(ch)
		ch <- err
	}

	// callback
	cb := func(_ context.Context, _ int32, pong *NanoRPCResponse) error {
		h(ResponseAsError(pong))
		return nil
	}

	_, err := c.enqueue(m, nil, cb)
	if err != nil {
		h(err)
	}

	return ch
}

// Request enqueues a NanoRPC request optionally converting path to path_hash
// if [ClientOptions].AlwaysHashPaths was set.
func (c *Client) Request(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   c.getPathOneOf(path),
	}

	return c.enqueue(m, msg, cb)
}

// RequestByHash enqueues a NanoRPC request using a given path_hash.
func (c *Client) RequestByHash(path uint32, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &NanoRPCRequest_PathHash{
			PathHash: path,
		},
	}

	return c.enqueue(m, msg, cb)
}

// RequestWithHash enqueues a NanoRPC request using the hash of the given path
func (c *Client) RequestWithHash(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	return c.RequestByHash(c.hc.Hash(path), msg, cb)
}

// Subscribe enqueues a NanoRPC subscription request
// optionally converting path to path_hash
// if [ClientOptions].AlwaysHashPaths was set.
func (c *Client) Subscribe(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   c.getPathOneOf(path),
	}

	return c.enqueue(m, msg, cb)
}

// SubscribeByHash enqueues a NanoRPC request using a given path_hash.
func (c *Client) SubscribeByHash(path uint32, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &NanoRPCRequest_PathHash{
			PathHash: path,
		},
	}

	return c.enqueue(m, msg, cb)
}

// SubscribeWithHash enqueues a NanoRPC request using the hash of the given path.
func (c *Client) SubscribeWithHash(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	return c.SubscribeByHash(c.hc.Hash(path), msg, cb)
}

func (c *Client) enqueue(m *NanoRPCRequest, msg proto.Message, cb RequestCallback) (int32, error) {
	cs, err := c.getSession()
	if err != nil {
		return 0, err
	}

	err = cs.Send(m, msg, cb)
	return m.RequestId, err
}
