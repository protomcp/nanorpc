package nanorpc

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"
)

type protocolRequestTypeTestCase struct {
	name  string
	value NanoRPCRequest_Type
}

func (tc protocolRequestTypeTestCase) test(t *testing.T) {
	req := &NanoRPCRequest{RequestType: tc.value}
	AssertEqual(t, tc.value, req.RequestType, "Request type mismatch")
}

func newProtocolRequestTypeTestCase(name string, value NanoRPCRequest_Type) protocolRequestTypeTestCase {
	return protocolRequestTypeTestCase{name, value}
}

func protocolRequestTypeTestCases() []protocolRequestTypeTestCase {
	return S(
		newProtocolRequestTypeTestCase("TYPE_UNSPECIFIED", NanoRPCRequest_TYPE_UNSPECIFIED),
		newProtocolRequestTypeTestCase("TYPE_PING", NanoRPCRequest_TYPE_PING),
		newProtocolRequestTypeTestCase("TYPE_REQUEST", NanoRPCRequest_TYPE_REQUEST),
		newProtocolRequestTypeTestCase("TYPE_SUBSCRIBE", NanoRPCRequest_TYPE_SUBSCRIBE),
	)
}

func newResponseTypeTestCase(name string, value NanoRPCResponse_Type) responseTypeTestCase {
	return responseTypeTestCase{name, value}
}

func responseTypeTestCases() []responseTypeTestCase {
	return S(
		newResponseTypeTestCase("TYPE_UNSPECIFIED", NanoRPCResponse_TYPE_UNSPECIFIED),
		newResponseTypeTestCase("TYPE_PONG", NanoRPCResponse_TYPE_PONG),
		newResponseTypeTestCase("TYPE_RESPONSE", NanoRPCResponse_TYPE_RESPONSE),
		newResponseTypeTestCase("TYPE_UPDATE", NanoRPCResponse_TYPE_UPDATE),
	)
}

// TestProtocolTypes verifies all request and response types are properly defined
func TestProtocolTypes(t *testing.T) {
	for _, tc := range protocolRequestTypeTestCases() {
		t.Run("Request_"+tc.name, tc.test)
	}
	for _, tc := range responseTypeTestCases() {
		t.Run("Response_"+tc.name, tc.test)
	}
}

type responseTypeTestCase struct {
	name  string
	value NanoRPCResponse_Type
}

func (tc responseTypeTestCase) test(t *testing.T) {
	res := &NanoRPCResponse{ResponseType: tc.value}
	AssertEqual(t, tc.value, res.ResponseType, "Response type mismatch")
}

type statusCodeTestCase struct {
	name    string
	value   NanoRPCResponse_Status
	isError bool
}

func (tc statusCodeTestCase) test(t *testing.T) {
	res := &NanoRPCResponse{
		ResponseStatus:  tc.value,
		ResponseMessage: "test message",
	}

	err := ResponseAsError(res)
	if tc.isError {
		if err == nil {
			t.Errorf("Expected error for status %v", tc.value)
		}
	} else if err != nil {
		t.Errorf("Expected no error for status %v, got: %v", tc.value, err)
	}
}

func newStatusCodeTestCase(name string, value NanoRPCResponse_Status, isError bool) statusCodeTestCase {
	return statusCodeTestCase{name, value, isError}
}

func statusCodeTestCases() []statusCodeTestCase {
	return S(
		newStatusCodeTestCase("STATUS_UNSPECIFIED", NanoRPCResponse_STATUS_UNSPECIFIED, true),
		newStatusCodeTestCase("STATUS_OK", NanoRPCResponse_STATUS_OK, false),
		newStatusCodeTestCase("STATUS_NOT_FOUND", NanoRPCResponse_STATUS_NOT_FOUND, true),
		newStatusCodeTestCase("STATUS_NOT_AUTHORIZED", NanoRPCResponse_STATUS_NOT_AUTHORIZED, true),
		newStatusCodeTestCase("STATUS_INTERNAL_ERROR", NanoRPCResponse_STATUS_INTERNAL_ERROR, true),
	)
}

