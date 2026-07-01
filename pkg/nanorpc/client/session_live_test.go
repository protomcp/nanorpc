package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/net/reconnect"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/client"
	"protomcp.org/nanorpc/pkg/nanorpc/mock/server"
)

// liveTimeout bounds every blocking wait in these fixtures so a routing
// regression fails as a timeout instead of hanging the suite.
const liveTimeout = 2 * time.Second

// cbEvent records one RequestCallback invocation.
type cbEvent struct {
	resp *nanorpc.NanoRPCResponse
	id   int32
}

// liveRecordingCallback returns a RequestCallback that forwards each
// invocation to ch. The channel is buffered by the caller so the session's
// reporting goroutines never block on a test that has stopped reading.
func liveRecordingCallback(ch chan cbEvent) client.RequestCallback {
	return func(_ context.Context, id int32, resp *nanorpc.NanoRPCResponse) error {
		ch <- cbEvent{resp: resp, id: id}
		return nil
	}
}

// mustRecvLiveEvent waits for the next callback event, failing on timeout.
func mustRecvLiveEvent(t *testing.T, ch <-chan cbEvent, what string) cbEvent {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(liveTimeout):
		t.Fatalf("timed out waiting for the %s callback", what)
		return cbEvent{}
	}
}

// newLiveResponse builds a server response for the given request id.
func newLiveResponse(id int32, rt nanorpc.NanoRPCResponse_Type,
	st nanorpc.NanoRPCResponse_Status) *nanorpc.NanoRPCResponse {
	return &nanorpc.NanoRPCResponse{
		RequestId:      id,
		ResponseType:   rt,
		ResponseStatus: st,
	}
}

// liveFixture is a real [client.Client] connected to a mock [server.Server]
// over a loopback TCP connection, exercising the production session factory
// and its reconnect.Client deadline hooks — the paths the in-memory pipe
// fixture deliberately skips. conn is the server side of the dialled
// connection, for scripting responses.
type liveFixture struct {
	c           *client.Client
	srv         *server.Server
	conn        *server.Conn
	connects    chan struct{}
	disconnects chan struct{}
}

// newLiveFixture stands up the mock server, dials a real client at it, and
// confirms the session is published before returning. Confirming the
// connection before any caller drives it is the binding rule for the live
// client: a tear-down landing between dial and setConn would otherwise park
// OnSession on a deadline-less read (see x/net/reconnect NEXT.md). OnConnect
// and OnDisconnect record onto buffered channels so the callbacks never
// block.
func newLiveFixture(t *testing.T) *liveFixture {
	t.Helper()

	f := &liveFixture{
		srv:         server.New(t),
		connects:    make(chan struct{}, 1),
		disconnects: make(chan struct{}, 1),
	}

	f.c = newLiveClient(t, f.srv, client.Config{
		OnConnect: func(context.Context, reconnect.WorkGroup) error {
			trySignal(f.connects)
			return nil
		},
		OnDisconnect: func(context.Context) error {
			trySignal(f.disconnects)
			return nil
		},
	})

	core.AssertMustNoError(t, f.c.Connect(), "Connect")

	// Retrieve the server side of the dialled connection, then confirm the
	// client published its session before any test drives it.
	f.conn = f.srv.Accept()

	ctx, cancel := context.WithTimeout(context.Background(), liveTimeout)
	defer cancel()
	core.AssertMustNoError(t, f.c.WaitConnected(ctx), "WaitConnected")
	f.mustSignal(t, f.connects, "OnConnect")

	return f
}

// newLiveClient builds a real client aimed at srv from cfg, filling in the
// loopback Context and Remote and registering the shutdown cleanup. It does
// not confirm the connection: callers that need a live session call
// WaitConnected. Shutting the client down before the server's own cleanup
// closes the listener (t.Cleanup is LIFO, and this registers after
// server.New) stops the reconnect loop instead of redialling a dead listener.
func newLiveClient(t *testing.T, srv *server.Server, cfg client.Config) *client.Client {
	t.Helper()

	cfg.Context = context.Background()
	cfg.Remote = srv.Addr()

	c, err := cfg.New()
	core.AssertMustNoError(t, err, "cfg.New")

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), liveTimeout)
		defer cancel()
		_ = c.Shutdown(ctx)
	})

	return c
}

// mustSignal waits for one signal on ch, failing on timeout.
func (*liveFixture) mustSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(liveTimeout):
		t.Fatalf("timed out waiting for the %s callback", what)
	}
}

// trySignal delivers a non-blocking signal, dropping it when one is already
// buffered: the tests assert a callback fired at least once, not a count.
func trySignal(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// TestLiveClient_Request_roundTrip drives a request over a real connection:
// dialling fires onReconnectConnect (session build) and onReconnectSession
// (Spawn + user OnConnect), and the idle pre-read deadline hook runs before
// any callback is queued. The server's TYPE_RESPONSE resolves the registered
// callback.
func TestLiveClient_Request_roundTrip(t *testing.T) {
	f := newLiveFixture(t)

	events := make(chan cbEvent, 4)
	id, err := f.c.Request("/echo", nil, liveRecordingCallback(events))
	core.AssertMustNoError(t, err, "Request")

	req := f.conn.Recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_REQUEST, req.RequestType,
		"request_type")
	core.AssertEqual(t, "/echo", req.GetPath(), "path")
	core.AssertEqual(t, id, req.RequestId, "request_id")

	f.conn.Reply(newLiveResponse(id, nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		nanorpc.NanoRPCResponse_STATUS_OK))

	ev := mustRecvLiveEvent(t, events, "request")
	core.AssertEqual(t, id, ev.id, "callback_id")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, ev.resp.ResponseStatus,
		"response_status")
}

