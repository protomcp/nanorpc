package client

import (
	"context"
	"errors"
	"testing"

	"darvaza.org/core"
	"google.golang.org/protobuf/proto"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// invokeSubscribeCallback wires newSubscribeCallback to the supplied typed
// callback and feeds it the given response, returning the out value plus
// the err the typed callback observed.
//
// A is fixed to *nanorpc.NanoRPCResponse purely because it is a convenient
// proto.Message available in this package; the success row marshals one
// into the wire Data so the test can verify the decoded payload.
func invokeSubscribeCallback(t *testing.T,
	res *nanorpc.NanoRPCResponse) (observedOut *nanorpc.NanoRPCResponse, observedErr error) {
	t.Helper()

	newOut := func() (*nanorpc.NanoRPCResponse, error) {
		return &nanorpc.NanoRPCResponse{}, nil
	}
	return invokeSubscribeCallbackWith(t, res, newOut)
}

// invokeSubscribeCallbackWith is invokeSubscribeCallback with a caller-chosen
// newOut factory, so tests can exercise the typed-nil and error paths of
// callNewOut.
func invokeSubscribeCallbackWith(t *testing.T, res *nanorpc.NanoRPCResponse,
	newOut func() (*nanorpc.NanoRPCResponse, error),
) (observedOut *nanorpc.NanoRPCResponse, observedErr error) {
	t.Helper()

	cb := func(_ context.Context, _ int32,
		out *nanorpc.NanoRPCResponse, err error) error {
		observedOut = out
		observedErr = err
		return nil
	}

	raw := newSubscribeCallback(cb, newOut)
	core.AssertMustNoError(t, raw(context.Background(), res.GetRequestId(), res),
		"raw callback")
	return observedOut, observedErr
}

// TestSubscribeCallback_ACKSurfacesEstablished confirms that a TYPE_RESPONSE
// with STATUS_OK on a subscription channel reaches the typed callback as
// [nanorpc.ErrSubscriptionEstablished], with a freshly allocated empty out.
func TestSubscribeCallback_ACKSurfacesEstablished(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}

	out, err := invokeSubscribeCallback(t, res)
	core.AssertErrorIs(t, err, nanorpc.ErrSubscriptionEstablished, "ACK err")
	core.AssertNotNil(t, out, "ACK out")
	core.AssertEqual(t, int32(0), out.GetRequestId(), "ACK out unpopulated")
}

// TestSubscribeCallback_ACKErrorStatusSurfacesRealError confirms that a
// TYPE_RESPONSE acknowledgement with a non-OK status surfaces as the
// corresponding [nanorpc.ResponseError], not as
// [nanorpc.ErrSubscriptionEstablished].
func TestSubscribeCallback_ACKErrorStatusSurfacesRealError(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:       100,
		ResponseType:    nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus:  nanorpc.NanoRPCResponse_STATUS_NOT_FOUND,
		ResponseMessage: "unknown path",
	}

	out, err := invokeSubscribeCallback(t, res)
	core.AssertFalse(t, nanorpc.IsSubscriptionEstablished(err),
		"established sentinel must not be returned on error ACK")
	core.AssertTrue(t, nanorpc.IsNotFound(err), "IsNotFound on NOT_FOUND ACK")
	core.AssertNotNil(t, out, "error ACK out")
	core.AssertEqual(t, int32(0), out.GetRequestId(), "error ACK out unpopulated")
}

// TestSubscribeCallback_UpdateWithDataIsDelivered confirms that a TYPE_UPDATE
// carrying a valid payload reaches the typed callback with err == nil and a
// non-zero out decoded from the wire bytes.
func TestSubscribeCallback_UpdateWithDataIsDelivered(t *testing.T) {
	inner := &nanorpc.NanoRPCResponse{RequestId: 42}
	data, err := proto.Marshal(inner)
	core.AssertNoError(t, err, "marshal inner")

	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_UPDATE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
		Data:           data,
	}

	out, err := invokeSubscribeCallback(t, res)
	core.AssertNoError(t, err, "update err")
	core.AssertNotNil(t, out, "update out")
	core.AssertEqual(t, int32(42), out.GetRequestId(), "decoded RequestId")
}

// TestSubscribeCallback_UpdateWithoutDataIsErrNoResponse confirms that a
// TYPE_UPDATE delivery carrying no payload preserves the historical
// [nanorpc.ErrNoResponse] surfacing — the ACK fix must not change update
// semantics.
func TestSubscribeCallback_UpdateWithoutDataIsErrNoResponse(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_UPDATE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}

	_, err := invokeSubscribeCallback(t, res)
	core.AssertTrue(t, nanorpc.IsNoResponse(err),
		"empty TYPE_UPDATE should remain ErrNoResponse")
	core.AssertFalse(t, nanorpc.IsSubscriptionEstablished(err),
		"established sentinel must not fire on TYPE_UPDATE")
}

// TestSubscribeCallback_UpdateWithBadDataSurfacesDecodeError confirms that a
// TYPE_UPDATE whose Data fails to unmarshal surfaces the proto error rather
// than [nanorpc.ErrNoResponse] or the established sentinel.
func TestSubscribeCallback_UpdateWithBadDataSurfacesDecodeError(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_UPDATE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
		Data:           []byte{0xff, 0xff, 0xff, 0xff},
	}

	_, err := invokeSubscribeCallback(t, res)
	core.AssertError(t, err, "decode err")
	core.AssertFalse(t, nanorpc.IsNoResponse(err),
		"decode error must not be reduced to ErrNoResponse")
	core.AssertFalse(t, nanorpc.IsSubscriptionEstablished(err),
		"decode error must not surface as established")
}

// TestSubscribeCallback_NewOutTypedNilRejected confirms that a newOut factory
// returning a typed-nil message is rejected with core.ErrInvalid rather than
// handing a (*T)(nil) to the typed callback.
func TestSubscribeCallback_NewOutTypedNilRejected(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}
	newOut := func() (*nanorpc.NanoRPCResponse, error) { return nil, nil }

	_, err := invokeSubscribeCallbackWith(t, res, newOut)
	core.AssertErrorIs(t, err, core.ErrInvalid, "typed-nil newOut rejected")
}

// TestSubscribeCallback_NewOutErrorSurfaces confirms that an error from the
// newOut factory reaches the typed callback unchanged.
func TestSubscribeCallback_NewOutErrorSurfaces(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}
	boom := errors.New("factory boom")
	newOut := func() (*nanorpc.NanoRPCResponse, error) { return nil, boom }

	_, err := invokeSubscribeCallbackWith(t, res, newOut)
	core.AssertErrorIs(t, err, boom, "factory error surfaced")
}

// TestSubscribeCallback_UpdateNewOutErrorSurfaces confirms that a newOut
// factory error on the TYPE_UPDATE delivery path reaches the typed callback
// unchanged — the update path must guard the factory just like the ACK path.
func TestSubscribeCallback_UpdateNewOutErrorSurfaces(t *testing.T) {
	res := &nanorpc.NanoRPCResponse{
		RequestId:      100,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_UPDATE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}
	boom := errors.New("factory boom")
	newOut := func() (*nanorpc.NanoRPCResponse, error) { return nil, boom }

	_, err := invokeSubscribeCallbackWith(t, res, newOut)
	core.AssertErrorIs(t, err, boom, "update factory error surfaced")
}
