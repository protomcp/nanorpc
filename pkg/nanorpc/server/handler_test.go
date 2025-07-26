package server

import (
	"context"
	"testing"

	"github.com/amery/nanorpc/pkg/nanorpc"
	"github.com/amery/nanorpc/pkg/nanorpc/common/testutils"
)

// Test helpers and factories

func newTestSession() *mockSession {
	return &mockSession{
		id:         "test-session",
		remoteAddr: "127.0.0.1:12345",
	}
}

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

func registerTestHandler(t testutils.T, handler *DefaultMessageHandler, path string, response []byte) {
	t.Helper()
	err := handler.RegisterHandlerFunc(path, func(_ context.Context, req *RequestContext) error {
		return req.SendOK(response)
	})
	testutils.AssertNoError(t, err, "register handler for %s", path)
}

func verifyResponse(t testutils.T, resp *nanorpc.NanoRPCResponse,
	expectStatus nanorpc.NanoRPCResponse_Status, expectData string) {
	t.Helper()
	testutils.AssertNotNil(t, resp, "response")
	testutils.AssertEqual(t, expectStatus, resp.ResponseStatus, "response status")

	if expectData != "" {
		testutils.AssertEqual(t, expectData, string(resp.Data), "response data")
	}
}

func TestDefaultMessageHandler_HandlePing(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	conn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	session := NewDefaultSession(conn, handler, nil)

	req := &nanorpc.NanoRPCRequest{
		RequestId:   123,
		RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
	}

	err := handler.HandleMessage(context.Background(), session, req)
	testutils.AssertNoError(t, err, "handle ping")

	// Verify response was written
	testutils.AssertTrue(t, len(conn.WriteData) > 0, "response data written")

	// Decode and verify the response
	response, _, err := nanorpc.DecodeResponse(conn.WriteData)
	testutils.AssertNoError(t, err, "decode response")

	testutils.AssertEqual(t, nanorpc.NanoRPCResponse_TYPE_PONG, response.ResponseType, "response type")
	testutils.AssertEqual(t, 123, response.RequestId, "request ID")
	testutils.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, response.ResponseStatus, "response status")
}

func TestDefaultMessageHandler_HandleUnsupportedType(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	conn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
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
	if len(conn.WriteData) > 0 {
		t.Fatal("Expected no response data for unsupported type")
	}
}

// hashCacheTestCase represents a test case for hash-based path handling
type hashCacheTestCase struct {
	name         string
	registerPath string
	requestPath  string
	useHash      bool
	expectFound  bool
}

// test runs the hash cache test case
func (tc *hashCacheTestCase) test(t *testing.T) {
	t.Helper()

	// Setup
	hashCache := &nanorpc.HashCache{}
	handler := NewDefaultMessageHandler(hashCache)
	registerTestHandler(t, handler, tc.registerPath, []byte("success"))
	session := newTestSession()

	// Create request
	var pathOneof any
	if tc.useHash {
		hash, err := hashCache.Hash(tc.requestPath)
		testutils.AssertNoError(t, err, "hash path")
		pathOneof = hash
	} else {
		pathOneof = tc.requestPath
	}
	req := newTestRequest(100, pathOneof)

	// Execute
	err := handler.HandleMessage(context.Background(), session, req)
	testutils.AssertNoError(t, err, "handle message")

	// Verify
	expectedStatus := nanorpc.NanoRPCResponse_STATUS_OK
	if !tc.expectFound {
		expectedStatus = nanorpc.NanoRPCResponse_STATUS_NOT_FOUND
	}
	verifyResponse(t, session.lastResponse, expectedStatus, "")
}

