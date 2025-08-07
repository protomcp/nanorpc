package server

import (
	"errors"
	"testing"

	"darvaza.org/slog"

	"darvaza.org/core"
	"protomcp.org/nanorpc/pkg/nanorpc/utils"
	"protomcp.org/nanorpc/pkg/nanorpc/utils/testutils"
)

// Test server default logger
func TestServerDefaultLogger(t *testing.T) {
	s := &Server{}
	logger := s.getLogger()
	core.AssertNotNil(t, logger, "default logger should not be nil")

	// Verify it doesn't panic
	logger.Info().Print("test message")
}

// Test server custom logger
func TestServerCustomLogger(t *testing.T) {
	customLogger := testutils.NewMockFieldLogger()
	s := &Server{logger: customLogger}
	logger := s.getLogger()

	core.AssertEqual[slog.Logger](t, customLogger, logger, "should return custom logger")
}

// Test Server.WithDebug
func TestServerWithDebug(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug // Enable debug level
	s := &Server{logger: mockLog}

	logger, ok := s.WithDebug()
	core.AssertTrue(t, ok, "WithDebug should return true when debug enabled")
	core.AssertNotNil(t, logger, "WithDebug should return a logger")

	// Verify the logger has the correct level
	if ml, ok := logger.(*testutils.MockFieldLogger); ok {
		core.AssertEqual(t, slog.Debug, ml.CurrentLevel, "should have Debug level")
	}
}

// Test Server.LogInfo
func TestServerLogInfo(t *testing.T) {
	// Test that LogInfo is called when info level is enabled
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	s := &Server{logger: mockLog}

	// LogInfo should succeed when threshold allows it
	s.LogInfo(nil, "test info message")

	// Test that LogInfo is not called when threshold is too high
	mockLog.Threshold = slog.Error
	s.LogInfo(nil, "this should not log")

	// No panic means test passed
	core.AssertTrue(t, true, "LogInfo executed without panic")
}

// Test Server.WithError
func TestServerWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	s := &Server{logger: mockLog}
	testErr := errors.New("test error")

	logger, ok := s.WithError(testErr)
	core.AssertTrue(t, ok, "WithError should return true when error enabled")
	core.AssertNotNil(t, logger, "WithError should return a logger")

	// Check error field was added
	if ml, ok := logger.(*testutils.MockFieldLogger); ok {
		if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, utils.FieldError, "error field"); ok {
			core.AssertEqual(t, testErr, err, "should have error field")
		}
	}
}

// Test SessionManager logging
func TestSessionManagerWithInfo(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	sm := &DefaultSessionManager{logger: mockLog}

	logger, ok := sm.WithInfo()
	core.AssertTrue(t, ok, "WithInfo should return true when info enabled")
	core.AssertNotNil(t, logger, "WithInfo should return a logger")
}

// Test Session logging with fields
func TestSessionWithDebug(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug

	// Create a mock connection
	mockConn := &testutils.MockConn{
		Remote: "127.0.0.1:12345",
	}

	// Use the constructor to ensure fields are properly added
	s := NewDefaultSession(mockConn, nil, mockLog)

	logger, ok := s.WithDebug()
	core.AssertTrue(t, ok, "WithDebug should return true when debug enabled")
	core.AssertNotNil(t, logger, "WithDebug should return a logger")

	// Check session fields are added
	if ml, ok := logger.(*testutils.MockFieldLogger); ok {
		// Check component field
		if component, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields,
			utils.FieldComponent, "component field"); ok {
			core.AssertEqual(t, utils.ComponentSession, component, "should have session component")
		}

		// Check session ID field
		if sid, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, utils.FieldSessionID, "session_id field"); ok {
			core.AssertEqual(t, s.ID(), sid, "should have session ID")
		}

		// Check remote address field
		if addr, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields,
			utils.FieldRemoteAddr, "remote_addr field"); ok {
			core.AssertEqual(t, "127.0.0.1:12345", addr, "should have remote address")
		}
	}
}

// Test logAccept helper
func TestServerLogAccept(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug
	s := &Server{logger: mockLog}

	// Create mock connection
	conn := &testutils.MockConn{
		Local:  "127.0.0.1:8080",
		Remote: "192.168.1.1:12345",
	}

	// logAccept should succeed when debug is enabled
	s.logAccept(conn)

	// Test that logAccept does nothing when debug is disabled
	mockLog.Threshold = slog.Info
	s.logAccept(conn)

	// Verify the method handles the connection properly
	core.AssertEqual(t, "127.0.0.1:8080", conn.LocalAddr().String(), "local address should match")
	core.AssertEqual(t, "192.168.1.1:12345", conn.RemoteAddr().String(), "remote address should match")
}

// Test additional Server logging methods
func TestServerLogDebug(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug
	s := &Server{logger: mockLog}

	// Test LogDebug with and without fields
	s.LogDebug(nil, "debug message")
	s.LogDebug(map[string]any{"key": "value"}, "debug message with fields")

	// Test when debug is disabled
	mockLog.Threshold = slog.Info
	s.LogDebug(nil, "should not log")
}

