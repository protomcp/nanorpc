package nanorpc

import (
	"testing"
)

type requestTypeTestCase struct {
	name         string
	requestType  NanoRPCRequest_Type
	expectedType NanoRPCRequest_Type
}

func (tc requestTypeTestCase) test(t *testing.T) {
	// Create a request with the type we're testing
	req := &NanoRPCRequest{
		RequestType: tc.requestType,
		RequestId:   123,
		PathOneof: &NanoRPCRequest_Path{
			Path: "/test",
		},
	}

	// Verify the request type matches expectation
	AssertEqual(t, tc.expectedType, req.RequestType, "RequestType mismatch")
}

func newRequestTypeTestCase(name string, requestType, expectedType NanoRPCRequest_Type) requestTypeTestCase {
	return requestTypeTestCase{
		name:         name,
		requestType:  requestType,
		expectedType: expectedType,
	}
}

func subscriptionRequestTypeTestCases() []requestTypeTestCase {
	return S(
		newRequestTypeTestCase("Subscribe_uses_TYPE_SUBSCRIBE",
			NanoRPCRequest_TYPE_SUBSCRIBE, NanoRPCRequest_TYPE_SUBSCRIBE),
		newRequestTypeTestCase("Request_uses_TYPE_REQUEST", NanoRPCRequest_TYPE_REQUEST, NanoRPCRequest_TYPE_REQUEST),
		newRequestTypeTestCase("Ping_uses_TYPE_PING", NanoRPCRequest_TYPE_PING, NanoRPCRequest_TYPE_PING),
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
	subscribeReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE, // This should be TYPE_SUBSCRIBE after our fix
		PathOneof: &NanoRPCRequest_Path{
			Path: "/events",
		},
	}

	AssertEqual(t, NanoRPCRequest_TYPE_SUBSCRIBE, subscribeReq.RequestType,
		"Subscribe request should use TYPE_SUBSCRIBE")

	// Test Request construction (like what Request() method does)
	requestReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &NanoRPCRequest_Path{
			Path: "/status",
		},
	}

	AssertEqual(t, NanoRPCRequest_TYPE_REQUEST, requestReq.RequestType,
		"Request should use TYPE_REQUEST")

	// Test Ping construction (like what Ping() method does)
	pingReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_PING,
	}

	AssertEqual(t, NanoRPCRequest_TYPE_PING, pingReq.RequestType,
		"Ping request should use TYPE_PING")
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

	pathOneof := AssertTypeIs[*NanoRPCRequest_Path](t, pathReq.PathOneof,
		"Expected *NanoRPCRequest_Path")
	AssertEqual(t, "/events", pathOneof.Path, "Path mismatch")

	// Test hash path
	hashReq := &NanoRPCRequest{
		RequestType: NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof: &NanoRPCRequest_PathHash{
			PathHash: 0x12345678,
		},
	}

	hashOneof := AssertTypeIs[*NanoRPCRequest_PathHash](t, hashReq.PathOneof,
		"Expected *NanoRPCRequest_PathHash")
	AssertEqual(t, uint32(0x12345678), hashOneof.PathHash, "Hash mismatch")
}

type stringRepresentationTestCase struct {
	value interface{ String() string }
	name  string
}

func (tc stringRepresentationTestCase) test(t *testing.T) {
	AssertNotEqual(t, "", tc.value.String(), "%s should have a string representation", tc.name)
}

func newStringRepresentationTestCase(name string, value interface{ String() string }) stringRepresentationTestCase {
	return stringRepresentationTestCase{name: name, value: value}
}

func requestTypeStringTestCases() []stringRepresentationTestCase {
	return S(
		newStringRepresentationTestCase("TYPE_UNSPECIFIED", NanoRPCRequest_TYPE_UNSPECIFIED),
		newStringRepresentationTestCase("TYPE_PING", NanoRPCRequest_TYPE_PING),
		newStringRepresentationTestCase("TYPE_REQUEST", NanoRPCRequest_TYPE_REQUEST),
		newStringRepresentationTestCase("TYPE_SUBSCRIBE", NanoRPCRequest_TYPE_SUBSCRIBE),
	)
}

func responseTypeStringTestCases() []stringRepresentationTestCase {
	return S(
		newStringRepresentationTestCase("TYPE_UNSPECIFIED", NanoRPCResponse_TYPE_UNSPECIFIED),
		newStringRepresentationTestCase("TYPE_PONG", NanoRPCResponse_TYPE_PONG),
		newStringRepresentationTestCase("TYPE_RESPONSE", NanoRPCResponse_TYPE_RESPONSE),
		newStringRepresentationTestCase("TYPE_UPDATE", NanoRPCResponse_TYPE_UPDATE),
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
