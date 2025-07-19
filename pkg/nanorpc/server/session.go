package server

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"darvaza.org/core"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

// DefaultSession implements Session interface
type DefaultSession struct {
	id         string
	conn       net.Conn
	handler    MessageHandler
	remoteAddr string
}

// NewDefaultSession creates a new session
func NewDefaultSession(conn net.Conn, handler MessageHandler) *DefaultSession {
	return &DefaultSession{
		id:         generateSessionID(conn),
		conn:       conn,
		handler:    handler,
		remoteAddr: conn.RemoteAddr().String(),
	}
}

// ID returns the session identifier
func (s *DefaultSession) ID() string {
	return s.id
}

// RemoteAddr returns the remote address
func (s *DefaultSession) RemoteAddr() string {
	return s.remoteAddr
}

// Handle processes messages for this session
func (s *DefaultSession) Handle(ctx context.Context) error {
	defer s.Close()

	scanner := bufio.NewScanner(s.conn)
	scanner.Split(nanorpc.Split)

	for {
		if err := s.processNextMessage(ctx, scanner); err != nil {
			if err == nanorpc.ErrSessionClosed {
				return nil
			}
			return err
		}
	}
}

// processNextMessage reads and processes a single message
func (s *DefaultSession) processNextMessage(ctx context.Context, scanner *bufio.Scanner) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set read deadline to prevent indefinite blocking
	if deadline, ok := ctx.Deadline(); ok {
		if err := s.conn.SetReadDeadline(deadline); err != nil {
			return core.Wrap(err, "SetReadDeadline")
		}
	}

	// Read next message
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}
		return nanorpc.ErrSessionClosed // EOF
	}

	// Decode and handle
	return s.decodeAndHandle(ctx, scanner.Bytes())
}

// decodeAndHandle decodes a request and passes it to the handler
func (s *DefaultSession) decodeAndHandle(ctx context.Context, data []byte) error {
	req, _, err := nanorpc.DecodeRequest(data)
	if err != nil {
		return core.Wrap(err, "decode")
	}

	if err := s.handler.HandleMessage(ctx, s, req); err != nil {
		// TODO: Add proper logging for handler errors
		return nil // Continue on handler errors
	}

	return nil
}

// Close closes the session
func (s *DefaultSession) Close() error {
	return s.conn.Close()
}

// Write sends data to the client
func (s *DefaultSession) Write(data []byte) (int, error) {
	return s.conn.Write(data)
}

// generateSessionID creates a unique session identifier
func generateSessionID(conn net.Conn) string {
	return fmt.Sprintf("session-%s", conn.RemoteAddr().String())
}
