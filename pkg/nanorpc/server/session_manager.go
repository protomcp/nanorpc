package server

import (
	"context"
	"net"
	"sync"
)

// DefaultSessionManager implements SessionManager interface
type DefaultSessionManager struct {
	sessions map[string]Session
	handler  MessageHandler
	mu       sync.RWMutex
}

// NewDefaultSessionManager creates a new session manager
func NewDefaultSessionManager(handler MessageHandler) *DefaultSessionManager {
	return &DefaultSessionManager{
		sessions: make(map[string]Session),
		handler:  handler,
	}
}

// AddSession creates a new session for the connection
func (sm *DefaultSessionManager) AddSession(conn net.Conn) Session {
	session := NewDefaultSession(conn, sm.handler)

	sm.mu.Lock()
	sm.sessions[session.ID()] = session
	sm.mu.Unlock()

	return session
}

// RemoveSession removes a session by ID
func (sm *DefaultSessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()
}

// GetSession retrieves a session by ID
func (sm *DefaultSessionManager) GetSession(sessionID string) Session {
	sm.mu.RLock()
	session := sm.sessions[sessionID]
	sm.mu.RUnlock()
	return session
}

// Shutdown gracefully closes all sessions
func (sm *DefaultSessionManager) Shutdown(_ context.Context) error {
	sm.mu.Lock()
	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	sm.sessions = make(map[string]Session)
	sm.mu.Unlock()

	// Close all sessions
	for _, session := range sessions {
		if err := session.Close(); err != nil {
			// TODO: log error when logger is available
			_ = err
		}
	}

	return nil
}
