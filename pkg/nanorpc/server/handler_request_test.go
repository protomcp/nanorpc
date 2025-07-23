package server

import (
	"context"
	"testing"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

const (
	// Test paths with semantic meaning
	pathUnregistered = "/api/unregistered" // Path with no handler
	pathEcho         = "/api/echo"         // Path that echoes back request data
)

// handleRequestTestCase represents a test case for TYPE_REQUEST handling
type handleRequestTestCase struct {
	name           string
	path           string
	requestData    []byte
	expectResponse nanorpc.NanoRPCResponse_Status
	expectError    bool
}

// echoHandler handles echo requests
func (*handleRequestTestCase) echoHandler(_ context.Context, req *RequestContext) error {
	// Echo back the request data using the new helper method
	return req.SendOK(req.GetData())
}

// test runs the test case
func (tc *handleRequestTestCase) test(t *testing.T) {
	t.Helper()

	// Create handler with test configuration
	handler := NewDefaultMessageHandler()
	_ = handler.RegisterHandlerFunc(pathEcho, tc.echoHandler)

	// Create mock session
	session := &mockSession{
		id:         "test-session",
		remoteAddr: "127.0.0.1:12345",
	}

	// Create request
	req := &nanorpc.NanoRPCRequest{
		RequestId:   100,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof: &nanorpc.NanoRPCRequest_Path{
			Path: tc.path,
		},
		Data: tc.requestData,
	}

	// Handle message
	err := handler.HandleMessage(context.Background(), session, req)

	// Check error expectation
	if tc.expectError && err == nil {
		t.Fatalf("expected error but got none")
	}
	if !tc.expectError && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check response (simplified based on PR feedback)
	if session.lastResponse == nil {
		t.Fatal("expected response but got none")
	}
	if session.lastResponse.ResponseStatus != tc.expectResponse {
		t.Fatalf("expected status %v, got %v",
			tc.expectResponse, session.lastResponse.ResponseStatus)
	}
}

// TestDefaultMessageHandler_HandleRequest tests TYPE_REQUEST handling
func TestDefaultMessageHandler_HandleRequest(t *testing.T) {
	tests := []handleRequestTestCase{
		{
			name:           "request with no registered handler",
			path:           pathUnregistered,
			requestData:    []byte("test data"),
			expectResponse: nanorpc.NanoRPCResponse_STATUS_NOT_FOUND,
			expectError:    false,
		},
		{
			name:           "request with registered handler",
			path:           pathEcho,
			requestData:    []byte("echo this"),
			expectResponse: nanorpc.NanoRPCResponse_STATUS_OK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// mockSession implements Session interface for testing
type mockSession struct {
	id           string
	remoteAddr   string
	lastResponse *nanorpc.NanoRPCResponse
	lastData     []byte
}

func (m *mockSession) ID() string {
	return m.id
}

func (m *mockSession) RemoteAddr() string {
	return m.remoteAddr
}

func (*mockSession) Handle(_ context.Context) error {
	return nil
}

func (*mockSession) Close() error {
	return nil
}

// SendResponse captures the response for testing
func (m *mockSession) SendResponse(req *nanorpc.NanoRPCRequest, response *nanorpc.NanoRPCResponse) error {
	// Fill envelope fields from request if provided
	if req != nil && response.RequestId == 0 {
		response.RequestId = req.RequestId
	}

	// Encode the response
	data, err := nanorpc.EncodeResponse(response, nil)
	if err != nil {
		return err
	}

	m.lastData = data
	m.lastResponse = response
	return nil
}