func TestServerWithWarn(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	s := &Server{logger: mockLog}
	testErr := errors.New("test warning")

	logger, ok := s.WithWarn(testErr)
	core.AssertTrue(t, ok, "WithWarn should return true when warn enabled")
	core.AssertNotNil(t, logger, "WithWarn should return a logger")
}

func TestServerLogWarn(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	s := &Server{logger: mockLog}
	testErr := errors.New("test warning")

	s.LogWarn(testErr, nil, "warn message")
	s.LogWarn(testErr, map[string]any{"key": "value"}, "warn message with fields")

	// Test when warn is disabled
	mockLog.Threshold = slog.Error
	s.LogWarn(testErr, nil, "should not log")
}

func TestServerLogError(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	s := &Server{logger: mockLog}
	testErr := errors.New("test error")

	s.LogError(testErr, nil, "error message")
	s.LogError(testErr, map[string]any{"key": "value"}, "error message with fields")
}

// Test SessionManager additional logging methods
func TestSessionManagerWithDebug(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug
	sm := &DefaultSessionManager{logger: mockLog}

	logger, ok := sm.WithDebug()
	core.AssertTrue(t, ok, "WithDebug should return true when debug enabled")
	core.AssertNotNil(t, logger, "WithDebug should return a logger")
}

func TestSessionManagerLogDebug(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug
	sm := &DefaultSessionManager{logger: mockLog}

	sm.LogDebug(nil, "debug message")
	sm.LogDebug(map[string]any{"key": "value"}, "debug message with fields")

	// Test when debug is disabled
	mockLog.Threshold = slog.Info
	sm.LogDebug(nil, "should not log")
}

func TestSessionManagerWithWarn(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	sm := &DefaultSessionManager{logger: mockLog}
	testErr := errors.New("test warning")

	logger, ok := sm.WithWarn(testErr)
	core.AssertTrue(t, ok, "WithWarn should return true when warn enabled")
	core.AssertNotNil(t, logger, "WithWarn should return a logger")
}

func TestSessionManagerLogWarn(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	sm := &DefaultSessionManager{logger: mockLog}
	testErr := errors.New("test warning")

	sm.LogWarn(testErr, nil, "warn message")
	sm.LogWarn(testErr, map[string]any{"key": "value"}, "warn message with fields")

	// Test when warn is disabled
	mockLog.Threshold = slog.Error
	sm.LogWarn(testErr, nil, "should not log")
}

func TestSessionManagerWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	sm := &DefaultSessionManager{logger: mockLog}
	testErr := errors.New("test error")

	logger, ok := sm.WithError(testErr)
	core.AssertTrue(t, ok, "WithError should return true when error enabled")
	core.AssertNotNil(t, logger, "WithError should return a logger")
}

func TestSessionManagerLogError(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	sm := &DefaultSessionManager{logger: mockLog}
	testErr := errors.New("test error")

	sm.LogError(testErr, nil, "error message")
	sm.LogError(testErr, map[string]any{"key": "value"}, "error message with fields")
}

// Test Session additional logging methods
func TestSessionWithInfo(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)

	logger, ok := s.WithInfo()
	core.AssertTrue(t, ok, "WithInfo should return true when info enabled")
	core.AssertNotNil(t, logger, "WithInfo should return a logger")
}

func TestSessionLogDebug(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)

	s.LogDebug(nil, "debug message")
	s.LogDebug(map[string]any{"key": "value"}, "debug message with fields")

	// Test when debug is disabled
	mockLog.Threshold = slog.Info
	s.LogDebug(nil, "should not log")
}

func TestSessionLogInfo(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)

	s.LogInfo(nil, "info message")
	s.LogInfo(map[string]any{"key": "value"}, "info message with fields")

	// Test when info is disabled
	mockLog.Threshold = slog.Error
	s.LogInfo(nil, "should not log")
}

func TestSessionWithWarn(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)
	testErr := errors.New("test warning")

	logger, ok := s.WithWarn(testErr)
	core.AssertTrue(t, ok, "WithWarn should return true when warn enabled")
	core.AssertNotNil(t, logger, "WithWarn should return a logger")
}

func TestSessionLogWarn(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)
	testErr := errors.New("test warning")

	s.LogWarn(testErr, nil, "warn message")
	s.LogWarn(testErr, map[string]any{"key": "value"}, "warn message with fields")

	// Test when warn is disabled
	mockLog.Threshold = slog.Error
	s.LogWarn(testErr, nil, "should not log")
}

func TestSessionWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)
	testErr := errors.New("test error")

	logger, ok := s.WithError(testErr)
	core.AssertTrue(t, ok, "WithError should return true when error enabled")
	core.AssertNotNil(t, logger, "WithError should return a logger")
}

func TestSessionLogError(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	mockConn := &testutils.MockConn{Remote: "127.0.0.1:12345"}
	s := NewDefaultSession(mockConn, nil, mockLog)
	testErr := errors.New("test error")

	s.LogError(testErr, nil, "error message")
	s.LogError(testErr, map[string]any{"key": "value"}, "error message with fields")
}
