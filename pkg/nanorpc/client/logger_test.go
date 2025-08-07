package client

import (
	"net"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"

	"protomcp.org/nanorpc/pkg/nanorpc/utils/testutils"
)

// mockAddr implements net.Addr for testing
type mockAddr struct {
	network string
	address string
}

func (m mockAddr) Network() string { return m.network }
func (m mockAddr) String() string  { return m.address }

// Helper factory functions
func newTestClient(logger slog.Logger) *Client {
	return &Client{logger: logger}
}

func newTestAddr(addr string) mockAddr {
	return mockAddr{network: "tcp", address: addr}
}

func TestClientDefaultLogger(t *testing.T) {
	c := &Client{}
	logger := c.getLogger()
	core.AssertNotNil(t, logger, "logger")

	// Verify it doesn't panic
	logger.Info().Print("test message")
}

func TestClientCustomLogger(t *testing.T) {
	customLogger := testutils.NewMockFieldLogger()
	c := newTestClient(customLogger)
	logger := c.getLogger()

	core.AssertEqual[slog.Logger](t, customLogger, logger, "logger")
}

type clientWithMethodTestCase struct {
	method      func(*Client, net.Addr) (slog.Logger, bool)
	expectValue any
	name        string
	address     string
	expectField string
}

func (tc *clientWithMethodTestCase) test(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	addr := newTestAddr(tc.address)

	logger, ok := tc.method(c, addr)
	core.AssertTrue(t, ok, tc.name+" ok")
	core.AssertNotNil(t, logger, tc.name+" logger")

	// Check expected field
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	val, ok := testutils.GetField[string, any](ml.Fields, tc.expectField)
	core.AssertTrue(t, ok, "field %s", tc.expectField)
	core.AssertEqual(t, tc.expectValue, val, tc.expectField)
}

func TestClientWithMethods(t *testing.T) {
	tests := []clientWithMethodTestCase{
		{
			name:        "WithDebug",
			address:     "127.0.0.1:8080",
			method:      (*Client).WithDebug,
			expectField: "remote_addr",
			expectValue: "127.0.0.1:8080",
		},
		{
			name:        "WithInfo",
			address:     "192.168.1.1:9090",
			method:      (*Client).WithInfo,
			expectField: "remote_addr",
			expectValue: "192.168.1.1:9090",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func TestClientWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	addr := newTestAddr("10.0.0.1:443")
	testErr := core.ErrInvalid

	logger, ok := c.WithError(addr, testErr)
	core.AssertTrue(t, ok, "WithError ok")
	core.AssertNotNil(t, logger, "logger")

	// Check fields
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	if addr, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, "remote_addr", "remote_addr"); ok {
		core.AssertEqual(t, "10.0.0.1:443", addr, "remote_addr")
	}
	if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, "error", "error"); ok {
		core.AssertEqual(t, testErr, err, "error")
	}
}

func TestClientGetErrorLogger(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	testErr := core.ErrNilReceiver

	logger, ok := c.getErrorLogger(testErr)
	core.AssertTrue(t, ok, "getErrorLogger ok")
	core.AssertNotNil(t, logger, "logger")

	// Check error field
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, "error", "error"); ok {
		core.AssertEqual(t, testErr, err, "error")
	}
}

type thresholdTestCase struct {
	method    func(*Client, net.Addr) (slog.Logger, bool)
	name      string
	threshold slog.LogLevel
	expected  bool
}

func (tc *thresholdTestCase) test(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = tc.threshold
	c := newTestClient(mockLog)
	addr := newTestAddr("127.0.0.1:8080")

	_, ok := tc.method(c, addr)
	if tc.expected {
		core.AssertTrue(t, ok, tc.name+" enabled")
	} else {
		core.AssertFalse(t, ok, tc.name+" disabled")
	}
}

