package client

import (
	"context"
	"io"
	"net"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/net/reconnect"
	"google.golang.org/protobuf/proto"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/utils"
)

// clientRequest combines in a single object the fields needed for
// Encoding a nanorpc.NanoRPCRequest when using [reconnect.StreamSession]
type clientRequest struct {
	r *nanorpc.NanoRPCRequest
	d proto.Message
}

type clientRequestQueue struct {
	Callback     RequestCallback
	RequestType  nanorpc.NanoRPCRequest_Type
	RequestID    int32
	Acknowledged bool
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

		if cb := cs.popRequestCallback(resp); cb != nil {
			// report
			cs.ss.Go(func(ctx context.Context) error {
				return cb(ctx, reqID, resp)
			})
		}
	}

	return nil
}

// popRequestCallback locates the callback for an incoming response. A
// TYPE_UPDATE always routes to the SUBSCRIBE entry and leaves it queued.
// Any other response prefers a non-SUBSCRIBE entry (plain request, ping,
// or unsubscribe acknowledgement) and removes it; when that entry was
// shadowing a SUBSCRIBE entry for the same request ID, the SUBSCRIBE
// entry is removed too because the subscription is terminated. A
// non-update response that resolves only a SUBSCRIBE entry is handled by
// unsafeResolveSubscribeResponse against the subscription lifecycle.
func (cs *Session) popRequestCallback(resp *nanorpc.NanoRPCResponse) RequestCallback {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	subIdx, otherIdx := cs.unsafeIndexCallbacks(resp.RequestId)

	if resp.ResponseType == nanorpc.NanoRPCResponse_TYPE_UPDATE {
		if subIdx < 0 {
			return nil
		}
		return cs.cb[subIdx].Callback
	}

	if otherIdx >= 0 {
		cb := cs.cb[otherIdx].Callback
		cs.unsafeRemoveResolved(subIdx, otherIdx)
		return cb
	}

	if subIdx < 0 {
		return nil
	}
	return cs.unsafeResolveSubscribeResponse(subIdx, resp)
}

// unsafeResolveSubscribeResponse handles a non-update response that matched
// only a SUBSCRIBE entry, following the subscription lifecycle in
// NANORPC_PROTOCOL.md §6.1. A Pending subscription takes exactly one
// acknowledgement: STATUS_OK moves it to Active (Acknowledged flips, the
// entry stays queued for updates, and the callback fires the
// establishment), while a non-OK ack moves it to Terminated (the entry is
// dropped — otherwise a phantom would linger and wrongly satisfy the
// unsubscribe guard in checkUnsubscribeTarget — and the callback fires the
// failure). An Active subscription has no acknowledgement left to take: the
// lifecycle leaves Active only via Unsubscribe() or session end, never via
// a server response, so any further TYPE_RESPONSE is anomalous and is
// ignored to keep the live subscription intact rather than tearing it down
// or re-firing. cs.mu must be held.
func (cs *Session) unsafeResolveSubscribeResponse(subIdx int,
	resp *nanorpc.NanoRPCResponse) RequestCallback {
	if cs.cb[subIdx].Acknowledged {
		// already Active: no valid transition on a server response
		return nil
	}

	cb := cs.cb[subIdx].Callback
	if nanorpc.ResponseAsError(resp) != nil {
		// Pending -> Terminated: subscribe rejected, drop the entry
		cs.cb = append(cs.cb[:subIdx], cs.cb[subIdx+1:]...)
		return cb
	}

	// Pending -> Active: subscription established
	cs.cb[subIdx].Acknowledged = true
	return cb
}

// Close reports session termination to every outstanding callback and
// clears the queue. Each callback fires once with a nil response, which
// surfaces as [nanorpc.ErrNoResponse], so waiters in [GetResponse] and
// [Client.Pong] unblock at disconnect instead of lingering — Pong has no
// context of its own to fall back on.
//
// Terminations are delivered synchronously with ctx. Close runs as
// onReconnectSession unwinds, once the session workgroup has already
// stopped, where scheduling through cs.ss.Go would be a silent no-op.
// Callbacks fire without cs.mu held, so a callback may re-enter the session
// safely.
func (cs *Session) Close(ctx context.Context) error {
	cs.mu.Lock()
	pending := cs.cb
	cs.cb = nil
	cs.mu.Unlock()

	for _, x := range pending {
		_ = x.Callback(ctx, x.RequestID, nil)
	}
	return nil
}

// Send enqueues the request and registers cb when non-nil.
// A zero RequestID is replaced by a unique sequence number; a
// negative RequestID becomes zero; a positive RequestID is
// preserved.
//
// A nil req is rejected with [ErrNilRequest]. TYPE_REQUEST and
// TYPE_SUBSCRIBE require a non-nil cb, else [ErrMissingCallback];
// TYPE_PING does not; other request types yield [ErrInvalidRequestType].
//
// A TYPE_REQUEST carrying a positive RequestID is the unsubscribe
// form (see [Client.Unsubscribe]); Send rejects it with
// [ErrNoSubscription] when no subscription matches the RequestID, or
// [ErrSubscriptionPending] when the subscription is not yet
// acknowledged.
func (cs *Session) Send(req *nanorpc.NanoRPCRequest, payload proto.Message, cb RequestCallback) error {
	if err := validateSendArgs(req, cb); err != nil {
		return err
	}

	if isUnsubscribeShape(req) {
		if err := cs.checkUnsubscribeTarget(req.RequestId); err != nil {
			return err
		}
	}

	cs.normaliseRequestID(req)

	if cb != nil {
		// remember callback
		cs.registerCallback(clientRequestQueue{
			RequestID:   req.RequestId,
			RequestType: req.RequestType,
			Callback:    cb,
		})
	}

	return cs.ss.Send(clientRequest{req, payload})
}

