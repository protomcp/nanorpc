package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/utils/testutils"
)

func TestDefaultSessionManager_AddSession(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	sm := NewDefaultSessionManager(handler, nil)

	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := sm.AddSession(conn)

	if !core.AssertNotNil(t, session, "session created") {
		t.FailNow()
	}

	if !core.AssertNotEqual(t, "", session.ID(), "session ID") {
		t.FailNow()
	}

	// Verify session was stored
	retrievedSession := sm.GetSession(session.ID())
	if !core.AssertEqual(t, session, retrievedSession, "retrieved session") {
		t.FailNow()
	}
}

func TestDefaultSessionManager_RemoveSession(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	sm := NewDefaultSessionManager(handler, nil)

	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := sm.AddSession(conn)
	sessionID := session.ID()

	// Verify session exists
	if sm.GetSession(sessionID) == nil {
		t.Fatal("Expected session to exist")
	}

	// Remove session
	sm.RemoveSession(sessionID)

	// Verify session was removed
	if sm.GetSession(sessionID) != nil {
		t.Fatal("Expected session to be removed")
	}
}

func TestDefaultSessionManager_GetSession_NotFound(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	sm := NewDefaultSessionManager(handler, nil)

	session := sm.GetSession("non-existent")
	if session != nil {
		t.Fatal("Expected nil for non-existent session")
	}
}

func TestDefaultSessionManager_Shutdown(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	sm := NewDefaultSessionManager(handler, nil)

	// Add multiple sessions
	conn1 := &mockConn{remoteAddr: "127.0.0.1:12345"}
	conn2 := &mockConn{remoteAddr: "127.0.0.1:12346"}

	session1 := sm.AddSession(conn1)
	session2 := sm.AddSession(conn2)

	// Verify sessions exist
	if sm.GetSession(session1.ID()) == nil {
		t.Fatal("Expected session1 to exist")
	}
	if sm.GetSession(session2.ID()) == nil {
		t.Fatal("Expected session2 to exist")
	}

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := sm.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Expected no error during shutdown, got %v", err)
	}

	// Verify sessions were removed
	if sm.GetSession(session1.ID()) != nil {
		t.Fatal("Expected session1 to be removed")
	}
	if sm.GetSession(session2.ID()) != nil {
		t.Fatal("Expected session2 to be removed")
	}

	// Verify connections were closed
	if !conn1.closed {
		t.Fatal("Expected conn1 to be closed")
	}
	if !conn2.closed {
		t.Fatal("Expected conn2 to be closed")
	}
}

// Test session manager with logger to cover logging paths
func TestDefaultSessionManager_WithLogger(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	logger := testutils.NewMockFieldLogger()
	sm := NewDefaultSessionManager(handler, logger)

	// Test AddSession with logger
	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := sm.AddSession(conn)

	if !core.AssertNotNil(t, session, "session created") {
		t.FailNow()
	}

	// Test RemoveSession with logger
	sessionID := session.ID()
	sm.RemoveSession(sessionID)

	// Verify session was removed
	if !core.AssertNil(t, sm.GetSession(sessionID), "session removed") {
		t.FailNow()
	}
}

// Test getLogger lazy initialization path
func TestDefaultSessionManager_GetLoggerLazy(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	sm := &DefaultSessionManager{
		handler:  handler,
		logger:   nil, // Force lazy init
		sessions: make(map[string]Session),
	}

	logger := sm.getLogger()
	if !core.AssertNotNil(t, logger, "lazy logger initialized") {
		t.FailNow()
	}
}

// Test shutdown with failing session close
func TestDefaultSessionManager_ShutdownWithErrors(t *testing.T) {
	handler := NewDefaultMessageHandler(nil)
	logger := testutils.NewMockFieldLogger()
	sm := NewDefaultSessionManager(handler, logger)

	// Add session with failing close
	conn := &mockConn{
		remoteAddr: "127.0.0.1:12345",
		closeErr:   errors.New("close failed"),
	}
	session := sm.AddSession(conn)

	// Shutdown should handle close error
	ctx := context.Background()
	err := sm.Shutdown(ctx)
	if !core.AssertNoError(t, err, "shutdown with errors") {
		t.FailNow()
	}

	// Verify session was still removed from map
	if !core.AssertNil(t, sm.GetSession(session.ID()), "session removed despite close error") {
		t.FailNow()
	}
}

// Test RemoveSession with subscription manager
func TestDefaultSessionManager_RemoveSessionWithSubscriptions(t *testing.T) {
	// Create handler that implements SubscriptionManager
	subMgr := &mockSubscriptionManager{}
	sm := NewDefaultSessionManager(subMgr, nil)

	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := sm.AddSession(conn)
	sessionID := session.ID()

	// Remove session should clean up subscriptions
	sm.RemoveSession(sessionID)

	if !core.AssertTrue(t, subMgr.cleanupCalled, "subscription cleanup called") {
		t.FailNow()
	}
	if !core.AssertEqual(t, sessionID, subMgr.cleanupSessionID, "cleanup session ID") {
		t.FailNow()
	}
}

// Mock subscription manager for testing
type mockSubscriptionManager struct {
	cleanupSessionID string
	cleanupCalled    bool
}

func (*mockSubscriptionManager) HandleMessage(_ context.Context, _ Session, _ *nanorpc.NanoRPCRequest) error {
	// No-op for testing
	return nil
}

func (m *mockSubscriptionManager) RemoveSubscriptionsForSession(sessionID string) {
	m.cleanupCalled = true
	m.cleanupSessionID = sessionID
}
