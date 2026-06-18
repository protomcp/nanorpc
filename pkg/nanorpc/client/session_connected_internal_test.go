package client

import (
	"bufio"
	"context"
	"io"
	"net"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/net/reconnect"
	"darvaza.org/x/sync/workgroup"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// recvTimeout bounds every blocking wait in these fixtures so a routing
// regression fails as a timeout instead of hanging the suite.
const recvTimeout = 2 * time.Second

// fakeServer drives the server end of a net.Pipe: a workgroup-supervised
// reader decodes inbound NanoRPC requests onto reqs, and reply scripts the
// matching responses back to the client session under test. The workgroup
// runs the reader and stores its first error as the cancellation cause, so
// the clean-up's Wait surfaces a genuine fault — while a test-initiated
// shutdown (Cancel before the pipe closes) leaves the cause nil.
type fakeServer struct {
	conn net.Conn
	wg   *workgroup.Group
	t    *testing.T
	reqs chan *nanorpc.NanoRPCRequest
}

// read decodes the length-delimited request stream, forwarding each request
// to recv, until the workgroup is cancelled or the stream ends. It mirrors a
// server's inbound half: nanorpc.Split frames, nanorpc.DecodeRequest
// decodes. A stream error seen after cancellation is the shutdown the test
// initiated — reported as nil — so only an error on a still-running group
// becomes the workgroup's stored cause.
func (fs *fakeServer) read(ctx context.Context) error {
	sc := bufio.NewScanner(fs.conn)
	sc.Split(nanorpc.Split)
	for sc.Scan() {
		req, _, err := nanorpc.DecodeRequest(sc.Bytes())
		if err != nil {
			return stopErr(ctx, err)
		}
		select {
		case fs.reqs <- req:
		case <-ctx.Done():
			return nil
		}
	}

	return stopErr(ctx, sc.Err())
}

// stopErr drops a stream error once the group is cancelled: nanorpc framing
// surfaces a boundary close as io.ErrUnexpectedEOF, indistinguishable from a
// real truncation by value alone, so intent decides. Before cancellation the
// error is genuine and propagates as the workgroup's cause.
func stopErr(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return nil
	}
	return err
}

// recv returns the next request the client sent, failing the test if none
// arrives within recvTimeout.
func (fs *fakeServer) recv() *nanorpc.NanoRPCRequest {
	fs.t.Helper()
	select {
	case req := <-fs.reqs:
		return req
	case <-time.After(recvTimeout):
		fs.t.Fatal("timed out waiting for a request from the client")
		return nil
	}
}

// reply writes a response back to the client session. The pipe write is
// consumed by the session's reader goroutine, so it cannot deadlock while
// the session is spawned.
func (fs *fakeServer) reply(res *nanorpc.NanoRPCResponse) {
	fs.t.Helper()
	_, err := nanorpc.EncodeResponseTo(fs.conn, res, nil)
	core.AssertMustNoError(fs.t, err, "EncodeResponseTo")
}

// newResponse builds a server response for the given request id.
func newResponse(id int32, rt nanorpc.NanoRPCResponse_Type,
	st nanorpc.NanoRPCResponse_Status) *nanorpc.NanoRPCResponse {
	return &nanorpc.NanoRPCResponse{
		RequestId:      id,
		ResponseType:   rt,
		ResponseStatus: st,
	}
}

// cbEvent records one RequestCallback invocation.
type cbEvent struct {
	resp *nanorpc.NanoRPCResponse
	id   int32
}

// recordingCallback returns a RequestCallback that forwards each
// invocation to ch. The channel is buffered by the caller so the
// session's reporting goroutines never block on a test that has stopped
// reading.
func recordingCallback(ch chan cbEvent) RequestCallback {
	return func(_ context.Context, id int32, resp *nanorpc.NanoRPCResponse) error {
		ch <- cbEvent{resp: resp, id: id}
		return nil
	}
}

