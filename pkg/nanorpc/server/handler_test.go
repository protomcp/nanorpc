package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"darvaza.org/core"
	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/utils/testutils"
)

// Test helpers and factories

func registerTestHandler(t core.T, handler *DefaultMessageHandler, path string, response []byte) {
	t.Helper()
	err := handler.RegisterHandlerFunc(path, func(_ context.Context, req *RequestContext) error {
		return req.SendOK(response)
	})
	core.AssertNoError(t, err, "register %s", path)
}

func verifyResponse(t core.T, resp *nanorpc.NanoRPCResponse,
	expectStatus nanorpc.NanoRPCResponse_Status, expectData string) {
	t.Helper()
	core.AssertNotNil(t, resp, "response")
	core.AssertEqual(t, expectStatus, resp.ResponseStatus, "response status")

	if expectData != "" {
		core.AssertEqual(t, expectData, string(resp.Data), "response data")
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
	core.AssertNoError(t, err, "ping")

	// Verify response was written
	core.AssertTrue(t, len(conn.WriteData) > 0, "response written")

	// Decode and verify the response
	response, _, err := nanorpc.DecodeResponse(conn.WriteData)
	core.AssertNoError(t, err, "decode")

	core.AssertEqual(t, nanorpc.NanoRPCResponse_TYPE_PONG, response.ResponseType, "response type")
	core.AssertEqual(t, 123, response.RequestId, "request ID")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, response.ResponseStatus, "response status")
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
	session := newTestSession("", 0)

	// Create request
	var pathOneof any
	if tc.useHash {
		hash, err := hashCache.Hash(tc.requestPath)
		core.AssertNoError(t, err, "hash")
		pathOneof = hash
	} else {
		pathOneof = tc.requestPath
	}
	req := newTestRequest(100, pathOneof)

	// Execute
	err := handler.HandleMessage(context.Background(), session, req)
	core.AssertNoError(t, err, "handler")

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
		core.AssertNoError(t, err, "register %s", path)
	}

	// Execute
	session := newTestSession("", 0)
	req := newTestRequest(400, tc.requestHash)
	err := handler.HandleMessage(context.Background(), session, req)
	core.AssertNoError(t, err, "handler")

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
	core.AssertError(t, err, "duplicate path")

	// Verify hash-based request still works
	session := newTestSession("", 0)
	hash1, _ := hashCache.Hash(path1)
	req := newTestRequest(100, hash1)

	err = handler.HandleMessage(context.Background(), session, req)
	core.AssertNoError(t, err, "request")

	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_OK, "path1")
}

// Test factories for unsubscribe protocol

const testSubscriptionPath = "/test/subscription"

func newUnsubscribeRequest(requestID int32) *nanorpc.NanoRPCRequest {
	return &nanorpc.NanoRPCRequest{
		RequestId:   requestID,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   nanorpc.GetPathOneOfString(testSubscriptionPath),
		Data:        []byte{}, // Empty data = unsubscribe
	}
}

func countSubscriptions(t core.T, handler *DefaultMessageHandler, pathHash uint32) int {
	t.Helper()
	handler.mu.RLock()
	defer handler.mu.RUnlock()

	subs := handler.subscriptions[pathHash]
	if subs == nil {
		return 0
	}
	return subs.Len()
}

// testUnsubscribeRemovesSubscription verifies unsubscribe removes subscription
func testUnsubscribeRemovesSubscription(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session := newTestSession("test-session", 0)
	testPath := testSubscriptionPath
	pathHash, _ := handler.hashCache.Hash(testPath)

	// Subscribe first
	subReq := newTestSubscribeRequest(42, testPath, []byte("filter-data"))
	err := handler.HandleMessage(context.Background(), session, subReq)
	core.AssertNoError(t, err, "subscribe")
	core.AssertEqual(t, 1, countSubscriptions(t, handler, pathHash),
		"subscriptions")

	// Unsubscribe using the same request ID as the subscription
	unsubscribeReq := newUnsubscribeRequest(42)
	err = handler.HandleMessage(context.Background(), session, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Verify removed
	core.AssertEqual(t, 0, countSubscriptions(t, handler, pathHash),
		"subscriptions")
	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_OK, "")
}

