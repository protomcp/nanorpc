package nanorpc

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"
)

// TestProtocolTypes verifies all request and response types are properly defined
func TestProtocolTypes(t *testing.T) {
	requestTypes := []struct {
		name  string
		value NanoRPCRequest_Type
	}{
		{"TYPE_UNSPECIFIED", NanoRPCRequest_TYPE_UNSPECIFIED},
		{"TYPE_PING", NanoRPCRequest_TYPE_PING},
		{"TYPE_REQUEST", NanoRPCRequest_TYPE_REQUEST},
		{"TYPE_SUBSCRIBE", NanoRPCRequest_TYPE_SUBSCRIBE},
	}

	for _, rt := range requestTypes {
		t.Run("Request_"+rt.name, func(t *testing.T) {
			req := &NanoRPCRequest{RequestType: rt.value}
			if req.RequestType != rt.value {
				t.Errorf("Request type mismatch: expected %v, got %v", rt.value, req.RequestType)
			}
		})
	}

	responseTypes := []struct {
		name  string
		value NanoRPCResponse_Type
	}{
		{"TYPE_UNSPECIFIED", NanoRPCResponse_TYPE_UNSPECIFIED},
		{"TYPE_PONG", NanoRPCResponse_TYPE_PONG},
		{"TYPE_RESPONSE", NanoRPCResponse_TYPE_RESPONSE},
		{"TYPE_UPDATE", NanoRPCResponse_TYPE_UPDATE},
	}

	for _, rt := range responseTypes {
		t.Run("Response_"+rt.name, func(t *testing.T) {
			res := &NanoRPCResponse{ResponseType: rt.value}
			if res.ResponseType != rt.value {
				t.Errorf("Response type mismatch: expected %v, got %v", rt.value, res.ResponseType)
			}
		})
	}
}

// TestStatusCodes verifies all status codes are properly defined
func TestStatusCodes(t *testing.T) {
	statusCodes := []struct {
		name    string
		value   NanoRPCResponse_Status
		isError bool
	}{
		{"STATUS_UNSPECIFIED", NanoRPCResponse_STATUS_UNSPECIFIED, true},
		{"STATUS_OK", NanoRPCResponse_STATUS_OK, false},
		{"STATUS_NOT_FOUND", NanoRPCResponse_STATUS_NOT_FOUND, true},
		{"STATUS_NOT_AUTHORIZED", NanoRPCResponse_STATUS_NOT_AUTHORIZED, true},
		{"STATUS_INTERNAL_ERROR", NanoRPCResponse_STATUS_INTERNAL_ERROR, true},
	}

	for _, sc := range statusCodes {
		t.Run(sc.name, func(t *testing.T) {
			res := &NanoRPCResponse{
				ResponseStatus:  sc.value,
				ResponseMessage: "test message",
			}

			err := ResponseAsError(res)
			if sc.isError && err == nil {
				t.Errorf("Expected error for status %v, got nil", sc.value)
			}
			if !sc.isError && err != nil {
				t.Errorf("Expected no error for status %v, got %v", sc.value, err)
			}
		})
	}
}

// ExtendedRequestTestCase represents a test case for extended request encoding/decoding
type ExtendedRequestTestCase struct {
	name    string
	request *NanoRPCRequest
}

func (tc ExtendedRequestTestCase) test(t *testing.T) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		// Test with no payload
		tc.testWithPayload(t, nil)

		// Test with payload - simple string data instead of proto message
		tc.testWithPayload(t, []byte("test payload"))
	})
}

