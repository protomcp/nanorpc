package testutils

import (
	"sync"
	"testing"
)

// TestMockT_Helper tests the Helper method
func TestMockT_Helper(t *testing.T) {
	mt := &MockT{}

	// Initially should be 0
	AssertEqual(t, 0, mt.HelperCalled, "initial HelperCalled count")

	// Call Helper once
	mt.Helper()
	AssertEqual(t, 1, mt.HelperCalled, "HelperCalled after one call")

	// Call Helper multiple times
	mt.Helper()
	mt.Helper()
	AssertEqual(t, 3, mt.HelperCalled, "HelperCalled after three calls")
}

// TestMockT_Fatal tests the Fatal method
func TestMockT_Fatal(t *testing.T) {
	mt := &MockT{}

	// Test initial state
	AssertFalse(t, mt.Failed, "initial Failed state")
	AssertFalse(t, mt.FatalCalled, "initial FatalCalled state")
	_, hasError := mt.LastError()
	AssertFalse(t, hasError, "should not have error initially")

	// Test Fatal with no args
	mt.Fatal()
	AssertTrue(t, mt.Failed, "Failed after Fatal")
	AssertTrue(t, mt.FatalCalled, "FatalCalled after Fatal")
	lastError, hasError := mt.LastError()
	AssertTrue(t, hasError, "should have error after Fatal")
	AssertEqual(t, "", lastError, "empty message from Fatal()")

	// Reset and test Fatal with args
	mt.Reset()
	mt.Fatal("test error", 42)
	AssertTrue(t, mt.Failed, "Failed after Fatal with args")
	AssertTrue(t, mt.FatalCalled, "FatalCalled after Fatal with args")
	lastError, hasError = mt.LastError()
	AssertTrue(t, hasError, "should have error after Fatal with args")
	AssertEqual(t, "test error42", lastError, "message from Fatal with args")
}

// TestMockT_Fatalf tests the Fatalf method
func TestMockT_Fatalf(t *testing.T) {
	mt := &MockT{}

	// Test Fatalf with format string
	mt.Fatalf("error: %s, code: %d", "test", 42)
	AssertTrue(t, mt.Failed, "Failed after Fatalf")
	AssertTrue(t, mt.FatalCalled, "FatalCalled after Fatalf")
	lastError, hasError := mt.LastError()
	AssertTrue(t, hasError, "should have error after Fatalf")
	AssertEqual(t, "error: test, code: 42", lastError, "formatted message from Fatalf")

	// Test Fatalf with no format args
	mt.Reset()
	mt.Fatalf("simple message")
	lastError, hasError = mt.LastError()
	AssertTrue(t, hasError, "should have error after Fatalf with simple message")
	AssertEqual(t, "simple message", lastError, "simple message from Fatalf")
}

// TestMockT_Error tests the Error method
func TestMockT_Error(t *testing.T) {
	mt := &MockT{}

	// Test Error with no args
	mt.Error()
	lastError, hasError := mt.LastError()
	AssertTrue(t, hasError, "should have error after Error")
	AssertEqual(t, "", lastError, "empty message from Error()")
	_, hasLog := mt.LastLog()
	AssertFalse(t, hasLog, "Error should not add to Logs")
	AssertTrue(t, mt.Failed, "Failed should be true after Error")
	AssertFalse(t, mt.FatalCalled, "FatalCalled should remain false after Error")

	// Reset and test Error with args
	mt.Reset()
	mt.Error("error message", 123)
	lastError, hasError = mt.LastError()
	AssertTrue(t, hasError, "should have error after Error with args")
	AssertEqual(t, "error message123", lastError, "message from Error with args")
	_, hasLog = mt.LastLog()
	AssertFalse(t, hasLog, "Error should not add to Logs")
}

// TestMockT_Errorf tests the Errorf method
func TestMockT_Errorf(t *testing.T) {
	mt := &MockT{}

	// Test Errorf with format string
	mt.Errorf("error: %s, value: %d", "test", 42)
	lastError, hasError := mt.LastError()
	AssertTrue(t, hasError, "should have error after Errorf")
	AssertEqual(t, "error: test, value: 42", lastError, "formatted message from Errorf")
	_, hasLog := mt.LastLog()
	AssertFalse(t, hasLog, "Errorf should not add to Logs")
	AssertTrue(t, mt.Failed, "Failed should be true after Errorf")
	AssertFalse(t, mt.FatalCalled, "FatalCalled should remain false after Errorf")
}