// mustRecvEvent waits for the next callback event, failing on timeout.
func mustRecvEvent(t *testing.T, ch <-chan cbEvent, what string) cbEvent {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(recvTimeout):
		t.Fatalf("timed out waiting for the %s callback", what)
		return cbEvent{}
	}
}

// newConnectedSession wires a spawned [Session] to a [fakeServer] over an
// in-memory net.Pipe and attaches it to the [Client], so the public
// request API (Request/Subscribe/Ping/Unsubscribe) flows through Send, the
// run loop and handleResponse against scripted server responses. It builds
// the session directly rather than via newClientSession because the
// production factory binds the StreamSession to a live reconnect.Client and
// its deadline hooks; the pipe needs neither.
func newConnectedSession(t *testing.T) (*Client, *fakeServer) {
	t.Helper()
	c := newClientForTest(t)

	cliConn, srvConn := net.Pipe()

	ss := &reconnect.StreamSession[*nanorpc.NanoRPCResponse, clientRequest]{
		QueueSize: 1,
		Conn:      cliConn,
		Context:   context.Background(),
		Split:     nanorpc.Split,
		MarshalTo: func(r clientRequest, w io.Writer) error {
			_, err := nanorpc.EncodeRequestTo(w, r.r, r.d)
			return err
		},
		Unmarshal: func(data []byte) (*nanorpc.NanoRPCResponse, error) {
			resp, _, err := nanorpc.DecodeResponse(data)
			return resp, err
		},
	}
	cs := &Session{c: c, ss: ss, WorkGroup: ss}

	core.AssertMustNoError(t, cs.Spawn(), "Spawn")
	core.AssertMustNoError(t, c.setSession(cs), "setSession")

	srv := &fakeServer{
		conn: srvConn,
		wg:   workgroup.New(context.Background()),
		t:    t,
		reqs: make(chan *nanorpc.NanoRPCRequest, 16),
	}
	core.AssertMustNoError(t, srv.wg.GoCatch(srv.read, nil), "server reader")

	t.Cleanup(func() {
		// Cancel before closing the pipe so the read unwinds with the group
		// already cancelled, keeping the cause nil; closing srvConn unblocks
		// the read itself.
		srv.wg.Cancel(nil)
		_ = srvConn.Close()
		_ = ss.Close()
		_ = cs.Wait()
		core.AssertNoError(t, srv.wg.Wait(), "server read loop")
	})

	return c, srv
}

// TestSession_Ping_whenConnected drives Send's accept path for a
// callback-less request: Ping returns true and the server observes a
// TYPE_PING on the wire.
func TestSession_Ping_whenConnected(t *testing.T) {
	c, srv := newConnectedSession(t)

	core.AssertTrue(t, c.Ping(), "Ping")

	req := srv.recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_PING, req.RequestType,
		"request_type")
}

// TestSession_Pong_acknowledged drives a ping through the run loop: the
// server's TYPE_PONG resolves the registered callback, so the Pong channel
// reports success.
func TestSession_Pong_acknowledged(t *testing.T) {
	c, srv := newConnectedSession(t)

	ch := c.Pong()

	req := srv.recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_PING, req.RequestType,
		"request_type")
	srv.reply(newResponse(req.RequestId, respPong, statusOK))

	select {
	case err := <-ch:
		core.AssertNoError(t, err, "Pong")
	case <-time.After(recvTimeout):
		t.Fatal("timed out waiting for the pong")
	}
}

// TestSession_Request_roundTrip drives a request from Send through the run
// loop and handleResponse: the server's TYPE_RESPONSE pops the registered
// callback, which fires once with the matching response.
func TestSession_Request_roundTrip(t *testing.T) {
	c, srv := newConnectedSession(t)

	events := make(chan cbEvent, 4)
	id, err := c.Request("/echo", nil, recordingCallback(events))
	core.AssertMustNoError(t, err, "Request")

	req := srv.recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_REQUEST, req.RequestType,
		"request_type")
	core.AssertEqual(t, "/echo", req.GetPath(), "path")
	core.AssertEqual(t, id, req.RequestId, "request_id")

	srv.reply(newResponse(req.RequestId, respResponse, statusOK))

	ev := mustRecvEvent(t, events, "request")
	core.AssertEqual(t, id, ev.id, "callback_id")
	core.AssertEqual(t, statusOK, ev.resp.ResponseStatus, "response_status")
}

