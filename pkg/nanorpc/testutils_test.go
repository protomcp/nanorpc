package nanorpc

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"
)

// S creates a slice from variadic arguments
func S[T any](items ...T) []T {
	return items
}

// AssertEqual checks if two values are equal
func AssertEqual[T comparable](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if expected != actual {
		msg := fmt.Sprintf("Expected %v, got %v", expected, actual)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + ": " + msg
		}
		t.Errorf("%s", msg)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual[T comparable](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if expected == actual {
		msg := fmt.Sprintf("Expected values to be different, both are %v", expected)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + ": " + msg
		}
		t.Errorf("%s", msg)
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value any, msgAndArgs ...any) {
	t.Helper()
	if value != nil {
		msg := fmt.Sprintf("Expected nil, got %v", value)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + ": " + msg
		}
		t.Errorf("%s", msg)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value any, msgAndArgs ...any) {
	t.Helper()
	if value == nil {
		msg := "Expected non-nil value"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf("%s", msg)
	}
}

// AssertTrue checks if a condition is true
//
//revive:disable-next-line:flag-parameter
func AssertTrue(t *testing.T, condition bool, msgAndArgs ...any) {
	t.Helper()
	if !condition {
		msg := "Expected condition to be true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf("%s", msg)
	}
}

// AssertFalse checks if a condition is false
//
//revive:disable-next-line:flag-parameter
func AssertFalse(t *testing.T, condition bool, msgAndArgs ...any) {
	t.Helper()
	if condition {
		msg := "Expected condition to be false"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf("%s", msg)
	}
}

// AssertNoError checks if error is nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		msg := fmt.Sprintf("Expected no error, got %v", err)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + ": " + msg
		}
		t.Errorf("%s", msg)
	}
}

// AssertError checks if error is not nil
func AssertError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		msg := "Expected error, got nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf("%s", msg)
	}
}

// AssertTypeIs checks if value is of expected type
func AssertTypeIs[T any](t *testing.T, value any, msgAndArgs ...any) T {
	t.Helper()
	result, ok := value.(T)
	if !ok {
		msg := fmt.Sprintf("Expected type %T, got %T", *new(T), value)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + ": " + msg
		}
		t.Errorf("%s", msg)
	}
	return result
}

// EncodeDecodeTestHelper helps with request/response round-trip testing
type EncodeDecodeTestHelper struct {
	t *testing.T
}

// NewEncodeDecodeTestHelper creates a new encode/decode test helper
func NewEncodeDecodeTestHelper(t *testing.T) *EncodeDecodeTestHelper {
	t.Helper()
	return &EncodeDecodeTestHelper{t: t}
}

// TestRequestRoundTrip tests encoding and decoding of a request
//
//revive:disable-next-line:cognitive-complexity
func (h *EncodeDecodeTestHelper) TestRequestRoundTrip(req *NanoRPCRequest, payloadData proto.Message) {
	h.t.Helper()

	// Encode
	encoded, err := EncodeRequest(req, payloadData)
	AssertNoError(h.t, err, "EncodeRequest failed")

	// Decode
	decoded, n, err := DecodeRequest(encoded)
	AssertNoError(h.t, err, "DecodeRequest failed")
	AssertEqual(h.t, len(encoded), n, "DecodeRequest length mismatch")

	// Verify fields
	AssertEqual(h.t, req.RequestId, decoded.RequestId, "RequestId mismatch")
	AssertEqual(h.t, req.RequestType, decoded.RequestType, "RequestType mismatch")

	// Verify path
	h.assertPathEqual(req.PathOneof, decoded.PathOneof)

	// Verify payload
	if payloadData != nil {
		AssertTrue(h.t, len(decoded.Data) > 0, "Expected payload data, got empty")

		// Marshal the original payload for comparison
		expectedData, err := proto.Marshal(payloadData)
		AssertNoError(h.t, err, "Failed to marshal original payload")
		h.assertBytesEqual(expectedData, decoded.Data, "Payload data mismatch")
	}
}