// TestLiveClient_Subscribe_receivesUpdate covers the active branch of the
// read-deadline hook: once the subscription is queued the reader's pre-read
// hook takes ResetReadDeadline instead of the idle SetReadDeadline. The
// STATUS_OK acknowledgement moves the subscription Pending -> Active, then a
// TYPE_UPDATE routes to the same callback while Active.
func TestLiveClient_Subscribe_receivesUpdate(t *testing.T) {
	f := newLiveFixture(t)

	events := make(chan cbEvent, 4)
	id, err := f.c.Subscribe("/sensors/temp", nil, liveRecordingCallback(events))
	core.AssertMustNoError(t, err, "Subscribe")

	sub := f.conn.Recv()
	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE, sub.RequestType,
		"subscribe_type")

	// Pending -> Active: the acknowledgement fires the callback.
	f.conn.Reply(newLiveResponse(id, nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		nanorpc.NanoRPCResponse_STATUS_OK))
	ack := mustRecvLiveEvent(t, events, "subscribe acknowledgement")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, ack.resp.ResponseStatus,
		"ack_status")

	// Active: an update routes to the same callback and stays queued.
	f.conn.Reply(newLiveResponse(id, nanorpc.NanoRPCResponse_TYPE_UPDATE,
		nanorpc.NanoRPCResponse_STATUS_OK))
	upd := mustRecvLiveEvent(t, events, "update")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_TYPE_UPDATE, upd.resp.ResponseType,
		"update_type")
}

// TestLiveClient_OnDisconnect_firesOnServerClose closes the server while the
// session is live, driving onReconnectDisconnect and the user OnDisconnect
// callback. Closing the whole server also drops the listener, so the client
// cannot redial and the disconnect is observed exactly once.
func TestLiveClient_OnDisconnect_firesOnServerClose(t *testing.T) {
	f := newLiveFixture(t)

	core.AssertMustNoError(t, f.srv.Close(), "server close")

	f.mustSignal(t, f.disconnects, "OnDisconnect")
}

// TestLiveClient_Disconnect_noCallback covers onReconnectDisconnect's
// fn == nil arm: with no OnDisconnect configured the server-side close still
// tears the session down through the debug-log-only branch. The read error
// from that tear-down surfaces through onReconnectError, an edge we wait on
// rather than polling for the state change; the session is cleared before
// that error fires, so IsConnected then reports false. The do-not-reconnect
// waiter stops the retry loop so the disconnect error is the first observed.
func TestLiveClient_Disconnect_noCallback(t *testing.T) {
	srv := server.New(t)

	errs := make(chan error, 1)
	c := newLiveClient(t, srv, client.Config{
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),
		OnError: func(_ context.Context, err error) error {
			select {
			case errs <- err:
			default:
			}
			return err
		},
	})

	core.AssertMustNoError(t, c.Connect(), "Connect")
	srv.Accept()

	ctx, cancel := context.WithTimeout(context.Background(), liveTimeout)
	defer cancel()
	core.AssertMustNoError(t, c.WaitConnected(ctx), "WaitConnected")

	core.AssertMustNoError(t, srv.Close(), "server close")

	err := mustRecvError(t, errs, "OnError")
	core.AssertError(t, err, "disconnect error")
	core.AssertFalse(t, c.IsConnected(), "IsConnected after disconnect")
}

// TestLiveClient_OnConnect_error covers onReconnectSession's OnConnect-error
// arm: an OnConnect that returns an error aborts the session before Wait, and
// the error surfaces through onReconnectError to the user OnError callback. A
// do-not-reconnect waiter halts the retry loop so the OnConnect error is the
// first one observed.
func TestLiveClient_OnConnect_error(t *testing.T) {
	srv := server.New(t)

	wantErr := errors.New("refusing the connection")
	errs := make(chan error, 1)

	c := newLiveClient(t, srv, client.Config{
		WaitReconnect: reconnect.NewDoNotReconnectWaiter(nil),
		OnConnect: func(context.Context, reconnect.WorkGroup) error {
			return wantErr
		},
		OnError: func(_ context.Context, err error) error {
			select {
			case errs <- err:
			default:
			}
			return err
		},
	})

	core.AssertMustNoError(t, c.Connect(), "Connect")
	srv.Accept()

	err := mustRecvError(t, errs, "OnError")
	core.AssertErrorIs(t, err, wantErr, "OnError cause")
}

// mustRecvError waits for the next error delivered to ch, failing on timeout.
func mustRecvError(t *testing.T, ch <-chan error, what string) error {
	t.Helper()

	select {
	case err := <-ch:
		return err
	case <-time.After(liveTimeout):
		t.Fatalf("timed out waiting for the %s callback", what)
		return nil
	}
}
