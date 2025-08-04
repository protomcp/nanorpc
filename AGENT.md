# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Repository Overview

`nanorpc` is a lightweight RPC framework designed for embedded systems and
resource-constrained environments. It provides efficient binary protocol
communication using Protocol Buffers (protobuf) with nanopb for embedded C
targets and Go for server/client applications.

## Prerequisites

Before starting development, ensure you have:

- Go 1.23 or later installed (check with `go version`).
- `make` command available (usually pre-installed on Unix systems).
- Protocol Buffers compiler (`protoc`) installed.
- `pnpm` for JavaScript/TypeScript tooling (preferred over npm).
- Git configured for proper line endings.

## Common Development Commands

```bash
# Full build cycle (get deps, generate, tidy, build)
make all

# Run tests
make test

# Run tests with coverage
make test GOTEST_FLAGS="-cover"

# Run tests with verbose output and coverage
make test GOTEST_FLAGS="-v -cover"

# Generate coverage reports
make coverage

# Generate Codecov configuration and upload scripts
make codecov

# Format code and tidy dependencies (run before committing)
make tidy

# Clean build artifacts
make clean

# Update dependencies
make up

# Run go:generate directives
make generate
```

## Build System Features

### Multi-Module Support

The project uses a sophisticated build system that handles multiple Go modules:

- **Root module**: `protomcp.org/nanorpc`
- **Submodules**: `pkg/generator`, `pkg/nanopb`, `pkg/nanorpc`
- **Dynamic rules**: Generated via `internal/build/gen_mk.sh`
- **Dependency tracking**: Handles inter-module dependencies

### Tool Integration

The build system includes comprehensive tooling:

#### Linting and Quality

- **golangci-lint**: Go code linting with version selection
- **revive**: Additional Go linting with custom rules
- **markdownlint**: Markdown formatting and style checking
- **shellcheck**: Shell script analysis
- **cspell**: Spell checking for documentation and code
- **languagetool**: Grammar checking for Markdown files

#### Coverage and Testing

- **Coverage collection**: Automated across all modules
- **Codecov integration**: Multi-module coverage reporting
- **Test execution**: Parallel testing with dependency management

#### Development Tools

- **Whitespace fixing**: Automated trailing whitespace removal
- **EOF handling**: Ensures files end with newlines
- **Dynamic tool detection**: Tools auto-detected via pnpx

### Configuration Files

Tool configurations are stored in `internal/build/`:

- `markdownlint.json`: Markdown linting rules (80-char lines)
- `cspell.json`: Spell checking dictionary and rules
- `languagetool.cfg`: Grammar checking configuration
- `revive.toml`: Go linting rules and thresholds

## Project Architecture

### Core Components

- **nanorpc protocol**: Binary RPC protocol with protobuf serialization
- **Client implementation**: Complete client with reconnection and subscriptions
- **Server implementation**: (Planned) Server to complement client
- **Generator**: Protobuf code generation utilities
- **nanopb integration**: Protocol Buffers for embedded C

### Protocol Design

The nanorpc protocol supports:

- **Request types**: `TYPE_PING`, `TYPE_REQUEST`, `TYPE_SUBSCRIBE`
- **Response types**: `TYPE_PONG`, `TYPE_RESPONSE`, `TYPE_UPDATE`
- **Path optimization**: Both string paths and FNV-1a hashes
- **Pub/sub messaging**: Subscription-based updates
- **Error handling**: Structured error responses

See the [NANORPC_PROTOCOL.md](NANORPC_PROTOCOL.md) document for the complete
protocol specification.

### Key Features

- **Embedded-friendly**: Designed for resource-constrained environments
- **Binary protocol**: Efficient serialization using protobuf
- **Reconnection**: Automatic client reconnection handling
- **Hash-based paths**: Reduced memory usage for embedded targets
- **Zero-copy**: Efficient message handling where possible

## Development Workflow

### MANDATORY: Test-Driven Development (TDD)

**ALL DEVELOPMENT MUST FOLLOW TDD**:

1. **Write failing tests first** - Define expected behaviour in tests
2. **Implement minimal code** - Write just enough to pass tests
3. **Refactor for quality** - Improve code while maintaining tests
4. **Repeat cycle** - Continue until feature is complete

**Test Infrastructure**:

