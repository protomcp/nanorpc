// Package testutils provides test utilities for nanorpc including assertion helpers
// and mock implementations for testing.
package testutils

import (
	"fmt"
	"reflect"
	"strings"

	"darvaza.org/core"
)

// S is a generic slice constructor helper that creates a slice from variadic arguments.
// This is useful for creating inline slices in test assertions without declaring variables.
//
// Example:
//
//	AssertEqual(t, S(1, 2, 3), result)
//	AssertEqual(t, S("a", "b"), strings)
func S[T any](items ...T) []T {
	return items
}

// AssertNotNil fails the test if value is nil.
// This is useful for checking that pointers, interfaces, maps, slices, or channels are not nil.
//
// Example:
//
//	AssertNotNil(t, result)
//	AssertNotNil(t, err, "function should return an error")
func AssertNotNil(t T, value any, name string, args ...any) {
	t.Helper()
	ok := !core.IsNil(value)
	if !ok {
		doFatal(t, name, args, "expected non-nil value")
	}
}

// AssertNoError fails the test if error is not nil.
// This is the most common assertion for checking that functions completed successfully.
//
// Example:
//
//	err := doSomething()
//	AssertNoError(t, err)
//	AssertNoError(t, err, "doSomething failed")
func AssertNoError(t T, err error, name string, args ...any) {
	t.Helper()
	ok := err == nil
	if !ok {
		doFatal(t, name, args, "unexpected error: %v", err)
	}
}

// AssertError fails the test if error is nil.
// Use this when you expect a function to return an error.
//
// Example:
//
//	err := validateInput("")
//	AssertError(t, err)
//	AssertError(t, err, "empty input should cause error")
func AssertError(t T, err error, name string, args ...any) {
	t.Helper()
	ok := err != nil
	if !ok {
		doFatal(t, name, args, "expected an error but got nil")
	}
}

// AssertEqual fails the test if expected and actual are not equal.
// Uses reflect.DeepEqual for comparison, so it works with any type including slices, maps, and structs.
//
// Example:
//
//	AssertEqual(t, 42, result)
//	AssertEqual(t, []int{1, 2, 3}, slice)
//	AssertEqual(t, expected, actual, "values should match")
func AssertEqual[V any](t T, expected, actual V, name string, args ...any) {
	t.Helper()
	ok := reflect.DeepEqual(expected, actual)
	if !ok {
		doFatal(t, name, args, "expected %v, got %v", expected, actual)
	}
}

// AssertTrue fails the test if value is not true.
// Use this for boolean conditions that should be true.
//
// Example:
//
//	AssertTrue(t, len(slice) > 0)
//	AssertTrue(t, isValid, "value should be valid")
//
//revive:disable-next-line:flag-parameter
func AssertTrue(t T, value bool, name string, args ...any) {
	t.Helper()
	AssertEqual(t, true, value, name, args...)
}

// AssertFalse fails the test if value is not false.
// Use this for boolean conditions that should be false.
//
// Example:
//
//	AssertFalse(t, hasError)
//	AssertFalse(t, found, "item should not be found")
//
//revive:disable-next-line:flag-parameter
func AssertFalse(t T, value bool, name string, args ...any) {
	t.Helper()
	AssertEqual(t, false, value, name, args...)
}

// AssertNil fails the test if value is not nil.
// This is useful for checking that pointers, interfaces, maps, slices, or channels are nil.
//
// Example:
//
//	AssertNil(t, err)
//	AssertNil(t, result, "result should be nil on error")
func AssertNil(t T, value any, name string, args ...any) {
	t.Helper()
	ok := core.IsNil(value)
	if !ok {
		doFatal(t, name, args, "expected nil, got %v", value)
	}
}

// AssertNotEqual fails the test if expected and actual are equal.
// Uses reflect.DeepEqual for comparison, so it works with any type.
//
// Example:
//
//	AssertNotEqual(t, oldValue, newValue)
//	AssertNotEqual(t, 0, count, "count should not be zero")
func AssertNotEqual[V any](t T, expected, actual V, name string, args ...any) {
	t.Helper()
	ok := !reflect.DeepEqual(expected, actual)
	if !ok {
		doFatal(t, name, args, "expected values to be different, both were %v", expected)
	}
}

// AssertContains fails the test if the string doesn't contain the substring.
// This is useful for checking error messages or log output.
//
// Example:
//
//	AssertContains(t, err.Error(), "invalid")
//	AssertContains(t, output, "success", "output should indicate success")
func AssertContains(t T, str, substr string, name string, args ...any) {
	t.Helper()
	ok := strings.Contains(str, substr)
	if !ok {
		doFatal(t, name, args, "expected %q to contain %q", str, substr)
	}
}

// AssertTypeIs fails the test if value is not of the expected type.
// It returns the value cast to the expected type if successful.
//
// Example:
//
//	req := getSomeInterface()
//	typedReq := AssertTypeIs[*MyRequest](t, req, "expected *MyRequest")
func AssertTypeIs[U any](t T, value any, name string, args ...any) U {
	t.Helper()
	result, ok := value.(U)
	if !ok {
		var zero U
		doFatal(t, name, args, "expected type %T but got %T", zero, value)
	}
	return result
}

// doFatal builds a formatted error message and calls t.Fatal.
// It combines an optional prefix message with a main message.
func doFatal(t T, prefixFormat string, prefixArgs []any, messageFormat string, args ...any) {
	var msg string
	if prefixFormat != "" {
		prefix := fmt.Sprintf(prefixFormat, prefixArgs...)
		msg = fmt.Sprintf("%s: %s", prefix, fmt.Sprintf(messageFormat, args...))
	} else {
		msg = fmt.Sprintf(messageFormat, args...)
	}
	t.Fatal(msg)
}
