package server

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

func TestDefaultSession_ID(t *testing.T) {
	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	handler := NewDefaultMessageHandler(nil)
	session := NewDefaultSession(conn, handler, nil)

	id := session.ID()
	if id == "" {
		t.Fatal("Session ID should not be empty")
	}

	expectedSuffix := "@127.0.0.1:12345"
	if !strings.HasSuffix(id, expectedSuffix) {
		t.Fatalf("Expected ID to end with %q, got %q", expectedSuffix, id)
	}

	// Check that the ID has the expected xid@addr format
	parts := strings.SplitN(id, "@", 2)
	if len(parts) != 2 {
		t.Fatalf("Expected ID format 'xid@addr', got %q", id)
	}
	if parts[0] == "" {
		t.Fatal("Session ID xid part should not be empty")
	}
	if parts[1] != "127.0.0.1:12345" {
		t.Fatalf("Expected address part to be '127.0.0.1:12345', got %q", parts[1])
	}
}

func TestDefaultSession_RemoteAddr(t *testing.T) {
	expectedAddr := "127.0.0.1:12345"
	conn := &mockConn{remoteAddr: expectedAddr}
	handler := NewDefaultMessageHandler(nil)
	session := NewDefaultSession(conn, handler, nil)

	addr := session.RemoteAddr()
	if addr != expectedAddr {
		t.Fatalf("Expected remote addr %q, got %q", expectedAddr, addr)
	}
}

func TestDefaultSession_Close(t *testing.T) {
	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	handler := NewDefaultMessageHandler(nil)
	session := NewDefaultSession(conn, handler, nil)

	err := session.Close()
	if err != nil {
		t.Fatalf("Expected no error closing session, got %v", err)
	}

	if !conn.closed {
		t.Fatal("Expected connection to be closed")
	}
}

// mockConn implements net.Conn for testing
type mockConn struct {
	closeErr   error
	remoteAddr string
	localAddr  string
	data       []byte
	writeData  []byte
	readPos    int
	closed     bool
}

func (m *mockConn) Read(b []byte) (int, error) {
	if m.closed {
		return 0, net.ErrClosed
	}
	if m.readPos >= len(m.data) {
		// Return 0 to simulate EOF for test
		return 0, nil
	}
	n := copy(b, m.data[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockConn) Write(b []byte) (int, error) {
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return m.closeErr
}

func (m *mockConn) LocalAddr() net.Addr {
	return &mockAddr{addr: m.localAddr}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &mockAddr{addr: m.remoteAddr}
}

func (*mockConn) SetDeadline(_ time.Time) error      { return nil }
func (*mockConn) SetReadDeadline(_ time.Time) error  { return nil }
func (*mockConn) SetWriteDeadline(_ time.Time) error { return nil }

type mockAddr struct {
	addr string
}

func (*mockAddr) Network() string  { return "tcp" }
func (m *mockAddr) String() string { return m.addr }

func TestDefaultSession_HandlePing(t *testing.T) {
	// Create a ping request
	pingReq := &nanorpc.NanoRPCRequest{
		RequestId:   42,
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	// Encode the ping request
	pingData, err := nanorpc.EncodeRequest(pingReq, nil)
	if err != nil {
		t.Fatalf("Failed to encode ping request: %v", err)
	}

	// Create mock connection with ping data
	conn := &mockConn{
		remoteAddr: "127.0.0.1:12345",
		data:       pingData,
	}

	handler := NewDefaultMessageHandler(nil)
	session := NewDefaultSession(conn, handler, nil)

	// Handle the session
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = session.Handle(ctx)
	if err != nil {
		t.Logf("Session handle returned error: %v", err)
	}

	// Verify pong response was written
	if len(conn.writeData) == 0 {
		t.Fatal("Expected response data to be written")
	}

	// Decode and verify the response
	response, _, err := nanorpc.DecodeResponse(conn.writeData)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ResponseType != nanorpc.NanoRPCResponse_TYPE_PONG {
		t.Fatalf("Expected PONG response, got %v", response.ResponseType)
	}

	if response.RequestId != 42 {
		t.Fatalf("Expected request ID 42, got %d", response.RequestId)
	}

	if response.ResponseStatus != nanorpc.NanoRPCResponse_STATUS_OK {
		t.Fatalf("Expected STATUS_OK, got %v", response.ResponseStatus)
	}
}
