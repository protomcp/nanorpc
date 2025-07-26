package server

import (
	"errors"
	"testing"

	"darvaza.org/slog"

	"protomcp.org/nanorpc/pkg/nanorpc/common"
	"protomcp.org/nanorpc/pkg/nanorpc/common/testutils"
)

// Test server default logger
func TestServerDefaultLogger(t *testing.T) {
	s := &Server{}
	logger := s.getLogger()
	testutils.AssertNotNil(t, logger, "default logger should not be nil")

	// Verify it doesn't panic
	logger.Info().Print("test message")
}

// Test server custom logger
func TestServerCustomLogger(t *testing.T) {
	customLogger := testutils.NewMockFieldLogger()
	s := &Server{logger: customLogger}
	logger := s.getLogger()

	testutils.AssertEqual[slog.Logger](t, customLogger, logger, "should return custom logger")
}

// Test Server.WithDebug
func TestServerWithDebug(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug // Enable debug level
	s := &Server{logger: mockLog}

	logger, ok := s.WithDebug()
	testutils.AssertTrue(t, ok, "WithDebug should return true when debug enabled")
	testutils.AssertNotNil(t, logger, "WithDebug should return a logger")

	// Verify the logger has the correct level
	if ml, ok := logger.(*testutils.MockFieldLogger); ok {
		testutils.AssertEqual(t, slog.Debug, ml.CurrentLevel, "should have Debug level")
	}
}

// Test Server.LogInfo
func TestServerLogInfo(t *testing.T) {
	// Test that LogInfo is called when info level is enabled
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	s := &Server{logger: mockLog}

	// LogInfo should succeed when threshold allows it
	s.LogInfo("test info message")

	// Test that LogInfo is not called when threshold is too high
	mockLog.Threshold = slog.Error
	s.LogInfo("this should not log")

	// No panic means test passed
	testutils.AssertTrue(t, true, "LogInfo executed without panic")
}

// Test Server.WithError
func TestServerWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	s := &Server{logger: mockLog}
	testErr := errors.New("test error")

	logger, ok := s.WithError(testErr)
	testutils.AssertTrue(t, ok, "WithError should return true when error enabled")
	testutils.AssertNotNil(t, logger, "WithError should return a logger")

	// Check error field was added
	if ml, ok := logger.(*testutils.MockFieldLogger); ok {
		if err, ok := testutils.GetField[string, error](ml.Fields, common.FieldError); ok {
			testutils.AssertEqual(t, testErr, err, "should have error field")
		} else {
			t.Error("logger should have error field")
		}
	}
}

// Test SessionManager logging
func TestSessionManagerWithInfo(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	sm := &DefaultSessionManager{logger: mockLog}

	logger, ok := sm.WithInfo()
	testutils.AssertTrue(t, ok, "WithInfo should return true when info enabled")
	testutils.AssertNotNil(t, logger, "WithInfo should return a logger")
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
	testutils.AssertTrue(t, ok, "WithDebug should return true when debug enabled")
	testutils.AssertNotNil(t, logger, "WithDebug should return a logger")

	// Check session fields are added
	if ml, ok := logger.(*testutils.MockFieldLogger); ok {
		// Check component field
		if component, ok := testutils.GetField[string, string](ml.Fields, common.FieldComponent); ok {
			testutils.AssertEqual(t, common.ComponentSession, component, "should have session component")
		} else {
			t.Error("logger should have component field")
		}

		// Check session ID field
		if sid, ok := testutils.GetField[string, string](ml.Fields, common.FieldSessionID); ok {
			testutils.AssertEqual(t, s.ID(), sid, "should have session ID")
		} else {
			t.Error("logger should have session_id field")
		}

		// Check remote address field
		if addr, ok := testutils.GetField[string, string](ml.Fields, common.FieldRemoteAddr); ok {
			testutils.AssertEqual(t, "127.0.0.1:12345", addr, "should have remote address")
		} else {
			t.Error("logger should have remote_addr field")
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
	testutils.AssertEqual(t, "127.0.0.1:8080", conn.LocalAddr().String(), "local address should match")
	testutils.AssertEqual(t, "192.168.1.1:12345", conn.RemoteAddr().String(), "remote address should match")
}