func (tc ExtendedRequestTestCase) testWithPayload(t *testing.T, payload []byte) {
	t.Helper()
	// Set payload if provided
	req := proto.Clone(tc.request).(*NanoRPCRequest)
	req.Data = payload

	// Encode
	encoded, err := EncodeRequest(req, nil)
	if err != nil {
		t.Fatalf("EncodeRequest failed: %v", err)
	}

	// Decode
	decoded, n, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeRequest failed: %v", err)
	}

	if n != len(encoded) {
		t.Errorf("DecodeRequest length mismatch: expected %d, got %d", len(encoded), n)
	}

	// Verify request fields
	if decoded.RequestId != req.RequestId {
		t.Errorf("RequestId mismatch: expected %d, got %d", req.RequestId, decoded.RequestId)
	}

	if decoded.RequestType != req.RequestType {
		t.Errorf("RequestType mismatch: expected %v, got %v", req.RequestType, decoded.RequestType)
	}

	// Verify path information
	if req.PathOneof != nil {
		if decoded.PathOneof == nil {
			t.Errorf("PathOneof is nil in decoded request")
		} else {
			switch originalPath := req.PathOneof.(type) {
			case *NanoRPCRequest_Path:
				if decodedPath, ok := decoded.PathOneof.(*NanoRPCRequest_Path); ok {
					if decodedPath.Path != originalPath.Path {
						t.Errorf("Path mismatch: expected %s, got %s", originalPath.Path, decodedPath.Path)
					}
				} else {
					t.Errorf("PathOneof type mismatch: expected Path, got %T", decoded.PathOneof)
				}
			case *NanoRPCRequest_PathHash:
				if decodedHash, ok := decoded.PathOneof.(*NanoRPCRequest_PathHash); ok {
					if decodedHash.PathHash != originalPath.PathHash {
						t.Errorf("PathHash mismatch: expected %d, got %d", originalPath.PathHash, decodedHash.PathHash)
					}
				} else {
					t.Errorf("PathOneof type mismatch: expected PathHash, got %T", decoded.PathOneof)
				}
			}
		}
	}

	// Verify payload
	if payload != nil {
		if len(decoded.Data) == 0 {
			t.Errorf("Expected payload data, got empty")
		} else if !bytes.Equal(decoded.Data, payload) {
			t.Errorf("Payload mismatch: expected %v, got %v", payload, decoded.Data)
		}
	}
}

// Test cases for extended request encoding/decoding
var extendedRequestTestCases = []ExtendedRequestTestCase{
	{
		name: "ping_request",
		request: &NanoRPCRequest{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_PING,
		},
	},
	{
		name: "request_with_path",
		request: &NanoRPCRequest{
			RequestId:   456,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			PathOneof: &NanoRPCRequest_Path{
				Path: "/test/path",
			},
		},
	},
	{
		name: "request_with_hash",
		request: &NanoRPCRequest{
			RequestId:   789,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			PathOneof: &NanoRPCRequest_PathHash{
				PathHash: 0x12345678,
			},
		},
	},
	{
		name: "subscribe_with_path",
		request: &NanoRPCRequest{
			RequestId:   101,
			RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
			PathOneof: &NanoRPCRequest_Path{
				Path: "/events",
			},
		},
	},
	{
		name: "subscribe_with_hash",
		request: &NanoRPCRequest{
			RequestId:   102,
			RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
			PathOneof: &NanoRPCRequest_PathHash{
				PathHash: 0xABCDEF01,
			},
		},
	},
}

// TestEncodeDecodeRequestExtended extends the existing test with more request types
func TestEncodeDecodeRequestExtended(t *testing.T) {
	for _, tc := range extendedRequestTestCases {
		tc.test(t)
	}
}

// ResponseTestCase represents a test case for response encoding/decoding
type ResponseTestCase struct {
	name     string
	response *NanoRPCResponse
}

func (tc ResponseTestCase) test(t *testing.T) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		// Test with no payload
		tc.testWithPayload(t, nil)

		// Test with payload
		tc.testWithPayload(t, []byte("response payload"))
	})
}

