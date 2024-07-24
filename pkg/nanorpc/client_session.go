package nanorpc

import (
	"context"
	"net"
	"os"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/net/reconnect"

	"google.golang.org/protobuf/proto"
)

// clientRequest combines in a single object the fields needed for
// Encoding a NanoRPCRequest when using [reconnect.StreamSession]
type clientRequest struct {
	r *NanoRPCRequest
	d proto.Message
}

type clientRequestQueue struct {
	RequestID   int32
	RequestType NanoRPCRequest_Type
	Callback    RequestCallback
}

// ClientSession represents a connection to a NanoRPC server.
type ClientSession struct {
	reconnect.WorkGroup

	mu sync.Mutex
	c  *Client
	rc *reconnect.Client
	ra net.Addr

	ss *reconnect.StreamSession[*NanoRPCResponse, clientRequest]

	cb []clientRequestQueue
}

// Spawn starts the required workers to handle the session
func (cs *ClientSession) Spawn() error {
	if err := cs.ss.Spawn(); err != nil {
		return err
	}

	cs.ss.Go(cs.run)
	return nil
}

func (cs *ClientSession) run(ctx context.Context) error {
	for {
		if err := cs.doRunPass(ctx); err != nil {
			return err
		}
	}
}

func (cs *ClientSession) doRunPass(ctx context.Context) error {
	select {
	case resp := <-cs.ss.Recv():
		return cs.handleResponse(resp)
	case <-ctx.Done():
		return ctx.Err()
	}
}

//
// workflow
//

func (cs *ClientSession) handleResponse(resp *NanoRPCResponse) error {
	if resp != nil && resp.RequestId > 0 {
		reqID := resp.RequestId

		if cb := cs.popRequestCallback(reqID); cb != nil {
			// report
			cs.ss.Go(func(ctx context.Context) error {
				return cb(ctx, reqID, resp)
			})
		}
	}

	return nil
}

// popRequestCallback searches the queue for a request ID and returns
// the callback if registered. unless the message is the response to a subscription,
// the callback will be removed.
func (cs *ClientSession) popRequestCallback(reqID int32) RequestCallback {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	i, found := cs.unsafeIndexRequestCallback(reqID)
	if !found {
		return nil
	}

	x := cs.cb[i]
	if x.RequestType != NanoRPCRequest_TYPE_SUBSCRIBE {
		// remove unless it's a subscription
		cs.cb = append(cs.cb[:i], cs.cb[i+1:]...)
	}

	return x.Callback
}

// Close clears out any outstanding request callback
func (cs *ClientSession) Close() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, x := range cs.cb {
		cs.unsafePopOne(x)
	}

	cs.cb = cs.cb[:0]
	return nil
}
func (cs *ClientSession) unsafePopOne(x clientRequestQueue) {
	cb := x.Callback
	reqID := x.RequestID

	// report termination
	cs.ss.Go(func(ctx context.Context) error {
		_ = cb(ctx, reqID, nil)
		return nil
	})
}

// Send stores the optional callback and enqueues the request.
// A zero RequestID will be replaced by a unique sequence number.
// A negative RequestID will become zero.
func (cs *ClientSession) Send(req *NanoRPCRequest, payload proto.Message, cb RequestCallback) error {
	switch req.RequestType {
	case NanoRPCRequest_TYPE_PING:
		// no further checks
	case NanoRPCRequest_TYPE_REQUEST, NanoRPCRequest_TYPE_SUBSCRIBE:
		// callback required
		if cb == nil {
			return core.QuietWrap(os.ErrInvalid, "missing callback")
		}
	default:
		// invalid type
		return core.QuietWrap(os.ErrInvalid, "%v: invalid request type", int(req.RequestType))
	}

	switch {
	case req.RequestId < 0:
		req.RequestId = 0
	case req.RequestId == 0:
		req.RequestId = cs.nextRequestID()
	}

	if cb != nil {
		// remember callback
		x := clientRequestQueue{
			RequestID:   req.RequestId,
			RequestType: req.RequestType,
			Callback:    cb,
		}

		cs.mu.Lock()
		cs.cb = append(cs.cb, x)
		cs.mu.Unlock()
	}

	return cs.ss.Send(clientRequest{req, payload})
}

func (cs *ClientSession) nextRequestID() int32 {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for {
		next := cs.c.reqCounter.Next()

		_, dupe := cs.unsafeIndexRequestCallback(next)
		if !dupe {
			return next
		}
	}
}

// unsafeIndexRequestCallback finds the position of the callback for a RequestID.
func (cs *ClientSession) unsafeIndexRequestCallback(reqID int32) (int, bool) {
	for i, x := range cs.cb {
		if x.RequestID == reqID {
			return i, true
		}
	}

	return -1, false
}

//
// StreamSession callbacks
//

func (cs *ClientSession) onSetReadDeadline() error {
	return cs.rc.ResetReadDeadline()
}

func (cs *ClientSession) onSetWriteDeadline() error {
	return cs.rc.ResetWriteDeadline()
}

func (cs *ClientSession) onUnsetReadDeadline() error {
	return cs.rc.SetReadDeadline(0)
}

func (cs *ClientSession) onUnsetWriteDeadline() error {
	return cs.rc.SetReadDeadline(0)
}

func (cs *ClientSession) onError(err error) {
	cs.LogError(err, "error")
}

//
// factory
//

func newClientSession(ctx context.Context, c *Client, queueSize uint, conn net.Conn) *ClientSession {
	ss := &reconnect.StreamSession[*NanoRPCResponse, clientRequest]{
		QueueSize: queueSize,
		Conn:      c.rc,
		Context:   ctx,

		Split: Split,
		Marshal: func(r clientRequest) ([]byte, error) {
			return EncodeRequest(r.r, r.d)
		},
		Unmarshal: func(data []byte) (*NanoRPCResponse, error) {
			resp, _, err := DecodeResponse(data)
			return resp, err
		},
	}

	cs := &ClientSession{
		c:  c,
		rc: c.rc,
		ra: conn.RemoteAddr(),

		ss:        ss,
		WorkGroup: ss,
	}

	cs.ss.SetReadDeadline = cs.onSetReadDeadline
	cs.ss.SetWriteDeadline = cs.onSetWriteDeadline
	cs.ss.UnsetReadDeadline = cs.onUnsetReadDeadline
	cs.ss.UnsetWriteDeadline = cs.onUnsetWriteDeadline
	cs.ss.OnError = cs.onError
	return cs
}
