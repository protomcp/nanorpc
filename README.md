# NanoRPC

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov-link]

NanoRPC is a lightweight RPC framework designed for embedded systems and
resource-constrained environments.

## Overview

The `nanorpc` package provides efficient binary protocol communication using
Protocol Buffers with [nanopb][nanopb-url] for embedded C targets and Go for
server/client applications. It is optimised for memory-constrained devices
while maintaining flexibility for standard server environments.

## Features

- **Embedded-friendly**: Optimized for resource-constrained environments
- **Binary protocol**: Efficient serialization using Protocol Buffers
- **Multi-language support**: C (via nanopb) and Go implementations
- **Pub/sub messaging**: Subscription-based updates with filtering
- **Hash-based paths**: Reduced memory usage for embedded targets
- **Flexible connectivity**: TCP and UNIX socket support
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

See [NANORPC_PROTOCOL.md](NANORPC_PROTOCOL.md) for the
complete protocol specification.

## Using NanoRPC in Your Project

### With Buf (Recommended)

The nanorpc protocol definitions are available on the Buf Schema Registry:

```yaml
# buf.yaml
version: v2
deps:
  - buf.build/protomcp/nanorpc  # Includes nanopb dependency
```

Then import in your proto files:

```protobuf
import "nanorpc.proto";

service MyService {
  rpc GetData(GetDataRequest) returns (GetDataResponse) {
    option (nanorpc).request_path = "/api/data";
  }
}
```

### Manual Download

Alternatively, download proto files from the repository for use with protoc.

## Components

### Go Client Library

The [`pkg/nanorpc/client`](pkg/nanorpc/client/) package provides a complete Go
client implementation with:

- Connection management and automatic reconnection
- Request/response handling with correlation IDs
- Subscription management with callback support
- Hash-based path optimization for embedded systems
- Thread-safe operation with request queuing
- Comprehensive error handling

### Embedded C Support

Integration with [nanopb][nanopb-url] for embedded C applications with
helpers for nanopb generation (C code) - implementation planned.

### Go Server Library

The [`pkg/nanorpc/server`](pkg/nanorpc/server/) package provides a complete Go
server implementation with decoupled architecture design:

- Clean separation of concerns (Listener, SessionManager, MessageHandler)
- Request/response and ping-pong protocol support
- Extensible message handling with RequestContext
- Graceful shutdown and session management
- Comprehensive test coverage

### Protocol Buffer Generation

The [`pkg/generator`](pkg/generator/) package provides utilities for
generating Protocol Buffer code (implementation planned).

### Shared Types

The [`pkg/nanorpc`](pkg/nanorpc/) package provides shared types and utilities:

- Protocol buffer definitions and generated code
- Hash cache implementation for path hashing
- Request/response encoding and decoding utilities
- Type aliases for protocol buffer internals

## Build (Development)

**Note**: This section is for nanorpc development. To *use* nanorpc in your
project, see "Using NanoRPC in Your Project" above.

You need Go 1.23 or later. The project uses protoc and a comprehensive
Makefile for all development tasks.

### Basic Build

```sh
make         # Equivalent to 'make all'
make all     # Full build cycle: get deps, generate, tidy, build
```

### Testing and Coverage

**IMPORTANT**: Always run `make all coverage` before committing changes.

```sh
make test              # Run tests
make coverage          # Generate coverage reports
make all coverage      # Full build with coverage (required before commits)
```

### Other Commands

```sh
make tidy              # Format code and fix issues
make clean             # Remove build artifacts
make up                # Update dependencies
make lint              # Run linting (included in 'make tidy')
```

## Development

For development guidelines, please refer to [AGENTS.md](AGENTS.md).

## License

See [LICENCE.txt](LICENCE.txt) for licensing information.

## See also

- <https://github.com/nanopb/nanopb>
- <https://github.com/amery/protogen>

[godoc-badge]: https://pkg.go.dev/badge/protomcp.org/nanorpc.svg
[godoc-link]: https://pkg.go.dev/protomcp.org/nanorpc
[goreportcard-badge]: https://goreportcard.com/badge/protomcp.org/nanorpc
[goreportcard-link]: https://goreportcard.com/report/protomcp.org/nanorpc
[codecov-badge]: https://codecov.io/gh/protomcp/nanorpc/graph/badge.svg
[codecov-link]: https://codecov.io/gh/protomcp/nanorpc
[nanopb-url]: https://github.com/nanopb/nanopb
