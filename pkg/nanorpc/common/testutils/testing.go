// Package testutils provides test utilities for nanorpc.
package testutils

import "testing"

// T is a minimal interface for test assertion functions.
// It matches the subset of testing.T that our assertions need.
// This allows for easier mocking in tests and reduces coupling to the testing package.
//
// Any type that implements Helper() and Fatal() can be used with our assertion functions.
// The standard testing.T type implements this interface.
type T interface {
	// Helper marks the calling function as a test helper function.
	// When printing file and line information, that function will be skipped.
	Helper()
	// Fatal is equivalent to Log followed by FailNow.
	// It prints the args and stops the test execution.
	Fatal(args ...any)
}

// Ensure testing.T implements our T interface.
// This compile-time check guarantees that testing.T can be used wherever T is expected.
var _ T = (*testing.T)(nil)