// TestResponseRoundTrip tests encoding and decoding of a response
func (h *EncodeDecodeTestHelper) TestResponseRoundTrip(res *NanoRPCResponse, payloadData proto.Message) {
	h.t.Helper()

	// Encode
	encoded, err := EncodeResponse(res, payloadData)
	AssertNoError(h.t, err, "EncodeResponse failed")

	// Decode
	decoded, n, err := DecodeResponse(encoded)
	AssertNoError(h.t, err, "DecodeResponse failed")
	AssertEqual(h.t, len(encoded), n, "DecodeResponse length mismatch")

	// Verify fields
	AssertEqual(h.t, res.RequestId, decoded.RequestId, "RequestId mismatch")
	AssertEqual(h.t, res.ResponseType, decoded.ResponseType, "ResponseType mismatch")
	AssertEqual(h.t, res.ResponseStatus, decoded.ResponseStatus, "ResponseStatus mismatch")
	AssertEqual(h.t, res.ResponseMessage, decoded.ResponseMessage, "ResponseMessage mismatch")

	// Verify payload
	if payloadData != nil {
		AssertTrue(h.t, len(decoded.Data) > 0, "Expected payload data, got empty")

		// Marshal the original payload for comparison
		expectedData, err := proto.Marshal(payloadData)
		AssertNoError(h.t, err, "Failed to marshal original payload")
		h.assertBytesEqual(expectedData, decoded.Data, "Payload data mismatch")
	}
}

// assertPathEqual compares two PathOneof fields
func (h *EncodeDecodeTestHelper) assertPathEqual(expected, actual isNanoRPCRequest_PathOneof) {
	h.t.Helper()

	if expected == nil && actual == nil {
		return
	}

	AssertNotNil(h.t, expected, "Expected PathOneof should not be nil")
	AssertNotNil(h.t, actual, "Actual PathOneof should not be nil")

	switch expectedPath := expected.(type) {
	case *NanoRPCRequest_Path:
		actualPath := AssertTypeIs[*NanoRPCRequest_Path](h.t, actual, "PathOneof type mismatch")
		AssertEqual(h.t, expectedPath.Path, actualPath.Path, "Path mismatch")
	case *NanoRPCRequest_PathHash:
		actualHash := AssertTypeIs[*NanoRPCRequest_PathHash](h.t, actual, "PathOneof type mismatch")
		AssertEqual(h.t, expectedPath.PathHash, actualHash.PathHash, "PathHash mismatch")
	default:
		h.t.Errorf("Unknown PathOneof type: %T", expected)
	}
}

// assertBytesEqual compares two byte slices
func (h *EncodeDecodeTestHelper) assertBytesEqual(expected, actual []byte, msg string) {
	h.t.Helper()
	AssertEqual(h.t, len(expected), len(actual), "%s: length mismatch", msg)
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			h.t.Errorf("%s: byte mismatch at index %d: expected %02x, got %02x", msg, i, expected[i], actual[i])
			return
		}
	}
}

// AssertRawPayload tests encoding/decoding with raw byte payload
func AssertRawPayload(t *testing.T, testPayload []byte,
	encode func([]byte) ([]byte, error),
	decode func([]byte) ([]byte, int, error)) {
	t.Helper()

	// Encode
	encoded, err := encode(testPayload)
	AssertNoError(t, err, "Encode failed")

	// Decode
	decodedData, n, err := decode(encoded)
	AssertNoError(t, err, "Decode failed")
	AssertEqual(t, len(encoded), n, "Decode length mismatch")

	// Verify payload
	AssertTrue(t, len(decodedData) > 0, "Expected payload data, got empty")
	for i, b := range testPayload {
		if i >= len(decodedData) || decodedData[i] != b {
			t.Errorf("Payload mismatch at byte %d: expected %02x, got %02x", i, b, decodedData[i])
			break
		}
	}
}