// testUnsubscribeWithoutSubscription verifies behaviour when no subscription exists
func testUnsubscribeWithoutSubscription(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session := newTestSession("test-session", 0)

	// Unsubscribe without subscribing first
	unsubscribeReq := newUnsubscribeRequest(44)
	err := handler.HandleMessage(context.Background(), session, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Should get NOT_FOUND since no handler registered
	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_NOT_FOUND, "")
}

// testUnsubscribeIsSessionSpecific verifies only the session's subscriptions are removed
func testUnsubscribeIsSessionSpecific(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session1 := newTestSession("session-1", 0)
	session2 := newTestSession("session-2", 0)
	testPath := testSubscriptionPath
	pathHash, _ := handler.hashCache.Hash(testPath)

	// Both sessions subscribe
	subscribeReq1 := newTestSubscribeRequest(42, testPath, []byte("filter-1"))
	subscribeReq2 := newTestSubscribeRequest(52, testPath, []byte("filter-2"))

	err := handler.HandleMessage(context.Background(), session1, subscribeReq1)
	core.AssertNoError(t, err, "subscribe")
	err = handler.HandleMessage(context.Background(), session2, subscribeReq2)
	core.AssertNoError(t, err, "subscribe")

	core.AssertEqual(t, 2, countSubscriptions(t, handler, pathHash),
		"subscriptions")

	// Session1 unsubscribes its subscription (request ID 42)
	unsubscribeReq := newUnsubscribeRequest(42)
	err = handler.HandleMessage(context.Background(), session1, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Only session2's subscription should remain
	core.AssertEqual(t, 1, countSubscriptions(t, handler, pathHash),
		"subscriptions")
}

// TestUnsubscribeProtocol tests that TYPE_REQUEST with empty data unsubscribes
func TestUnsubscribeProtocol(t *testing.T) {
	t.Run("unsubscribe removes subscription", testUnsubscribeRemovesSubscription)
	t.Run("unsubscribe without subscription", testUnsubscribeWithoutSubscription)
	t.Run("unsubscribe is session-specific", testUnsubscribeIsSessionSpecific)
}

// TestUnsubscribeEdgeCases tests comprehensive unsubscribe edge cases
func TestUnsubscribeEdgeCases(t *testing.T) {
	t.Run("hash-based unsubscribe", testUnsubscribeWithHashBasedRequest)
	t.Run("non-existent subscription", testUnsubscribeNonExistentSubscription)
	t.Run("mismatched request ID", testUnsubscribeWithMismatchedRequestID)
	t.Run("zero path hash", testUnsubscribeWithZeroPathHash)
}

// testUnsubscribeWithHashBasedRequest tests unsubscribe using hash-based path
func testUnsubscribeWithHashBasedRequest(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session := newTestSession("hash-session", 0)
	testPath := testSubscriptionPath
	pathHash, _ := handler.hashCache.Hash(testPath)

	// Subscribe first using string path
	subReq := newTestSubscribeRequest(100, testPath, []byte("hash-filter"))
	err := handler.HandleMessage(context.Background(), session, subReq)
	core.AssertNoError(t, err, "subscribe")
	core.AssertEqual(t, 1, countSubscriptions(t, handler, pathHash), "subscriptions")

	// Unsubscribe using hash-based request
	unsubscribeReq := &nanorpc.NanoRPCRequest{
		RequestId:   100,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   nanorpc.GetPathOneOfHash(pathHash),
		Data:        []byte{}, // Empty data = unsubscribe
	}
	err = handler.HandleMessage(context.Background(), session, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Verify subscription removed
	core.AssertEqual(t, 0, countSubscriptions(t, handler, pathHash), "subscriptions")
	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_OK, "")
}

// testUnsubscribeNonExistentSubscription tests unsubscribe when no subscription exists
func testUnsubscribeNonExistentSubscription(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session := newTestSession("non-existent-session", 0)
	testPath := testSubscriptionPath
	pathHash, _ := handler.hashCache.Hash(testPath)

	// Register a handler for the path but don't subscribe
	err := handler.RegisterHandlerFunc(testPath, func(_ context.Context, reqCtx *RequestContext) error {
		return reqCtx.SendOK(nil)
	})
	core.AssertNoError(t, err, "register")

	// Try to unsubscribe non-existent subscription
	unsubscribeReq := newUnsubscribeRequest(200)
	err = handler.HandleMessage(context.Background(), session, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Should succeed with OK status (normal request handling)
	core.AssertEqual(t, 0, countSubscriptions(t, handler, pathHash), "subscriptions")
	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_OK, "")
}

// testUnsubscribeWithMismatchedRequestID tests unsubscribe with wrong request ID
func testUnsubscribeWithMismatchedRequestID(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session := newTestSession("mismatch-session", 0)
	testPath := testSubscriptionPath
	pathHash, _ := handler.hashCache.Hash(testPath)

	// Subscribe with request ID 300
	subReq := newTestSubscribeRequest(300, testPath, []byte("mismatch-filter"))
	err := handler.HandleMessage(context.Background(), session, subReq)
	core.AssertNoError(t, err, "subscribe")
	core.AssertEqual(t, 1, countSubscriptions(t, handler, pathHash), "subscriptions")

	// Try to unsubscribe with different request ID (no handler registered)
	unsubscribeReq := newUnsubscribeRequest(301)
	err = handler.HandleMessage(context.Background(), session, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Original subscription should remain (request ID mismatch)
	core.AssertEqual(t, 1, countSubscriptions(t, handler, pathHash), "subscriptions")
	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_NOT_FOUND, "")
}

// testUnsubscribeWithZeroPathHash tests unsubscribe with zero path hash
func testUnsubscribeWithZeroPathHash(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	session := newTestSession("zero-hash-session", 0)

	// Try to unsubscribe with zero path hash
	unsubscribeReq := &nanorpc.NanoRPCRequest{
		RequestId:   400,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   nanorpc.GetPathOneOfHash(0),
		Data:        []byte{}, // Empty data = unsubscribe
	}
	err := handler.HandleMessage(context.Background(), session, unsubscribeReq)
	core.AssertNoError(t, err, "unsubscribe")

	// Should return NOT_FOUND (no handler for empty path)
	core.AssertEqual(t, 0, countSubscriptions(t, handler, 0), "subscriptions")
	verifyResponse(t, session.lastResponse, nanorpc.NanoRPCResponse_STATUS_NOT_FOUND, "")
}

// TestUnsubscribeConcurrency tests concurrent unsubscribe operations
func TestUnsubscribeConcurrency(t *testing.T) {
	handler, pathHash := setupConcurrentTestHandler(t)

	const numGoroutines = 50
	const numSubscriptionsPerGoroutine = 5

	// Phase 1: Concurrent subscribe operations
	runConcurrentSubscriptions(t, handler, pathHash, numGoroutines, numSubscriptionsPerGoroutine)

	// Phase 2: Concurrent unsubscribe operations
	runConcurrentUnsubscribeOperations(t, handler, pathHash, numGoroutines, numSubscriptionsPerGoroutine)
}

// setupConcurrentTestHandler creates and configures a handler for concurrent testing
func setupConcurrentTestHandler(t *testing.T) (*DefaultMessageHandler, uint32) {
	t.Helper()

	handler := NewDefaultMessageHandler(nil)
	testPath := testSubscriptionPath
	pathHash, _ := handler.hashCache.Hash(testPath)

	// Register handler for the path
	err := handler.RegisterHandlerFunc(testPath, func(_ context.Context, reqCtx *RequestContext) error {
		return reqCtx.SendOK(nil)
	})
	core.AssertNoError(t, err, "register")

	return handler, pathHash
}

// runConcurrentSubscriptions executes concurrent subscribe operations
func runConcurrentSubscriptions(t *testing.T, handler *DefaultMessageHandler, pathHash uint32,
	numGoroutines, numSubscriptionsPerGoroutine int) {
	t.Helper()

	subscribeHelper := &testutils.ConcurrentTestHelper{
		NumGoroutines: numGoroutines,
		Timeout:       10 * time.Second,
		TestFunc: func(id int) error {
			return performSubscriptions(handler, id, numSubscriptionsPerGoroutine)
		},
	}

	subscribeErrors := subscribeHelper.Run()
	for i, err := range subscribeErrors {
		core.AssertNoError(t, err, "subscribe goroutine %d", i)
	}

	// Verify all subscriptions were created
	expectedSubscriptions := numGoroutines * numSubscriptionsPerGoroutine
	core.AssertEqual(t, expectedSubscriptions, countSubscriptions(t, handler, pathHash),
		"subscriptions")
}

// runConcurrentUnsubscribeOperations executes concurrent unsubscribe operations
func runConcurrentUnsubscribeOperations(t *testing.T, handler *DefaultMessageHandler, pathHash uint32,
	numGoroutines, numSubscriptionsPerGoroutine int) {
	t.Helper()

	unsubscribeHelper := &testutils.ConcurrentTestHelper{
		NumGoroutines: numGoroutines,
		Timeout:       10 * time.Second,
		TestFunc: func(id int) error {
			return performUnsubscribeOperations(handler, id, numSubscriptionsPerGoroutine)
		},
	}

	unsubscribeErrors := unsubscribeHelper.Run()
	for i, err := range unsubscribeErrors {
		core.AssertNoError(t, err, "unsubscribe goroutine %d", i)
	}

	// Verify all subscriptions were removed
	testutils.AssertWaitForCondition(t,
		func() bool { return countSubscriptions(t, handler, pathHash) == 0 },
		5*time.Second, "subscriptions removed")
}

// performSubscriptions creates subscriptions for a single goroutine
func performSubscriptions(handler *DefaultMessageHandler, id, numSubscriptionsPerGoroutine int) error {
	session := newTestSession(fmt.Sprintf("concurrent-session-%d", id), 0)
	testPath := testSubscriptionPath

	for i := 0; i < numSubscriptionsPerGoroutine; i++ {
		requestID := int32(id*1000 + i)
		subReq := newTestSubscribeRequest(requestID, testPath, []byte(fmt.Sprintf("filter-%d-%d", id, i)))
		if err := handler.HandleMessage(context.Background(), session, subReq); err != nil {
			return fmt.Errorf("subscribe failed for session %d, req %d: %w", id, i, err)
		}
	}
	return nil
}

// performUnsubscribeOperations removes subscriptions for a single goroutine
func performUnsubscribeOperations(handler *DefaultMessageHandler, id, numSubscriptionsPerGoroutine int) error {
	session := newTestSession(fmt.Sprintf("concurrent-session-%d", id), 0)

	for i := 0; i < numSubscriptionsPerGoroutine; i++ {
		requestID := int32(id*1000 + i)
		unsubscribeReq := newUnsubscribeRequest(requestID)
		if err := handler.HandleMessage(context.Background(), session, unsubscribeReq); err != nil {
			return fmt.Errorf("unsubscribe failed for session %d, req %d: %w", id, i, err)
		}
	}
	return nil
}

// TestUnsubscribeByRequestIDEdgeCases tests the unsubscribeByRequestID method directly
func TestUnsubscribeByRequestIDEdgeCases(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)

	t.Run("nil handler", testUnsubscribeByRequestIDNilHandler)
	t.Run("empty session ID", func(t *testing.T) {
		testUnsubscribeByRequestIDEmptySession(t, handler)
	})
	t.Run("zero path hash", func(t *testing.T) {
		testUnsubscribeByRequestIDZeroHash(t, handler)
	})
	t.Run("non-existent path hash", func(t *testing.T) {
		testUnsubscribeByRequestIDNonExistentHash(t, handler)
	})
}

func testUnsubscribeByRequestIDNilHandler(t *testing.T) {
	var nilHandler *DefaultMessageHandler
	removed := nilHandler.unsubscribeByRequestID("session", 123, 456)
	core.AssertEqual(t, false, removed, "removed")
}

func testUnsubscribeByRequestIDEmptySession(t *testing.T, handler *DefaultMessageHandler) {
	removed := handler.unsubscribeByRequestID("", 123, 456)
	core.AssertEqual(t, false, removed, "removed")
}

func testUnsubscribeByRequestIDZeroHash(t *testing.T, handler *DefaultMessageHandler) {
	removed := handler.unsubscribeByRequestID("session", 123, 0)
	core.AssertEqual(t, false, removed, "removed")
}

func testUnsubscribeByRequestIDNonExistentHash(t *testing.T, handler *DefaultMessageHandler) {
	removed := handler.unsubscribeByRequestID("session", 123, 999999)
	core.AssertEqual(t, false, removed, "removed")
}
