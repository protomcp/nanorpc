package client

import (
	"sync"
	"testing"
)

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

// Run executes the given function concurrently in multiple goroutines
func (h *ConcurrentTestHelper) Run(fn func(int) (any, error)) {
	h.wg.Add(h.numRoutines)
	for i := range h.numRoutines {
		go func(idx int) {
			defer h.wg.Done()
			result, err := fn(idx)
			h.mutex.Lock()
			h.results[idx] = result
			h.errors[idx] = err
			h.mutex.Unlock()
		}(i)
	}
	h.wg.Wait()
}

// AssertNoErrors checks that no errors occurred during concurrent execution
func (h *ConcurrentTestHelper) AssertNoErrors() {
	h.t.Helper()
	for i, err := range h.errors {
		if err != nil {
			h.t.Errorf("Error at index %d: %v", i, err)
		}
	}
}

// GetResults returns the results and errors from concurrent execution
func (h *ConcurrentTestHelper) GetResults() ([]any, []error) {
	return h.results, h.errors
}
