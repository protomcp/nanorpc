package server

import (
	"context"
	"net"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// Listener handles connection acceptance
type Listener interface {
	// Accept waits for and returns the next connection
	Accept() (net.Conn, error)
	// Close closes the listener
	Close() error
	// Addr returns the listener's network address
	Addr() net.Addr
}

// SessionManager manages connection lifecycle and tracking
type SessionManager interface {
	// AddSession creates a new session for the connection
	AddSession(conn net.Conn) Session
	// RemoveSession removes a session by ID
	RemoveSession(sessionID string)
	// GetSession retrieves a session by ID
	GetSession(sessionID string) Session
	// Shutdown gracefully closes all sessions
	Shutdown(ctx context.Context) error
}

// Session represents a single client connection
type Session interface {
	// ID returns the unique session identifier
	ID() string
	// RemoteAddr returns the remote address
	RemoteAddr() string
	// Handle processes messages for this session
	Handle(ctx context.Context) error
	// SendResponse sends a NanoRPC response to the client
	// If req is provided, it will be used to fill envelope fields like RequestID
	SendResponse(req *nanorpc.NanoRPCRequest, response *nanorpc.NanoRPCResponse) error
	// Close closes the session
	Close() error
}

// MessageHandler processes protocol messages
type MessageHandler interface {
	// HandleMessage processes a decoded request
	HandleMessage(ctx context.Context, session Session, req *nanorpc.NanoRPCRequest) error
}

// RequestHandler handles incoming requests for a specific path
type RequestHandler interface {
	Handle(ctx context.Context, req *RequestContext) error
}

// RequestHandlerFunc is an adapter to allow ordinary functions to be used as RequestHandlers
type RequestHandlerFunc func(context.Context, *RequestContext) error

// Handle calls the function with the given context and request
func (f RequestHandlerFunc) Handle(ctx context.Context, req *RequestContext) error {
	if f == nil {
		return core.ErrNilReceiver
	}

	return f(ctx, req)
}
