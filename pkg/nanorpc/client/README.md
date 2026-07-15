# NanoRPC Client Package

The `client` package provides a reconnecting client implementation for the
NanoRPC protocol, designed for reliable communication with embedded systems.

## Overview

The NanoRPC client features:

- Automatic reconnection with configurable retry strategies
- Support for both string paths and path hashes (32-bit FNV-1a)
- Request/response pattern with callbacks
- Subscription support for real-time updates
- Ping/pong for connection health monitoring
- Thread-safe operation with request queuing

## Basic Usage

```go
import "protomcp.org/nanorpc/pkg/nanorpc/client"

// Create a simple client
c, err := client.NewClient(context.Background(), "localhost:8080")
if err != nil {
    log.Fatal(err)
}
defer c.Close()

// Start the client
if err := c.Start(); err != nil {
    log.Fatal(err)
}

// Make a request
err = c.Request("/api/status", nil, func(ctx context.Context, reqID int32,
    resp *nanorpc.NanoRPCResponse) error {
    // Handle response
    if err := nanorpc.ResponseAsError(resp); err != nil {
        return err
    }

    var status StatusMessage
    if err := nanorpc.DecodeResponseData(resp, &status); err != nil {
        return err
    }

    fmt.Printf("Status: %v\n", status)
    return nil
})
```

## Connection Types

The client automatically detects connection type from the remote address:

```go
// Network addresses
client.NewClient(ctx, "localhost:8080")
client.NewClient(ctx, "[::1]:9000")

// File system paths
client.NewClient(ctx, "/var/run/app.sock")
client.NewClient(ctx, "@service-name")  // Abstract socket
```

## Advanced Configuration

```go
cfg := &client.Config{
    Remote:          "device.local:8080",
    Context:         ctx,
    Logger:          myLogger,

    // Timeouts
    DialTimeout:     2 * time.Second,
    ReadTimeout:     2 * time.Second,
    WriteTimeout:    2 * time.Second,
    IdleTimeout:     10 * time.Second,
    KeepAlive:       5 * time.Second,

    // Reconnection
    ReconnectDelay:  5 * time.Second,
    WaitReconnect:   reconnect.NewExponentialWaiter(time.Second,
                         30*time.Second),

    // Path handling
    AlwaysHashPaths: true,  // Use path hashes instead of strings
    HashCache:       myHashCache,

    // Callbacks
    OnConnect: func(ctx context.Context, wg reconnect.WorkGroup) error {
        log.Info("Connected")
        return nil
    },
    OnDisconnect: func(ctx context.Context) error {
        log.Info("Disconnected")
        return nil
    },
    OnError: func(ctx context.Context, err error) error {
        log.Error("Connection error:", err)
        return err
    },
}

client, err := cfg.New()
```

## Path Hashing

The client supports both string paths and path hashes. Path hashing is useful
for embedded systems with limited bandwidth:

```go
// Use string paths (default)
cfg := &client.Config{
    Remote: "device:8080",
    AlwaysHashPaths: false,
}

// Use path hashes (more efficient)
cfg := &client.Config{
    Remote: "device:8080",
    AlwaysHashPaths: true,
    HashCache: nanorpc.NewHashCache(), // Optional: custom cache
}
```

## Subscriptions

Subscribe to paths for real-time updates:

```go
err := c.Subscribe("/events/temperature", nil, func(ctx context.Context,
    reqID int32, resp *nanorpc.NanoRPCResponse) error {
    if resp == nil {
        // Subscription ended
        return nil
    }

    var temp TemperatureEvent
    if err := nanorpc.DecodeResponseData(resp, &temp); err != nil {
        return err
    }

    fmt.Printf("Temperature: %.1f°C\n", temp.Value)
    return nil
})
```

## Connection Management

The client automatically manages connections and reconnections:

```go
// Wait for connection
<-c.Connected()

// Or wait with a context — returns nil on connect, ctx.Err() on timeout
if err := c.WaitConnected(ctx); err != nil {
    return err
}

// Check connection status
if c.IsConnected() {
    // Make requests
}

// Graceful shutdown
c.Close()
c.Wait() // Wait for all goroutines to finish
```

## Testing

The package includes test utilities for writing unit tests:

```go
import (
    "protomcp.org/nanorpc/pkg/nanorpc/client"
    "protomcp.org/nanorpc/pkg/nanorpc/utils/testutils"
)

func TestMyClient(t *testing.T) {
    cfg := &client.Config{
        Remote: "test:8080",
    }

    c, err := cfg.New()
    testutils.AssertNoError(t, err, "Failed to create client")
    testutils.AssertNotNil(t, c, "Client should not be nil")

    // Test with concurrent operations
    helper := client.NewConcurrentTestHelper(t, 10)
    helper.Run(func(idx int) (any, error) {
        return c.Ping()
    })
    helper.AssertNoErrors()
}
```

## Error Handling

Invalid-argument failures are reported through exported sentinels, each
wrapping `darvaza.org/core.ErrInvalid`. Match a specific cause directly, or
the whole family with `IsInvalid`:

```go
_, err := client.Subscribe(c, "/events", nil, cb, newOut)
switch {
case errors.Is(err, client.ErrMissingCallback):
    // a callback-bearing request type was sent without one
case client.IsInvalid(err):
    // any invalid-argument failure (also matches os.ErrInvalid)
}
```

The sentinels are:

- `ErrNilRequest` - `Send` was called with a nil request.
- `ErrMissingCallback` - a request type that requires a callback got none.
- `ErrInvalidRequestType` - an unsupported request type.
- `ErrNoSubscription` - an unsubscribe targeted an unregistered request ID.
- `ErrSubscriptionPending` - an unsubscribe targeted a subscription not yet
  acknowledged.
- `ErrNoSession` / `ErrSessionAttached` - session attach guards.
- `ErrMissingClient` / `ErrMissingOut` - nil arguments to `GetResponse`.
- `ErrMissingNewOut` / `ErrNilOut` - missing or nil-returning `newOut`
  factory.

Call sites add dynamic context by wrapping a sentinel, e.g.
`core.QuietWrap(client.ErrNoSubscription, "request_id %d", id)`; both the
sentinel and `core.ErrInvalid` still match through the wrap.

## Architecture

The client package is organized as follows:

- `client.go` - Main client type and public API
- `session.go` - Session management for active connections
- `config.go` - Configuration structure and defaults
- `reconnect.go` - Reconnection logic and connection lifecycle
- `request.go` - Request handling methods
- `errors.go` - Invalid-argument sentinels and the `IsInvalid` predicate
- `logger.go` - Structured logging support
- `request_counter.go` - Thread-safe request ID generation

## Thread Safety

All public methods of the Client type are thread-safe. The client uses internal
locking to ensure safe concurrent access to shared state. Request callbacks are
executed in separate goroutines and should not block.
