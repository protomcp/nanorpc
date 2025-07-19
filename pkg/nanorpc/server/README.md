# nanorpc/server

[![pkg.go.dev][godoc-badge]][godoc-url]
[![Go Report Card][goreportcard-badge]][goreportcard-url]
[![codecov][codecov-badge]][codecov-url]

Go server library for the NanoRPC protocol - a lightweight RPC framework
designed for embedded systems and resource-constrained environments.

This package provides the server-side implementation with a clean, decoupled
architecture that separates concerns into distinct, testable components.

## Architecture

The server implementation uses a modular design with clear separation of
responsibilities:

### Core Interfaces

- **`Listener`** - Handles connection acceptance
- **`SessionManager`** - Manages connection lifecycle and tracking
- **`Session`** - Represents individual client connections
- **`MessageHandler`** - Processes protocol messages

### Components

- **`ListenerAdapter`** - Wraps `net.Listener` for our interface
- **`DefaultSessionManager`** - Multi-session coordination
- **`DefaultSession`** - Individual connection management
- **`DefaultMessageHandler`** - Protocol parsing and request routing
- **`Server`** - Main orchestrator using dependency injection

## Features

- **Decoupled Architecture**: Clean separation of concerns for better
  testability
- **Ping-Pong Protocol**: Built-in health check and connection validation
- **Graceful Shutdown**: Proper session clean-up and resource management
- **Session Management**: Automatic session lifecycle tracking
- **Extensible Handlers**: Easy to add new message types via `MessageHandler`
- **Thread Safety**: Safe for concurrent use across multiple goroutines
- **Comprehensive Testing**: 82.3% test coverage with unit and integration tests

## Installation

```bash
go get github.com/amery/nanorpc/pkg/nanorpc/server
```

## Quick Start

### Basic Server

```go
package main

import (
    "context"
    "log"
    "net"

    "github.com/amery/nanorpc/pkg/nanorpc/server"
)

func main() {
    // Create a TCP listener
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()

    // Create server with default components
    srv := server.NewDefaultServer(listener)

    // Start serving
    ctx := context.Background()
    log.Println("Server starting on :8080")
    if err := srv.Serve(ctx); err != nil {
        log.Printf("Server stopped: %v", err)
    }
}
```

### Custom Server with Dependency Injection

```go
package main

import (
    "context"
    "log"
    "net"

    "github.com/amery/nanorpc/pkg/nanorpc/server"
)

func main() {
    // Create components
    netListener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer netListener.Close()

    listener := server.NewListenerAdapter(netListener)
    handler := server.NewDefaultMessageHandler()
    sessionManager := server.NewDefaultSessionManager(handler)

    // Create server with custom components
    srv := server.NewServer(listener, sessionManager, handler)

    // Start serving
    ctx := context.Background()
    if err := srv.Serve(ctx); err != nil {
        log.Printf("Server error: %v", err)
    }
}
```

### Graceful Shutdown

```go
package main

import (
    "context"
    "log"
    "net"
    "time"

    "github.com/amery/nanorpc/pkg/nanorpc/server"
)

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }

    srv := server.NewDefaultServer(listener)

    // Start server in background
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        if err := srv.Serve(ctx); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    // ... signal handling code ...

    // Graceful shutdown
    log.Println("Shutting down server...")
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(),
        5*time.Second)
    defer shutdownCancel()

    cancel() // Cancel serve context
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    } else {
        log.Println("Server stopped gracefully")
    }
}
```

## Protocol Support

Currently supports the ping-pong protocol pattern:

### Ping-Pong Flow

```text
Client: TYPE_PING (request_id=42)
Server: TYPE_PONG (request_id=42, status=STATUS_OK)
```

The server automatically responds to `TYPE_PING` requests with `TYPE_PONG`
responses, echoing back the client's request ID.

## Extending the Server

### Custom Message Handler