- Common test utilities in `pkg/nanorpc/common/testutils/`
- Server-specific mocks in `pkg/nanorpc/server/testutils_test.go`
- Client-specific mocks in `pkg/nanorpc/client/testutils_test.go`

### Before Starting Work

1. **Read the plans**: Check existing plan documents for context.
2. **Understand modules**: Review go.mod files in each package.
3. **Check dependencies**: Understand inter-module relationships.
4. **Review tests**: Examine existing test patterns.
5. **Write tests first**: Always start with failing tests.

### Code Quality Standards

The project enforces quality through:

- **Go standards**: Standard Go conventions and formatting
- **Field alignment**: Structs optimized for memory efficiency

  ```bash
  # Fix field alignment issues (exclude generated files like *.pb.go)
  GOXTOOLS="golang.org/x/tools/go/analysis/passes"
  FA="$GOXTOOLS/fieldalignment/cmd/fieldalignment"
  # Only run on hand-written files, not generated ones
  go -C pkg/nanorpc run "$FA@latest" -fix \
    errors.go hashcache.go path.go utils.go request_counter.go

  # For client files (in separate package)
  go -C pkg/nanorpc/client run "$FA@latest" -fix \
    client.go client_*.go

  # For server files (in separate package)
  go -C pkg/nanorpc/server run "$FA@latest" -fix \
    handler.go subscription.go interfaces.go

  # For test files with complex types, create a temporary file:
  # 1. Copy struct definitions to a temp.go file with simplified types
  # 2. Run fieldalignment on the temp file
  # 3. Apply the suggested field ordering to the test files
  # 4. Remove the temp file
  ```

- **Linting rules**: Comprehensive linting via golangci-lint and revive
- **Test coverage**: Aim for high test coverage across modules
- **Documentation**: All public APIs must be documented

### Protocol Implementation

When working with the nanorpc protocol:

1. **Follow existing patterns**: Use client implementation as reference
2. **Handle errors properly**: Use structured error responses
3. **Support both paths**: String paths and hash-based paths
4. **Test thoroughly**: Protocol changes affect multiple components

## Testing Guidelines

### Test Structure

- **Table-driven tests**: Preferred for comprehensive coverage
- **Module isolation**: Each module should test independently
- **Integration tests**: Cross-module functionality testing
- **Protocol tests**: End-to-end protocol validation

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
make test GOTEST_FLAGS="-race"

# Run specific tests
make test GOTEST_FLAGS="-run TestSpecific"

# Generate coverage
make coverage

# Test specific module
make test-nanorpc
```

## Important Notes

### Build System

- Go 1.23 is the minimum required version
- The Makefile dynamically generates rules for submodules
- Tool versions are selected based on Go version
- All tools are auto-detected with fallback to no-op

### Protocol Considerations

- This is a binary protocol optimized for embedded systems
- Hash-based paths reduce memory usage on embedded targets
- Subscription handling must be efficient for resource constraints
- Error responses should be structured and informative

### Development Environment

- Always use `pnpm` instead of `npm` for JavaScript/TypeScript tooling
- Protocol buffer files are in `protos/` directory
- Generated code should not be manually edited
- Use `make generate` after protocol changes

## Pre-commit Checklist

1. **ALWAYS run `make tidy` first** - Fix ALL issues before committing:
   - Go code formatting and whitespace clean-up
   - Markdown files checked with markdownlint and cspell
   - Shell scripts checked with shellcheck
   - Protocol buffer files regenerated if needed
2. **Verify all tests pass** with `make test`
3. **Check coverage** with `make coverage` if adding new code
4. **Update documentation** if changing public APIs
5. **Run `make generate`** if protocol definitions changed

## Git Usage Guidelines

**CRITICAL**: Always follow these git practices to avoid accidental commits:

1. **NEVER use bulk operations** - Always explicitly specify files:

   ```bash
   # CORRECT - explicitly specify files
   git add file1.go file2.go
   git commit -s file1.go file2.go -m "commit message"

   # WRONG - bulk staging/committing
   git add .
   git add -A
   git add -u
   git commit -s -m "commit message"
   git commit -a -m "commit message"
   ```

2. **Use `-s` when doing commits** - Don't take credit for the work

3. **Check what you're committing**:

   ```bash
   git status --porcelain  # Check current state
   git diff --cached       # Review staged changes before committing
   ```

4. **Atomic commits** - Each commit should contain only related changes for a
   single purpose

## Development Patterns with darvaza.org

### API Verification First

Always verify API usage with `go doc` before implementing:

```bash
# Check package documentation
go -C pkg/nanorpc doc server

