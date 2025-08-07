package nanorpc

import (
	"sync"
	"testing"

	"darvaza.org/core"
	"google.golang.org/protobuf/proto"
)

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
	if !core.AssertNoError(h.t, err, "EncodeRequest") {
		h.t.FailNow()
	}

	// Decode
	decoded, n, err := DecodeRequest(encoded)
	if !core.AssertNoError(h.t, err, "DecodeRequest") {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, len(encoded), n, "decode length") {
		h.t.FailNow()
	}

	// Verify fields
	if !core.AssertEqual(h.t, req.RequestId, decoded.RequestId, "RequestId") {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, req.RequestType, decoded.RequestType, "RequestType") {
		h.t.FailNow()
	}

	// Verify path
	h.assertPathEqual(req.PathOneof, decoded.PathOneof)

	// Verify payload
	if payloadData != nil {
		if !core.AssertTrue(h.t, len(decoded.Data) > 0, "payload data") {
			h.t.FailNow()
		}

		// Marshal the original payload for comparison
		expectedData, err := proto.Marshal(payloadData)
		if !core.AssertNoError(h.t, err, "marshal payload") {
			h.t.FailNow()
		}
		h.assertBytesEqual(expectedData, decoded.Data, "Payload data mismatch")
	}
}

// TestResponseRoundTrip tests encoding and decoding of a response
func (h *EncodeDecodeTestHelper) TestResponseRoundTrip(res *NanoRPCResponse, payloadData proto.Message) {
	h.t.Helper()

	encoded := h.encodeResponse(res, payloadData)
	decoded := h.decodeResponse(encoded)
	h.verifyResponseFields(res, decoded)
	h.verifyResponsePayload(decoded, payloadData)
}

func (h *EncodeDecodeTestHelper) encodeResponse(res *NanoRPCResponse, payloadData proto.Message) []byte {
	h.t.Helper()
	encoded, err := EncodeResponse(res, payloadData)
	if !core.AssertNoError(h.t, err, "EncodeResponse") {
		h.t.FailNow()
	}
	return encoded
}

func (h *EncodeDecodeTestHelper) decodeResponse(encoded []byte) *NanoRPCResponse {
	h.t.Helper()
	decoded, n, err := DecodeResponse(encoded)
	if !core.AssertNoError(h.t, err, "DecodeResponse") {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, len(encoded), n, "decode length") {
		h.t.FailNow()
	}
	return decoded
}

func (h *EncodeDecodeTestHelper) verifyResponseFields(expected, actual *NanoRPCResponse) {
	h.t.Helper()
	if !core.AssertEqual(h.t, expected.RequestId, actual.RequestId, "RequestId") {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, expected.ResponseType, actual.ResponseType, "ResponseType") {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, expected.ResponseStatus, actual.ResponseStatus, "ResponseStatus") {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, expected.ResponseMessage, actual.ResponseMessage, "ResponseMessage") {
		h.t.FailNow()
	}
}

func (h *EncodeDecodeTestHelper) verifyResponsePayload(decoded *NanoRPCResponse, payloadData proto.Message) {
	h.t.Helper()
	if payloadData == nil {
		return
	}

	if !core.AssertTrue(h.t, len(decoded.Data) > 0, "payload data") {
		h.t.FailNow()
	}

	expectedData, err := proto.Marshal(payloadData)
	if !core.AssertNoError(h.t, err, "marshal payload") {
		h.t.FailNow()
	}
	h.assertBytesEqual(expectedData, decoded.Data, "Payload data mismatch")
}

// assertPathEqual compares two PathOneof fields
func (h *EncodeDecodeTestHelper) assertPathEqual(expected, actual isNanoRPCRequest_PathOneof) {
	h.t.Helper()

	if expected == nil && actual == nil {
		return
	}

	h.assertPathOneofNotNil(expected, actual)
	h.assertPathOneofValues(expected, actual)
}

func (h *EncodeDecodeTestHelper) assertPathOneofNotNil(expected, actual isNanoRPCRequest_PathOneof) {
	h.t.Helper()
	if !core.AssertNotNil(h.t, expected, "PathOneof") {
		h.t.FailNow()
	}
	if !core.AssertNotNil(h.t, actual, "actual PathOneof") {
		h.t.FailNow()
	}
}

