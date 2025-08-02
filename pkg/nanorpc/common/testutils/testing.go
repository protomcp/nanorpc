// Package testutils provides test utilities for nanorpc.
package testutils

import "testing"

// T is an interface for test assertion functions.
// It matches the subset of testing.T that our assertions need.
// This allows for easier mocking in tests and reduces coupling to the testing package.
//
// Any type that implements these methods can be used with our assertion functions.
// The standard testing.T type implements this interface.
type T interface {
	// Helper marks the calling function as a test helper function.
	// When printing file and line information, that function will be skipped.
	Helper()
	// Fatal is equivalent to Log followed by FailNow.
	// It prints the args and stops the test execution.
	Fatal(args ...any)
	// Fatalf is equivalent to Logf followed by FailNow.
	// It formats according to a format specifier and stops the test execution.
	Fatalf(format string, args ...any)
	// Error is equivalent to Log followed by Fail.
	// It logs an error message without stopping the test.
	Error(args ...any)
	// Errorf is equivalent to Logf followed by Fail.
	// It formats according to a format specifier and logs an error message.
	Errorf(format string, args ...any)
	// Log formats its arguments using default formatting and records the text.
	Log(args ...any)
	// Logf formats its arguments according to the format and records the text.
	Logf(format string, args ...any)
}

// Ensure testing.T implements our T interface.
// This compile-time check guarantees that testing.T can be used wherever T is expected.
var _ T = (*testing.T)(nil)

// TestCase represents a test case that can be run by RunTestCases.
// This interface follows the TESTING.md pattern for table-driven tests
// with named test types and test methods.
type TestCase interface {
	// Name returns the name of the test case for use with t.Run()
	Name() string
	// Test executes the test case logic
	Test(t *testing.T)
}

// RunTestCases runs a slice of test cases using t.Run() for each case.
// This is a generic helper that works with any type implementing TestCase.
// It follows the TESTING.md pattern of using named functions with t.Run().
//
// Example usage:
//
//	type myTestCase struct {
//		name     string
//		input    string
//		expected string
//		wantErr  bool
//	}
//
//	func (tc myTestCase) Name() string { return tc.name }
//	func (tc myTestCase) Test(t *testing.T) {
//		result, err := doSomething(tc.input)
//		// ... test logic
//	}
//
//	func TestSomething(t *testing.T) {
//		tests := []myTestCase{
//			{name: "valid input", input: "test", expected: "TEST"},
//			{name: "empty input", input: "", wantErr: true},
//		}
//		RunTestCases(t, tests)
//	}
func RunTestCases[V TestCase](t *testing.T, testCases []V) {
	for _, tc := range testCases {
		t.Run(tc.Name(), tc.Test)
	}
}
