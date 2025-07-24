// Package common provides shared constants and types used by both client and server.
package common

import "darvaza.org/slog"

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
