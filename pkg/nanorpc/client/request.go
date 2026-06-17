package client

import (
	"context"

	"darvaza.org/core"
	"google.golang.org/protobuf/proto"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/utils"
)

// SubscribeCallback is a function given to [Subscribe] to be called on every
// subscription event. The server's TYPE_RESPONSE acknowledgement surfaces
// with err == [nanorpc.ErrSubscriptionEstablished] on STATUS_OK or a
// [nanorpc.ResponseError] for any other status; TYPE_UPDATE deliveries pass
// the decoded payload with err == nil, or [nanorpc.ErrNoResponse] when the
// update carried no data. Decode and factory errors surface through err.
//
// res is a freshly allocated value from the newOut factory passed to
// [Subscribe]; it is never a typed nil and has no fields populated on the
// acknowledgement.
type SubscribeCallback[A proto.Message] func(ctx context.Context, id int32, res A, err error) error

// Requester is a view of the [Client] that only allows [Client.Request] calls
type Requester interface {
	Request(string, proto.Message, RequestCallback) (int32, error)
}

// Subscriber is a view of the [Client] that only allows [Client.Subscribe] calls
type Subscriber interface {
	Subscribe(string, proto.Message, RequestCallback) (int32, error)
}

// Unsubscriber is a view of the [Client] that only allows [Client.Unsubscribe] calls
type Unsubscriber interface {
	Unsubscribe(string, int32, RequestCallback) error
}

// Ping sends a ping message to keep the connection alive.
// Ping returns false if the [Client] isn't connected.
func (c *Client) Ping() bool {
	// assemble header
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	_, err := c.enqueue(m, nil, nil)
	return err == nil
}

// Pong returns a channel that waits until a ping
// is answered.
// the channel returns nil on success or ErrPingTimeout
// if not connected or disconnected before answered.
func (c *Client) Pong() <-chan error {
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	// size 1 so we can write even if no-one is listening.
	ch := make(chan error, 1)

	// handler
	h := func(err error) {
		defer close(ch)
		ch <- err
	}

	// callback
	cb := func(_ context.Context, _ int32, pong *nanorpc.NanoRPCResponse) error {
		h(nanorpc.ResponseAsError(pong))
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
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   c.getPathOneOf(path),
	}

	return c.enqueue(m, msg, cb)
}

// RequestByHash enqueues a NanoRPC request using a given path_hash.
func (c *Client) RequestByHash(path uint32, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &nanorpc.NanoRPCRequest_PathHash{
			PathHash: path,
		},
	}

	return c.enqueue(m, msg, cb)
}

// RequestWithHash enqueues a NanoRPC request using the hash of the given path
func (c *Client) RequestWithHash(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	hash, err := c.hc.Hash(path)
	if err != nil {
		// Fall back to string path on hash collision
		if logger, ok := c.getErrorLogger(err); ok {
			logger.WithField(utils.FieldPath, path).
				Print("Falling back to string path to maintain compatibility")
		}
		return c.Request(path, msg, cb)
	}
	return c.RequestByHash(hash, msg, cb)
}

// Subscribe enqueues a NanoRPC subscription request
// optionally converting path to path_hash
// if [ClientOptions].AlwaysHashPaths was set.
func (c *Client) Subscribe(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof:   c.getPathOneOf(path),
	}

	return c.enqueue(m, msg, cb)
}

// SubscribeByHash enqueues a NanoRPC request using a given path_hash.
func (c *Client) SubscribeByHash(path uint32, msg proto.Message, cb RequestCallback) (int32, error) {
	// assemble header
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof: &nanorpc.NanoRPCRequest_PathHash{
			PathHash: path,
		},
	}

	return c.enqueue(m, msg, cb)
}

// SubscribeWithHash enqueues a NanoRPC request using the hash of the given path.
func (c *Client) SubscribeWithHash(path string, msg proto.Message, cb RequestCallback) (int32, error) {
	hash, err := c.hc.Hash(path)
	if err != nil {
		// Fall back to string path on hash collision
		if logger, ok := c.getErrorLogger(err); ok {
			logger.WithField(utils.FieldPath, path).
				Print("Falling back to string path to maintain compatibility")
		}
		return c.Subscribe(path, msg, cb)
	}
	return c.SubscribeByHash(hash, msg, cb)
}

