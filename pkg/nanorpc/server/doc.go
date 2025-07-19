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
// Example usage:
//
//	listener, err := net.Listen("tcp", ":8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	server := NewDefaultServer(listener)
//	if err := server.Serve(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
package server
