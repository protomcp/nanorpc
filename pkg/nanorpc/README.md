# nanorpc

[![pkg.go.dev][godoc-badge]][godoc-url]
[![Go Report Card][goreportcard-badge]][goreportcard-url]
[![codecov][codecov-badge]][codecov-url]

Go client library for the NanoRPC protocol - a lightweight RPC framework
designed for embedded systems and resource-constrained environments.

This package provides the client-side implementation. For server implementation,
see the companion `server` package.

## Features

- **Connection Management**: Automatic reconnection with configurable backoff
- **Request/Response**: Synchronous RPC calls with correlation IDs
- **Pub/Sub**: Event-driven messaging with subscription callbacks
- **Hash Optimization**: Reduced memory usage with path hashing
- **Protocol Support**: Binary protocol using Protocol Buffers
- **Error Handling**: Structured error responses and connection recovery

## Installation

```bash
go get protomcp.org/nanorpc/pkg/nanorpc
```

## Quick Start

### Basic Client Usage

```go
package main

import (
    "context"
    "log"

    "darvaza.org/slog"

    "protomcp.org/nanorpc/pkg/nanorpc"
)

func main() {
    // Create client configuration
    config := &nanorpc.ClientConfig{
        Remote: "localhost:8080",
        Logger: slog.Default(),
    }

    // Create and connect client
    client, err := config.New()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Make a request
    ctx := context.Background()
    ch := make(chan *nanorpc.NanoRPCResponse, 1)
    callback := func(ctx context.Context, id int32,
        res *nanorpc.NanoRPCResponse) error {
        ch <- res
        return nil
    }

    _, err = client.Request("/status", nil, callback)
    if err != nil {
        log.Fatal(err)
    }

    // Wait for response
    response := <-ch
    log.Printf("Response: %+v", response)
}
```

### Subscription Example

```go
// Subscribe to events
callback := func(ctx context.Context, id int32,
    update *nanorpc.NanoRPCResponse) error {
    log.Printf("Received update: %+v", update)
    return nil
}

requestID, err := client.Subscribe("/events", nil, callback)
if err != nil {
    log.Fatal(err)
}

log.Printf("Subscribed with request ID: %d", requestID)
```

## Protocol

The nanorpc protocol supports three communication patterns:

For the complete protocol specification, see
[NANORPC_PROTOCOL.md](../../NANORPC_PROTOCOL.md).

### Ping-Pong

Health checks and connection validation:

```go
// Simple ping (fire-and-forget)
if !client.Ping() {
    log.Printf("Client not connected")
}

// Ping with response (waits for pong)
ch := client.Pong()
select {
case err := <-ch:
    if err != nil {
        log.Printf("Ping failed: %v", err)
    } else {
        log.Printf("Ping successful")
    }
case <-time.After(5 * time.Second):
    log.Printf("Ping timeout")
}
```

### Request-Response

Synchronous RPC calls:

```go
callback := func(ctx context.Context, id int32,
    res *nanorpc.NanoRPCResponse) error {
    // Handle response
    return nil
}
requestID, err := client.Request("/api/data", requestData, callback)
```

### Pub/Sub

Event-driven messaging:

```go
callback := func(ctx context.Context, id int32,
    update *nanorpc.NanoRPCResponse) error {
    // Handle update
    return nil
}
requestID, err := client.Subscribe("/events", filter, callback)
```

## Configuration

### Client Options

```go
config := &nanorpc.ClientConfig{
    Remote:          "localhost:8080", // Server address
    Logger:          slog.Default(),   // Logger instance

    // Connection settings
    DialTimeout:     2 * time.Second,  // Default: 2s
    ReadTimeout:     2 * time.Second,  // Default: 2s
    WriteTimeout:    2 * time.Second,  // Default: 2s
    IdleTimeout:     10 * time.Second, // Default: 10s
    KeepAlive:       5 * time.Second,  // Default: 5s

    // Reconnection settings
    ReconnectDelay:  5 * time.Second,  // Default: 5s

    // Hash optimization
    AlwaysHashPaths: false,            // Use path hashing
    QueueSize:       0,                // Request queue size
}
```

### Hash Optimization

For embedded targets with limited memory, paths can be hashed:

```go
// Pre-register path to compute hash
nanorpc.RegisterPath("/long/api/path")

// Client can use string path (hashed automatically if AlwaysHashPaths=true)
requestID, err := client.Request("/long/api/path", data, callback)

// Or use pre-computed hash directly
requestID, err := client.RequestWithHash("/long/api/path", data, callback)

// Or use hash value directly
hash := nanorpc.HashCache{}.Hash("/long/api/path")
requestID, err := client.RequestByHash(hash, data, callback)
```

## Error Handling

The library provides structured error handling:

```go
callback := func(ctx context.Context, id int32,
    response *nanorpc.NanoRPCResponse) error {
    // Check for protocol errors first
    if err := nanorpc.ResponseAsError(response); err != nil {
        if nanorpc.IsNotFound(err) {
            log.Printf("Endpoint not found")
        } else if nanorpc.IsNotAuthorized(err) {
            log.Printf("Not authorized: %v", err)
        } else if nanorpc.IsNoResponse(err) {
            log.Printf("No response received: %v", err)
        } else {
            log.Printf("Server error: %v", err)
        }
        return err
    }

    // Response is OK, process data
    log.Printf("Success: %+v", response)
    return nil
}

_, err := client.Request("/api/endpoint", data, callback)
if err != nil {
    log.Printf("Request failed: %v", err)
}
```

## Thread Safety

The client is thread-safe and supports concurrent usage:

```go
var wg sync.WaitGroup

// Multiple goroutines can safely use the same client
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()

        callback := func(ctx context.Context, reqID int32,
            res *nanorpc.NanoRPCResponse) error {
            log.Printf("Goroutine %d got response for request %d", id, reqID)
            return nil
        }

        _, err := client.Request("/parallel", data, callback)
        if err != nil {
            log.Printf("Goroutine %d request failed: %v", id, err)
        }
    }(i)
}

wg.Wait()
```

[godoc-badge]: https://pkg.go.dev/badge/protomcp.org/nanorpc/pkg/nanorpc.svg
[godoc-url]: https://pkg.go.dev/protomcp.org/nanorpc/pkg/nanorpc
[goreportcard-badge]: https://goreportcard.com/badge/protomcp.org/nanorpc/pkg/nanorpc
[goreportcard-url]: https://goreportcard.com/report/protomcp.org/nanorpc/pkg/nanorpc
[codecov-badge]: https://codecov.io/gh/protomcp/nanorpc/branch/main/graph/badge.svg?flag=nanorpc
[codecov-url]: https://codecov.io/gh/protomcp/nanorpc?flag=nanorpc