func TestClientLoggingThreshold(t *testing.T) {
	tests := []thresholdTestCase{
		{
			name:      "Debug with Warn threshold",
			threshold: slog.Warn,
			method:    (*Client).WithDebug,
			expected:  false,
		},
		{
			name:      "Info with Warn threshold",
			threshold: slog.Warn,
			method:    (*Client).WithInfo,
			expected:  false,
		},
		{
			name:      "Error with Warn threshold",
			threshold: slog.Warn,
			method: func(c *Client, addr net.Addr) (slog.Logger, bool) {
				return c.WithError(addr, core.ErrInvalid)
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// Helper factory for test sessions
func newTestSession(client *Client, addr string) *Session {
	return &Session{
		c:  client,
		ra: newTestAddr(addr),
	}
}

func TestSessionGetLogger(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	cs := newTestSession(c, "10.0.0.1:8080")

	logger := cs.getLogger()
	core.AssertNotNil(t, logger, "logger")

	// Check fields
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	component, ok := testutils.GetField[string, string](ml.Fields, "component")
	core.AssertTrue(t, ok, "component field")
	core.AssertEqual(t, "session", component, "component")
	addr, ok := testutils.GetField[string, string](ml.Fields, "remote_addr")
	core.AssertTrue(t, ok, "remote_addr field")
	core.AssertEqual(t, "10.0.0.1:8080", addr, "remote_addr")
}

func TestSessionWithDebug(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	cs := newTestSession(c, "192.168.1.100:9000")

	logger, ok := cs.WithDebug()
	core.AssertTrue(t, ok, "WithDebug ok")
	core.AssertNotNil(t, logger, "logger")
}

func TestSessionLogMethods(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	cs := newTestSession(c, "localhost:7777")

	// Test LogDebug
	cs.LogDebug(nil, "debug message")

	// Test LogInfo
	cs.LogInfo(nil, "info message")

	// Test LogError
	testErr := core.ErrInvalid
	cs.LogError(testErr, nil, "error message")

	// Verify no panic
	core.AssertNotNil(t, cs, "session")
}

// Test missing Client logging methods

func TestClientLogDebug(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Debug
	c := newTestClient(mockLog)
	addr := newTestAddr("127.0.0.1:8080")

	// Test LogDebug with and without fields
	c.LogDebug(addr, nil, "debug message")
	c.LogDebug(addr, map[string]any{"key": "value"}, "debug message with fields")

	// Test when debug is disabled
	mockLog.Threshold = slog.Info
	c.LogDebug(addr, nil, "should not log")
}

func TestClientLogInfo(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	c := newTestClient(mockLog)
	addr := newTestAddr("127.0.0.1:8080")

	// Test LogInfo with and without fields
	c.LogInfo(addr, nil, "info message")
	c.LogInfo(addr, map[string]any{"key": "value"}, "info message with fields")

	// Test when info is disabled
	mockLog.Threshold = slog.Error
	c.LogInfo(addr, nil, "should not log")
}

func TestClientWithWarn(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	c := newTestClient(mockLog)
	addr := newTestAddr("127.0.0.1:8080")
	testErr := core.ErrInvalid

	logger, ok := c.WithWarn(addr, testErr)
	core.AssertTrue(t, ok, "WithWarn ok")
	core.AssertNotNil(t, logger, "logger")

	// Check fields are added
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, "error", "error"); ok {
		core.AssertEqual(t, testErr, err, "error")
	}
	if addrVal, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, "remote_addr", "remote_addr"); ok {
		core.AssertEqual(t, "127.0.0.1:8080", addrVal, "remote_addr")
	}
}

func TestClientLogWarn(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	c := newTestClient(mockLog)
	addr := newTestAddr("127.0.0.1:8080")
	testErr := core.ErrInvalid

	// Test LogWarn with and without fields
	c.LogWarn(addr, testErr, nil, "warn message")
	c.LogWarn(addr, testErr, map[string]any{"key": "value"}, "warn message with fields")

	// Test when warn is disabled
	mockLog.Threshold = slog.Error
	c.LogWarn(addr, testErr, nil, "should not log")
}

// Test Session WithWarn and LogWarn methods

func TestSessionWithWarn(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	c := newTestClient(mockLog)
	cs := newTestSession(c, "127.0.0.1:8080")
	testErr := core.ErrInvalid

	logger, ok := cs.WithWarn(testErr)
	core.AssertTrue(t, ok, "WithWarn ok")
	core.AssertNotNil(t, logger, "logger")

	// Check error field is added
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, "error", "error"); ok {
		core.AssertEqual(t, testErr, err, "error")
	}
}

func TestSessionLogWarn(_ *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Warn
	c := newTestClient(mockLog)
	cs := newTestSession(c, "127.0.0.1:8080")
	testErr := core.ErrInvalid

	// Test LogWarn with and without fields
	cs.LogWarn(testErr, nil, "warn message")
	cs.LogWarn(testErr, map[string]any{"key": "value"}, "warn message with fields")

	// Test when warn is disabled
	mockLog.Threshold = slog.Error
	cs.LogWarn(testErr, nil, "should not log")
}

// Test Session WithInfo method which was missing
func TestSessionWithInfo(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Info
	c := newTestClient(mockLog)
	cs := newTestSession(c, "127.0.0.1:8080")

	logger, ok := cs.WithInfo()
	core.AssertTrue(t, ok, "WithInfo ok")
	core.AssertNotNil(t, logger, "logger")
}

// Test Session WithError method which was missing
func TestSessionWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	mockLog.Threshold = slog.Error
	c := newTestClient(mockLog)
	cs := newTestSession(c, "127.0.0.1:8080")
	testErr := core.ErrInvalid

	logger, ok := cs.WithError(testErr)
	core.AssertTrue(t, ok, "WithError ok")
	core.AssertNotNil(t, logger, "logger")

	// Check error field is added
	ml, ok := core.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "MockFieldLogger")
	if !ok {
		return
	}
	if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, "error", "error"); ok {
		core.AssertEqual(t, testErr, err, "error")
	}
}