# Check specific types and methods
go -C pkg/nanorpc doc server.RequestContext
go -C pkg/nanorpc doc -src HashCache.Hash

# Check darvaza.org dependencies (accessible from nanorpc directory)
go -C pkg/nanorpc doc darvaza.org/core.Catch
go -C pkg/nanorpc doc darvaza.org/slog.Logger
go -C pkg/nanorpc doc darvaza.org/x/sync/workgroup.Group
```

The `-C` flag allows checking documentation from the correct module context:

```bash
# ✅ CORRECT - Check from module directory
go -C pkg/nanorpc doc HashCache

# ❌ WRONG - May give incorrect results
go doc protomcp.org/nanorpc/pkg/nanorpc.HashCache
```

### Error Handling

Use darvaza.org/core for consistent error handling:

```go
// Check nil receivers
if obj == nil {
    return core.ErrNilReceiver
}

// Wrap errors with context
if err != nil {
    return core.Wrapf(err, "failed to process %s", path)
}

// Create formatted errors
return core.Errorf("invalid path hash: 0x%08x", hash)

// Panic recovery in critical sections
err := core.Catch(func() error {
    // Code that might panic
    return riskyOperation()
})
```

### Structured Logging

Use darvaza.org/slog with proper patterns. For standardized field names,
refer to `pkg/nanorpc/common/fields.go` which defines constants for field
names and helper functions for safe field addition:

```go
// Basic logging with levels
logger.Info().Print("server started")
logger.Error().WithField(slog.ErrorFieldName, err).Print("operation failed")

// Check if logging is enabled before expensive operations
if logger, ok := logger.Debug().WithEnabled(); ok {
    logger.WithField("data", expensiveDebugData()).Print("debug info")
}

// Request-scoped logging using standard field names
logger := baseLogger.WithField(common.FieldRequestID, req.RequestId)
logger.Info().Print("processing request")

// Component logging using standard constants
logger := common.WithComponent(baseLogger, common.ComponentServer)

// Enhanced helper methods with structured fields
server.LogInfo(slog.Fields{"port": 8080}, "server started on port %d", 8080)
client.LogError(addr, err, slog.Fields{common.FieldAttempt: 3},
    "connection failed after %d attempts", 3)
session.LogDebug(slog.Fields{common.FieldRequestID: reqID},
    "processing request %s", reqID)
```

#### Logger Helper Methods

Client, Server, Session, and SessionManager types provide enhanced logging
helper methods:

```go
// Client methods
LogDebug(addr net.Addr, fields slog.Fields, msg string, args ...any)
LogInfo(addr net.Addr, fields slog.Fields, msg string, args ...any)
LogWarn(addr net.Addr, err error, fields slog.Fields, msg string, args ...any)
LogError(addr net.Addr, err error, fields slog.Fields, msg string, args ...any)

// Server/Session methods
LogDebug(fields slog.Fields, msg string, args ...any)
LogInfo(fields slog.Fields, msg string, args ...any)
LogWarn(err error, fields slog.Fields, msg string, args ...any)
LogError(err error, fields slog.Fields, msg string, args ...any)

// Usage examples
server.LogInfo(slog.Fields{"port": 8080}, "server started on port %d", 8080)
client.LogError(addr, err, nil, "failed to connect")
session.LogDebug(slog.Fields{"path": req.Path}, "processing request for %s",
    req.Path)
```

### Concurrent Operations

Use darvaza.org/x/sync/workgroup for goroutine management:

```go
// Create workgroup
wg := workgroup.New(ctx)
defer wg.Close()

// Launch basic goroutine
err := wg.Go(func(ctx context.Context) {
    // Task implementation
})

// Launch with error handling
err := wg.GoCatch(
    func(ctx context.Context) error {
        return doWork(ctx)
    },
    func(ctx context.Context, err error) error {
        if err != nil && !errors.Is(err, context.Canceled) {
            logger.Error().WithField(slog.ErrorFieldName, err).
                Print("task failed")
        }
        return nil // Don't propagate to keep other workers running
    },
)

