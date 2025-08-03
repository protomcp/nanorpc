package client

import (
	"net"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"

	"protomcp.org/nanorpc/pkg/nanorpc/common/testutils"
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
	testutils.AssertNotNil(t, logger, "default logger should not be nil")

	// Verify it doesn't panic
	logger.Info().Print("test message")
}

func TestClientCustomLogger(t *testing.T) {
	customLogger := testutils.NewMockFieldLogger()
	c := newTestClient(customLogger)
	logger := c.getLogger()

	testutils.AssertEqual[slog.Logger](t, customLogger, logger, "should return custom logger")
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
	testutils.AssertTrue(t, ok, tc.name+" should return true")
	testutils.AssertNotNil(t, logger, tc.name+" should return a logger")

	// Check expected field
	ml := testutils.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "expected MockFieldLogger")
	val, ok := testutils.GetField[string, any](ml.Fields, tc.expectField)
	testutils.AssertTrue(t, ok, "logger should have %s field", tc.expectField)
	testutils.AssertEqual(t, tc.expectValue, val, "should have "+tc.expectField+" field")
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
	testutils.AssertTrue(t, ok, "WithError should return true")
	testutils.AssertNotNil(t, logger, "error logger should not be nil")

	// Check fields
	ml := testutils.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "expected MockFieldLogger")
	if addr, ok := testutils.GetField[string, string](ml.Fields, "remote_addr"); ok {
		testutils.AssertEqual(t, "10.0.0.1:443", addr, "should have remote address")
	} else {
		t.Error("logger should have remote_addr field")
	}
	if err, ok := testutils.GetField[string, error](ml.Fields, "error"); ok {
		testutils.AssertEqual(t, testErr, err, "should have error field")
	} else {
		t.Error("logger should have error field")
	}
}

func TestClientGetErrorLogger(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	testErr := core.ErrNilReceiver

	logger, ok := c.getErrorLogger(testErr)
	testutils.AssertTrue(t, ok, "getErrorLogger should return true")
	testutils.AssertNotNil(t, logger, "error logger should not be nil")

	// Check error field
	ml := testutils.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "expected MockFieldLogger")
	if err, ok := testutils.GetField[string, error](ml.Fields, "error"); ok {
		testutils.AssertEqual(t, testErr, err, "should have error field")
	} else {
		t.Error("logger should have error field")
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
		testutils.AssertTrue(t, ok, tc.name+" should be enabled")
	} else {
		testutils.AssertFalse(t, ok, tc.name+" should be disabled")
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
	testutils.AssertNotNil(t, logger, "session logger should not be nil")

	// Check fields
	ml := testutils.AssertTypeIs[*testutils.MockFieldLogger](t, logger, "expected MockFieldLogger")
	component, ok := testutils.GetField[string, string](ml.Fields, "component")
	testutils.AssertTrue(t, ok, "logger should have component field")
	testutils.AssertEqual(t, "session", component, "should have session component")
	addr, ok := testutils.GetField[string, string](ml.Fields, "remote_addr")
	testutils.AssertTrue(t, ok, "logger should have remote_addr field")
	testutils.AssertEqual(t, "10.0.0.1:8080", addr, "should have remote address field")
}

func TestSessionWithDebug(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	c := newTestClient(mockLog)
	cs := newTestSession(c, "192.168.1.100:9000")

	logger, ok := cs.WithDebug()
	testutils.AssertTrue(t, ok, "WithDebug should return true for debug level")
	testutils.AssertNotNil(t, logger, "debug logger should not be nil")
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
	testutils.AssertNotNil(t, cs, "session should not be nil")
}
