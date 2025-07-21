// Package server implements the NanoRPC server for handling lightweight RPC
// requests in embedded systems and resource-constrained environments.
//
// The server provides a modular architecture with clear separation of concerns:
//
//   - Listener: Handles connection acceptance
//   - SessionManager: Manages client session lifecycle
//   - MessageHandler: Processes protocol messages
//   - RequestHandler: Handles path-specific requests
//
// The server supports the NanoRPC protocol with:
//   - TYPE_PING/TYPE_PONG for connection health checking
//   - TYPE_REQUEST/TYPE_RESPONSE for request-response patterns
//   - TYPE_SUBSCRIBE/TYPE_UPDATE for pub/sub messaging (planned)
//
// # Logging
//
// The server supports structured logging via darvaza.org/slog. Pass a
// configured logger to NewServer or NewDefaultServer. If no logger is
// provided, a discard logger is lazily initialized on first use to ensure
// safe operation.
//
//	logger := myLogger.WithField("service", "nanorpc")
//	server := NewDefaultServer(listener, logger)
//
// The server uses consistent field names exported as constants to allow
// callers to filter or transform logs. See the Field* constants for
// available fields.
//
// Example usage:
//
//	listener, err := net.Listen("tcp", ":8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	server := NewDefaultServer(listener, nil)
//	if err := server.Serve(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
package server
