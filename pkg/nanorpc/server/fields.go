package server

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

	// Request fields
	FieldRequestID   = "request_id"
	FieldRequestType = "request_type"
	FieldPath        = "path"
	FieldPathHash    = "path_hash"

	// Response fields
	FieldResponseType   = "response_type"
	FieldResponseStatus = "response_status"

	// Performance fields
	FieldDuration  = "duration_ms"
	FieldStartTime = "start_time"

	// Error field (using slog standard)
	FieldError = slog.ErrorFieldName // "error"

	// Connection fields
	FieldLocalAddr = "local_addr"
	FieldNetwork   = "network"

	// Handler fields
	FieldHandlerName = "handler_name"
	FieldHandlerPath = "handler_path"
)

// Component name constants for the FieldComponent field
const (
	ComponentServer         = "server"
	ComponentSessionManager = "session-manager"
	ComponentSession        = "session"
	ComponentMessageHandler = "message-handler"
	ComponentListener       = "listener"
)
