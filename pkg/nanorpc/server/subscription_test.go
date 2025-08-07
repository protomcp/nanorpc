package server

import (
	"context"
	"sync"
	"testing"
	"time"
	"unsafe"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// Test factory helpers

func newTestSubscription(sessionID string, requestID int32, pathHash uint32) *ActiveSubscription {
	return &ActiveSubscription{
		Session:   newTestSession(sessionID, 0),
		RequestID: requestID,
		PathHash:  pathHash,
		CreatedAt: time.Now(),
		Filter:    []byte("filter-" + sessionID),
	}
}

func newTestSubscriptionWithFilter(session Session, requestID int32, pathHash uint32,
	filter []byte) *ActiveSubscription {
	return &ActiveSubscription{
		Session:   session,
		RequestID: requestID,
		PathHash:  pathHash,
		CreatedAt: time.Now(),
		Filter:    filter,
	}
}

func newTestSubscribeRequest(requestID int32, path string, filter []byte) *nanorpc.NanoRPCRequest {
	return &nanorpc.NanoRPCRequest{
		RequestId:   requestID,
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof:   nanorpc.GetPathOneOfString(path),
		Data:        filter,
	}
}

func newTestSubscribeRequestWithHash(requestID int32, pathHash uint32, filter []byte) *nanorpc.NanoRPCRequest {
	return &nanorpc.NanoRPCRequest{
		RequestId:   requestID,
		RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
		PathOneof:   nanorpc.GetPathOneOfHash(pathHash),
		Data:        filter,
	}
}

const (
	sessionID1          = "session1"
	sessionID2          = "session2"
	sessionID3          = "session3"
	concurrentSessionID = "concurrent-session"
)

func TestSubscriptionMapOperations(t *testing.T) {
	sm := make(SubscriptionMap)

	// Test empty map
	core.AssertNil(t, sm.GetSubscribers(123), "subscribers")

	// Create test subscriptions
	sub1 := newTestSubscription(sessionID1, 1, 123)
	sub2 := newTestSubscription(sessionID2, 2, 123) // Same path hash
	sub3 := newTestSubscription(sessionID1, 3, 456) // Different path hash

	// Test AddSubscription
	sm.AddSubscription(123, sub1)
	subList := sm.GetSubscribers(123)
	core.AssertNotNil(t, subList, "subscription list")
	core.AssertEqual(t, 1, subList.Len(), "subscription count")

	// Add second subscription to same path
	sm.AddSubscription(123, sub2)
	subList = sm.GetSubscribers(123)
	core.AssertEqual(t, 2, subList.Len(), "subscription count")

	// Add subscription to different path
	sm.AddSubscription(456, sub3)
	subList456 := sm.GetSubscribers(456)
	core.AssertNotNil(t, subList456, "subscription list")
	core.AssertEqual(t, 1, subList456.Len(), "subscription count")

	// Verify original path still has two subscriptions
	subList = sm.GetSubscribers(123)
	core.AssertEqual(t, 2, subList.Len(), "subscription count")

	// Test RemoveForSession - remove session1
	sm.RemoveForSession(sessionID1)

	// Path 123 should now have only one subscription (session2)
	subList = sm.GetSubscribers(123)
	core.AssertNotNil(t, subList, "subscription list")
	core.AssertEqual(t, 1, subList.Len(), "subscription count")

	// Verify it's session2's subscription
	var foundSession2 bool
	subList.ForEach(func(sub *ActiveSubscription) bool {
		if sub.Session.ID() == sessionID2 {
			foundSession2 = true
		}
		return true
	})
	core.AssertTrue(t, foundSession2, "session2 found")

	// Path 456 should be removed entirely (it only had session1's subscription)
	subList456 = sm.GetSubscribers(456)
	core.AssertNil(t, subList456, "path removed")

	// Remove session2
	sm.RemoveForSession(sessionID2)

	// Path 123 should now be removed entirely
	subList = sm.GetSubscribers(123)
	core.AssertNil(t, subList, "path removed")

	// Test RemoveForSession with non-existent session (should not panic)
	sm.RemoveForSession("non-existent")
}

func TestActiveSubscriptionFieldAlignment(t *testing.T) {
	// Verify field alignment for memory efficiency
	var sub ActiveSubscription

	// Check that 8-byte aligned fields come first
	sessionOffset := unsafe.Offsetof(sub.Session)
	createdAtOffset := unsafe.Offsetof(sub.CreatedAt)
	filterOffset := unsafe.Offsetof(sub.Filter)

	// Check that 4-byte aligned fields come after
	requestIDOffset := unsafe.Offsetof(sub.RequestID)
	pathHashOffset := unsafe.Offsetof(sub.PathHash)

	// Session and CreatedAt should be 8-byte aligned
	core.AssertEqual(t, uintptr(0), sessionOffset%8, "Session alignment")
	core.AssertEqual(t, uintptr(0), createdAtOffset%8, "CreatedAt alignment")
	core.AssertEqual(t, uintptr(0), filterOffset%8, "Filter alignment")

	// RequestID and PathHash should be 4-byte aligned
	core.AssertEqual(t, uintptr(0), requestIDOffset%4, "RequestID alignment")
	core.AssertEqual(t, uintptr(0), pathHashOffset%4, "PathHash alignment")

	// Verify the order is optimal (8-byte fields first)
	core.AssertTrue(t, sessionOffset < requestIDOffset, "field ordering")
	core.AssertTrue(t, createdAtOffset < requestIDOffset, "field ordering")
	core.AssertTrue(t, filterOffset < requestIDOffset, "field ordering")
}

// subscribeTestCase represents a test case for the Subscribe method
type subscribeTestCase struct {
	setupHandler   func() *DefaultMessageHandler
	setupSession   func() *mockSession
	request        *nanorpc.NanoRPCRequest
	verifyFunc     func(t *testing.T, h *DefaultMessageHandler, s *mockSession)
	name           string
	expectedStatus nanorpc.NanoRPCResponse_Status
	expectError    bool
}

// test executes the test case
func (tc *subscribeTestCase) test(t *testing.T) {
	handler := tc.setupHandler()
	session := tc.setupSession()

	err := handler.Subscribe(context.Background(), session, tc.request)

	if tc.expectError {
		core.AssertError(t, err, "error")
		return
	}

	core.AssertNoError(t, err, "error")

	// Verify response
	response := session.GetLastResponse()
	core.AssertNotNil(t, response, "response")
	core.AssertEqual(t, tc.request.RequestId, response.RequestId, "request ID")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_TYPE_RESPONSE, response.ResponseType, "response type")
	core.AssertEqual(t, tc.expectedStatus, response.ResponseStatus, "response status")

	// Run additional verifications
	if tc.verifyFunc != nil {
		tc.verifyFunc(t, handler, session)
	}
}

