package server

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"

	"github.com/amery/nanorpc/pkg/nanorpc/common"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

// DefaultSession implements Session interface
type DefaultSession struct {
	conn       net.Conn
	handler    MessageHandler
	logger     slog.Logger
	id         string
	remoteAddr string
	mu         sync.Mutex
}

// NewDefaultSession creates a new session
func NewDefaultSession(conn net.Conn, handler MessageHandler, logger slog.Logger) *DefaultSession {
	sessionID := generateSessionID(conn)
	remoteAddr := conn.RemoteAddr().String()

	// Create annotated logger with session fields if logger is provided
	var sessionLogger slog.Logger
	if logger != nil {
		sessionLogger = logger.
			WithField(common.FieldComponent, common.ComponentSession).
			WithField(common.FieldSessionID, sessionID).
			WithField(common.FieldRemoteAddr, remoteAddr)
	}

	return &DefaultSession{
		id:         sessionID,
		conn:       conn,
		handler:    handler,
		remoteAddr: remoteAddr,
		logger:     sessionLogger,
	}
}

// getLogger returns the configured logger or lazily initializes a discard logger
func (s *DefaultSession) getLogger() slog.Logger {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logger == nil {
		s.logger = discard.New()
	}
	return s.logger
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
		s.getLogger().Error().
			WithField(common.FieldError, err).
			WithField("data_length", len(data)).
			WithField("data_preview", hexDump(data, 32)).
			Print("Failed to decode request")
		return core.Wrap(err, "decode")
	}

	if err := s.handler.HandleMessage(ctx, s, req); err != nil {
		s.getLogger().Error().
			WithField(common.FieldRequestID, req.GetRequestId()).
			WithField(common.FieldError, err).
			Print("Handler error")
		return nil // Continue on handler errors
	}

	return nil
}

// Close closes the session
func (s *DefaultSession) Close() error {
	return s.conn.Close()
}

// SendResponse sends a NanoRPC response to the client
func (s *DefaultSession) SendResponse(req *nanorpc.NanoRPCRequest, response *nanorpc.NanoRPCResponse) error {
	// Fill envelope fields from request if provided
	if req != nil && response.RequestId == 0 {
		response.RequestId = req.RequestId
	}

	// Encode the response
	data, err := nanorpc.EncodeResponse(response, nil)
	if err != nil {
		return err
	}

	// Send to client
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.conn.Write(data)
	return err
}

// generateSessionID creates a unique session identifier
func generateSessionID(conn net.Conn) string {
	return fmt.Sprintf("session-%s", conn.RemoteAddr().String())
}

// hexDump returns a hex dump of data up to maxBytes, space-delimited
func hexDump(data []byte, maxBytes int) string {
	preview := data
	if len(preview) > maxBytes {
		preview = preview[:maxBytes]
	}

	hexStr := strings.ToUpper(hex.EncodeToString(preview))

	// Add spaces between bytes
	var spaced strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			_ = spaced.WriteByte(' ')
		}
		_, _ = spaced.WriteString(hexStr[i : i+2])
	}

	if len(data) > maxBytes {
		_, _ = spaced.WriteString(" ...")
	}

	return spaced.String()
}