// validateSendArgs rejects a Send call whose request is nil, whose type
// is unknown, or whose callback is missing on the types that need one. It
// runs first in Send so every later step (isUnsubscribeShape,
// normaliseRequestID, registerCallback) can dereference req unguarded.
func validateSendArgs(req *nanorpc.NanoRPCRequest, cb RequestCallback) error {
	if req == nil {
		return ErrNilRequest
	}

	switch req.RequestType {
	case nanorpc.NanoRPCRequest_TYPE_PING:
		// no further checks
		return nil
	case nanorpc.NanoRPCRequest_TYPE_REQUEST, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE:
		// callback required
		if cb == nil {
			return ErrMissingCallback
		}
		return nil
	default:
		// invalid type
		return core.QuietWrap(ErrInvalidRequestType, "%v", int(req.RequestType))
	}
}

// isUnsubscribeShape reports whether req is the unsubscribe form from the
// protocol: a TYPE_REQUEST whose caller already chose a positive RequestId
// to reuse an existing subscription's id.
func isUnsubscribeShape(req *nanorpc.NanoRPCRequest) bool {
	return req.RequestType == nanorpc.NanoRPCRequest_TYPE_REQUEST &&
		req.RequestId > 0
}

// normaliseRequestID assigns a fresh id when the caller supplied zero and
// folds a negative id to zero so the server sees no client-side sentinel.
func (cs *Session) normaliseRequestID(req *nanorpc.NanoRPCRequest) {
	switch {
	case req.RequestId < 0:
		req.RequestId = 0
	case req.RequestId == 0:
		req.RequestId = cs.nextRequestID()
	default:
		// keep caller-supplied positive RequestId
	}
}

// checkUnsubscribeTarget verifies that an unsubscribe targets an
// acknowledged subscription before its callback is registered. The check
// and the later registerCallback do not share a single lock hold; a
// subsequent removal between this check and registerCallback only causes
// the unsubscribe ack to fire the registered callback against an
// already-empty queue, which is the same observable outcome as a clean
// unsubscribe.
func (cs *Session) checkUnsubscribeTarget(reqID int32) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	subIdx, _ := cs.unsafeIndexCallbacks(reqID)
	if subIdx < 0 {
		return core.QuietWrap(ErrNoSubscription, "request_id %d", reqID)
	}
	if !cs.cb[subIdx].Acknowledged {
		return core.QuietWrap(ErrSubscriptionPending, "request_id %d", reqID)
	}
	return nil
}

// registerCallback appends a queue entry under cs.mu.
func (cs *Session) registerCallback(x clientRequestQueue) {
	cs.mu.Lock()
	cs.cb = append(cs.cb, x)
	cs.mu.Unlock()
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

// unsafeIndexCallbacks splits the matches for a RequestID by request type:
// subIdx is the first SUBSCRIBE entry, otherIdx is the first non-SUBSCRIBE
// entry, and either is -1 when no such match exists. cs.mu must be held.
func (cs *Session) unsafeIndexCallbacks(reqID int32) (subIdx, otherIdx int) {
	subIdx, otherIdx = -1, -1
	for i, x := range cs.cb {
		if x.RequestID != reqID {
			continue
		}
		subIdx, otherIdx = selectMatchSlot(subIdx, otherIdx, i, x.RequestType)
	}
	return subIdx, otherIdx
}

// selectMatchSlot stores the index of a matched queue entry in the right
// slot for its request type, leaving an already-filled slot untouched so
// the first match wins. Two SUBSCRIBE entries never share a request id
// (ids are unique per session), so the already-filled SUBSCRIBE arm is
// defensive; a repeated non-SUBSCRIBE id is reachable through concurrent
// unsubscribes reusing the same subscription id.
func selectMatchSlot(subIdx, otherIdx, i int,
	reqType nanorpc.NanoRPCRequest_Type) (nextSub, nextOther int) {
	if reqType == nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE {
		if subIdx < 0 {
			return i, otherIdx
		}
		return subIdx, otherIdx
	}
	if otherIdx < 0 {
		return subIdx, i
	}
	return subIdx, otherIdx
}

// unsafeRemoveResolved drops the entries resolved by a non-update response:
// the matched non-SUBSCRIBE entry at otherIdx, and the SUBSCRIBE entry it
// shadowed at subIdx when present (subIdx == -1 skips it). Removing the
// higher index first keeps the lower index valid. cs.mu must be held.
//
// In practice subIdx < otherIdx, since a subscription is always registered
// before the unsubscribe that shadows it; the swap only tolerates the
// reverse order defensively.
func (cs *Session) unsafeRemoveResolved(subIdx, otherIdx int) {
	high, low := otherIdx, subIdx
	if low > high {
		high, low = low, high
	}
	cs.cb = append(cs.cb[:high], cs.cb[high+1:]...)
	if low >= 0 {
		cs.cb = append(cs.cb[:low], cs.cb[low+1:]...)
	}
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
	cs.LogError(err, nil, "session run loop: %v", err)
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
	sessionLogger := utils.WithComponent(c.getLogger(), utils.ComponentSession)
	sessionLogger = utils.WithRemoteAddr(sessionLogger, conn.RemoteAddr())

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