// Test case factory functions for Subscribe method

func testSuccessfulSubscriptionWithStringPath() subscribeTestCase {
	return subscribeTestCase{
		name:           "successful subscription with string path",
		setupHandler:   func() *DefaultMessageHandler { return NewDefaultMessageHandler(nil) },
		setupSession:   func() *mockSession { return newTestSession("", 0) },
		request:        newTestSubscribeRequest(123, "/test/path", []byte("filter-data")),
		expectedStatus: nanorpc.NanoRPCResponse_STATUS_OK,
		verifyFunc: func(t *testing.T, h *DefaultMessageHandler, s *mockSession) {
			// Verify subscription was added
			pathHash, err := h.hashCache.Hash("/test/path")
			core.AssertNoError(t, err, "hash error")

			subList := h.subscriptions.GetSubscribers(pathHash)
			core.AssertNotNil(t, subList, "subscription list")
			core.AssertEqual(t, 1, subList.Len(), "subscription count")

			// Verify subscription details
			var foundSub *ActiveSubscription
			subList.ForEach(func(sub *ActiveSubscription) bool {
				foundSub = sub
				return true
			})
			core.AssertNotNil(t, foundSub, "subscription")
			core.AssertEqual(t, s.ID(), foundSub.Session.ID(), "session ID")
			core.AssertEqual(t, int32(123), foundSub.RequestID, "request ID")
			core.AssertEqual(t, pathHash, foundSub.PathHash, "path hash")
			core.AssertEqual(t, "filter-data", string(foundSub.Filter), "filter")
		},
	}
}

func testSuccessfulSubscriptionWithHashPath() subscribeTestCase {
	// Pre-compute the hash and handler to ensure consistency
	testPath := "/test/path"
	handler := NewDefaultMessageHandler(nil)
	pathHash, _ := handler.hashCache.Hash(testPath)

	return subscribeTestCase{
		name: "successful subscription with hash path",
		setupHandler: func() *DefaultMessageHandler {
			// Return the same handler instance with pre-populated cache
			return handler
		},
		setupSession:   func() *mockSession { return newTestSession(sessionID2, 2002) },
		request:        newTestSubscribeRequestWithHash(456, pathHash, []byte("filter-data-2")),
		expectedStatus: nanorpc.NanoRPCResponse_STATUS_OK,
		verifyFunc: func(t *testing.T, h *DefaultMessageHandler, _ *mockSession) {
			subList := h.subscriptions.GetSubscribers(pathHash)
			core.AssertNotNil(t, subList, "subscription list")
			core.AssertEqual(t, 1, subList.Len(), "subscription count")

			// Verify subscription details
			var foundSub *ActiveSubscription
			subList.ForEach(func(sub *ActiveSubscription) bool {
				foundSub = sub
				return true
			})
			core.AssertNotNil(t, foundSub, "subscription")
			core.AssertEqual(t, pathHash, foundSub.PathHash, "path hash")
			core.AssertEqual(t, "filter-data-2", string(foundSub.Filter), "filter")
		},
	}
}

