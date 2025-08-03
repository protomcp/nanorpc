package client

import (
	"context"
	"io"
	"net"
	"os"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/net/reconnect"
	"google.golang.org/protobuf/proto"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/common"
)

// clientRequest combines in a single object the fields needed for
// Encoding a nanorpc.NanoRPCRequest when using [reconnect.StreamSession]
type clientRequest struct {
	r *nanorpc.NanoRPCRequest
	d proto.Message
}

type clientRequestQueue struct {
	Callback    RequestCallback
	RequestType nanorpc.NanoRPCRequest_Type
	RequestID   int32
}

// Session represents a connection to a NanoRPC server.
type Session struct {
	reconnect.WorkGroup

	c      *Client
	rc     *reconnect.Client
	ra     net.Addr
	ss     *reconnect.StreamSession[*nanorpc.NanoRPCResponse, clientRequest]
	logger slog.Logger

	cb []clientRequestQueue
	mu sync.Mutex
}

// Spawn starts the required workers to handle the session
func (cs *Session) Spawn() error {
	if err := cs.ss.Spawn(); err != nil {
		return err
	}

	cs.ss.Go(cs.run)
	return nil
}

func (cs *Session) run(ctx context.Context) error {
	for {
		if err := cs.doRunPass(ctx); err != nil {
			return err
		}
	}
}

func (cs *Session) doRunPass(ctx context.Context) error {
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

// IsActive indicates the session has registered callbacks
func (cs *Session) IsActive() bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return len(cs.cb) > 0
}

func (cs *Session) handleResponse(resp *nanorpc.NanoRPCResponse) error {
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
func (cs *Session) popRequestCallback(reqID int32) RequestCallback {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	i, found := cs.unsafeIndexRequestCallback(reqID)
	if !found {
		return nil
	}

	x := cs.cb[i]
	if x.RequestType != nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE {
		// remove unless it's a subscription
		cs.cb = append(cs.cb[:i], cs.cb[i+1:]...)
	}

	return x.Callback
}

// Close clears out any outstanding request callback
func (cs *Session) Close() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, x := range cs.cb {
		cs.unsafePopOne(x)
	}

	cs.cb = cs.cb[:0]
	return nil
}

func (cs *Session) unsafePopOne(x clientRequestQueue) {
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
func (cs *Session) Send(req *nanorpc.NanoRPCRequest, payload proto.Message, cb RequestCallback) error {
	switch req.RequestType {
	case nanorpc.NanoRPCRequest_TYPE_PING:
		// no further checks
	case nanorpc.NanoRPCRequest_TYPE_REQUEST, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE:
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

func (cs *Session) nextRequestID() int32 {
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
func (cs *Session) unsafeIndexRequestCallback(reqID int32) (int, bool) {
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

func (cs *Session) onSetReadDeadline() error {
	if cs.IsActive() {
		return cs.rc.ResetReadDeadline()
	}

	d := cs.c.idleReadTimeout
	return cs.rc.SetReadDeadline(d)
}

func (cs *Session) onSetWriteDeadline() error {
	return cs.rc.ResetDeadline()
}

func (cs *Session) onUnsetReadDeadline() error {
	return cs.rc.SetReadDeadline(0)
}

func (cs *Session) onUnsetWriteDeadline() error {
	return cs.rc.SetWriteDeadline(0)
}

func (cs *Session) onError(err error) {
	cs.LogError(err, nil, "error")
}

//
// factory
//

func newClientSession(ctx context.Context, c *Client, queueSize uint, conn net.Conn) *Session {
	ss := &reconnect.StreamSession[*nanorpc.NanoRPCResponse, clientRequest]{
		QueueSize: queueSize,
		Conn:      c.rc,
		Context:   ctx,

		Split: nanorpc.Split,
		MarshalTo: func(r clientRequest, w io.Writer) error {
			_, err := nanorpc.EncodeRequestTo(w, r.r, r.d)
			return err
		},
		Unmarshal: func(data []byte) (*nanorpc.NanoRPCResponse, error) {
			resp, _, err := nanorpc.DecodeResponse(data)
			return resp, err
		},
	}

	// Create session logger with fields added once
	sessionLogger := common.WithComponent(c.getLogger(), common.ComponentSession)
	sessionLogger = common.WithRemoteAddr(sessionLogger, conn.RemoteAddr())

	cs := &Session{
		c:      c,
		rc:     c.rc,
		ra:     conn.RemoteAddr(),
		logger: sessionLogger,

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