func (tc ResponseTestCase) testWithPayload(t *testing.T, payload []byte) {
	t.Helper()
	// Set payload if provided
	res := proto.Clone(tc.response).(*NanoRPCResponse)
	res.Data = payload

	// Encode
	encoded, err := EncodeResponse(res, nil)
	if err != nil {
		t.Fatalf("EncodeResponse failed: %v", err)
	}

	// Decode
	decoded, n, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if n != len(encoded) {
		t.Errorf("DecodeResponse length mismatch: expected %d, got %d", len(encoded), n)
	}

	// Verify response fields
	if decoded.RequestId != res.RequestId {
		t.Errorf("RequestId mismatch: expected %d, got %d", res.RequestId, decoded.RequestId)
	}

	if decoded.ResponseType != res.ResponseType {
		t.Errorf("ResponseType mismatch: expected %v, got %v", res.ResponseType, decoded.ResponseType)
	}

	if decoded.ResponseStatus != res.ResponseStatus {
		t.Errorf("ResponseStatus mismatch: expected %v, got %v", res.ResponseStatus, decoded.ResponseStatus)
	}

	if decoded.ResponseMessage != res.ResponseMessage {
		t.Errorf("ResponseMessage mismatch: expected %s, got %s", res.ResponseMessage, decoded.ResponseMessage)
	}

	// Verify payload
	if payload != nil {
		if len(decoded.Data) == 0 {
			t.Errorf("Expected payload data, got empty")
		} else if !bytes.Equal(decoded.Data, payload) {
			t.Errorf("Payload mismatch: expected %v, got %v", payload, decoded.Data)
		}
	}
}

// Test cases for response encoding/decoding
var responseTestCases = []ResponseTestCase{
	{
		name: "pong_response",
		response: &NanoRPCResponse{
			RequestId:      123,
			ResponseType:   NanoRPCResponse_TYPE_PONG,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		},
	},
	{
		name: "ok_response",
		response: &NanoRPCResponse{
			RequestId:      456,
			ResponseType:   NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		},
	},
	{
		name: "not_found_response",
		response: &NanoRPCResponse{
			RequestId:       789,
			ResponseType:    NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus:  NanoRPCResponse_STATUS_NOT_FOUND,
			ResponseMessage: "Path not found",
		},
	},
	{
		name: "update_response",
		response: &NanoRPCResponse{
			RequestId:      101,
			ResponseType:   NanoRPCResponse_TYPE_UPDATE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		},
	},
}

// TestEncodeDecodeResponse tests response encoding/decoding
func TestEncodeDecodeResponse(t *testing.T) {
	for _, tc := range responseTestCases {
		tc.test(t)
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
	if err != nil {
		t.Fatalf("Failed to encode request: %v", err)
	}

	// Test Split function
	advance, msg, err := Split(encoded, true)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if advance != len(encoded) {
		t.Errorf("Split advance mismatch: expected %d, got %d", len(encoded), advance)
	}

	if !bytes.Equal(msg, encoded) {
		t.Errorf("Split message mismatch")
	}

	// Test partial data (should return 0, nil, nil when atEOF=false)
	partialData := encoded[:len(encoded)/2]
	advance, msg, err = Split(partialData, false)
	if err != nil {
		t.Errorf("Split should not error on partial data when atEOF=false: %v", err)
	}
	if advance != 0 || msg != nil {
		t.Errorf("Split should return 0,nil,nil for partial data when atEOF=false, got %d,%v,%v", advance, msg, err)
	}
}

// TestErrorHandling tests error conversion and helper functions
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		response *NanoRPCResponse
		checkFn  func(error) bool
	}{
		{
			name:     "nil response",
			response: nil,
			checkFn:  IsNoResponse,
		},
		{
			name: "not found",
			response: &NanoRPCResponse{
				ResponseStatus: NanoRPCResponse_STATUS_NOT_FOUND,
			},
			checkFn: IsNotFound,
		},
		{
			name: "not authorized",
			response: &NanoRPCResponse{
				ResponseStatus: NanoRPCResponse_STATUS_NOT_AUTHORIZED,
			},
			checkFn: IsNotAuthorized,
		},
		{
			name: "ok status",
			response: &NanoRPCResponse{
				ResponseStatus: NanoRPCResponse_STATUS_OK,
			},
			checkFn: func(err error) bool { return err == nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ResponseAsError(tt.response)
			if !tt.checkFn(err) {
				t.Errorf("Error check failed for %s: got %v", tt.name, err)
			}
		})
	}
}
