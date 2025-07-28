package server

import (
	"context"
	"fmt"
	"sync"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/common"
)

// mockSession implements Session interface for testing
type mockSession struct {
	id           string
	remoteAddr   string
	lastResponse *nanorpc.NanoRPCResponse
	lastData     []byte
	responses    []*nanorpc.NanoRPCResponse // Store all responses for subscription testing
	mu           sync.Mutex                 // Protect concurrent access
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

	// Encode the response to verify it's valid
	if _, err := nanorpc.EncodeResponse(response, nil); err != nil {
		return err
	}

	// Store response for verification (thread-safe)
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastResponse = response
	if response.Data != nil {
		m.lastData = response.Data
	}

	// Store all responses for subscription testing
	if m.responses == nil {
		m.responses = make([]*nanorpc.NanoRPCResponse, 0)
	}
	m.responses = append(m.responses, response)

	return nil
}

// GetLastResponse returns the last response sent
func (m *mockSession) GetLastResponse() *nanorpc.NanoRPCResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastResponse
}

// GetAllResponses returns all responses sent (for subscription testing)
func (m *mockSession) GetAllResponses() []*nanorpc.NanoRPCResponse {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to avoid race conditions
	if m.responses == nil {
		return nil
	}
	result := make([]*nanorpc.NanoRPCResponse, len(m.responses))
	copy(result, m.responses)
	return result
}

// ClearResponses clears the response history
func (m *mockSession) ClearResponses() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = common.ClearSlice(m.responses)
	m.lastResponse = nil
	m.lastData = nil
}

// mockSessionWithError extends mockSession to simulate send errors
type mockSessionWithError struct {
	sendError error
	mockSession
}

func (m *mockSessionWithError) SendResponse(req *nanorpc.NanoRPCRequest, response *nanorpc.NanoRPCResponse) error {
	if m.sendError != nil {
		return m.sendError
	}
	return m.mockSession.SendResponse(req, response)
}

// Test helper functions

// newTestSession creates a new mock session for testing
func newTestSession(id string, port uint16) *mockSession {
	if id == "" {
		id = "test-session"
	}
	if port == 0 {
		port = 12345
	}

	return &mockSession{
		id:         id,
		remoteAddr: fmt.Sprintf("127.0.0.1:%d", port),
	}
}

// newTestRequest creates a test request
func newTestRequest(id int32, pathOneOf any) *nanorpc.NanoRPCRequest {
	req := &nanorpc.NanoRPCRequest{
		RequestId:   id,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
	}

	switch p := pathOneOf.(type) {
	case string:
		req.PathOneof = nanorpc.GetPathOneOfString(p)
	case uint32:
		req.PathOneof = nanorpc.GetPathOneOfHash(p)
	case nanorpc.PathOneOf:
		req.PathOneof = p
	default:
		// This shouldn't happen in tests, but be defensive
		req.PathOneof = nil
	}

	return req
}
