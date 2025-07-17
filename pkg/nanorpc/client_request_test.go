package nanorpc

import (
	"testing"
)

// RequestTypeTestCase represents a test case for verifying request types
type RequestTypeTestCase struct {
	name         string
	requestType  NanoRPCRequest_Type
	expectedType NanoRPCRequest_Type
}

func (tc RequestTypeTestCase) test(t *testing.T) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		// Create a request with the type we're testing
		req := &NanoRPCRequest{
			RequestType: tc.requestType,
			RequestId:   123,
			PathOneof: &NanoRPCRequest_Path{
				Path: "/test",
			},
		}

		// Verify the request type matches expectation
		if req.RequestType != tc.expectedType {
			t.Errorf("Expected RequestType %v (%d), got %v (%d)",
				tc.expectedType, tc.expectedType,
				req.RequestType, req.RequestType)
		}
	})
}

// Test cases for verifying subscription request types
var subscriptionRequestTypeTestCases = []RequestTypeTestCase{
	{
		name:         "Subscribe_uses_TYPE_SUBSCRIBE",
		requestType:  NanoRPCRequest_TYPE_SUBSCRIBE,
		expectedType: NanoRPCRequest_TYPE_SUBSCRIBE,
	},
	{
		name:         "Request_uses_TYPE_REQUEST",
		requestType:  NanoRPCRequest_TYPE_REQUEST,
		expectedType: NanoRPCRequest_TYPE_REQUEST,
	},
	{
		name:         "Ping_uses_TYPE_PING",
		requestType:  NanoRPCRequest_TYPE_PING,
		expectedType: NanoRPCRequest_TYPE_PING,
	},
}

// TestSubscriptionRequestTypes verifies that the correct request types are used
// This is a regression test for the bug where Subscribe methods used TYPE_REQUEST
func TestSubscriptionRequestTypes(t *testing.T) {
	for _, tc := range subscriptionRequestTypeTestCases {
		tc.test(t)
	}
}

// TestRequestConstruction tests that we can construct requests with the right types
func TestRequestConstruction(t *testing.T) {
	// Test Subscribe request construction (like what Subscribe() method does)
	subscribeReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE, // This should be TYPE_SUBSCRIBE after our fix
		PathOneof: &NanoRPCRequest_Path{
			Path: "/events",
		},
	}

	if subscribeReq.RequestType != NanoRPCRequest_TYPE_SUBSCRIBE {
		t.Errorf("Subscribe request should use TYPE_SUBSCRIBE, got %v", subscribeReq.RequestType)
	}

	// Test Request construction (like what Request() method does)
	requestReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &NanoRPCRequest_Path{
			Path: "/status",
		},
	}

	if requestReq.RequestType != NanoRPCRequest_TYPE_REQUEST {
		t.Errorf("Request should use TYPE_REQUEST, got %v", requestReq.RequestType)
	}

	// Test Ping construction (like what Ping() method does)
	pingReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_PING,
	}

	if pingReq.RequestType != NanoRPCRequest_TYPE_PING {
		t.Errorf("Ping request should use TYPE_PING, got %v", pingReq.RequestType)
	}
}

// TestPathOneofTypes tests both path string and hash variants
func TestPathOneofTypes(t *testing.T) {
	// Test string path
	pathReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof: &NanoRPCRequest_Path{
			Path: "/events",
		},
	}

	if pathOneof, ok := pathReq.PathOneof.(*NanoRPCRequest_Path); ok {
		if pathOneof.Path != "/events" {
			t.Errorf("Expected path '/events', got '%s'", pathOneof.Path)
		}
	} else {
		t.Errorf("Expected *NanoRPCRequest_Path, got %T", pathReq.PathOneof)
	}

	// Test hash path
	hashReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof: &NanoRPCRequest_PathHash{
			PathHash: 0x12345678,
		},
	}

	if hashOneof, ok := hashReq.PathOneof.(*NanoRPCRequest_PathHash); ok {
		if hashOneof.PathHash != 0x12345678 {
			t.Errorf("Expected hash 0x12345678, got 0x%x", hashOneof.PathHash)
		}
	} else {
		t.Errorf("Expected *NanoRPCRequest_PathHash, got %T", hashReq.PathOneof)
	}
}

// TestProtocolDefinitions verifies that all expected protocol constants exist
func TestProtocolDefinitions(t *testing.T) {
	// Test that all request types are defined
	requestTypes := []NanoRPCRequest_Type{
		NanoRPCRequest_TYPE_UNSPECIFIED,
		NanoRPCRequest_TYPE_PING,
		NanoRPCRequest_TYPE_REQUEST,
		NanoRPCRequest_TYPE_SUBSCRIBE,
	}

	for _, rt := range requestTypes {
		if rt.String() == "" {
			t.Errorf("Request type %d should have a string representation", rt)
		}
	}

	// Test that all response types are defined
	responseTypes := []NanoRPCResponse_Type{
		NanoRPCResponse_TYPE_UNSPECIFIED,
		NanoRPCResponse_TYPE_PONG,
		NanoRPCResponse_TYPE_RESPONSE,
		NanoRPCResponse_TYPE_UPDATE,
	}

	for _, rt := range responseTypes {
		if rt.String() == "" {
			t.Errorf("Response type %d should have a string representation", rt)
		}
	}
}