func (h *EncodeDecodeTestHelper) assertPathOneofValues(expected, actual isNanoRPCRequest_PathOneof) {
	h.t.Helper()
	switch expectedPath := expected.(type) {
	case *NanoRPCRequest_Path:
		h.assertPathString(expectedPath, actual)
	case *NanoRPCRequest_PathHash:
		h.assertPathHash(expectedPath, actual)
	default:
		h.t.Errorf("Unknown PathOneof type: %T", expected)
	}
}

func (h *EncodeDecodeTestHelper) assertPathString(expectedPath *NanoRPCRequest_Path,
	actual isNanoRPCRequest_PathOneof) {
	h.t.Helper()
	actualPath, ok := core.AssertTypeIs[*NanoRPCRequest_Path](h.t, actual, "PathOneof type")
	if !ok {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, expectedPath.Path, actualPath.Path, "Path") {
		h.t.FailNow()
	}
}

func (h *EncodeDecodeTestHelper) assertPathHash(expectedPath *NanoRPCRequest_PathHash,
	actual isNanoRPCRequest_PathOneof) {
	h.t.Helper()
	actualHash, ok := core.AssertTypeIs[*NanoRPCRequest_PathHash](h.t, actual, "PathOneof type")
	if !ok {
		h.t.FailNow()
	}
	if !core.AssertEqual(h.t, expectedPath.PathHash, actualHash.PathHash, "PathHash") {
		h.t.FailNow()
	}
}

// assertBytesEqual compares two byte slices
func (h *EncodeDecodeTestHelper) assertBytesEqual(expected, actual []byte, msg string) {
	h.t.Helper()
	if !core.AssertEqual(h.t, len(expected), len(actual), msg) {
		h.t.FailNow()
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			h.t.Errorf("%s: byte mismatch at index %d: expected %02x, got %02x", msg, i, expected[i], actual[i])
			return
		}
	}
}

// AssertRawPayload tests encoding/decoding with raw byte payload
func AssertRawPayload(t core.T, testPayload []byte,
	encode func([]byte) ([]byte, error),
	decode func([]byte) ([]byte, int, error)) {
	t.Helper()

	// Encode
	encoded, err := encode(testPayload)
	if !core.AssertNoError(t, err, "Encode") {
		return
	}

	// Decode
	decodedData, n, err := decode(encoded)
	if !core.AssertNoError(t, err, "Decode") {
		return
	}
	if !core.AssertEqual(t, len(encoded), n, "decode length") {
		return
	}

	// Verify payload
	if !core.AssertTrue(t, len(decodedData) > 0, "payload data") {
		return
	}
	core.AssertSliceEqual(t, testPayload, decodedData, "payload data")
}

// ConcurrentTestHelper helps with concurrent test execution
type ConcurrentTestHelper struct {
	t           *testing.T
	results     []any
	errors      []error
	wg          sync.WaitGroup
	mutex       sync.Mutex
	numRoutines int
}

// NewConcurrentTestHelper creates a new concurrent test helper
func NewConcurrentTestHelper(t *testing.T, numRoutines int) *ConcurrentTestHelper {
	t.Helper()
	return &ConcurrentTestHelper{
		t:           t,
		numRoutines: numRoutines,
		results:     make([]any, numRoutines),
		errors:      make([]error, numRoutines),
	}
}

// Run executes the test function concurrently
func (h *ConcurrentTestHelper) Run(testFunc func(int) (any, error)) {
	h.t.Helper()
	h.wg.Add(h.numRoutines)

	for i := 0; i < h.numRoutines; i++ {
		go func(idx int) {
			defer h.wg.Done()
			result, err := testFunc(idx)

			h.mutex.Lock()
			h.results[idx] = result
			h.errors[idx] = err
			h.mutex.Unlock()
		}(i)
	}

	h.wg.Wait()
}

// GetResults returns all results and errors
func (h *ConcurrentTestHelper) GetResults() ([]any, []error) {
	h.t.Helper()
	return h.results, h.errors
}

// AssertNoErrors checks that no goroutines returned errors
func (h *ConcurrentTestHelper) AssertNoErrors() {
	h.t.Helper()
	for i, err := range h.errors {
		if err != nil {
			h.t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}
}

// GetResult returns the result at index with type assertion
func GetResult[T any](values []any, index int) (T, bool) {
	if index < 0 || index >= len(values) {
		var zero T
		return zero, false
	}
	result, ok := values[index].(T)
	return result, ok
}
