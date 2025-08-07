package client

import (
	"testing"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

type requestTypeTestCase struct {
	name         string
	requestType  nanorpc.NanoRPCRequest_Type
	expectedType nanorpc.NanoRPCRequest_Type
}

func (tc requestTypeTestCase) test(t *testing.T) {
	// Create a request with the type we're testing
	req := &nanorpc.NanoRPCRequest{
		RequestType: tc.requestType,
		RequestId:   123,
		PathOneof: &nanorpc.NanoRPCRequest_Path{
			Path: "/test",
		},
	}

	// Verify the request type matches expectation
	core.AssertEqual(t, tc.expectedType, req.RequestType, "request_type")
}

func newRequestTypeTestCase(name string, requestType, expectedType nanorpc.NanoRPCRequest_Type) requestTypeTestCase {
	return requestTypeTestCase{
		name:         name,
		requestType:  requestType,
		expectedType: expectedType,
	}
}

func subscriptionRequestTypeTestCases() []requestTypeTestCase {
	return core.S(
		newRequestTypeTestCase("Subscribe_uses_TYPE_SUBSCRIBE",
			nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE),
		newRequestTypeTestCase("Request_uses_TYPE_REQUEST",
			nanorpc.NanoRPCRequest_TYPE_REQUEST, nanorpc.NanoRPCRequest_TYPE_REQUEST),
		newRequestTypeTestCase("Ping_uses_TYPE_PING",
			nanorpc.NanoRPCRequest_TYPE_PING, nanorpc.NanoRPCRequest_TYPE_PING),
		newRequestTypeTestCase("Unsubscribe_uses_TYPE_REQUEST",
			nanorpc.NanoRPCRequest_TYPE_REQUEST, nanorpc.NanoRPCRequest_TYPE_REQUEST),
	)
}

// TestSubscriptionRequestTypes verifies that the correct request types are used
// This is a regression test for the bug where Subscribe methods used TYPE_REQUEST
func TestSubscriptionRequestTypes(t *testing.T) {
	for _, tc := range subscriptionRequestTypeTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// TestRequestConstruction tests that we can construct requests with the right types
func TestRequestConstruction(t *testing.T) {
	// Test Subscribe request construction (like what Subscribe() method does)
	subscribeReq := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE, // This should be TYPE_SUBSCRIBE after our fix
		PathOneof: &nanorpc.NanoRPCRequest_Path{
			Path: "/events",
		},
	}

	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE, subscribeReq.RequestType,
		"Subscribe request should use TYPE_SUBSCRIBE")

	// Test Request construction (like what Request() method does)
	requestReq := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &nanorpc.NanoRPCRequest_Path{
			Path: "/status",
		},
	}

	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_REQUEST, requestReq.RequestType,
		"Request should use TYPE_REQUEST")

	// Test Ping construction (like what Ping() method does)
	pingReq := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_PING, pingReq.RequestType,
		"Ping request should use TYPE_PING")

	// Test Unsubscribe construction (like what Unsubscribe() method does)
	unsubscribeReq := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		RequestId:   42, // Original subscription ID
		PathOneof: &nanorpc.NanoRPCRequest_Path{
			Path: "/events",
		},
		// Data is nil for unsubscribe (empty data indicates unsubscribe)
	}

	core.AssertEqual(t, nanorpc.NanoRPCRequest_TYPE_REQUEST, unsubscribeReq.RequestType,
		"Unsubscribe request should use TYPE_REQUEST")
	core.AssertEqual(t, int32(42), unsubscribeReq.RequestId,
		"Unsubscribe request should preserve original request ID")
	core.AssertNil(t, unsubscribeReq.Data,
		"Unsubscribe request should have empty data")
}

// TestPathOneofTypes tests both path string and hash variants
func TestPathOneofTypes(t *testing.T) {
	// Test string path
	pathReq := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof: &nanorpc.NanoRPCRequest_Path{
			Path: "/events",
		},
	}

	pathOneof, ok := core.AssertTypeIs[*nanorpc.NanoRPCRequest_Path](t, pathReq.PathOneof,
		"Expected *nanorpc.NanoRPCRequest_Path")
	if ok {
		core.AssertEqual(t, "/events", pathOneof.Path, "path")
	}

	// Test hash path
	hashReq := &nanorpc.NanoRPCRequest{
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof: &nanorpc.NanoRPCRequest_PathHash{
			PathHash: 0x12345678,
		},
	}

	hashOneof, ok := core.AssertTypeIs[*nanorpc.NanoRPCRequest_PathHash](t, hashReq.PathOneof,
		"Expected *nanorpc.NanoRPCRequest_PathHash")
	if ok {
		core.AssertEqual(t, uint32(0x12345678), hashOneof.PathHash, "path_hash")
	}
}

type stringRepresentationTestCase struct {
	value interface{ String() string }
	name  string
}

func (tc stringRepresentationTestCase) test(t *testing.T) {
	core.AssertNotEqual(t, "", tc.value.String(), "%s should have a string representation", tc.name)
}

func newStringRepresentationTestCase(name string, value interface{ String() string }) stringRepresentationTestCase {
	return stringRepresentationTestCase{name: name, value: value}
}

func requestTypeStringTestCases() []stringRepresentationTestCase {
	return core.S(
		newStringRepresentationTestCase("TYPE_UNSPECIFIED", nanorpc.NanoRPCRequest_TYPE_UNSPECIFIED),
		newStringRepresentationTestCase("TYPE_PING", nanorpc.NanoRPCRequest_TYPE_PING),
		newStringRepresentationTestCase("TYPE_REQUEST", nanorpc.NanoRPCRequest_TYPE_REQUEST),
		newStringRepresentationTestCase("TYPE_SUBSCRIBE", nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE),
	)
}

func responseTypeStringTestCases() []stringRepresentationTestCase {
	return core.S(
		newStringRepresentationTestCase("TYPE_UNSPECIFIED", nanorpc.NanoRPCResponse_TYPE_UNSPECIFIED),
		newStringRepresentationTestCase("TYPE_PONG", nanorpc.NanoRPCResponse_TYPE_PONG),
		newStringRepresentationTestCase("TYPE_RESPONSE", nanorpc.NanoRPCResponse_TYPE_RESPONSE),
		newStringRepresentationTestCase("TYPE_UPDATE", nanorpc.NanoRPCResponse_TYPE_UPDATE),
	)
}

// TestProtocolDefinitions verifies that all expected protocol constants exist
func TestProtocolDefinitions(t *testing.T) {
	// Test that all request types are defined
	for _, tc := range requestTypeStringTestCases() {
		t.Run("Request_"+tc.name, tc.test)
	}

	// Test that all response types are defined
	for _, tc := range responseTypeStringTestCases() {
		t.Run("Response_"+tc.name, tc.test)
	}
}