// TestMockT_Log tests the Log method
func TestMockT_Log(t *testing.T) {
	mt := &MockT{}

	// Test Log with no args
	mt.Log()
	lastLog, hasLog := mt.LastLog()
	AssertTrue(t, hasLog, "should have log after Log")
	AssertEqual(t, "", lastLog, "empty message from Log()")
	_, hasError := mt.LastError()
	AssertFalse(t, hasError, "Log should not add to Errors")
	AssertFalse(t, mt.Failed, "Failed should remain false after Log")
	AssertFalse(t, mt.FatalCalled, "FatalCalled should remain false after Log")

	// Reset and test Log with args
	mt.Reset()
	mt.Log("log message", 456)
	lastLog, hasLog = mt.LastLog()
	AssertTrue(t, hasLog, "should have log after Log with args")
	AssertEqual(t, "log message456", lastLog, "message from Log with args")
	_, hasError = mt.LastError()
	AssertFalse(t, hasError, "Log should not add to Errors")
}

// TestMockT_Logf tests the Logf method
func TestMockT_Logf(t *testing.T) {
	mt := &MockT{}

	// Test Logf with format string
	mt.Logf("log: %s, number: %d", "test", 42)
	lastLog, hasLog := mt.LastLog()
	AssertTrue(t, hasLog, "should have log after Logf")
	AssertEqual(t, "log: test, number: 42", lastLog, "formatted message from Logf")
	_, hasError := mt.LastError()
	AssertFalse(t, hasError, "Logf should not add to Errors")
	AssertFalse(t, mt.Failed, "Failed should remain false after Logf")
	AssertFalse(t, mt.FatalCalled, "FatalCalled should remain false after Logf")
}

// TestMockT_LastLog tests the LastLog method
func TestMockT_LastLog(t *testing.T) {
	mt := &MockT{}

	// Test with no logs
	msg, ok := mt.LastLog()
	AssertFalse(t, ok, "LastLog should return false when no logs exist")
	AssertEqual(t, "", msg, "LastLog message should be empty when no logs exist")

	// Test with one log
	mt.Log("first log")
	msg, ok = mt.LastLog()
	AssertTrue(t, ok, "LastLog should return true when logs exist")
	AssertEqual(t, "first log", msg, "LastLog should return first log")

	// Test with multiple logs
	mt.Log("second log")
	mt.Logf("third log: %d", 3)
	msg, ok = mt.LastLog()
	AssertTrue(t, ok, "LastLog should return true when multiple logs exist")
	AssertEqual(t, "third log: 3", msg, "LastLog should return most recent log")
}

// TestMockT_LastError tests the LastError method
func TestMockT_LastError(t *testing.T) {
	mt := &MockT{}

	// Test with no errors
	msg, ok := mt.LastError()
	AssertFalse(t, ok, "LastError should return false when no errors exist")
	AssertEqual(t, "", msg, "LastError message should be empty when no errors exist")

	// Test with one error
	mt.Error("first error")
	msg, ok = mt.LastError()
	AssertTrue(t, ok, "LastError should return true when errors exist")
	AssertEqual(t, "first error", msg, "LastError should return first error")

	// Test with multiple errors
	mt.Error("second error")
	mt.Errorf("third error: %d", 3)
	msg, ok = mt.LastError()
	AssertTrue(t, ok, "LastError should return true when multiple errors exist")
	AssertEqual(t, "third error: 3", msg, "LastError should return most recent error")
}

