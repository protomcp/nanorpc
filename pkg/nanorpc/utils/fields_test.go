package utils

import (
	"errors"
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

func TestWithRemoteAddr(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	addr := mockAddr{network: "tcp", address: "127.0.0.1:8080"}

	// Test with valid logger and addr
	result := WithRemoteAddr(mockLog, addr)
	core.AssertNotNil(t, result, "should return a logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if remoteAddr, ok := testutils.GetField[string, string](ml.Fields, FieldRemoteAddr); ok {
			core.AssertEqual(t, "127.0.0.1:8080", remoteAddr, "should have remote address")
		} else {
			t.Error("logger should have remote_addr field")
		}
	}

	// Test with nil logger
	result = WithRemoteAddr(nil, addr)
	core.AssertNil(t, result, "should return nil for nil logger")

	// Test with nil addr
	result = WithRemoteAddr(mockLog, nil)
	core.AssertEqual[slog.Logger](t, mockLog, result, "should return original logger for nil addr")
}

func TestWithLocalAddr(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	addr := mockAddr{network: "tcp", address: "192.168.1.1:9090"}

	result := WithLocalAddr(mockLog, addr)
	core.AssertNotNil(t, result, "should return a logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if localAddr, ok := testutils.GetField[string, string](ml.Fields, FieldLocalAddr); ok {
			core.AssertEqual(t, "192.168.1.1:9090", localAddr, "should have local address")
		} else {
			t.Error("logger should have local_addr field")
		}
	}
}

func TestWithConnAddrs(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()

	// Create a mock connection
	conn := &testutils.MockConn{
		Local:  "127.0.0.1:8080",
		Remote: "192.168.1.1:9090",
	}

	result := WithConnAddrs(mockLog, conn)
	core.AssertNotNil(t, result, "should return a logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		// Check remote address
		if remoteAddr, ok := testutils.GetField[string, string](ml.Fields, FieldRemoteAddr); ok {
			core.AssertEqual(t, "192.168.1.1:9090", remoteAddr, "should have remote address")
		} else {
			t.Error("logger should have remote_addr field")
		}

		// Check local address
		if localAddr, ok := testutils.GetField[string, string](ml.Fields, FieldLocalAddr); ok {
			core.AssertEqual(t, "127.0.0.1:8080", localAddr, "should have local address")
		} else {
			t.Error("logger should have local_addr field")
		}
	}

	// Test with nil connection
	result = WithConnAddrs(mockLog, nil)
	core.AssertEqual[slog.Logger](t, mockLog, result, "should return original logger for nil conn")
}

func TestWithComponent(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	component := ComponentServer

	result := WithComponent(mockLog, component)
	core.AssertNotNil(t, result, "should return a logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if comp, ok := testutils.GetField[string, string](ml.Fields, FieldComponent); ok {
			core.AssertEqual(t, ComponentServer, comp, "should have component field")
		} else {
			t.Error("logger should have component field")
		}
	}

	// Test with nil logger
	result = WithComponent(nil, component)
	core.AssertNil(t, result, "should return nil for nil logger")
}

func TestWithSessionID(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	sessionID := "session-123"

	result := WithSessionID(mockLog, sessionID)
	core.AssertNotNil(t, result, "should return a logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if sid, ok := testutils.GetField[string, string](ml.Fields, FieldSessionID); ok {
			core.AssertEqual(t, "session-123", sid, "should have session ID")
		} else {
			t.Error("logger should have session_id field")
		}
	}
}

func TestWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	testErr := errors.New("test error")

	result := WithError(mockLog, testErr)
	core.AssertNotNil(t, result, "should return a logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if err, ok := testutils.GetField[string, error](ml.Fields, FieldError); ok {
			core.AssertEqual(t, testErr, err, "should have error field")
		} else {
			t.Error("logger should have error field")
		}
	}

	// Test with nil error
	result = WithError(mockLog, nil)
	core.AssertEqual[slog.Logger](t, mockLog, result, "should return original logger for nil error")

	// Test with nil logger
	result = WithError(nil, testErr)
	core.AssertNil(t, result, "should return nil for nil logger")
}
