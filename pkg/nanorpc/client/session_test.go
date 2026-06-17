package client

import (
	"context"
	"testing"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// Compile-time verification that test case types implement TestCase
var _ core.TestCase = routingTestCase{}
var _ core.TestCase = sendGuardTestCase{}
var _ core.TestCase = sendRejectTestCase{}

// Local aliases keep the data tables readable.
const (
	reqUnspecified = nanorpc.NanoRPCRequest_TYPE_UNSPECIFIED
	reqRequest     = nanorpc.NanoRPCRequest_TYPE_REQUEST
	reqSubscribe   = nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE
	reqPing        = nanorpc.NanoRPCRequest_TYPE_PING

	respResponse = nanorpc.NanoRPCResponse_TYPE_RESPONSE
	respUpdate   = nanorpc.NanoRPCResponse_TYPE_UPDATE
	respPong     = nanorpc.NanoRPCResponse_TYPE_PONG

	statusOK       = nanorpc.NanoRPCResponse_STATUS_OK
	statusNotFound = nanorpc.NanoRPCResponse_STATUS_NOT_FOUND
)

// routingSeed describes one entry in a seeded callback queue. It also
// doubles as the expected residue shape after a dispatch, since both
// carry the same three observable fields.
type routingSeed struct {
	requestType  nanorpc.NanoRPCRequest_Type
	requestID    int32
	acknowledged bool
}

// Seed constructors give the data tables a short vocabulary.
func sub(id int32) routingSeed  { return routingSeed{reqSubscribe, id, false} }
func req(id int32) routingSeed  { return routingSeed{reqRequest, id, false} }
func ping(id int32) routingSeed { return routingSeed{reqPing, id, false} }

func subAcknowledged(id int32) routingSeed {
	return routingSeed{reqSubscribe, id, true}
}

// routingTestCase exercises (*Session).popRequestCallback's routing
// matrix: seeded queue, dispatched (reqID, respType, respStatus), expected
// fired callback and queue residue.
type routingTestCase struct {
	name       string
	seed       []routingSeed
	residue    []routingSeed
	firedIdx   int
	respType   nanorpc.NanoRPCResponse_Type
	respStatus nanorpc.NanoRPCResponse_Status
	reqID      int32
}

func (tc routingTestCase) Name() string { return tc.name }

func (tc routingTestCase) Test(t *testing.T) {
	t.Helper()

	fired := -1
	cs := newRoutingSession(tc.seed, &fired)

	resp := &nanorpc.NanoRPCResponse{
		RequestId:      tc.reqID,
		ResponseType:   tc.respType,
		ResponseStatus: tc.respStatus,
	}
	cb := cs.popRequestCallback(resp)

	assertRoutingFired(t, cb, &fired, tc.firedIdx, tc.reqID)
	assertResidue(t, cs.cb, tc.residue)
}

// newRoutingTestCase narrates a success-status row: name, seed queue,
// dispatched (reqID, respType), expected (firedIdx, residue). The routed
// response carries STATUS_OK — the realistic status for every non-failure
// path; the failure path is covered by newRoutingErrTestCase.
//
//revive:disable-next-line:argument-limit
func newRoutingTestCase(name string, seed []routingSeed,
	reqID int32, respType nanorpc.NanoRPCResponse_Type,
	firedIdx int, residue []routingSeed) routingTestCase {
	return routingTestCase{
		name:       name,
		seed:       seed,
		residue:    residue,
		firedIdx:   firedIdx,
		respType:   respType,
		respStatus: statusOK,
		reqID:      reqID,
	}
}

// newRoutingErrTestCase narrates a failure-status row, where respStatus is
// load-bearing: a non-OK subscribe acknowledgement must drop its SUBSCRIBE
// entry instead of marking it Acknowledged.
//
//revive:disable-next-line:argument-limit
func newRoutingErrTestCase(name string, seed []routingSeed,
	reqID int32, respType nanorpc.NanoRPCResponse_Type,
	respStatus nanorpc.NanoRPCResponse_Status,
	firedIdx int, residue []routingSeed) routingTestCase {
	return routingTestCase{
		name:       name,
		seed:       seed,
		residue:    residue,
		firedIdx:   firedIdx,
		respType:   respType,
		respStatus: respStatus,
		reqID:      reqID,
	}
}

func routingTestCases() []routingTestCase {
	return []routingTestCase{
		newRoutingTestCase("plain_request_response",
			core.S(req(5)), 5, respResponse,
			0, nil),
		newRoutingTestCase("ping_pong",
			core.S(ping(7)), 7, respPong,
			0, nil),
		newRoutingTestCase("pending_ack_ok_activates",
			core.S(sub(9)), 9, respResponse,
			0, core.S(subAcknowledged(9))),
		newRoutingErrTestCase("pending_ack_error_drops_entry",
			core.S(sub(9)), 9, respResponse, statusNotFound,
			0, nil),
		newRoutingTestCase("active_duplicate_ok_ignored",
			core.S(subAcknowledged(9)), 9, respResponse,
			-1, core.S(subAcknowledged(9))),
		newRoutingErrTestCase("active_error_response_ignored",
			core.S(subAcknowledged(9)), 9, respResponse, statusNotFound,
			-1, core.S(subAcknowledged(9))),
		newRoutingTestCase("subscribe_update_acknowledged",
			core.S(subAcknowledged(9)), 9, respUpdate,
			0, core.S(subAcknowledged(9))),
		newRoutingTestCase("in_flight_update_routes_to_subscribe",
			core.S(subAcknowledged(9), req(9)), 9, respUpdate,
			0, core.S(subAcknowledged(9), req(9))),
		newRoutingTestCase("unsubscribe_response_drops_both",
			core.S(subAcknowledged(9), req(9)), 9, respResponse,
			1, nil),
		newRoutingTestCase("double_unsubscribe_first_wins",
			core.S(subAcknowledged(9), req(9), req(9)), 9, respResponse,
			1, core.S(req(9))),
		newRoutingTestCase("update_without_subscribe_returns_nil",
			core.S(req(9)), 9, respUpdate,
			-1, core.S(req(9))),
		newRoutingTestCase("unknown_id_returns_nil",
			core.S(subAcknowledged(9)), 99, respResponse,
			-1, core.S(subAcknowledged(9))),
	}
}

// TestSession_popRequestCallback exercises the routing matrix.
func TestSession_popRequestCallback(t *testing.T) {
	core.RunTestCases(t, routingTestCases())
}

// sendGuardTestCase exercises (*Session).checkUnsubscribeTarget: the
// guard that rejects unsubscribe-shape TYPE_REQUEST entries whose
// target subscription is missing or still pending.
type sendGuardTestCase struct {
	name    string
	wantErr error
	seed    []routingSeed
	reqID   int32
}

func (tc sendGuardTestCase) Name() string { return tc.name }

func (tc sendGuardTestCase) Test(t *testing.T) {
	t.Helper()

	cs := newRoutingSession(tc.seed, new(int))
	initialLen := len(cs.cb)

	err := cs.checkUnsubscribeTarget(tc.reqID)

	if tc.wantErr == nil {
		core.AssertNoError(t, err, "err")
	} else {
		core.AssertErrorIs(t, err, tc.wantErr, "err")
		core.AssertErrorIs(t, err, core.ErrInvalid, "err_family")
	}

	core.AssertEqual(t, initialLen, len(cs.cb), "queue_unchanged")
}

// newSendGuardTestCase narrates a row: name, seed queue, dispatched
// reqID, expected sentinel (nil means the guard accepts).
func newSendGuardTestCase(name string, seed []routingSeed,
	reqID int32, wantErr error) sendGuardTestCase {
	return sendGuardTestCase{
		name:    name,
		wantErr: wantErr,
		seed:    seed,
		reqID:   reqID,
	}
}

func sendGuardTestCases() []sendGuardTestCase {
	return []sendGuardTestCase{
		newSendGuardTestCase("no_matching_subscription",
			nil, 5, ErrNoSubscription),
		newSendGuardTestCase("subscription_pending",
			core.S(sub(5)), 5, ErrSubscriptionPending),
		newSendGuardTestCase("subscription_acknowledged",
			core.S(subAcknowledged(5)), 5, nil),
	}
}

// TestSession_checkUnsubscribeTarget exercises the Send guard.
func TestSession_checkUnsubscribeTarget(t *testing.T) {
	core.RunTestCases(t, sendGuardTestCases())
}

// sendRejectTestCase drives (*Session).Send through the guards that
// return before the request reaches the wire: an unknown request type, a
// missing callback, or an unsubscribe-shape request whose target
// subscription is missing or still pending. Only rejection rows belong
// here — an accepted Send reaches cs.ss and needs a connected session.
type sendRejectTestCase struct {
	cb      RequestCallback
	name    string
	wantErr error
	seed    []routingSeed
	reqType nanorpc.NanoRPCRequest_Type
	reqID   int32
}

func (tc sendRejectTestCase) Name() string { return tc.name }

func (tc sendRejectTestCase) Test(t *testing.T) {
	t.Helper()

	cs := newRoutingSession(tc.seed, new(int))
	initialLen := len(cs.cb)

	req := &nanorpc.NanoRPCRequest{RequestType: tc.reqType, RequestId: tc.reqID}
	err := cs.Send(req, nil, tc.cb)

	core.AssertErrorIs(t, err, tc.wantErr, "err")
	core.AssertErrorIs(t, err, core.ErrInvalid, "err_family")
	core.AssertEqual(t, initialLen, len(cs.cb), "queue_unchanged")
}

// newSendRejectTestCase narrates a row: name, seed queue, request type and
// id, the callback to offer, and the expected sentinel.
//
//revive:disable-next-line:argument-limit
func newSendRejectTestCase(name string, seed []routingSeed,
	reqType nanorpc.NanoRPCRequest_Type, reqID int32,
	cb RequestCallback, wantErr error) sendRejectTestCase {
	return sendRejectTestCase{
		name:    name,
		wantErr: wantErr,
		seed:    seed,
		cb:      cb,
		reqType: reqType,
		reqID:   reqID,
	}
}

func sendRejectTestCases() []sendRejectTestCase {
	return []sendRejectTestCase{
		newSendRejectTestCase("invalid_request_type",
			nil, reqUnspecified, 0, nil, ErrInvalidRequestType),
		newSendRejectTestCase("request_missing_callback",
			nil, reqRequest, 0, nil, ErrMissingCallback),
		newSendRejectTestCase("subscribe_missing_callback",
			nil, reqSubscribe, 0, nil, ErrMissingCallback),
		newSendRejectTestCase("unsubscribe_no_subscription",
			nil, reqRequest, 5, discardCallback(), ErrNoSubscription),
		newSendRejectTestCase("unsubscribe_pending",
			core.S(sub(5)), reqRequest, 5, discardCallback(),
			ErrSubscriptionPending),
	}
}

// TestSession_Send_rejections exercises Send's pre-wire rejection paths.
func TestSession_Send_rejections(t *testing.T) {
	core.RunTestCases(t, sendRejectTestCases())
}

// TestSession_Send_nilRequest covers the exported guard against a nil
// request, which the table cases cannot express since they always build a
// concrete request. Without the guard Send would panic dereferencing
// req.RequestType.
func TestSession_Send_nilRequest(t *testing.T) {
	cs := newRoutingSession(nil, new(int))

	err := cs.Send(nil, nil, discardCallback())

	core.AssertErrorIs(t, err, ErrNilRequest, "err")
	core.AssertErrorIs(t, err, core.ErrInvalid, "err_family")
	core.AssertEqual(t, 0, len(cs.cb), "queue_unchanged")
}

// discardCallback is a no-op RequestCallback for rows that must offer a
// callback to pass validateSendArgs before a later guard rejects them.
func discardCallback() RequestCallback {
	return func(context.Context, int32, *nanorpc.NanoRPCResponse) error {
		return nil
	}
}

// newRoutingSession builds a *Session with its callback queue seeded
// from seed[]. Each seeded callback writes its index into *fired when
// invoked, so the test can identify which callback popRequestCallback
// returned.
func newRoutingSession(seed []routingSeed, fired *int) *Session {
	cs := &Session{}
	for i, s := range seed {
		cs.cb = append(cs.cb, clientRequestQueue{
			RequestID:    s.requestID,
			RequestType:  s.requestType,
			Acknowledged: s.acknowledged,
			Callback:     newSeedCallback(i, fired),
		})
	}
	return cs
}

func newSeedCallback(idx int, fired *int) RequestCallback {
	return func(_ context.Context, _ int32, _ *nanorpc.NanoRPCResponse) error {
		*fired = idx
		return nil
	}
}

// assertRoutingFired checks the callback returned by popRequestCallback:
// firedIdx == -1 expects nil, otherwise the callback is invoked and the
// captured sentinel must match firedIdx.
func assertRoutingFired(t *testing.T, cb RequestCallback, fired *int,
	firedIdx int, reqID int32) {
	t.Helper()

	if firedIdx < 0 {
		core.AssertNil(t, cb, "cb")
		return
	}
	if !core.AssertNotNil(t, cb, "cb") {
		return
	}
	_ = cb(context.Background(), reqID, &nanorpc.NanoRPCResponse{})
	core.AssertEqual(t, firedIdx, *fired, "fired_idx")
}

// assertResidue checks the surviving queue entries by RequestType,
// RequestID, and Acknowledged. The order of want must match the order
// the entries appear in got.
func assertResidue(t *testing.T, got []clientRequestQueue, want []routingSeed) {
	t.Helper()

	if !core.AssertEqual(t, len(want), len(got), "residue_len") {
		return
	}
	for i, w := range want {
		core.AssertEqual(t, w.requestType, got[i].RequestType,
			"residue_%d_type", i)
		core.AssertEqual(t, w.requestID, got[i].RequestID,
			"residue_%d_id", i)
		core.AssertEqual(t, w.acknowledged, got[i].Acknowledged,
			"residue_%d_acknowledged", i)
	}
}