```go
type CustomHandler struct {
    // Your custom fields
}

func (h *CustomHandler) HandleMessage(ctx context.Context,
    session server.Session, req *nanorpc.NanoRPCRequest) error {
    switch req.RequestType {
    case nanorpc.NanoRPCRequest_TYPE_PING:
        return h.handlePing(session, req)
    case nanorpc.NanoRPCRequest_TYPE_REQUEST:
        return h.handleRequest(session, req)
    case nanorpc.NanoRPCRequest_TYPE_SUBSCRIBE:
        return h.handleSubscribe(session, req)
    default:
        return nil // Ignore unknown types
    }
}

func (h *CustomHandler) handleRequest(session server.Session,
    req *nanorpc.NanoRPCRequest) error {
    // Your custom request handling logic
    response := &nanorpc.NanoRPCResponse{
        RequestId:      req.RequestId,
        ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
        ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
        // Add your response data
    }

    responseData, err := nanorpc.EncodeResponse(response, yourData)
    if err != nil {
        return err
    }

    // Send response using interface assertion
    if sessionWriter, ok := session.(io.Writer); ok {
        _, err = sessionWriter.Write(responseData)
        return err
    }
    return nil
}
```

### Custom Session Manager

```go
type CustomSessionManager struct {
    // Your session tracking logic
}

func (sm *CustomSessionManager) AddSession(conn net.Conn) server.Session {
    // Create custom session with your logic
    return &CustomSession{conn: conn}
}

// Implement other SessionManager methods...
```

## Testing

The package includes comprehensive testing utilities:

### Unit Tests

Run individual component tests:

```bash
go test -v github.com/amery/nanorpc/pkg/nanorpc/server
```

### Integration Tests

The test suite includes full client-server integration tests that verify
the complete ping-pong protocol flow over real network connections.

### Test Coverage

Current test coverage: **82.3%**

- Session management: ✅ Complete
- Message handling: ✅ Complete
- Server lifecycle: ✅ Complete
- Integration flows: ✅ Complete

## Performance Characteristics

- **Concurrent Connections**: Supports multiple simultaneous clients
- **Goroutine Per Connection**: Each client gets dedicated goroutine
- **Memory Efficient**: Session IDs generated from connection addresses
- **Graceful Degradation**: Continues operating on individual connection errors

## Thread Safety

All components are designed for concurrent use:

- **Server**: Thread-safe for multiple goroutines
- **SessionManager**: Protected with RWMutex for concurrent access
- **Sessions**: Each session handles one connection in its own goroutine
- **MessageHandler**: Stateless design safe for concurrent calls

## Error Handling

The server handles various error scenarios gracefully:

- **Connection Errors**: Logged but don't affect other sessions
- **Decode Errors**: Logged and skipped, connection continues
- **Handler Errors**: Logged and skipped, connection continues
- **Accept Errors**: Trigger server shutdown
- **Context Cancellation**: Graceful shutdown initiated

## Examples

For complete examples, see the test files:

- `server_test.go` - Integration tests with real network connections
- `session_test.go` - Session management examples
- `handler_test.go` - Message handling patterns

## Future Extensions

The decoupled architecture enables easy extension for:

- **Request-Response Patterns**: Custom request routing and responses
- **Pub/Sub Messaging**: Subscription management and event broadcasting
- **Authentication**: Session-based security layers
- **Rate Limiting**: Per-session or global rate controls
- **Metrics Collection**: Connection and request monitoring

[godoc-badge]: https://pkg.go.dev/badge/github.com/amery/nanorpc/pkg/nanorpc/server.svg
[godoc-url]: https://pkg.go.dev/github.com/amery/nanorpc/pkg/nanorpc/server
[goreportcard-badge]: https://goreportcard.com/badge/github.com/amery/nanorpc/pkg/nanorpc/server
[goreportcard-url]: https://goreportcard.com/report/github.com/amery/nanorpc/pkg/nanorpc/server
[codecov-badge]: https://codecov.io/gh/amery/nanorpc/branch/main/graph/badge.svg
[codecov-url]: https://codecov.io/gh/amery/nanorpc
