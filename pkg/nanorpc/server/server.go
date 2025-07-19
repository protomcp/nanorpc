package server

import (
	"context"
	"net"
	"sync"
)

// Server represents a decoupled NanoRPC server
type Server struct {
	listener       Listener
	sessionManager SessionManager
	messageHandler MessageHandler
	shutdown       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
}

// NewServer creates a new decoupled server
func NewServer(listener Listener, sessionManager SessionManager, messageHandler MessageHandler) *Server {
	return &Server{
		listener:       listener,
		sessionManager: sessionManager,
		messageHandler: messageHandler,
		shutdown:       make(chan struct{}),
	}
}

// NewDefaultServer creates a server with default components
func NewDefaultServer(netListener net.Listener) *Server {
	listener := NewListenerAdapter(netListener)
	handler := NewDefaultMessageHandler()
	sessionManager := NewDefaultSessionManager(handler)

	return NewServer(listener, sessionManager, handler)
}

// Serve starts serving requests
func (s *Server) Serve(ctx context.Context) error {
	connCh := make(chan net.Conn)
	errCh := make(chan error)

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
	session := s.sessionManager.AddSession(conn)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.sessionManager.RemoveSession(session.ID())

		if err := session.Handle(ctx); err != nil {
			// TODO: log session error when logger is available
			_ = err
		}
	}()
}

// gracefulShutdown performs graceful server shutdown
func (s *Server) gracefulShutdown(ctx context.Context) error {
	if err := s.listener.Close(); err != nil {
		// TODO: log error when logger is available
		_ = err
	}
	if err := s.sessionManager.Shutdown(ctx); err != nil {
		// TODO: log error when logger is available
		_ = err
	}
	s.wg.Wait()
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
