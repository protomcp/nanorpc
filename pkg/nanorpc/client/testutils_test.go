package client

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

// S is a generic slice constructor helper
func S[T any](items ...T) []T {
	return items
}

// AssertNotNil fails the test if value is nil
func AssertNotNil(t *testing.T, value any, msgAndArgs ...any) {
	t.Helper()
	if value == nil {
		msg := "expected non-nil value"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Fatal(msg)
	}
}

// AssertNoError fails the test if error is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		msg := fmt.Sprintf("unexpected error: %v", err)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s: %v", msg, err)
		}
		t.Fatal(msg)
	}
}

// AssertError fails the test if error is nil
func AssertError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		msg := "expected an error but got nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Fatal(msg)
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual[T any](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := fmt.Sprintf("expected %v, got %v", expected, actual)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s: expected %v, got %v", msg, expected, actual)
		}
		t.Fatal(msg)
	}
}

// AssertNotEqual fails the test if expected == actual
func AssertNotEqual[T any](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := fmt.Sprintf("expected values to be different but both were %v", expected)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Fatal(msg)
	}
}

// AssertTrue fails the test if value is not true
//
//revive:disable-next-line:flag-parameter
func AssertTrue(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if !value {
		msg := "expected true but got false"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Fatal(msg)
	}
}

// AssertFalse fails the test if value is not false
//
//revive:disable-next-line:flag-parameter
func AssertFalse(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if value {
		msg := "expected false but got true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Fatal(msg)
	}
}

// AssertTypeIs fails the test if value is not of the expected type
func AssertTypeIs[T any](t *testing.T, value any, msgAndArgs ...any) T {
	t.Helper()
	result, ok := value.(T)
	if !ok {
		var zero T
		msg := fmt.Sprintf("expected type %T but got %T", zero, value)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s: expected type %T but got %T", msg, zero, value)
		}
		t.Fatal(msg)
	}
	return result
}

// ConcurrentTestHelper provides utilities for executing test functions concurrently
// across multiple goroutines with safe collection of results and errors.
// It uses synchronization primitives to coordinate concurrent execution and
// protect shared state during result collection.
type ConcurrentTestHelper struct {
	t           *testing.T
	results     []any
	errors      []error
	wg          sync.WaitGroup
	mutex       sync.Mutex
	numRoutines int
}

// NewConcurrentTestHelper creates a new ConcurrentTestHelper for executing
// test functions concurrently across the specified number of goroutines.
// It pre-allocates slices for results and errors based on numRoutines.
func NewConcurrentTestHelper(t *testing.T, numRoutines int) *ConcurrentTestHelper {
	t.Helper()
	return &ConcurrentTestHelper{
		t:           t,
		numRoutines: numRoutines,
		results:     make([]any, numRoutines),
		errors:      make([]error, numRoutines),
	}
}

// Run executes the test function concurrently
func (h *ConcurrentTestHelper) Run(testFunc func(int) (any, error)) {
	h.t.Helper()
	h.wg.Add(h.numRoutines)

	for i := 0; i < h.numRoutines; i++ {
		go func(idx int) {
			defer h.wg.Done()
			result, err := testFunc(idx)

			h.mutex.Lock()
			h.results[idx] = result
			h.errors[idx] = err
			h.mutex.Unlock()
		}(i)
	}

	h.wg.Wait()
}

// GetResults returns all results and errors
func (h *ConcurrentTestHelper) GetResults() ([]any, []error) {
	h.t.Helper()
	return h.results, h.errors
}

// AssertNoErrors checks that no goroutines returned errors
func (h *ConcurrentTestHelper) AssertNoErrors() {
	h.t.Helper()
	for i, err := range h.errors {
		if err != nil {
			h.t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}
}

// GetResult returns the result at index with type assertion
func GetResult[T any](values []any, index int) (T, bool) {
	if index < 0 || index >= len(values) {
		var zero T
		return zero, false
	}
	result, ok := values[index].(T)
	return result, ok
}
