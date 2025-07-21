package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

func TestDecoupledServer_PingPong(t *testing.T) {
	// Setup server
	listener, server, serverErr := setupTestServer(t)
	defer listener.Close()

	// Connect client
	conn := connectToServer(t, listener.Addr().String())
	defer conn.Close()

	// Send ping and receive pong
	sendPingReceivePong(t, conn)

	// Shutdown and verify
	shutdownServer(t, server, serverErr)
}

// setupTestServer creates and starts a test server
func setupTestServer(t *testing.T) (net.Listener, *Server, <-chan error) {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	server := NewDefaultServer(listener, nil)
	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.Serve(context.Background())
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	return listener, server, serverErr
}

// connectToServer establishes a connection to the test server
func connectToServer(t *testing.T, addr string) net.Conn {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	return conn
}

// sendPingReceivePong sends a ping request and verifies the pong response
func sendPingReceivePong(t *testing.T, conn net.Conn) {
	t.Helper()

	// Send ping
	pingReq := &nanorpc.NanoRPCRequest{
		RequestId:   789,
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	pingData, err := nanorpc.EncodeRequest(pingReq, nil)
	if err != nil {
		t.Fatalf("Failed to encode ping: %v", err)
	}

	if _, err := conn.Write(pingData); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// Read response
	response := readResponse(t, conn)

	// Verify pong
	verifyPongResponse(t, response, 789)
}

// readResponse reads and decodes a response from the connection
func readResponse(t *testing.T, conn net.Conn) *nanorpc.NanoRPCResponse {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Logf("Failed to set read deadline: %v", err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	response, _, err := nanorpc.DecodeResponse(buffer[:n])
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return response
}

// verifyPongResponse checks that the response is a valid pong
func verifyPongResponse(t *testing.T, response *nanorpc.NanoRPCResponse, expectedID int32) {
	t.Helper()

	if response.ResponseType != nanorpc.NanoRPCResponse_TYPE_PONG {
		t.Fatalf("Expected PONG, got %v", response.ResponseType)
	}

	if response.RequestId != expectedID {
		t.Fatalf("Expected RequestId=%d, got %d", expectedID, response.RequestId)
	}

	if response.ResponseStatus != nanorpc.NanoRPCResponse_STATUS_OK {
		t.Fatalf("Expected STATUS_OK, got %v", response.ResponseStatus)
	}
}

// shutdownServer cancels the context and waits for server to stop
func shutdownServer(t *testing.T, server *Server, serverErr <-chan error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Logf("Shutdown error: %v", err)
	}

	select {
	case err := <-serverErr:
		if err != nil && err != context.Canceled {
			t.Fatalf("Server stopped with unexpected error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Server shutdown timeout")
	}
}

func TestDecoupledServer_Shutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	server := NewDefaultServer(listener, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown server: %v", err)
	}
}
