// Package testutils provides testing utilities for nanorpc.
//
// This package contains mock implementations and helper functions
// used in tests across the nanorpc codebase.
//
// # Assertions
//
// The package provides a comprehensive set of assertion functions that work
// with any type implementing the minimal T interface:
//
//	AssertEqual    - asserts two values are equal
//	AssertNotEqual - asserts two values are not equal
//	AssertTrue     - asserts a boolean is true
//	AssertFalse    - asserts a boolean is false
//	AssertNil      - asserts a value is nil
//	AssertNotNil   - asserts a value is not nil
//	AssertNoError  - asserts an error is nil
//	AssertError    - asserts an error is not nil
//	AssertContains - asserts a string contains a substring
//	AssertTypeIs   - asserts a value is of a specific type and returns it
//
// All assertion functions support optional formatted messages:
//
//	AssertEqual(t, got, want, "expected %q but got %q", want, got)
//
// # Testing Interface
//
// The T interface provides a minimal testing interface that is compatible
// with *testing.T but allows for mock implementations in tests:
//
//	type T interface {
//		Helper()
//		Fatal(args ...any)
//	}
//
// This design allows the testutils package to be tested without circular
// dependencies on the testing package.
package testutils
