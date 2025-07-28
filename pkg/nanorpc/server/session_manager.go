package server

import (
	"context"
	"net"
	"sync"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"

	"protomcp.org/nanorpc/pkg/nanorpc/common"
)

// DefaultSessionManager implements SessionManager interface
type DefaultSessionManager struct {
	handler  MessageHandler
	logger   slog.Logger
	sessions map[string]Session
	mu       sync.RWMutex
}

// NewDefaultSessionManager creates a new session manager
func NewDefaultSessionManager(handler MessageHandler, logger slog.Logger) *DefaultSessionManager {
	// Add session manager component field to logger using common helper
	logger = common.WithComponent(logger, common.ComponentSessionManager)

	return &DefaultSessionManager{
		sessions: make(map[string]Session),
		handler:  handler,
		logger:   logger,
	}
}

// getLogger returns the configured logger or lazily initializes a discard logger
func (sm *DefaultSessionManager) getLogger() slog.Logger {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.logger == nil {
		sm.logger = discard.New()
	}
	return sm.logger
}

// AddSession creates a new session for the connection
func (sm *DefaultSessionManager) AddSession(conn net.Conn) Session {
	// Create the session first
	session := NewDefaultSession(conn, sm.handler, nil)
	sessionID := session.ID()

	// Create session logger with all relevant fields using common helpers
	sessionLogger := common.WithSessionID(sm.getLogger(), sessionID)
	sessionLogger = common.WithRemoteAddr(sessionLogger, conn.RemoteAddr())
	sessionLogger = common.WithComponent(sessionLogger, common.ComponentSession)

	// Update session with the logger
	session.logger = sessionLogger

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	// Log session creation using common helpers
	if l, ok := sm.WithInfo(); ok {
		l = common.WithSessionID(l, sessionID)
		l = common.WithRemoteAddr(l, conn.RemoteAddr())
		l.Print("Session created")
	}

	return session
}

// RemoveSession removes a session by ID
func (sm *DefaultSessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()

	// Clean up subscriptions for this session
	if subMgr, ok := sm.handler.(SubscriptionManager); ok {
		subMgr.RemoveSubscriptionsForSession(sessionID)
	}

	// Log session removal using common helpers
	if l, ok := sm.WithInfo(); ok {
		l = common.WithSessionID(l, sessionID)
		l.Print("Session removed")
	}
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
			if l, ok := sm.WithError(err); ok {
				l = common.WithSessionID(l, session.ID())
				l.Print("Failed to close session")
			}
		}
	}

	return nil
}