// Wait for completion
if err := wg.Wait(); err != nil {
    logger.Error().WithField(slog.ErrorFieldName, err).Print("workgroup failed")
}
```

### Context Propagation

Always pass context through the call stack:

```go
// Always pass context through the call stack
func (s *Server) Handle(ctx context.Context, req Request) error {
    // Create child context with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Pass to next layer
    return s.processRequest(ctx, req)
}
```

### Testing Patterns

#### Table-Driven Tests

```go
type testCase struct {
    name     string
    input    string
    expected string
    wantErr  bool
}

func (tc testCase) test(t *testing.T) {
    t.Helper()

    result, err := processInput(tc.input)
    if tc.wantErr {
        require.Error(t, err)
        return
    }

    require.NoError(t, err)
    assert.Equal(t, tc.expected, result)
}

func TestProcess(t *testing.T) {
    tests := []testCase{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, tc.test)
    }
}
```

#### Concurrent Testing

```go
// Use existing test utilities
helper := &ConcurrentTestHelper{
    TestFunc: func(id int) error {
        return client.Ping()
    },
    NumGoroutines: 50,
}

results := helper.Run()
for i, err := range results {
    assert.NoError(t, err, "goroutine %d failed", i)
}
```

### Field Alignment

Always optimize struct field alignment:

```go
// ✅ GOOD - Aligned by size
type Session struct {
    // 8-byte fields first
    conn      net.Conn
    logger    slog.Logger
    createdAt time.Time

    // 4-byte fields
    id        int32

    // 1-byte fields
    active    bool

    // Strings last
    name      string
}

// ❌ BAD - Poor alignment
type Session struct {
    active    bool      // 1 byte + 7 padding
    conn      net.Conn  // 8 bytes
    id        int32     // 4 bytes + 4 padding
    logger    slog.Logger // 8 bytes
    name      string    // 16 bytes
    createdAt time.Time // 8 bytes
}
```

Use fieldalignment tool to check:

```bash
FA="golang.org/x/tools/go/analysis/passes/fieldalignment"
go run "$FA/cmd/fieldalignment@latest" -fix ./...
```

### Common Pitfalls to Avoid

#### 1. Incorrect slog Usage

```go
// ❌ WRONG - WithError doesn't exist
logger.WithError(err).Error("failed")

// ✅ CORRECT
logger.Error().WithField(slog.ErrorFieldName, err).Print("failed")
```

#### 2. Missing Nil Checks

```go
// ❌ WRONG - May panic
func (s *Server) Start() {
    s.logger.Info().Print("starting")
}

// ✅ CORRECT
func (s *Server) Start() {
    if s.logger != nil {
        s.logger.Info().Print("starting")
    }
}
```

#### 3. Context Ignorance

```go
// ❌ WRONG - Ignores context cancellation
func process() {
    for {
        doWork()
    }
}

// ✅ CORRECT
func process(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            doWork()
        }
    }
}
```

### Build and Quality Commands

```bash
# Full build cycle
make all

# Run tests with coverage
make test GOTEST_FLAGS="-cover"

# Fix code issues
make tidy

# Generate coverage reports
make coverage

# Update dependencies
make up
```

### Git Workflow Best Practices

```bash
# ✅ CORRECT - Explicit file specification
git add handler.go handler_test.go
git commit -s handler.go handler_test.go -m "feat: add response helpers"

# ❌ WRONG - Bulk operations
git add .
git commit -a -m "updates"
```

Always sign commits with `-s` flag and be explicit about files.

## Troubleshooting

### Common Issues

1. **Protocol buffer compilation**:
   - Ensure `protoc` is installed and in PATH
   - Check that nanopb submodule is properly included
   - Verify import paths in proto files

2. **Module dependencies**:
   - Run `make tidy` to fix go.mod issues
   - Check that replace directives are correct
   - Verify inter-module dependencies

3. **Tool detection failures**:
   - Install tools globally with `pnpm install -g <tool>`
   - Check that pnpx is available and functional
   - Tools fall back to no-op if not found

4. **Coverage issues**:
   - Ensure all modules have test files
   - Check that `.tmp/index` exists
   - Use `GOTEST_FLAGS` for additional test configuration

### Getting Help

- Check existing issues and documentation
- Review test files for usage examples
- Examine client implementation for protocol patterns
- Refer to Protocol Buffer documentation for schema changes
- Use `go doc` to verify API usage before implementation

This project focuses on providing efficient, reliable RPC communication for
embedded systems while maintaining clean, well-tested Go code.
