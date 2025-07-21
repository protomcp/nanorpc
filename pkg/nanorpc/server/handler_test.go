package server

import (
	"context"
	"testing"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

func TestDefaultMessageHandler_HandlePing(t *testing.T) {
	handler := NewDefaultMessageHandler()
	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := NewDefaultSession(conn, handler, nil)

	req := &nanorpc.NanoRPCRequest{
		RequestId:   123,
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	err := handler.HandleMessage(context.Background(), session, req)
	if err != nil {
		t.Fatalf("Expected no error handling ping, got %v", err)
	}

	// Verify response was written
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

	if response.RequestId != 123 {
		t.Fatalf("Expected request ID 123, got %d", response.RequestId)
	}

	if response.ResponseStatus != nanorpc.NanoRPCResponse_STATUS_OK {
		t.Fatalf("Expected STATUS_OK, got %v", response.ResponseStatus)
	}
}

func TestDefaultMessageHandler_HandleUnsupportedType(t *testing.T) {
	handler := NewDefaultMessageHandler()
	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := NewDefaultSession(conn, handler, nil)

	req := &nanorpc.NanoRPCRequest{
		RequestId:   456,
		RequestType: 99, // Invalid request type
	}

	err := handler.HandleMessage(context.Background(), session, req)
	if err != nil {
		t.Fatalf("Expected no error for unsupported type, got %v", err)
	}

	// Verify no response was written
	if len(conn.writeData) > 0 {
		t.Fatal("Expected no response data for unsupported type")
	}
}