// Unsubscribe sends an unsubscribe request for the given path and request ID.
// According to the NanoRPC protocol, unsubscribing is done by sending a TYPE_REQUEST
// with empty data to the same path using the original subscription request ID.
// The path will be converted to path_hash if [ClientOptions].AlwaysHashPaths was set.
//
// requestID must be the value returned by a prior [Client.Subscribe] call.
// The call fails with [ErrNoSubscription] when none is registered for
// requestID, or [ErrSubscriptionPending] when it is not yet acknowledged.
// cb fires once when the server acknowledges the unsubscribe.
func (c *Client) Unsubscribe(path string, requestID int32, cb RequestCallback) error {
	// assemble header
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		RequestId:   requestID,
		PathOneof:   c.getPathOneOf(path),
	}

	_, err := c.enqueue(m, nil, cb)
	return err
}

// UnsubscribeByHash sends an unsubscribe request using a given path_hash and request ID.
// According to the NanoRPC protocol, unsubscribing is done by sending a TYPE_REQUEST
// with empty data to the path using the original subscription request ID.
//
// requestID must be the value returned by a prior [Client.SubscribeByHash]
// call. The call fails with [ErrNoSubscription] when none is registered for
// requestID, or [ErrSubscriptionPending] when it is not yet acknowledged.
// cb fires once when the server acknowledges the unsubscribe.
func (c *Client) UnsubscribeByHash(pathHash uint32, requestID int32, cb RequestCallback) error {
	// assemble header
	m := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		RequestId:   requestID,
		PathOneof: &nanorpc.NanoRPCRequest_PathHash{
			PathHash: pathHash,
		},
	}

	_, err := c.enqueue(m, nil, cb)
	return err
}

// UnsubscribeWithHash sends an unsubscribe request using the hash of the given path and request ID.
// According to the NanoRPC protocol, unsubscribing is done by sending a TYPE_REQUEST
// with empty data to the path using the original subscription request ID.
//
// requestID must be the value returned by a prior [Client.SubscribeWithHash]
// call. The call fails with [ErrNoSubscription] when none is registered for
// requestID, or [ErrSubscriptionPending] when it is not yet acknowledged.
// cb fires once when the server acknowledges the unsubscribe.
func (c *Client) UnsubscribeWithHash(path string, requestID int32, cb RequestCallback) error {
	hash, err := c.hc.Hash(path)
	if err != nil {
		// Fall back to string path on hash collision
		if logger, ok := c.getErrorLogger(err); ok {
			logger.WithField(utils.FieldPath, path).
				Print("Falling back to string path to maintain compatibility")
		}
		return c.Unsubscribe(path, requestID, cb)
	}
	return c.UnsubscribeByHash(hash, requestID, cb)
}

func (c *Client) enqueue(m *nanorpc.NanoRPCRequest, msg proto.Message, cb RequestCallback) (int32, error) {
	cs, err := c.getSession()
	if err != nil {
		return 0, err
	}

	err = cs.Send(m, msg, cb)
	return m.RequestId, err
}

// GetResponse makes a [Client.Request] and waits for the response.
func GetResponse[Q, A proto.Message](ctx context.Context, c Requester, path string, req Q, out A) error {
	if core.IsNil(c) {
		return ErrMissingClient
	}
	if core.IsNil(out) {
		return ErrMissingOut
	}

	ch, cb := newGetResponseCallback(out)
	if _, err := c.Request(path, req, cb); err != nil {
		return err
	}
	return waitGetResponse(ctx, ch)
}

// waitGetResponse blocks until the response arrives on ch or ctx cancels,
// returning whichever error landed first.
func waitGetResponse(ctx context.Context, ch <-chan error) error {
	select {
	case e, ok := <-ch:
		if !ok {
			return nil
		}
		return e
	case <-ctx.Done():
		return ctx.Err()
	}
}

