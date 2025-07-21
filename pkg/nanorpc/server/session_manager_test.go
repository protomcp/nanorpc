package server

import (
	"context"
	"testing"
	"time"
)

func TestDefaultSessionManager_AddSession(t *testing.T) {
	handler := NewDefaultMessageHandler()
	sm := NewDefaultSessionManager(handler, nil)

	conn := &mockConn{remoteAddr: "127.0.0.1:12345"}
	session := sm.AddSession(conn)

	if session == nil {
		t.Fatal("Expected session to be created")
	}

	if session.ID() == "" {
		t.Fatal("Expected session to have an ID")
	}

	// Verify session was stored
	retrievedSession := sm.GetSession(session.ID())
	if retrievedSession != session {
		t.Fatal("Expected to retrieve the same session")
	}
}

func TestDefaultSessionManager_RemoveSession(t *testing.T) {
	handler := NewDefaultMessageHandler()
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
	handler := NewDefaultMessageHandler()
	sm := NewDefaultSessionManager(handler, nil)

	session := sm.GetSession("non-existent")
	if session != nil {
		t.Fatal("Expected nil for non-existent session")
	}
}

func TestDefaultSessionManager_Shutdown(t *testing.T) {
	handler := NewDefaultMessageHandler()
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
