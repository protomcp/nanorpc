package testutils

import (
	"testing"

	"darvaza.org/slog"
)

const (
	testValue1 = "value1"
	testValue2 = "value2"
)

// Helper factory functions
func newMockFieldLoggerWithFields(fields map[string]any) *MockFieldLogger {
	logger := NewMockFieldLogger()
	for k, v := range fields {
		logger.Fields[k] = v
	}
	return logger
}

func TestNewMockFieldLogger(t *testing.T) {
	logger := NewMockFieldLogger()
	if logger == nil {
		t.Fatal("NewMockFieldLogger returned nil")
	}
	if logger.Fields == nil {
		t.Error("Fields map should be initialized")
	}
	if logger.CurrentLevel != slog.UndefinedLevel {
		t.Error("CurrentLevel should be UndefinedLevel by default")
	}
	if logger.Threshold != slog.Debug {
		t.Error("Threshold should be Debug (accept all) by default")
	}
}

type withFieldTestCase struct {
	initial   map[string]any
	addValue  any
	expectVal any
	name      string
	addKey    string
	expectKey string
}

func (tc *withFieldTestCase) test(t *testing.T) {
	logger := newMockFieldLoggerWithFields(tc.initial)
	newLogger := logger.WithField(tc.addKey, tc.addValue)

	// Verify it returns MockFieldLogger
	ml, ok := newLogger.(*MockFieldLogger)
	if !ok {
		t.Fatal("WithField should return *MockFieldLogger")
	}

	// Check the expected field
	if ml.Fields[tc.expectKey] != tc.expectVal {
		t.Errorf("expected field %s=%v, got %v", tc.expectKey, tc.expectVal, ml.Fields[tc.expectKey])
	}
}

func TestMockFieldLoggerWithField(t *testing.T) {
	tests := []withFieldTestCase{
		{
			name:      "add single field",
			initial:   nil,
			addKey:    "key1",
			addValue:  testValue1,
			expectKey: "key1",
			expectVal: testValue1,
		},
		{
			name:      "add to existing fields",
			initial:   map[string]any{"existing": "value"},
			addKey:    "key2",
			addValue:  testValue2,
			expectKey: "key2",
			expectVal: testValue2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}

	// Test immutability separately
	t.Run("original not modified", func(t *testing.T) {
		logger := NewMockFieldLogger()
		_ = logger.WithField("key1", testValue1)
		if logger.Fields["key1"] != nil {
			t.Error("Original logger should not be modified")
		}
	})
}

func TestMockFieldLoggerWithFields(t *testing.T) {
	logger := NewMockFieldLogger()

	// Add multiple fields
	fields := map[string]any{
		"field1": testValue1,
		"field2": 42,
		"field3": true,
	}
	logger2 := logger.WithFields(fields)
	if ml, ok := logger2.(*MockFieldLogger); ok {
		if ml.Fields["field1"] != testValue1 {
			t.Error("field1 should be added")
		}
		if ml.Fields["field2"] != 42 {
			t.Error("field2 should be added")
		}
		if !ml.Fields["field3"].(bool) {
			t.Error("field3 should be added")
		}
	}
}

type levelMethodTestCase struct {
	method   func(slog.Logger) slog.Logger
	name     string
	expected slog.LogLevel
}

func (tc *levelMethodTestCase) test(t *testing.T) {
	logger := NewMockFieldLogger()
	result := tc.method(logger)

	ml, ok := result.(*MockFieldLogger)
	if !ok {
		t.Fatal("level method should return *MockFieldLogger")
	}

	if ml.CurrentLevel != tc.expected {
		t.Errorf("expected level %v, got %v", tc.expected, ml.CurrentLevel)
	}
}

func TestMockFieldLoggerLevelMethods(t *testing.T) {
	tests := []levelMethodTestCase{
		{
			name:     "Debug",
			method:   slog.Logger.Debug,
			expected: slog.Debug,
		},
		{
			name:     "Info",
			method:   slog.Logger.Info,
			expected: slog.Info,
		},
		{
			name:     "Warn",
			method:   slog.Logger.Warn,
			expected: slog.Warn,
		},
		{
			name:     "Error",
			method:   slog.Logger.Error,
			expected: slog.Error,
		},
		{
			name:     "Fatal",
			method:   slog.Logger.Fatal,
			expected: slog.Fatal,
		},
		{
			name:     "Panic",
			method:   slog.Logger.Panic,
			expected: slog.Panic,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func TestMockFieldLoggerEnabled(t *testing.T) {
	logger := NewMockFieldLogger()

	// Base logger without level should not be enabled
	if logger.Enabled() {
		t.Error("Logger without level should not be enabled")
	}

	// Debug logger with Debug threshold should be enabled
	debugLogger := logger.Debug()
	if !debugLogger.Enabled() {
		t.Error("Debug logger should be enabled with Debug threshold")
	}

	// Set threshold to Warn
	logger.Threshold = slog.Warn

	// Debug should be disabled (Debug > Warn in value)
	debugLogger = logger.Debug()
	if debugLogger.Enabled() {
		t.Error("Debug should be disabled when threshold is Warn")
	}

	// Info should be disabled
	infoLogger := logger.Info()
	if infoLogger.Enabled() {
		t.Error("Info should be disabled when threshold is Warn")
	}

	// Warn should be enabled
	warnLogger := logger.Warn()
	if !warnLogger.Enabled() {
		t.Error("Warn should be enabled when threshold is Warn")
	}

	// Error should be enabled (Error < Warn in value)
	errorLogger := logger.Error()
	if !errorLogger.Enabled() {
		t.Error("Error should be enabled when threshold is Warn")
	}
}

func TestMockFieldLoggerWithEnabled(t *testing.T) {
	logger := NewMockFieldLogger()
	logger.Threshold = slog.Info

	// Test with Info level
	infoLogger := logger.Info()
	returnedLogger, enabled := infoLogger.WithEnabled()
	if !enabled {
		t.Error("Info should be enabled with Info threshold")
	}
	if returnedLogger != infoLogger {
		t.Error("WithEnabled should return the same logger")
	}

	// Test with Debug level (should be disabled)
	debugLogger := logger.Debug()
	_, enabled = debugLogger.WithEnabled()
	if enabled {
		t.Error("Debug should be disabled with Info threshold")
	}
}

func TestMockFieldLoggerWithLevel(t *testing.T) {
	logger := NewMockFieldLogger()

	customLogger := logger.WithLevel(slog.Error)
	if ml, ok := customLogger.(*MockFieldLogger); ok {
		if ml.CurrentLevel != slog.Error {
			t.Error("WithLevel should set the specified level")
		}
	}
}

func TestMockFieldLoggerPrintMethods(_ *testing.T) {
	// Just verify they don't panic
	logger := NewMockFieldLogger()
	logger.Print("test")
	logger.Println("test")
	logger.Printf("test %s", "value")
}

func TestMockFieldLoggerWithStack(t *testing.T) {
	logger := NewMockFieldLogger()
	stackLogger := logger.WithStack(1)
	if stackLogger != logger {
		t.Error("WithStack should return the same logger (stack not tracked)")
	}
}
