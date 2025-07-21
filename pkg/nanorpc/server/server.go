package server

import (
	"context"
	"net"
	"sync"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

// Server represents a decoupled NanoRPC server
type Server struct {
	listener       Listener
	sessionManager SessionManager
	messageHandler MessageHandler
	logger         slog.Logger
	shutdown       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
}

// NewServer creates a new decoupled server
func NewServer(listener Listener, sessionManager SessionManager,
	messageHandler MessageHandler, logger slog.Logger) *Server {
	return &Server{
		listener:       listener,
		sessionManager: sessionManager,
		messageHandler: messageHandler,
		shutdown:       make(chan struct{}),
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
	connCh := make(chan net.Conn)
	errCh := make(chan error)

	s.getLogger().Info().
		WithField(FieldLocalAddr, s.listener.Addr().String()).
		Print("Server started")

	go s.acceptLoop(connCh, errCh)
	return s.serveLoop(ctx, connCh, errCh)
}

// acceptLoop runs the connection acceptance loop
func (s *Server) acceptLoop(connCh chan<- net.Conn, errCh chan<- error) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}
}

// serveLoop handles the main server event loop
func (s *Server) serveLoop(ctx context.Context, connCh <-chan net.Conn, errCh <-chan error) error {
	for {
		select {
		case <-ctx.Done():
			return s.gracefulShutdown(ctx)
		case <-s.shutdown:
			return s.gracefulShutdown(ctx)
		case conn := <-connCh:
			s.handleNewConnection(ctx, conn)
		case err := <-errCh:
			return s.handleAcceptError(ctx, err)
		}
	}
}

// handleNewConnection processes a new client connection
func (s *Server) handleNewConnection(ctx context.Context, conn net.Conn) {
	s.getLogger().Debug().
		WithField(FieldRemoteAddr, conn.RemoteAddr().String()).
		Print("Connection accepted")

	session := s.sessionManager.AddSession(conn)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.sessionManager.RemoveSession(session.ID())

		if err := session.Handle(ctx); err != nil {
			s.getLogger().Error().
				WithField(FieldSessionID, session.ID()).
				WithField(FieldError, err).
				Print("Session error")
		}
	}()
}

// gracefulShutdown performs graceful server shutdown
func (s *Server) gracefulShutdown(ctx context.Context) error {
	s.getLogger().Info().Print("Server shutting down")

	if err := s.listener.Close(); err != nil {
		s.getLogger().Warn().
			WithField(FieldError, err).
			Print("Failed to close listener")
	}
	if err := s.sessionManager.Shutdown(ctx); err != nil {
		s.getLogger().Warn().
			WithField(FieldError, err).
			Print("Session manager shutdown error")
	}
	s.wg.Wait()

	s.getLogger().Info().Print("Server shutdown complete")
	return nil
}

// handleAcceptError handles errors from the accept loop
func (s *Server) handleAcceptError(ctx context.Context, err error) error {
	select {
	case <-ctx.Done():
		return s.gracefulShutdown(ctx)
	case <-s.shutdown:
		return s.gracefulShutdown(ctx)
	default:
		return err
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.shutdown:
		return nil
	default:
		close(s.shutdown)
	}

	// Wait for graceful shutdown
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
