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
    server.go server_*.go client.go client_*.go \
    errors.go hashcache.go path.go utils.go request_counter.go

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

This project focuses on providing efficient, reliable RPC communication for
embedded systems while maintaining clean, well-tested Go code.