func TestDefaultMessageHandler_HashCache(t *testing.T) {
	tests := []hashCacheTestCase{
		{
			name:         "string path request",
			registerPath: "/api/test",
			requestPath:  "/api/test",
			useHash:      false,
			expectFound:  true,
		},
		{
			name:         "hash path request for registered path",
			registerPath: "/api/users",
			requestPath:  "/api/users",
			useHash:      true,
			expectFound:  true,
		},
		{
			name:         "hash path request for unregistered path",
			registerPath: "/api/products",
			requestPath:  "/api/unknown",
			useHash:      true,
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// requestContextPathHashTestCase tests RequestContext PathHash field population
type requestContextPathHashTestCase struct {
	capturedCtx  **RequestContext
	name         string
	path         string
	expectedHash uint32
	useHash      bool
}

// test runs the RequestContext PathHash test case
func (tc *requestContextPathHashTestCase) test(t *testing.T) {
	t.Helper()

	// Reset captured context
	*tc.capturedCtx = nil

	session := &mockSession{
		id:         "test-session",
		remoteAddr: "127.0.0.1:12345",
	}

	req := &nanorpc.NanoRPCRequest{
		RequestId:   200,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
	}

	if tc.useHash {
		req.PathOneof = nanorpc.GetPathOneOfHash(tc.expectedHash)
	} else {
		req.PathOneof = nanorpc.GetPathOneOfString(tc.path)
	}

	// Create handler and handle message
	hashCache := &nanorpc.HashCache{}
	handler := NewDefaultMessageHandler(hashCache)

	// Register handler that captures the RequestContext
	err := handler.RegisterHandlerFunc(tc.path, func(_ context.Context, req *RequestContext) error {
		*tc.capturedCtx = req
		return req.SendOK(nil)
	})
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	err = handler.HandleMessage(context.Background(), session, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify RequestContext was populated correctly
	if *tc.capturedCtx == nil {
		t.Fatal("Handler was not called")
	}

	if (*tc.capturedCtx).Path != tc.path {
		t.Errorf("Expected path %q, got %q", tc.path, (*tc.capturedCtx).Path)
	}

	if (*tc.capturedCtx).PathHash != tc.expectedHash {
		t.Errorf("Expected PathHash %d, got %d", tc.expectedHash, (*tc.capturedCtx).PathHash)
	}
}

func TestDefaultMessageHandler_RequestContext_PathHash(t *testing.T) {
	// Test that RequestContext properly populates PathHash field
	hashCache := &nanorpc.HashCache{}
	path := "/api/test/endpoint"

	// Get expected hash
	expectedHash, err := hashCache.Hash(path)
	if err != nil {
		t.Fatalf("Failed to hash path: %v", err)
	}

	var capturedContext *RequestContext

	tests := []requestContextPathHashTestCase{
		{
			name:         "string path request",
			path:         path,
			useHash:      false,
			capturedCtx:  &capturedContext,
			expectedHash: expectedHash,
		},
		{
			name:         "hash path request",
			path:         path,
			useHash:      true,
			capturedCtx:  &capturedContext,
			expectedHash: expectedHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// pathResolutionTestCase tests path resolution from hash
type pathResolutionTestCase struct {
	name          string
	expectPath    string
	registerPaths []string
	requestHash   uint32
	expectFound   bool
}

// test runs the path resolution test case
func (tc *pathResolutionTestCase) test(t *testing.T) {
	t.Helper()

	// Setup
	hashCache := &nanorpc.HashCache{}
	handler := NewDefaultMessageHandler(hashCache)

	// Register multiple paths
	for _, path := range tc.registerPaths {
		err := handler.RegisterHandlerFunc(path, func(_ context.Context, req *RequestContext) error {
			return req.SendOK([]byte(req.Path))
		})
		testutils.AssertNoError(t, err, "register handler for %s", path)
	}

	// Execute
	session := newTestSession()
	req := newTestRequest(400, tc.requestHash)
	err := handler.HandleMessage(context.Background(), session, req)
	testutils.AssertNoError(t, err, "handle message")

	// Verify
	if tc.expectFound {
		verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_OK, tc.expectPath)
	} else {
		verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_NOT_FOUND, "")
	}
}

func TestDefaultMessageHandler_PathResolution(t *testing.T) {
	// Pre-compute some hashes for testing
	hashCache := &nanorpc.HashCache{}
	apiUsersHash, _ := hashCache.Hash("/api/users")
	apiProductsHash, _ := hashCache.Hash("/api/products")
	unknownHash := uint32(0xDEADBEEF) // Random hash not in cache

	tests := []pathResolutionTestCase{
		{
			name:          "resolve registered path from hash",
			registerPaths: []string{"/api/users", "/api/products"},
			requestHash:   apiUsersHash,
			expectPath:    "/api/users",
			expectFound:   true,
		},
		{
			name:          "unknown hash returns not found",
			registerPaths: []string{"/api/users"},
			requestHash:   unknownHash,
			expectPath:    "",
			expectFound:   false,
		},
		{
			name:          "resolve second registered path",
			registerPaths: []string{"/api/users", "/api/products"},
			requestHash:   apiProductsHash,
			expectPath:    "/api/products",
			expectFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestDefaultMessageHandler_HashCollision tests hash collision handling
func TestDefaultMessageHandler_HashCollision(t *testing.T) {
	// This test verifies that hash collisions during handler registration
	// are properly handled. Since we can't easily force a real collision,
	// we test the error path by checking the hash integration works correctly.

	hashCache := &nanorpc.HashCache{}
	handler := NewDefaultMessageHandler(hashCache)

	// Register a handler
	path1 := "/api/test/path1"
	registerTestHandler(t, handler, path1, []byte("path1"))

	// Try to register the same path again (should fail with ErrExists)
	err := handler.RegisterHandlerFunc(path1, func(_ context.Context, req *RequestContext) error {
		return req.SendOK([]byte("duplicate"))
	})
	testutils.AssertError(t, err, "registering duplicate path should fail")

	// Verify hash-based request still works
	session := newTestSession()
	hash1, _ := hashCache.Hash(path1)
	req := newTestRequest(100, hash1)

	err = handler.HandleMessage(context.Background(), session, req)
	testutils.AssertNoError(t, err, "handle request")

	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_OK, "path1")
}
