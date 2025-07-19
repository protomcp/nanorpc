package server

import (
	"context"
	"sync"

	"darvaza.org/core"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

// RequestContext provides request information and response utilities
type RequestContext struct {
	Session Session
	Request *nanorpc.NanoRPCRequest
	Path    string // Resolved path (from string or hash)
}

// DefaultMessageHandler implements MessageHandler interface
type DefaultMessageHandler struct {
	handlers map[string]RequestHandler
	mu       sync.RWMutex
}

// NewDefaultMessageHandler creates a new message handler
func NewDefaultMessageHandler() *DefaultMessageHandler {
	return &DefaultMessageHandler{
		handlers: make(map[string]RequestHandler),
	}
}

// RegisterHandlerFunc registers a handler function for a specific path
func (h *DefaultMessageHandler) RegisterHandlerFunc(path string, fn RequestHandlerFunc) error {
	return h.RegisterHandler(path, fn)
}

// RegisterHandler registers a handler for a specific path
func (h *DefaultMessageHandler) RegisterHandler(path string, handler RequestHandler) error {
	if h == nil {
		return core.ErrNilReceiver
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.handlers == nil {
		h.handlers = make(map[string]RequestHandler)
	}

	if handler == nil {
		if _, exists := h.handlers[path]; exists {
			delete(h.handlers, path)
			return nil
		}
		return core.ErrNotExists
	} else if _, exists := h.handlers[path]; exists {
		return core.ErrExists
	}

	h.handlers[path] = handler
	return nil
}

// HandleMessage processes a decoded request
func (h *DefaultMessageHandler) HandleMessage(ctx context.Context, session Session, req *nanorpc.NanoRPCRequest) error {
	switch req.RequestType {
	case nanorpc.NanoRPCRequest_TYPE_PING:
		return h.handlePing(ctx, session, req)
	case nanorpc.NanoRPCRequest_TYPE_REQUEST:
		return h.handleRequest(ctx, session, req)
	default:
		// Ignore unsupported request types for now
		return nil
	}
}

// handlePing processes ping requests and sends pong responses
func (*DefaultMessageHandler) handlePing(_ context.Context, session Session, req *nanorpc.NanoRPCRequest) error {
	response := &nanorpc.NanoRPCResponse{
		RequestId:      req.RequestId,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_PONG,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}

	return session.SendResponse(req, response)
}

// handleRequest processes TYPE_REQUEST messages
func (h *DefaultMessageHandler) handleRequest(ctx context.Context, session Session, req *nanorpc.NanoRPCRequest) error {
	// Extract path from request
	var path string
	switch p := req.PathOneof.(type) {
	case *nanorpc.NanoRPCRequest_Path:
		path = p.Path
	case *nanorpc.NanoRPCRequest_PathHash:
		// TODO: resolve hash to path when hash cache is implemented
		path = "" // For now, unresolved
	}

	// Look up handler
	h.mu.RLock()
	handler, exists := h.handlers[path]
	h.mu.RUnlock()

	if !exists || handler == nil {
		// No handler registered, return NOT_FOUND
		response := &nanorpc.NanoRPCResponse{
			RequestId:       req.RequestId,
			ResponseType:    nanorpc.NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus:  nanorpc.NanoRPCResponse_STATUS_NOT_FOUND,
			ResponseMessage: "no handler registered for path",
		}

		return session.SendResponse(req, response)
	}

	// Create request context
	reqCtx := &RequestContext{
		Session: session,
		Request: req,
		Path:    path,
	}

	// Call the handler
	return handler.Handle(ctx, reqCtx)
}
