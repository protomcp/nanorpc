// Package testutils provides comprehensive testing utilities for nanorpc.
//
// This package contains assertion functions, mock implementations, and helper
// utilities used across the nanorpc codebase to enable consistent, reliable
// testing practices.
//
// # Assertions
//
// The package provides a comprehensive set of type-safe assertion functions
// that work with any type implementing the minimal T interface:
//
//	AssertEqual     - asserts two values are equal using reflect.DeepEqual
//	AssertNotEqual  - asserts two values are not equal
//	AssertTrue      - asserts a boolean is true
//	AssertFalse     - asserts a boolean is false
//	AssertNil       - asserts a value is nil (handles all nullable types)
//	AssertNotNil    - asserts a value is not nil
//	AssertNoError   - asserts an error is nil
//	AssertError     - asserts an error is not nil
//	AssertErrorIs   - asserts an error matches target using errors.Is
//	AssertContains  - asserts a string contains a substring
//	AssertTypeIs    - type-safe assertion with value return
//
// All assertion functions support optional formatted error messages:
//
//	AssertEqual(t, expected, actual, "values should match")
//	AssertTrue(t, condition, "condition failed: %s", reason)
//
// Generic assertions provide compile-time type safety:
//
//	value := AssertTypeIs[*MyStruct](t, result, "wrong type")
//	AssertEqual[string](t, "expected", actual, "string mismatch")
//
// # Testing Interface
//
// The T interface provides a comprehensive testing interface compatible with
// *testing.T but allows for mock implementations in testing the package itself:
//
//	type T interface {
//		Helper()
//		Fatal(args ...any)
//		Fatalf(format string, args ...any)
//		Error(args ...any)
//		Errorf(format string, args ...any)
//		Log(args ...any)
//		Logf(format string, args ...any)
//	}
//
// This design eliminates circular dependencies and enables comprehensive
// testing of the assertion functions themselves.
//
// # Mock Implementations
//
// ## MockT
//
// A complete T interface implementation for testing assertion functions:
//
//	mt := &MockT{}
//	AssertEqual(mt, expected, actual)
//
//	// Verify test results
//	if lastError, hasError := mt.LastError(); hasError {
//		// Assertion failed as expected
//	}
//
// MockT features:
//   - Complete T interface implementation
//   - Separate tracking of errors vs logs
//   - Thread-safe operations with mutex protection
//   - Fatal call detection with FatalCalled flag
//   - Convenient LastError(), LastLog(), LastFatal() accessors
//   - Reset() method for test reuse
//   - Proper field alignment for memory efficiency
//
// ## MockFieldLogger
//
// A complete slog.Logger implementation for testing logging behaviour:
//
//	logger := NewMockFieldLogger()
//	logger = logger.WithField("key", "value")
//	logger.Info().Print("message")
//
//	// Verify fields were added
//	if val, ok := GetField[string, string](logger.Fields, "key"); ok {
//		// Field was set correctly
//	}
//
// Features:
//   - Field tracking for verification
//   - Level-based filtering with configurable thresholds
//   - Full slog.Logger interface compliance
//   - No actual output (silent for tests)
//
// ## MockConn and MockAddr
//
// Network connection mocking for testing client-server interactions:
//
//	conn := &MockConn{
//		Local:  "127.0.0.1:8080",
//		Remote: "192.168.1.1:12345",
//		Data:   []byte("incoming data"),
//	}
//
//	// Test reading
//	buf := make([]byte, 1024)
//	n, err := conn.Read(buf)
//
//	// Test writing
//	conn.Write([]byte("outgoing data"))
//	// Verify with conn.WriteData
//
// MockConn features:
//   - Full net.Conn interface implementation
//   - Independent read/write data buffers
//   - Connection state tracking (open/closed)
//   - Configurable local/remote addresses
//   - Proper error handling for closed connections
//   - No-op deadline methods suitable for testing
//
// # Helper Functions
//
// ## GetField
//
// Type-safe field extraction from maps with any key type:
//
//	fields := map[string]any{"count": 42, "name": "test"}
//	count, ok := GetField[string, int](fields, "count")
//	name, ok := GetField[string, string](fields, "name")
//
// ## S (Slice Constructor)
//
// Generic slice constructor for inline test data:
//
//	testCases := S(
//		TestCase{name: "first", value: 1},
//		TestCase{name: "second", value: 2},
//	)
//
// This eliminates the need for verbose slice declarations in table-driven tests.
//
// # Design Principles
//
// The testutils package follows these design principles:
//
//   - Type Safety: Extensive use of generics for compile-time type checking
//   - Consistency: Uniform API patterns across all assertion functions
//   - Clarity: Clear, informative error messages with context
//   - Completeness: Full interface implementations for realistic testing
//   - Independence: No external dependencies beyond standard library and darvaza.org/slog
//   - Resource Free: Mock implementations don't require actual resources
//
// These utilities enable comprehensive testing of nanorpc components including
// network connections, logging behaviour, and complex data structures with
// consistent patterns and reliable verification.
package testutils
