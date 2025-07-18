# NanoRPC

[![codecov][codecov-badge]][codecov-url]

NanoRPC is a lightweight RPC framework designed for embedded systems and
resource-constrained environments. It provides efficient binary protocol
communication using Protocol Buffers with [nanopb][nanopb-url] for embedded
C targets and Go for server/client applications.

## Features

- **Embedded-friendly**: Optimized for resource-constrained environments
- **Binary protocol**: Efficient serialization using Protocol Buffers
- **Multi-language support**: C (via nanopb) and Go implementations
- **Pub/sub messaging**: Subscription-based updates with filtering
- **Hash-based paths**: Reduced memory usage for embedded targets
- **Reconnection handling**: Automatic client reconnection logic
- **Zero-copy**: Efficient message handling where possible

## Protocol Support

The nanorpc protocol supports multiple communication patterns:

- **Ping-Pong**: Health checks and connection validation
- **Request-Response**: Synchronous RPC calls with structured responses
- **Pub/Sub**: Event-driven messaging with subscription filtering

### Message Types

- Request types: `TYPE_PING`, `TYPE_REQUEST`, `TYPE_SUBSCRIBE`
- Response types: `TYPE_PONG`, `TYPE_RESPONSE`, `TYPE_UPDATE`
- Status codes: `STATUS_OK`, `STATUS_NOT_FOUND`, `STATUS_INTERNAL_ERROR`

## Components

### Go Client Library

The [`pkg/nanorpc`](pkg/nanorpc/) module provides a complete Go client
implementation with:

- Connection management and automatic reconnection
- Request/response handling with correlation IDs
- Subscription management with callback support
- Hash-based path optimization
- Comprehensive error handling

### Embedded C Support

Integration with [nanopb][nanopb-url] for embedded C applications
(implementation planned).

### Server Implementation

A complete Go server implementation with decoupled architecture design:

- Clean separation of concerns (Listener, SessionManager, MessageHandler)
- Ping-pong protocol support with extensible message handling
- Graceful shutdown and session management
- Comprehensive test coverage

## Development

### Lint

```sh
go install github.com/golangci/golangci-lint/cmd/...@latest
golangci-lint run
```

or

```sh
make get
make lint
```

## Build

```sh
go get -v ./...
go generate -v ./...
go build -v ./...
```

or

```sh
make get
make build
```

## See also

- <https://github.com/nanopb/nanopb>
- <https://github.com/amery/protogen>

[codecov-badge]: https://codecov.io/gh/amery/nanorpc/branch/main/graph/badge.svg
[codecov-url]: https://codecov.io/gh/amery/nanorpc
[nanopb-url]: https://github.com/nanopb/nanopb