func testSubscriptionWithInvalidPath() subscribeTestCase {
	return subscribeTestCase{
		name:         "subscription with invalid path",
		setupHandler: func() *DefaultMessageHandler { return NewDefaultMessageHandler(nil) },
		setupSession: func() *mockSession { return newTestSession(sessionID3, 3003) },
		request: &nanorpc.NanoRPCRequest{
			RequestId:   789,
			RequestType: nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE,
			PathOneof:   nil, // Invalid path
		},
		expectedStatus: nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR,
	}
}

func testNilHandlerReturnsError() subscribeTestCase {
	return subscribeTestCase{
		name:         "nil handler returns error",
		setupHandler: func() *DefaultMessageHandler { return nil },
		setupSession: func() *mockSession { return newTestSession("", 0) },
		request:      newTestSubscribeRequest(999, "/test", nil),
		expectError:  true,
	}
}

// subscribeTestCases returns test cases for Subscribe method
func subscribeTestCases() []subscribeTestCase {
	return []subscribeTestCase{
		testSuccessfulSubscriptionWithStringPath(),
		testSuccessfulSubscriptionWithHashPath(),
		testSubscriptionWithInvalidPath(),
		testNilHandlerReturnsError(),
	}
}

func TestSubscribeMethod(t *testing.T) {
	for _, tc := range subscribeTestCases() {
		t.Run(tc.name, tc.test)
	}
}

func TestPublishByHashNoSubscribers(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)

	// Test publishing to path with no subscribers
	err := handler.PublishByHash(12345, []byte("test-data"))
	core.AssertNoError(t, err, "publish error")

	// Test nil handler
	err = (*DefaultMessageHandler)(nil).PublishByHash(12345, []byte("test-data"))
	core.AssertError(t, err, "nil handler error")
}

// Helper function to verify update details
func verifyUpdateDetails(t *testing.T, updates []pendingUpdate) {
	for i, update := range updates {
		core.AssertNotNil(t, update.session, "session")
		core.AssertNotNil(t, update.message, "message")
		core.AssertEqual(t, nanorpc.NanoRPCResponse_TYPE_UPDATE, update.message.ResponseType,
			"response type")
		core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, update.message.ResponseStatus,
			"response status")
		core.AssertEqual(t, "update-data", string(update.message.Data), "data")

		// Check request ID matches subscription
		if update.session.ID() == sessionID1 {
			core.AssertEqual(t, int32(100), update.message.RequestId, "request ID")
		} else if update.session.ID() == sessionID2 {
			core.AssertEqual(t, int32(200), update.message.RequestId, "request ID")
		} else {
			t.Errorf("unexpected session ID: %s", update.session.ID())
		}
		_ = i // silence unused variable warning
	}
}

func TestCollectPendingUpdates(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)

	t.Run("NoSubscribers", func(t *testing.T) {
		updates := handler.collectPendingUpdates(12345, []byte("test-data"))
		core.AssertEqual(t, 0, len(updates), "update count")
	})

	t.Run("WithSubscribers", func(t *testing.T) {
		// Add some subscribers
		sub1 := newTestSubscriptionWithFilter(newTestSession(sessionID1, 1001), 100, 12345, []byte("filter1"))
		sub2 := newTestSubscriptionWithFilter(newTestSession(sessionID2, 1002), 200, 12345, []byte("filter2"))

		handler.subscriptions.AddSubscription(12345, sub1)
		handler.subscriptions.AddSubscription(12345, sub2)

		// Verify both subscriptions were added
		subList := handler.subscriptions.GetSubscribers(12345)
		core.AssertEqual(t, 2, subList.Len(), "subscription count")

		// Test collecting updates
		testData := []byte("update-data")
		updates := handler.collectPendingUpdates(12345, testData)
		core.AssertEqual(t, 2, len(updates), "update count")

		// Verify update details
		verifyUpdateDetails(t, updates)
	})

	t.Run("WithNilSession", func(t *testing.T) {
		// Test with subscription that has nil session
		sub3 := &ActiveSubscription{
			Session:   nil, // Nil session
			RequestID: 300,
			PathHash:  12345,
			CreatedAt: time.Now(),
			Filter:    []byte("filter3"),
		}

		handler.subscriptions.AddSubscription(12345, sub3)

		// Should still return only 2 updates (skipping nil session)
		updates := handler.collectPendingUpdates(12345, []byte("update-data"))
		core.AssertEqual(t, 2, len(updates), "update count")
	})
}