// TestMockT_FatalFlag tests that Fatal methods set FatalCalled flag
func TestMockT_FatalFlag(t *testing.T) {
	mt := &MockT{}

	// Test with non-fatal error
	mt.Error("non-fatal error")
	AssertFalse(t, mt.FatalCalled, "FatalCalled should be false after Error")
	lastError, hasError := mt.LastError()
	AssertTrue(t, hasError, "should have error after Error")
	AssertEqual(t, "non-fatal error", lastError, "should have non-fatal error message")

	// Test with fatal call
	mt.Fatal("fatal error")
	AssertTrue(t, mt.FatalCalled, "FatalCalled should be true after Fatal")
	lastError, hasError = mt.LastError()
	AssertTrue(t, hasError, "should have error after Fatal")
	AssertEqual(t, "fatal error", lastError, "should have fatal error message")

	// Test with Fatalf call
	mt.Reset()
	mt.Fatalf("fatal error: %d", 42)
	AssertTrue(t, mt.FatalCalled, "FatalCalled should be true after Fatalf")
	lastError, hasError = mt.LastError()
	AssertTrue(t, hasError, "should have error after Fatalf")
	AssertEqual(t, "fatal error: 42", lastError, "should have Fatalf error message")
}

// TestMockT_Reset tests the Reset method
func TestMockT_Reset(t *testing.T) {
	mt := &MockT{}

	// Populate with data
	mt.Helper()
	mt.Helper()
	mt.Fatal("fatal error")
	mt.Error("error")
	mt.Log("log")

	// Verify state before reset
	AssertEqual(t, 2, mt.HelperCalled, "HelperCalled before reset")
	AssertTrue(t, mt.Failed, "Failed should be true before reset")
	AssertTrue(t, mt.FatalCalled, "FatalCalled should be true before reset")
	_, hasError := mt.LastError()
	AssertTrue(t, hasError, "should have errors before reset")
	_, hasLog := mt.LastLog()
	AssertTrue(t, hasLog, "should have logs before reset")

	// Reset
	mt.Reset()

	// Verify state after reset
	AssertEqual(t, 0, mt.HelperCalled, "HelperCalled after reset")
	AssertFalse(t, mt.Failed, "Failed should be false after reset")
	AssertFalse(t, mt.FatalCalled, "FatalCalled should be false after reset")
	_, hasError = mt.LastError()
	AssertFalse(t, hasError, "should not have errors after reset")
	_, hasLog = mt.LastLog()
	AssertFalse(t, hasLog, "should not have logs after reset")
	AssertNil(t, mt.Errors, "Errors should be nil after reset")
	AssertNil(t, mt.Logs, "Logs should be nil after reset")
}

// TestMockT_Concurrent tests concurrent usage of MockT
func TestMockT_Concurrent(t *testing.T) {
	mt := &MockT{}
	const numGoroutines = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch goroutines that perform operations
	for i := range numGoroutines {
		go testMockTWorker(mt, &wg, i)
	}

	wg.Wait()

	// Verify basic functionality
	AssertTrue(t, mt.HelperCalled > 0, "should have Helper calls")
	_, hasLogs := mt.LastLog()
	AssertTrue(t, hasLogs, "should have logs")
	_, hasErrors := mt.LastError()
	AssertTrue(t, hasErrors, "should have errors")

	// Test Reset works
	mt.Reset()
	AssertEqual(t, 0, mt.HelperCalled, "should reset HelperCalled")
}

func testMockTWorker(mt *MockT, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	mt.Helper()
	mt.Log("log", id)
	mt.Error("error", id)
}

// TestMockT_InterfaceCompliance tests that MockT implements T interface
func TestMockT_InterfaceCompliance(t *testing.T) {
	var mt = &MockT{}
	var tt T = mt

	// Test that all interface methods can be called
	tt.Helper()
	tt.Fatal("test")

	// Reset for further testing
	mt.Reset()

	tt.Fatalf("test %d", 42)
	mt.Reset()

	tt.Error("test")
	tt.Errorf("test %d", 42)
	tt.Log("test")
	tt.Logf("test %d", 42)

	// Verify some basic functionality
	AssertTrue(t, mt.Failed, "Failed should be true after Error calls")
	_, hasErrors := mt.LastError()
	AssertTrue(t, hasErrors, "should have errors after Error calls")
	_, hasLogs := mt.LastLog()
	AssertTrue(t, hasLogs, "should have logs after Log calls")
}
