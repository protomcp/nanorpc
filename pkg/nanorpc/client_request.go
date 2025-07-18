package nanorpc

import (
	"context"

	"darvaza.org/core"

	"google.golang.org/protobuf/proto"
)

// SubscribeCallback is a function given to [Subscribe] to be called on every update
type SubscribeCallback[A proto.Message] func(ctx context.Context, id int32, res A, err error) error

// Requester is a view of the [Client] that only allows [Client.Request] calls
type Requester interface {
	Request(string, proto.Message, RequestCallback) (int32, error)
}

// Subscriber is a view of the [Client] that only allows [Client.Subscribe] calls
type Subscriber interface {
	Subscribe(string, proto.Message, RequestCallback) (int32, error)
}

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
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof:   c.getPathOneOf(path),
	}

	return c.enqueue(m, msg, cb)
}

// SubscribeByHash enqueues a NanoRPC request using a given path_hash.
func (c *Client) SubscribeByHash(path uint32, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
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

// GetResponse makes a [Client.Request] and waits for the response.
func GetResponse[Q, A proto.Message](ctx context.Context, c Requester, path string, req Q, out A) error {
	ch, cb := newGetResponseCallback(out)
	_, err := c.Request(path, req, cb)
	if err == nil {
		select {
		case e, ok := <-ch:
			if ok {
				err = e
			}
		case <-ctx.Done():
			err = ctx.Err()
		}
	}

	return err
}

func newGetResponseCallback(out proto.Message) (<-chan error, RequestCallback) {
	ch := make(chan error, 1)
	cb := func(_ context.Context, _ int32, res *NanoRPCResponse) error {
		defer close(ch)

		_, present, err := DecodeResponseData(res, out)
		if err == nil && !present {
			ch <- ErrNoResponse
		} else {
			ch <- err
		}

		return nil
	}
	return ch, cb
}

// Subscribe makes a subscription request and registers the given callback
// to be invoked on every update.
func Subscribe[Q, A proto.Message](c Subscriber, path string,
	req Q, cb SubscribeCallback[A], newOut func() (A, error)) (int32, error) {
	//
	if cb == nil {
		return 0, core.Wrap(core.ErrInvalid, "callback missing")
	}

	return c.Subscribe(path, req, newSubscribeCallback(cb, newOut))
}

func newSubscribeCallback[A proto.Message](cb SubscribeCallback[A], newOut func() (A, error)) RequestCallback {
	if newOut == nil {
		newOut = func() (A, error) {
			var zero A
			return zero, nil
		}
	}

	fn := func(ctx context.Context, id int32, res *NanoRPCResponse) error {
		out, err := newOut()
		if err != nil {
			return err
		}

		_, present, err := DecodeResponseData(res, out)
		if err == nil && !present {
			err = ErrNoResponse
		}

		return cb(ctx, id, out, err)
	}

	return fn
}
