package server

import (
	"context"
	"net"
	"testing"
	"time"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
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
	core.AssertMustNoError(t, err, "listen")

	server := NewDefaultServer(listener, nil, nil)
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
	core.AssertMustNoError(t, err, "dial")

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
	core.AssertMustNoError(t, err, "encode ping")

	_, err = conn.Write(pingData)
	core.AssertMustNoError(t, err, "send ping")

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
	core.AssertMustNoError(t, err, "read response")

	response, _, err := nanorpc.DecodeResponse(buffer[:n])
	core.AssertMustNoError(t, err, "decode response")

	return response
}

// verifyPongResponse checks that the response is a valid pong
func verifyPongResponse(t *testing.T, response *nanorpc.NanoRPCResponse, expectedID int32) {
	t.Helper()

	core.AssertEqual(t, nanorpc.NanoRPCResponse_TYPE_PONG, response.ResponseType,
		"response type")
	core.AssertEqual(t, expectedID, response.RequestId, "request id")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, response.ResponseStatus,
		"response status")
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
		if err != nil {
			core.AssertErrorIs(t, err, context.Canceled, "server stop error")
		}
	case <-ctx.Done():
		t.Fatal("Server shutdown timeout")
	}
}

// TestNewDefaultServer_UsesSuppliedHandler verifies that handlers registered
// on a [*DefaultMessageHandler] passed to [NewDefaultServer] are reachable
// from a live connection.
func TestNewDefaultServer_UsesSuppliedHandler(t *testing.T) {
	const path = "/api/echo"

	handler := NewDefaultMessageHandler(nil)
	err := handler.RegisterHandlerFunc(path,
		func(_ context.Context, req *RequestContext) error {
			return req.SendOK(req.GetData())
		})
	core.AssertMustNoError(t, err, "register handler")

	listener, err := net.Listen("tcp", "localhost:0")
	core.AssertMustNoError(t, err, "listen")

	server := NewDefaultServer(listener, handler, nil)
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.Serve(context.Background()) }()
	time.Sleep(50 * time.Millisecond)
	defer shutdownServer(t, server, serverErr)

	conn := connectToServer(t, listener.Addr().String())
	defer conn.Close()

	payload := []byte("hello")
	req := &nanorpc.NanoRPCRequest{
		RequestId:   42,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   nanorpc.GetPathOneOfString(path),
		Data:        payload,
	}
	reqData, err := nanorpc.EncodeRequest(req, nil)
	core.AssertMustNoError(t, err, "encode request")

	_, err = conn.Write(reqData)
	core.AssertMustNoError(t, err, "write request")

	response := readResponse(t, conn)
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, response.ResponseStatus,
		"response status")
	core.AssertEqual(t, string(payload), string(response.Data), "response data")
}

func TestDecoupledServer_Shutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	core.AssertMustNoError(t, err, "listen")

	server := NewDefaultServer(listener, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	core.AssertMustNoError(t, err, "shutdown server")
}
