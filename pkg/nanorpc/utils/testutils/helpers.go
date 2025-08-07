package testutils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"darvaza.org/core"
)

// GetField retrieves a typed value from a map[K]any
func GetField[K comparable, T any](m map[K]any, key K) (T, bool) {
	var zero T
	if v, ok := m[key]; ok {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	return zero, false
}

// ConcurrentTestHelper helps test concurrent operations with multiple goroutines.
// It's useful for testing race conditions and concurrent safety.
type ConcurrentTestHelper struct {
	// TestFunc is the function to run in each goroutine
	TestFunc func(id int) error
	// NumGoroutines is the number of concurrent goroutines to run
	NumGoroutines int
	// Timeout is the maximum time to wait for all goroutines to complete
	Timeout time.Duration
}

// Run executes the concurrent test and returns any errors from the goroutines.
// Returns a slice of errors, one for each goroutine (nil if no error).
func (h *ConcurrentTestHelper) Run() []error {
	h.setDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()

	errors := make([]error, h.NumGoroutines)
	var wg sync.WaitGroup

	for i := range h.NumGoroutines {
		wg.Add(1)
		go h.runWorker(ctx, i, &wg, errors)
	}

	wg.Wait()
	return errors
}

// setDefaults sets default values for helper configuration
func (h *ConcurrentTestHelper) setDefaults() {
	if h.NumGoroutines <= 0 {
		h.NumGoroutines = 10
	}
	if h.Timeout <= 0 {
		h.Timeout = 5 * time.Second
	}
}

// runWorker executes the test function in a goroutine
func (h *ConcurrentTestHelper) runWorker(ctx context.Context, id int, wg *sync.WaitGroup, errors []error) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		errors[id] = ctx.Err()
	default:
		if h.TestFunc != nil {
			errors[id] = h.TestFunc(id)
		}
	}
}

// WaitForCondition waits for a condition to become true within a timeout.
// It polls the condition function at regular intervals.
// Returns true if condition became true, false if timeout expired.
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	if interval <= 0 {
		interval = 10 * time.Millisecond
	}

	// Check once immediately
	if condition() {
		return true
	}

	return waitForConditionLoop(condition, timeout, interval)
}

// waitForConditionLoop performs the polling loop for WaitForCondition
func waitForConditionLoop(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if condition() {
				return true
			}
		}
	}
}

// AssertWaitForCondition is like WaitForCondition but fails the test if condition doesn't become true.
func AssertWaitForCondition(t core.T, condition func() bool, timeout time.Duration, name string, args ...any) bool {
	t.Helper()
	ok := WaitForCondition(condition, timeout, 0)
	if !ok {
		msg := fmt.Sprintf("condition did not become true within %v", timeout)
		if name != "" {
			prefix := fmt.Sprintf(name, args...)
			msg = fmt.Sprintf("%s: %s", prefix, msg)
		}
		t.Error(msg)
	}
	return ok
}

// AssertField asserts that a field exists in a map.
func AssertField(t core.T, m map[string]any, field, name string, args ...any) (any, bool) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		doError(t, name, args, "field %q not found", field)
	}
	return v, ok
}

// AssertNotField asserts that a field does not exist in a map.
func AssertNotField(t core.T, m map[string]any, field, name string, args ...any) bool {
	t.Helper()
	if v, ok := m[field]; ok {
		doError(t, name, args, "field %q should not exist, got %v", field, v)
		return false
	}
	return true
}

// AssertFieldTypeIs asserts that a field in a map has the expected type.
func AssertFieldTypeIs[T any](t core.T, m map[string]any, field, name string, args ...any) (T, bool) {
	t.Helper()
	var zero T

	vi, ok := AssertField(t, m, field, name, args...)
	if !ok {
		return zero, false
	}

	v, ok := vi.(T)
	if !ok {
		doError(t, name, args, "field %q type %T, expected %T", field, vi, zero)
	}

	return v, ok
}

func doError(t core.T, prefixFormat string, prefixArgs []any, messageFormat string, messageArgs ...any) {
	var prefix, msg string
	if prefixFormat != "" {
		prefix = fmt.Sprintf(prefixFormat, prefixArgs...)
	}
	msg = fmt.Sprintf(messageFormat, messageArgs...)
	if prefix != "" {
		msg = prefix + ": " + msg
	}
	t.Error(msg)
}