// TestStatusCodes verifies all status codes are properly defined
func TestStatusCodes(t *testing.T) {
	for _, tc := range statusCodeTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type extendedRequestTestCase struct {
	request *NanoRPCRequest
	name    string
}

func (tc extendedRequestTestCase) test(t *testing.T) {
	helper := NewEncodeDecodeTestHelper(t)

	// Test with no payload
	helper.TestRequestRoundTrip(tc.request, nil)

	// Test with raw byte payload
	testPayload := []byte("test payload")
	reqWithPayload, ok := proto.Clone(tc.request).(*NanoRPCRequest)
	if !ok {
		t.Fatalf("Failed to clone request")
	}
	reqWithPayload.Data = testPayload

	AssertRawPayload(t, testPayload,
		func([]byte) ([]byte, error) {
			return EncodeRequest(reqWithPayload, nil)
		},
		func(encoded []byte) ([]byte, int, error) {
			decoded, n, err := DecodeRequest(encoded)
			if err != nil {
				return nil, n, err
			}
			return decoded.Data, n, nil
		},
	)
}

func newExtendedRequestTestCase(name string, request *NanoRPCRequest) extendedRequestTestCase {
	return extendedRequestTestCase{
		name:    name,
		request: request,
	}
}

func extendedRequestTestCases() []extendedRequestTestCase {
	return S(
		newExtendedRequestTestCase("ping_request", &NanoRPCRequest{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_PING,
		}),
		newExtendedRequestTestCase("request_with_path", &NanoRPCRequest{
			RequestId:   456,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			PathOneof: &NanoRPCRequest_Path{
				Path: "/test/path",
			},
		}),
		newExtendedRequestTestCase("request_with_hash", &NanoRPCRequest{
			RequestId:   789,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			PathOneof: &NanoRPCRequest_PathHash{
				PathHash: 0x12345678,
			},
		}),
		newExtendedRequestTestCase("subscribe_with_path", &NanoRPCRequest{
			RequestId:   101,
			RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
			PathOneof: &NanoRPCRequest_Path{
				Path: "/events",
			},
		}),
		newExtendedRequestTestCase("subscribe_with_hash", &NanoRPCRequest{
			RequestId:   102,
			RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
			PathOneof: &NanoRPCRequest_PathHash{
				PathHash: 0xABCDEF01,
			},
		}),
	)
}

// TestEncodeDecodeRequestExtended extends the existing test with more request types
func TestEncodeDecodeRequestExtended(t *testing.T) {
	for _, tc := range extendedRequestTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type responseTestCase struct {
	response *NanoRPCResponse
	name     string
}

func (tc responseTestCase) test(t *testing.T) {
	helper := NewEncodeDecodeTestHelper(t)

	// Test with no payload
	helper.TestResponseRoundTrip(tc.response, nil)

	// Test with simple payload (raw bytes in response Data field)
	testPayload := []byte("response payload")
	resWithPayload, ok := proto.Clone(tc.response).(*NanoRPCResponse)
	if !ok {
		t.Fatalf("Failed to clone response")
	}
	resWithPayload.Data = testPayload

	AssertRawPayload(t, testPayload,
		func([]byte) ([]byte, error) {
			return EncodeResponse(resWithPayload, nil)
		},
		func(encoded []byte) ([]byte, int, error) {
			decoded, n, err := DecodeResponse(encoded)
			if err != nil {
				return nil, n, err
			}
			return decoded.Data, n, nil
		},
	)
}

func newResponseTestCase(name string, response *NanoRPCResponse) responseTestCase {
	return responseTestCase{
		name:     name,
		response: response,
	}
}

func responseTestCases() []responseTestCase {
	return S(
		newResponseTestCase("pong_response", &NanoRPCResponse{
			RequestId:      123,
			ResponseType:   NanoRPCResponse_TYPE_PONG,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		}),
		newResponseTestCase("ok_response", &NanoRPCResponse{
			RequestId:      456,
			ResponseType:   NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		}),
		newResponseTestCase("not_found_response", &NanoRPCResponse{
			RequestId:       789,
			ResponseType:    NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus:  NanoRPCResponse_STATUS_NOT_FOUND,
			ResponseMessage: "Path not found",
		}),
		newResponseTestCase("update_response", &NanoRPCResponse{
			RequestId:      101,
			ResponseType:   NanoRPCResponse_TYPE_UPDATE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		}),
	)
}

// TestEncodeDecodeResponse tests response encoding/decoding
func TestEncodeDecodeResponse(t *testing.T) {
	for _, tc := range responseTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// TestSplitMessage tests message framing
func TestSplitMessage(t *testing.T) {
	// Create a test request
	req := &NanoRPCRequest{
		RequestId:   123,
		RequestType: NanoRPCRequest_TYPE_PING,
	}

	// Encode it
	encoded, err := EncodeRequest(req, nil)
	AssertNoError(t, err, "Failed to encode request")

	// Test Split function
	advance, msg, err := Split(encoded, true)
	AssertNoError(t, err, "Split failed")
	AssertEqual(t, len(encoded), advance, "Split advance mismatch")
	AssertTrue(t, bytes.Equal(msg, encoded), "Split message mismatch: expected %v, got %v", encoded, msg)

	// Test partial data (should return 0, nil, nil when atEOF=false)
	partialData := encoded[:len(encoded)/2]
	advance, msg, err = Split(partialData, false)
	AssertNoError(t, err, "Split should not error on partial data when atEOF=false")
	AssertEqual(t, 0, advance, "Split should return 0 advance for partial data")
	AssertEqual(t, 0, len(msg), "Split should return empty/nil message for partial data")
}

type errorHandlingTestCase struct {
	response *NanoRPCResponse
	checkFn  func(error) bool
	name     string
}

func (tc errorHandlingTestCase) test(t *testing.T) {
	err := ResponseAsError(tc.response)
	AssertTrue(t, tc.checkFn(err), "Error check failed for %s: got %v", tc.name, err)
}

func newErrorHandlingTestCase(name string, response *NanoRPCResponse, checkFn func(error) bool) errorHandlingTestCase {
	return errorHandlingTestCase{
		name:     name,
		response: response,
		checkFn:  checkFn,
	}
}

func errorHandlingTestCases() []errorHandlingTestCase {
	return S(
		newErrorHandlingTestCase("nil response", nil, IsNoResponse),
		newErrorHandlingTestCase("not found", &NanoRPCResponse{
			ResponseStatus: NanoRPCResponse_STATUS_NOT_FOUND,
		}, IsNotFound),
		newErrorHandlingTestCase("not authorized", &NanoRPCResponse{
			ResponseStatus: NanoRPCResponse_STATUS_NOT_AUTHORIZED,
		}, IsNotAuthorized),
		newErrorHandlingTestCase("ok status", &NanoRPCResponse{
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		}, func(err error) bool { return err == nil }),
	)
}

// TestErrorHandling tests error conversion and helper functions
func TestErrorHandling(t *testing.T) {
	for _, tc := range errorHandlingTestCases() {
		t.Run(tc.name, tc.test)
	}
}
