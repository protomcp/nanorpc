// Package testutils provides testing utilities for nanorpc.
package testutils

import (
	"darvaza.org/slog"
)

// Compile-time check that MockFieldLogger implements slog.Logger
var _ slog.Logger = (*MockFieldLogger)(nil)

// MockFieldLogger is a test logger that tracks fields for verification
type MockFieldLogger struct {
	Fields       map[string]any
	CurrentLevel slog.LogLevel // the level this logger instance is set to
	Threshold    slog.LogLevel // minimum level that will produce output
}

// NewMockFieldLogger creates a new mock field logger with all levels enabled
func NewMockFieldLogger() *MockFieldLogger {
	return &MockFieldLogger{
		Fields:       make(map[string]any),
		CurrentLevel: slog.UndefinedLevel,
		Threshold:    slog.Debug, // Accept all levels by default
	}
}

// WithField returns a new logger with the given field added
func (m *MockFieldLogger) WithField(key string, value any) slog.Logger {
	newLogger := *m
	newLogger.Fields = make(map[string]any)
	for k, v := range m.Fields {
		newLogger.Fields[k] = v
	}
	newLogger.Fields[key] = value
	return &newLogger
}

// WithFields returns a new logger with the given fields added
func (m *MockFieldLogger) WithFields(fields map[string]any) slog.Logger {
	newLogger := *m
	newLogger.Fields = make(map[string]any)
	for k, v := range m.Fields {
		newLogger.Fields[k] = v
	}
	for k, v := range fields {
		newLogger.Fields[k] = v
	}
	return &newLogger
}

// Debug returns a debug logger
func (m *MockFieldLogger) Debug() slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = slog.Debug
	return &newLogger
}

// Info returns an info logger
func (m *MockFieldLogger) Info() slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = slog.Info
	return &newLogger
}

// Error returns an error logger
func (m *MockFieldLogger) Error() slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = slog.Error
	return &newLogger
}

// Warn returns a warn logger
func (m *MockFieldLogger) Warn() slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = slog.Warn
	return &newLogger
}

// Fatal returns a fatal logger (mock doesn't actually exit)
func (m *MockFieldLogger) Fatal() slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = slog.Fatal
	return &newLogger
}

// WithEnabled returns the logger and whether it's enabled
func (m *MockFieldLogger) WithEnabled() (slog.Logger, bool) {
	return m, m.Enabled()
}

// Print does nothing (test logger)
func (*MockFieldLogger) Print(_ ...any) {}

// Println does nothing (test logger)
func (*MockFieldLogger) Println(_ ...any) {}

// Printf does nothing (test logger)
func (*MockFieldLogger) Printf(_ string, _ ...any) {}

// Panic returns a panic logger (mock doesn't actually panic)
func (m *MockFieldLogger) Panic() slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = slog.Panic
	return &newLogger
}

// WithLevel returns a logger for the specified level
func (m *MockFieldLogger) WithLevel(level slog.LogLevel) slog.Logger {
	newLogger := *m
	newLogger.CurrentLevel = level
	return &newLogger
}

// WithStack returns the logger (test logger doesn't track stack)
func (m *MockFieldLogger) WithStack(_ int) slog.Logger {
	return m
}

// Enabled returns true if the current level meets the threshold
func (m *MockFieldLogger) Enabled() bool {
	// If no level is set, it's not enabled
	if m.CurrentLevel == slog.UndefinedLevel {
		return false
	}
	// Check if current level meets threshold (lower value = higher priority)
	return m.CurrentLevel <= m.Threshold
}