// TestSession_Subscribe_lifecycle walks the §6.1 subscription lifecycle
// over the wire: Pending -> Active on the STATUS_OK acknowledgement, a
// TYPE_UPDATE delivery while Active, then Unsubscribing -> Terminated on
// the unsubscribe acknowledgement, which drops both queue entries.
func TestSession_Subscribe_lifecycle(t *testing.T) {
	c, srv := newConnectedSession(t)
	cs, scErr := c.getSession()
	core.AssertMustNoError(t, scErr, "getSession")

	subEvents := make(chan cbEvent, 4)
	id, err := c.Subscribe("/sensors/temp", nil, recordingCallback(subEvents))
	core.AssertMustNoError(t, err, "Subscribe")

	sub := srv.recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE, sub.RequestType,
		"subscribe_type")

	// Pending -> Active: the acknowledgement fires the callback.
	srv.reply(newResponse(id, respResponse, statusOK))
	ack := mustRecvEvent(t, subEvents, "subscribe acknowledgement")
	core.AssertEqual(t, statusOK, ack.resp.ResponseStatus, "ack_status")
	core.AssertTrue(t, cs.IsActive(), "IsActive after ack")

	// Active: an update routes to the same callback and stays queued.
	srv.reply(newResponse(id, respUpdate, statusOK))
	upd := mustRecvEvent(t, subEvents, "update")
	core.AssertEqual(t, respUpdate, upd.resp.ResponseType, "update_type")
	core.AssertTrue(t, cs.IsActive(), "IsActive after update")

	// Unsubscribing: the unsubscribe request reuses the subscription id.
	unsubscribeEvents := make(chan cbEvent, 4)
	err = c.Unsubscribe("/sensors/temp", id, recordingCallback(unsubscribeEvents))
	core.AssertMustNoError(t, err, "Unsubscribe")

	unsubscribeReq := srv.recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_REQUEST,
		unsubscribeReq.RequestType, "unsubscribe_type")
	core.AssertEqual(t, id, unsubscribeReq.RequestId, "unsubscribe_id")

	// Terminated: the unsubscribe acknowledgement drops both entries.
	srv.reply(newResponse(id, respResponse, statusOK))
	done := mustRecvEvent(t, unsubscribeEvents, "unsubscribe acknowledgement")
	core.AssertEqual(t, id, done.id, "unsubscribe_callback_id")
	core.AssertFalse(t, cs.IsActive(), "IsActive after unsubscribe")
}

// TestSession_Subscribe_rejected covers the §6.1 Pending -> Terminated
// edge: a non-OK acknowledgement fires the callback with the failure and
// drops the SUBSCRIBE entry, so no phantom lingers to satisfy a later
// unsubscribe guard.
func TestSession_Subscribe_rejected(t *testing.T) {
	c, srv := newConnectedSession(t)
	cs, scErr := c.getSession()
	core.AssertMustNoError(t, scErr, "getSession")

	subEvents := make(chan cbEvent, 4)
	id, err := c.Subscribe("/sensors/temp", nil, recordingCallback(subEvents))
	core.AssertMustNoError(t, err, "Subscribe")

	sub := srv.recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE, sub.RequestType,
		"subscribe_type")

	srv.reply(newResponse(id, respResponse, statusNotFound))
	nack := mustRecvEvent(t, subEvents, "subscribe rejection")
	core.AssertEqual(t, statusNotFound, nack.resp.ResponseStatus, "nack_status")
	core.AssertFalse(t, cs.IsActive(), "IsActive after rejection")
}
