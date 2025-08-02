package testutils

import (
	"fmt"
	"sync"
)

// This compile-time check guarantees that MockT can be used wherever T is expected.
var _ T = (*MockT)(nil)

// MockT is a mock implementation of the T interface for testing assertion functions.
// It captures calls to Helper and Fatal for verification in tests.
// Writes are protected by mutex for concurrent testing; reads are unprotected.
type MockT struct {
	Errors       []string
	Logs         []string
	HelperCalled int
	mu           sync.Mutex
	Failed       bool
	FatalCalled  bool
}

// Helper marks the calling function as a test helper function.
// Increments the HelperCalled counter for verification in tests.
func (m *MockT) Helper() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HelperCalled++
}

// Fatal is equivalent to Log followed by FailNow.
// It captures the args and marks the test as failed.
func (m *MockT) Fatal(args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Failed = true
	m.FatalCalled = true

	// Format the message
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprint(args...)
	}

	m.Errors = append(m.Errors, msg)
}

// Fatalf is equivalent to Logf followed by FailNow.
// It formats according to a format specifier and marks the test as failed.
func (m *MockT) Fatalf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Failed = true
	m.FatalCalled = true

	msg := fmt.Sprintf(format, args...)
	m.Errors = append(m.Errors, msg)
}

// Error logs an error message without stopping the test.
// Adds the message to the Errors slice and marks the test as failed.
func (m *MockT) Error(args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Failed = true

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprint(args...)
	}

	m.Errors = append(m.Errors, msg)
}

// Errorf formats according to a format specifier and logs an error message without stopping the test.
// Adds the message to the Errors slice and marks the test as failed.
func (m *MockT) Errorf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Failed = true

	msg := fmt.Sprintf(format, args...)
	m.Errors = append(m.Errors, msg)
}

// Log logs a message.
// Adds the message to the Logs slice.
func (m *MockT) Log(args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprint(args...)
	}

	m.Logs = append(m.Logs, msg)
}

// Logf formats its arguments according to the format and logs the message.
// Adds the message to the Logs slice.
func (m *MockT) Logf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	m.Logs = append(m.Logs, msg)
}

// LastLog returns the last log message and whether one exists.
// This is not part of the T interface but is useful for testing.
func (m *MockT) LastLog() (string, bool) {
	if len(m.Logs) == 0 {
		return "", false
	}
	return m.Logs[len(m.Logs)-1], true
}

// LastError returns the last error message and whether one exists.
// This is not part of the T interface but is useful for testing.
func (m *MockT) LastError() (string, bool) {
	if len(m.Errors) == 0 {
		return "", false
	}
	return m.Errors[len(m.Errors)-1], true
}

// Reset clears the mock state, allowing reuse.
// This is not part of the T interface but is useful for testing.
func (m *MockT) Reset() {
	m.Errors = nil
	m.Logs = nil
	m.HelperCalled = 0
	m.Failed = false
	m.FatalCalled = false
}