func TestRemoveSubscriptionsForSession(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)

	// Add subscriptions for multiple sessions and paths
	sub1 := newTestSubscription(sessionID1, 100, 123)
	sub2 := newTestSubscription(sessionID2, 200, 123)
	sub3 := newTestSubscription(sessionID1, 300, 456)

	handler.subscriptions.AddSubscription(123, sub1)
	handler.subscriptions.AddSubscription(123, sub2)
	handler.subscriptions.AddSubscription(456, sub3)

	// Verify initial state
	core.AssertEqual(t, 2, handler.subscriptions.GetSubscribers(123).Len(), "subscription count")
	core.AssertEqual(t, 1, handler.subscriptions.GetSubscribers(456).Len(), "subscription count")

	// Remove session1's subscriptions
	handler.RemoveSubscriptionsForSession(sessionID1)

	// Path 123 should have only session2's subscription
	subList123 := handler.subscriptions.GetSubscribers(123)
	core.AssertNotNil(t, subList123, "subscription list")
	core.AssertEqual(t, 1, subList123.Len(), "subscription count")

	var foundSession2 bool
	subList123.ForEach(func(sub *ActiveSubscription) bool {
		if sub.Session.ID() == sessionID2 {
			foundSession2 = true
		}
		return true
	})
	core.AssertTrue(t, foundSession2, "session2 found")

	// Path 456 should be removed entirely
	subList456 := handler.subscriptions.GetSubscribers(456)
	core.AssertNil(t, subList456, "path removed")

	// Test with nil handler (should not panic)
	(*DefaultMessageHandler)(nil).RemoveSubscriptionsForSession("test")
}

// startPublishWorkers starts multiple goroutines that publish updates
// revive:disable-next-line:argument-limit
func startPublishWorkers(t *testing.T, wg *sync.WaitGroup, handler *DefaultMessageHandler,
	pathHash uint32, numWorkers, numOps int) {
	t.Helper()
	for range numWorkers {
		wg.Add(1)
		go func() {
			t.Helper()
			defer wg.Done()
			for range numOps {
				data := []byte("test-data")
				err := handler.PublishByHash(pathHash, data)
				core.AssertNoError(t, err, "publish error")
			}
		}()
	}
}

// startSubscribeWorkers starts multiple goroutines that add subscriptions
// revive:disable-next-line:argument-limit
func startSubscribeWorkers(t *testing.T, wg *sync.WaitGroup, handler *DefaultMessageHandler,
	testPath string, numWorkers, numOps int) {
	t.Helper()
	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			t.Helper()
			defer wg.Done()
			for j := range numOps {
				// Use Subscribe method with string path for thread safety
				session := newTestSession(concurrentSessionID, uint16(workerID))
				req := newTestSubscribeRequest(int32(workerID*1000+j), testPath, nil)
				err := handler.Subscribe(context.Background(), session, req)
				core.AssertNoError(t, err, "subscribe error")
			}
		}(i)
	}
}

func TestPublishByHashLockSafety(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)

	// Register a test path and get its hash
	testPath := "/test/concurrent/path"
	pathHash, err := handler.hashCache.Hash(testPath)
	core.AssertNoError(t, err, "hash error")

	// Add initial subscription using Subscribe method with string path
	session := newTestSession("", 0)
	req := newTestSubscribeRequest(123, testPath, []byte("filter"))
	err = handler.Subscribe(context.Background(), session, req)
	core.AssertNoError(t, err, "subscribe error")

	// Test concurrent operations
	var wg sync.WaitGroup
	const numGoroutines = 10
	const numOperations = 100

	startPublishWorkers(t, &wg, handler, pathHash, numGoroutines, numOperations)
	startSubscribeWorkers(t, &wg, handler, testPath, numGoroutines, numOperations)

	wg.Wait()

	// Verify results
	subList := handler.subscriptions.GetSubscribers(pathHash)
	core.AssertNotNil(t, subList, "subscription list")
	core.AssertTrue(t, subList.Len() >= 1, "subscription count")

	// Verify updates were received
	responses := session.GetAllResponses()
	core.AssertTrue(t, len(responses) > 0,
		"response count")

	// Count TYPE_UPDATE messages (skip the initial TYPE_RESPONSE from Subscribe)
	updateCount := 0
	for _, resp := range responses {
		if resp.ResponseType == nanorpc.NanoRPCResponse_TYPE_UPDATE {
			updateCount++
			core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, resp.ResponseStatus,
				"response status")
		}
	}
	core.AssertTrue(t, updateCount > 0, "update count")
}
