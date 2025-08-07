package utils

import (
	"net"

	"darvaza.org/slog"
)

// Field name constants for structured logging.
// These are exported to allow callers to perform field transformations
// or filtering based on known field names.
const (
	// Component fields
	FieldComponent = "component"

	// Session fields
	FieldSessionID  = "session_id"
	FieldRemoteAddr = "remote_addr"
	FieldLocalAddr  = "local_addr"

	// Request fields
	FieldRequestID   = "request_id"
	FieldRequestType = "request_type"
	FieldPath        = "path"
	FieldPathHash    = "path_hash"

	// Response fields
	FieldResponseType   = "response_type"
	FieldResponseStatus = "response_status"

	// Performance fields
	FieldDuration       = "duration_ms"
	FieldStartTime      = "start_time"
	FieldReconnectDelay = "reconnect_delay_ms"
	FieldAttempt        = "attempt"

	// Error field (using slog standard)
	FieldError = slog.ErrorFieldName // "error"

	// Connection fields
	FieldNetwork = "network"
	FieldState   = "state"

	// Queue fields
	FieldQueueSize  = "queue_size"
	FieldQueueDepth = "queue_depth"

	// Handler fields
	FieldHandlerName = "handler_name"
	FieldHandlerPath = "handler_path"

	// Subscription fields
	FieldSubscriptionCount = "subscription_count"
	FieldCallbackCount     = "callback_count"
)

// Component name constants for the FieldComponent field
const (
	// Server components
	ComponentServer         = "server"
	ComponentSessionManager = "session-manager"
	ComponentMessageHandler = "message-handler"
	ComponentListener       = "listener"

	// Client components
	ComponentClient          = "client"
	ComponentReconnect       = "reconnect"
	ComponentRequestQueue    = "request-queue"
	ComponentRequestCounter  = "request-counter"
	ComponentSubscriptionMgr = "subscription-mgr"

	// Shared components
	ComponentSession   = "session"
	ComponentHashCache = "hash-cache"
)

// State constants for the FieldState field
const (
	StateConnecting    = "connecting"
	StateConnected     = "connected"
	StateDisconnecting = "disconnecting"
	StateDisconnected  = "disconnected"
	StateReconnecting  = "reconnecting"
	StateShuttingDown  = "shutting_down"
)

// Logger helper functions for safe field addition

// WithRemoteAddr safely adds a remote address field to a logger.
// If logger or addr is nil, returns the original logger unchanged.
func WithRemoteAddr(logger slog.Logger, addr net.Addr) slog.Logger {
	if logger != nil && addr != nil {
		return logger.WithField(FieldRemoteAddr, addr.String())
	}
	return logger
}

// WithLocalAddr safely adds a local address field to a logger.
// If logger or addr is nil, returns the original logger unchanged.
func WithLocalAddr(logger slog.Logger, addr net.Addr) slog.Logger {
	if logger != nil && addr != nil {
		return logger.WithField(FieldLocalAddr, addr.String())
	}
	return logger
}

// WithConnAddrs safely adds both remote and local address fields from a connection.
// If logger, conn, or either address is nil, those fields are skipped.
func WithConnAddrs(logger slog.Logger, conn net.Conn) slog.Logger {
	if logger != nil && conn != nil {
		logger = WithRemoteAddr(logger, conn.RemoteAddr())
		logger = WithLocalAddr(logger, conn.LocalAddr())
	}
	return logger
}

// WithComponent adds a component field to a logger.
// If logger is nil, returns nil.
func WithComponent(logger slog.Logger, component string) slog.Logger {
	if logger != nil {
		return logger.WithField(FieldComponent, component)
	}
	return logger
}

// WithSessionID adds a session ID field to a logger.
// If logger is nil, returns nil.
func WithSessionID(logger slog.Logger, sessionID string) slog.Logger {
	if logger != nil {
		return logger.WithField(FieldSessionID, sessionID)
	}
	return logger
}

// WithError adds an error field to a logger.
// If logger or err is nil, returns the original logger unchanged.
func WithError(logger slog.Logger, err error) slog.Logger {
	if logger != nil && err != nil {
		return logger.WithField(FieldError, err)
	}
	return logger
}