func newGetResponseCallback(out proto.Message) (<-chan error, RequestCallback) {
	ch := make(chan error, 1)
	cb := func(_ context.Context, _ int32, res *nanorpc.NanoRPCResponse) error {
		defer close(ch)

		_, present, err := nanorpc.DecodeResponseData(res, out)
		if err == nil && !present {
			ch <- nanorpc.ErrNoResponse
		} else {
			ch <- err
		}

		return nil
	}
	return ch, cb
}

// Subscribe makes a subscription request and registers cb to be invoked on
// every subscription event. The newOut factory allocates a fresh message
// container for each callback invocation; both cb and newOut are required.
func Subscribe[Q, A proto.Message](c Subscriber, path string,
	req Q, cb SubscribeCallback[A], newOut func() (A, error)) (int32, error) {
	//
	if cb == nil {
		return 0, ErrMissingCallback
	}
	if newOut == nil {
		return 0, ErrMissingNewOut
	}

	return c.Subscribe(path, req, newSubscribeCallback(cb, newOut))
}

func newSubscribeCallback[A proto.Message](cb SubscribeCallback[A], newOut func() (A, error)) RequestCallback {
	return func(ctx context.Context, id int32, res *nanorpc.NanoRPCResponse) error {
		return dispatchSubscribeResponse(ctx, id, res, cb, newOut)
	}
}

// dispatchSubscribeResponse routes a single response to the typed callback:
// the TYPE_RESPONSE acknowledgement surfaces as a fresh out plus either the
// status's [nanorpc.ResponseError] or [nanorpc.ErrSubscriptionEstablished];
// TYPE_UPDATE payloads are decoded into a fresh out.
func dispatchSubscribeResponse[A proto.Message](ctx context.Context, id int32,
	res *nanorpc.NanoRPCResponse, cb SubscribeCallback[A],
	newOut func() (A, error)) error {
	if isSubscribeACK(res) {
		out, err := callNewOut(newOut)
		if err != nil {
			return cb(ctx, id, out, err)
		}
		return cb(ctx, id, out, subscribeACKErr(res))
	}
	out, err := decodeSubscribePayload(res, newOut)
	return cb(ctx, id, out, err)
}

// callNewOut allocates a fresh out message via the caller-supplied factory,
// normalising a misbehaving factory into an error so callers need no nil
// check of their own: when it returns a nil error, out is guaranteed to be a
// usable, non-nil message; on any error out must be ignored. A factory error
// passes through unchanged, while a typed-nil result with no error is promoted
// to [ErrNilOut] so it cannot masquerade as a valid message.
func callNewOut[A proto.Message](newOut func() (A, error)) (A, error) {
	out, err := newOut()
	switch {
	case err != nil:
		return out, err
	case core.IsNil(out):
		return out, ErrNilOut
	default:
		return out, nil
	}
}

// isSubscribeACK reports whether res is the TYPE_RESPONSE acknowledgement
// that follows a TYPE_SUBSCRIBE — as opposed to a TYPE_UPDATE delivery.
func isSubscribeACK(res *nanorpc.NanoRPCResponse) bool {
	return res != nil &&
		res.ResponseType == nanorpc.NanoRPCResponse_TYPE_RESPONSE
}

// subscribeACKErr maps the acknowledgement status to the error surfaced
// through the subscription callback: a [nanorpc.ResponseError] for any
// non-OK status, or the [nanorpc.ErrSubscriptionEstablished] sentinel on
// success.
func subscribeACKErr(res *nanorpc.NanoRPCResponse) error {
	if err := nanorpc.ResponseAsError(res); err != nil {
		return err
	}
	return nanorpc.ErrSubscriptionEstablished
}

// decodeSubscribePayload allocates a fresh out via [callNewOut] and decodes
// res into it, returning [nanorpc.ErrNoResponse] when the message carried
// no data.
func decodeSubscribePayload[A proto.Message](res *nanorpc.NanoRPCResponse,
	newOut func() (A, error)) (A, error) {
	out, err := callNewOut(newOut)
	if err != nil {
		return out, err
	}
	_, present, err := nanorpc.DecodeResponseData(res, out)
	if err == nil && !present {
		err = nanorpc.ErrNoResponse
	}
	return out, err
}
