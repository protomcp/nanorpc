package server_test

import (
	"context"
	"fmt"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/server"
)

// Example_jsonEchoHandler demonstrates using JSON helpers for request/response handling
func Example_jsonEchoHandler() {
	handler := server.NewDefaultMessageHandler(nil)

	// Register a JSON echo handler
	_ = handler.RegisterHandlerFunc("/api/json/echo", func(_ context.Context, req *server.RequestContext) error {
		// Define the expected request structure
		var requestData map[string]any

		// Unmarshal JSON request
		if err := req.UnmarshalRequestJSON(&requestData); err != nil {
			return req.SendBadRequest("invalid JSON in request body")
		}

		// Echo back the data with a timestamp
		// Note: Fixed timestamp for consistent examples (use time.Now() in real handlers)
		responseData := map[string]any{
			"echo":      requestData,
			"timestamp": "2024-01-15T10:30:00Z",
		}

		// Send JSON response
		return req.SendJSON(responseData)
	})

	fmt.Println("JSON echo handler registered at /api/json/echo")
	// Output: JSON echo handler registered at /api/json/echo
}

// Example_protobufHandler demonstrates using protobuf helpers
func Example_protobufHandler() {
	handler := server.NewDefaultMessageHandler(nil)

	// Register a protobuf handler
	_ = handler.RegisterHandlerFunc("/api/proto/ping", func(_ context.Context, req *server.RequestContext) error {
		// Unmarshal protobuf request
		var pingReq nanorpc.NanoRPCRequest
		if err := req.UnmarshalRequestProtobuf(&pingReq); err != nil {
			return req.SendBadRequest("invalid protobuf in request body")
		}

		// Create response
		pongResp := &nanorpc.NanoRPCResponse{
			RequestId:      pingReq.RequestId,
			ResponseType:   nanorpc.NanoRPCResponse_TYPE_PONG,
			ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
		}

		// Send protobuf response
		return req.SendProtobuf(pongResp)
	})

	fmt.Println("Protobuf handler registered at /api/proto/ping")
	// Output: Protobuf handler registered at /api/proto/ping
}

// Example_errorHandlers demonstrates various error response helpers
func Example_errorHandlers() {
	handler := server.NewDefaultMessageHandler(nil)

	// Handler that returns not found
	_ = handler.RegisterHandlerFunc("/api/users/get", func(_ context.Context, req *server.RequestContext) error {
		// Simulate user lookup failure
		return req.SendNotFound("user not found")
	})

	// Handler that checks authorization
	_ = handler.RegisterHandlerFunc("/api/admin/action", func(_ context.Context, req *server.RequestContext) error {
		// Simulate authorization check
		authorized := false
		if !authorized {
			return req.SendUnauthorized("admin access required")
		}
		return req.SendOK(nil)
	})

	// Handler that validates input
	_ = handler.RegisterHandlerFunc("/api/data/process", func(_ context.Context, req *server.RequestContext) error {
		if !req.HasData() {
			return req.SendBadRequest("request body is required")
		}

		// Simulate processing error
		return req.SendInternalError("database connection failed")
	})

	fmt.Println("Error handlers registered")
	// Output: Error handlers registered
}

// Example_dataAccessHelpers demonstrates request data access helpers
func Example_dataAccessHelpers() {
	handler := server.NewDefaultMessageHandler(nil)

	_ = handler.RegisterHandlerFunc("/api/info", func(_ context.Context, req *server.RequestContext) error {
		// Access request metadata
		requestID := req.GetRequestID()
		hasData := req.HasData()
		data := req.GetData()

		// Build response based on request info
		info := fmt.Sprintf("Request ID: %d, Has Data: %v, Data Length: %d",
			requestID, hasData, len(data))

		return req.SendOK([]byte(info))
	})

	fmt.Println("Info handler registered at /api/info")
	// Output: Info handler registered at /api/info
}
