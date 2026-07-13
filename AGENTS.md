# AGENTS.md

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

- Go 1.25 or later installed (check with `go version`).
- `make` command available (usually pre-installed on Unix systems).
- Protocol Buffers compiler (`protoc`) installed for code generation.
- **Buf CLI** (`buf`) for publishing to Buf Schema Registry (releases only).
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

# Generate coverage reports (self and integration coverage)
make coverage

# Generate merged coverage and upload scripts
make codecov

# Create merged repository-wide profiles
make merged-coverage

# Format code and tidy dependencies (run before committing)
make tidy

# Clean build artefacts
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

- **Dual coverage system**: Self-coverage and integration coverage perspectives
- **Coverage collection**: Automated across modules with hierarchical profiles
- **Codecov integration**: Intelligent upload with best coverage data
- **Test execution**: Parallel testing with dependency management
- **Merged profiles**: Repository-wide coverage aggregation

#### Development Tools

- **Whitespace fixing**: Automated trailing whitespace removal
- **EOF handling**: Ensures files end with newlines
- **Dynamic tool detection**: Tools auto-detected via `pnpm dlx`

### Configuration Files

Tool configurations are stored in `internal/build/`:

- `markdownlint.json`: Markdown linting rules (80-char lines)
- `cspell.json`: Spell checking dictionary and rules
- `languagetool.cfg`: Grammar checking configuration
- `revive.toml`: Go linting rules and thresholds

## Project Architecture

### Core Components

- **nanorpc protocol**: Binary RPC protocol with protobuf serialisation
- **Client implementation**: Complete client with reconnection and subscriptions
- **Server implementation**: Listener, session management, and
  subscription dispatch complementing the client
- **Generator**: Protobuf code generation utilities
- **nanopb integration**: Protocol Buffers for embedded C

### Protocol Design

The nanorpc protocol supports:

- **Request types**: `TYPE_PING`, `TYPE_REQUEST`, `TYPE_SUBSCRIBE`
- **Response types**: `TYPE_PONG`, `TYPE_RESPONSE`, `TYPE_UPDATE`
- **Path optimisation**: Both string paths and FNV-1a hashes
- **Pub/sub messaging**: Subscription-based updates
- **Error handling**: Structured error responses

See the [NANORPC_PROTOCOL.md](NANORPC_PROTOCOL.md) document for the complete
protocol specification.

### Key Features

- **Embedded-friendly**: Designed for resource-constrained environments
- **Binary protocol**: Efficient serialisation using protobuf
- **Reconnection**: Automatic client reconnection handling
- **Hash-based paths**: Reduced memory usage for embedded targets
- **Zero-copy**: Efficient message handling where possible

## Development Workflow

### Test-Driven Development (TDD)

The project follows a test-first workflow:

1. **Write failing tests first** - Define expected behaviour in tests
2. **Implement minimal code** - Write just enough to pass tests
3. **Refactor for quality** - Improve code while maintaining tests
4. **Repeat cycle** - Continue until feature is complete

**Test Infrastructure**:

- Common test utilities in `pkg/nanorpc/utils/testutils/`
- Reusable mock test doubles in `pkg/nanorpc/mock/` (`client`,
  `server`, `wire`)
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
- **Field alignment**: Structs optimised for memory efficiency — see the
  [Field Alignment](#field-alignment) section for the probe-file workflow.
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

# Run tests with race detection (dedicated target, CGO_ENABLED=1)
make race

# Run specific tests
make test GOTEST_FLAGS="-run TestSpecific"

# Generate coverage (self and integration coverage)
make coverage

# Generate specific module coverage
make coverage-nanorpc

# View merged coverage profiles
make merged-coverage

# Test specific module
make test-nanorpc

# View individual module coverage reports
cat .tmp/coverage/coverage_<module>.func             # Integration coverage
cat .tmp/coverage/coverage_<module>_self.func        # Self-coverage
grep -E 'class="(cov|miss)[0-9]*"' \
    .tmp/coverage/coverage_<module>.html             # Line-level analysis
```

## Important Notes

### Build System

- Go 1.25 is the minimum required version
- The Makefile dynamically generates rules for submodules
- Tool versions are selected based on Go version
- All tools are auto-detected with fallback to no-op

### Protocol Considerations

- This is a binary protocol optimised for embedded systems
- Hash-based paths reduce memory usage on embedded targets
- Subscription handling must be efficient for resource constraints
- Error responses should be structured and informative

### Development Environment

- Always use `pnpm` instead of `npm` for JavaScript/TypeScript tooling
- Protocol buffer files are in `proto/` directory (not `protos/`)
  - Subdirectories: `proto/nanopb/`, `proto/nanorpc/`, `proto/vendor/`

### Protocol Buffer Code Generation

- Protocol buffer files in `proto/` directory generate Go code via `go:generate`
- Generated `.pb.go` files are committed and included in Go modules
- Uses `protoc` via `internal/build/proto.sh` scripts
- **Do not edit `.pb.go` files manually** - they are generated
- Run `go generate ./...` or `make generate` after proto changes

## Pre-commit Checklist

1. **Run `make tidy` first** - fix all issues before committing:
   - Go code formatting and whitespace clean-up
   - Markdown files checked with markdownlint and cspell
   - Shell scripts checked with shellcheck
   - Protocol buffer files regenerated if needed
2. **Run `make all coverage`** - required before every commit
3. **Review staged changes**: `git diff --cached` - Know exactly what you're
   committing
4. **Verify commit message** - Ensure it describes ONLY the actual changes
5. **Check for sensitive data** - No secrets, keys, or credentials
6. **Ensure atomic commits** - Each commit should be self-contained

## Git Usage Guidelines

Follow these git practices to avoid accidental commits:

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

## Commit Message Requirements

Follow these requirements for commit messages:

1. **READ THE ACTUAL DIFF**: Always use `git diff` or `git diff --cached` to
   see exact changes before creating commit messages. Do not rely on memory or
   assumptions.

2. **FOCUS ON THE PATCH**: Describe what the code change actually does, not
   what you think it should do or what tasks were discussed.

3. **NO AI ATTRIBUTION**: Never include "Generated with Claude Code",
   "Co-Authored-By: Claude", or similar attribution lines.

4. **AVOID RECENCY BIAS**: Do not let recent conversation topics influence
   commit message content. Focus solely on the actual code changes.

5. **USE CONVENTIONAL COMMITS**: Follow the format: `type(scope): description`
   - Types: feat, fix, docs, style, refactor, test, chore, build, ci
   - Scope: Component or area affected
   - Description: Present tense, lowercase, no period

Example workflow:

```bash
# 1. Stage specific files
git add file1.go file2.go

# 2. ALWAYS review the diff
git diff --cached

# 3. Write commit message based ONLY on the diff
git commit -s -m "fix(client): handle nil response in ping handler"
```

## Development Patterns with darvaza.org

### API Verification First

Always verify API usage with `go doc` before implementing. Prefer
fully-qualified import paths — they resolve unambiguously from any
directory:

```bash
# Types and symbols, by full import path
go doc protomcp.org/nanorpc/pkg/nanorpc.HashCache
go doc protomcp.org/nanorpc/pkg/nanorpc/server.RequestContext
go doc -src protomcp.org/nanorpc/pkg/nanorpc.HashCache.Hash

# darvaza.org dependencies, likewise by full path
go doc darvaza.org/core.Catch
go doc darvaza.org/slog.Logger
go doc darvaza.org/x/sync/workgroup.Group
```

The `-C dir` flag does not scope `go doc`: for a package or symbol
argument it resolves the same as without `-C` (a workspace-wide
search), so `-C` cannot pin the lookup to one module. To use short
names, change into the module directory instead — a subshell keeps
your shell put:

```bash
(cd pkg/nanorpc && go doc HashCache)   # symbol in that package
(cd pkg/nanorpc && go doc ./server)    # relative subpackage path
```

Avoid a bare, unqualified package name: in the `go.work` workspace it
is resolved by a workspace-wide search that can land on a different
module's package of the same name — a bare `utils` may match another
module's `utils` instead of
`protomcp.org/nanorpc/pkg/nanorpc/utils`, regardless of the current
directory. Qualify with the full import path, or use a relative
`./utils` path from inside the module directory.

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

// QuietWrap matches a sentinel for errors.Is without prefixing its
// text ("invalid argument") onto the message.
return core.QuietWrap(core.ErrInvalid, "invalid path hash: 0x%08x", hash)

// Panic recovery in critical sections
err := core.Catch(func() error {
    // Code that might panic
    return riskyOperation()
})
```

### Structured Logging

Use darvaza.org/slog with proper patterns. For standardised field names,
refer to `pkg/nanorpc/utils/fields.go` which defines constants for field
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
logger := baseLogger.WithField(utils.FieldRequestID, req.RequestId)
logger.Info().Print("processing request")

// Component logging using standard constants
logger := utils.WithComponent(baseLogger, utils.ComponentServer)

// Enhanced helper methods with structured fields
server.LogInfo(slog.Fields{"port": 8080}, "server started on port %d", 8080)
client.LogError(addr, err, slog.Fields{utils.FieldAttempt: 3},
    "connection failed after %d attempts", 3)
session.LogDebug(slog.Fields{utils.FieldRequestID: reqID},
    "processing request %s", reqID)
```

#### Logger Helper Methods

Client, Server, `DefaultSession`, and `DefaultSessionManager` types provide
enhanced logging helper methods (the `Session` and `SessionManager` interfaces
do not expose them — call these on the concrete implementations):

```go
// Client methods
LogDebug(addr net.Addr, fields slog.Fields, msg string, args ...any)
LogInfo(addr net.Addr, fields slog.Fields, msg string, args ...any)
LogWarn(addr net.Addr, err error, fields slog.Fields, msg string, args ...any)
LogError(addr net.Addr, err error, fields slog.Fields, msg string, args ...any)

// Server, DefaultSession, and DefaultSessionManager methods
LogDebug(fields slog.Fields, msg string, args ...any)
LogInfo(fields slog.Fields, msg string, args ...any)
LogWarn(err error, fields slog.Fields, msg string, args ...any)
LogError(err error, fields slog.Fields, msg string, args ...any)
```

### Concurrent Operations

Use darvaza.org/x/sync/workgroup for goroutine management:

```go
// Create workgroup
wg := workgroup.New(ctx)
defer wg.Close()

// Launch basic goroutine
wg.Go(func(ctx context.Context) {
    // Task implementation
})

// Launch with error handling
wg.GoCatch(
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
// Always pass context through the call stack. A RequestHandlerFunc
// receives the caller's context; derive child contexts from it
// rather than reaching for context.Background().
func handleRequest(ctx context.Context, rc *RequestContext) error {
    // Create child context with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Pass to next layer
    return processRequest(ctx, rc)
}
```

### Testing Patterns

#### Table-Driven Tests

```go
// Compile-time check that the type implements core.TestCase.
var _ core.TestCase = processTestCase{}

type processTestCase struct {
    name     string
    input    string
    expected string
    wantErr  bool
}

func (tc processTestCase) Name() string {
    return tc.name
}

func (tc processTestCase) Test(t *testing.T) {
    t.Helper()

    result, err := processInput(tc.input)
    if tc.wantErr {
        core.AssertError(t, err, "error")
        return
    }

    core.AssertNoError(t, err, "process")
    core.AssertEqual(t, tc.expected, result, "result")
}

// newProcessTestCase keeps a logical parameter order, decoupled
// from the struct's field alignment.
func newProcessTestCase(name, input, expected string,
    wantErr bool) processTestCase {
    return processTestCase{
        name:     name,
        input:    input,
        expected: expected,
        wantErr:  wantErr,
    }
}

func TestProcess(t *testing.T) {
    cases := []processTestCase{
        newProcessTestCase("valid input", "test", "TEST", false),
        newProcessTestCase("empty input", "", "", true),
    }
    core.RunTestCases(t, cases)
}
```

#### Concurrent Testing

```go
// Use the shared testutils.ConcurrentTestHelper.
// Ping reports connectivity as a bool; the TestFunc contract wants an
// error, so surface the client's own not-connected sentinel
// (darvaza.org/x/net/reconnect.ErrNotConnected).
helper := &testutils.ConcurrentTestHelper{
    TestFunc: func(id int) error {
        if !client.Ping() {
            return reconnect.ErrNotConnected
        }
        return nil
    },
    NumGoroutines: 50,
}

errs := helper.Run()
for i, err := range errs {
    core.AssertNoError(t, err, "goroutine %d", i)
}
```

### Field Alignment

Always optimise struct field alignment:

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

`fieldalignment -fix` rewrites files in place and strips ALL comments from
every file it touches — never run it against the source tree or `./...`.
Use an isolated probe file instead:

```bash
# 1. Copy the structs to optimise into .tmp/probe.go as
#    `package probe` (comments are expendable in the probe).
# 2. Run the tool against just that file:
GOXTOOLS="golang.org/x/tools/go/analysis/passes"
FA="$GOXTOOLS/fieldalignment/cmd/fieldalignment"
go run "$FA@latest" -fix .tmp/probe.go
# 3. Diff the rewritten probe to read off the suggested order.
# 4. Apply that order to the real source by hand, preserving
#    all comments and doc strings.
# 5. Delete the probe, then run `make tidy`.
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

## Publishing to Buf Schema Registry

**Separate from development workflow**: We use buf only for publishing modules
to buf.build.

```bash
# Before publishing - validate workspace
buf build
buf lint

# Publish with version labels
buf push --label main --label v0.1.2
```

**Module Structure**:

- `buf.build/protomcp/nanorpc` - Core protocol with Go options
- `buf.build/protomcp/nanopb` - nanopb extensions for embedded C

Users consume these via buf dependencies, not our internal protoc workflow.

## Common Mistakes to Avoid

See [Commit Message Requirements](#commit-message-requirements) above.

### File Operations

- **DON'T**: Create files without explicit request
- **DO**: Ask before creating new files
- **DON'T**: Move or rename files without understanding impact
- **DO**: Check dependencies before structural changes

### Code Changes

- **DON'T**: Make changes based on assumptions
- **DO**: Verify with `go doc` or existing code first
- **DON'T**: Ignore existing patterns
- **DO**: Follow established conventions

## Task Management with TodoWrite

Use the TodoWrite tool for:

- Tasks with 3+ steps
- Complex implementations
- When explicitly requested

Do NOT use for:

- Simple, single-step tasks
- Information queries
- Trivial operations

## go:generate Scripts

<!-- cspell:ignore GOFILE GOPACKAGE PKGDIR toplevel Iproto -->

When creating shell scripts for go:generate:

- Use `$GOFILE` - filename that triggered generate
- Use `$GOPACKAGE` - package name of the file
- Always use `git rev-parse --show-toplevel` to find repo root
- Quote all variable expansions

Example:

```bash
#!/bin/sh
set -eu

DIR="$PWD"
cd "$(git rev-parse --show-toplevel)"
PKGDIR="${DIR#"$PWD"/}"

protoc -Iproto/nanopb -Iproto/nanorpc -Iproto/vendor \
    "--go_out=$PKGDIR" \
    --go_opt=paths=source_relative \
    "proto/$GOPACKAGE/${GOFILE%.go}.proto"
```

## Troubleshooting

### Common Issues

1. **Protocol buffer compilation**:
   - Ensure `protoc` is installed and in PATH
   - Check that nanopb submodule is properly included
   - Verify import paths in proto files

2. **Module dependencies**:
   - Run `make tidy` to fix go.mod issues
   - Check the `go.work` workspace is active for cross-module dev
     (resolution is via the workspace; no `replace` directives)
   - Verify inter-module dependencies

3. **Tool detection failures**:
   - Install tools globally with `pnpm install -g <tool>`
   - Check that `pnpm dlx` is available and functional
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
