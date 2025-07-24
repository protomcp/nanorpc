package server

import (
	"context"
	"errors"
	"net"
	"sync"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/sync/workgroup"

	"github.com/amery/nanorpc/pkg/nanorpc/common"
)

// Server represents a decoupled NanoRPC server
type Server struct {
	listener       Listener
	sessionManager SessionManager
	messageHandler MessageHandler
	logger         slog.Logger
	wg             workgroup.Group
	mu             sync.RWMutex
}

// NewServer creates a new decoupled server
func NewServer(listener Listener, sessionManager SessionManager,
	messageHandler MessageHandler, logger slog.Logger) *Server {
	// Add server component field to logger
	if logger != nil {
		logger = logger.WithField(common.FieldComponent, common.ComponentServer)
	}

	return &Server{
		listener:       listener,
		sessionManager: sessionManager,
		messageHandler: messageHandler,
		logger:         logger,
	}
}

// NewDefaultServer creates a server with default components
func NewDefaultServer(netListener net.Listener, logger slog.Logger) *Server {
	listener := NewListenerAdapter(netListener)
	handler := NewDefaultMessageHandler()
	sessionManager := NewDefaultSessionManager(handler, logger)

	return NewServer(listener, sessionManager, handler, logger)
}

// getLogger returns the configured logger or lazily initializes a discard logger
func (s *Server) getLogger() slog.Logger {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logger == nil {
		s.logger = discard.New()
	}
	return s.logger
}

// Serve starts serving requests
func (s *Server) Serve(ctx context.Context) error {
	// Configure workgroup
	s.wg.Parent = ctx
	s.wg.OnCancel = s.onGroupCancel

	if l, ok := s.WithInfo(); ok {
		l.WithField(common.FieldLocalAddr, s.listener.Addr().String()).
			Print("Server started")
	}

	// Start accept loop in workgroup with error catching
	if err := s.wg.GoCatch(s.acceptLoop, s.catchAcceptError); err != nil {
		return err
	}

	// Wait for completion or cancellation
	err := s.wg.Wait()

	// Log final status
	if err != nil && err != context.Canceled {
		s.LogError(err, "Server stopped with error")
	} else {
		s.LogInfo("Server stopped")
	}

	return err
}

// acceptLoop runs the connection acceptance loop
func (s *Server) acceptLoop(ctx context.Context) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		// Check for cancellation after successful accept
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return ctx.Err()
		default:
			s.handleNewConnection(ctx, conn)
		}
	}
}

// handleNewConnection processes a new client connection
func (s *Server) handleNewConnection(_ context.Context, conn net.Conn) {
	s.logAccept(conn)

	session := s.sessionManager.AddSession(conn)

	// Handle session in workgroup with error catching
	sid := session.ID()
	_ = s.wg.GoCatch(
		func(ctx context.Context) error {
			defer s.sessionManager.RemoveSession(sid)
			return session.Handle(ctx)
		},
		func(ctx context.Context, err error) error {
			return s.catchSessionError(ctx, err, sid)
		},
	)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.LogInfo("Server shutting down")

	// Close listener to stop accepting new connections
	if err := s.listener.Close(); err != nil {
		if l, ok := s.WithWarn(); ok {
			l.WithField(common.FieldError, err).
				Print("Failed to close listener")
		}
	}

	// Cancel workgroup to signal all goroutines to stop
	s.wg.Cancel(context.Canceled)

	// Shutdown session manager
	if err := s.sessionManager.Shutdown(ctx); err != nil {
		if l, ok := s.WithWarn(); ok {
			l.WithField(common.FieldError, err).
				Print("Session manager shutdown error")
		}
	}

	// Use context with timeout for shutdown
	done := s.wg.Done()
	select {
	case <-done:
		s.LogInfo("Server shutdown complete")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// onGroupCancel is called when the workgroup is cancelled
func (s *Server) onGroupCancel(_ context.Context, err error) {
	if err != nil && err != context.Canceled {
		s.LogError(err, "Server cancelled with error")
	}
	// Ensure listener is closed on cancel
	if closeErr := s.listener.Close(); closeErr != nil {
		if l, ok := s.WithWarn(); ok {
			l.WithField(common.FieldError, closeErr).
				Print("Failed to close listener during cancel")
		}
	}
}

// catchAcceptError filters accept loop errors
func (s *Server) catchAcceptError(_ context.Context, err error) error {
	if err == nil {
		return nil
	}

	// Check for expected shutdown error
	if s.isExpectedAcceptError(err) {
		if l, ok := s.WithDebug(); ok {
			l.WithField(common.FieldError, err).
				Print("Accept loop stopped due to closed listener")
		}
		return nil
	}

	return err
}

// isExpectedAcceptError checks if the error is an expected accept error during shutdown
func (*Server) isExpectedAcceptError(err error) bool {
	opErr, ok := err.(*net.OpError)
	return ok && opErr.Op == "accept" && errors.Is(opErr.Err, net.ErrClosed)
}

// catchSessionError handles session errors
func (s *Server) catchSessionError(_ context.Context, err error, sessionID string) error {
	if err != nil && err != context.Canceled {
		if l, ok := s.WithError(err); ok {
			l.WithField(common.FieldSessionID, sessionID).
				Print("Session error")
		}
	}
	// Don't propagate session errors - they shouldn't crash the server
	return nil
}
