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
	core.AssertNotNil(t, result, "logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if remoteAddr, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, FieldRemoteAddr, "remote_addr"); ok {
			core.AssertEqual(t, "127.0.0.1:8080", remoteAddr, "remote_addr")
		}
	}

	// Test with nil logger
	result = WithRemoteAddr(nil, addr)
	core.AssertNil(t, result, "nil logger result")

	// Test with nil addr
	result = WithRemoteAddr(mockLog, nil)
	core.AssertEqual[slog.Logger](t, mockLog, result, "original logger")
}

func TestWithLocalAddr(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	addr := mockAddr{network: "tcp", address: "192.168.1.1:9090"}

	result := WithLocalAddr(mockLog, addr)
	core.AssertNotNil(t, result, "logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if localAddr, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, FieldLocalAddr, "local_addr"); ok {
			core.AssertEqual(t, "192.168.1.1:9090", localAddr, "local_addr")
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
	core.AssertNotNil(t, result, "logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		// Check remote address
		if remoteAddr, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, FieldRemoteAddr, "remote_addr"); ok {
			core.AssertEqual(t, "192.168.1.1:9090", remoteAddr, "remote_addr")
		}

		// Check local address
		if localAddr, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, FieldLocalAddr, "local_addr"); ok {
			core.AssertEqual(t, "127.0.0.1:8080", localAddr, "local_addr")
		}
	}

	// Test with nil connection
	result = WithConnAddrs(mockLog, nil)
	core.AssertEqual[slog.Logger](t, mockLog, result, "original logger")
}

func TestWithComponent(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	component := ComponentServer

	result := WithComponent(mockLog, component)
	core.AssertNotNil(t, result, "logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if comp, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, FieldComponent, "component"); ok {
			core.AssertEqual(t, ComponentServer, comp, "component")
		}
	}

	// Test with nil logger
	result = WithComponent(nil, component)
	core.AssertNil(t, result, "nil logger result")
}

func TestWithSessionID(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	sessionID := "session-123"

	result := WithSessionID(mockLog, sessionID)
	core.AssertNotNil(t, result, "logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if sid, ok := testutils.AssertFieldTypeIs[string](t, ml.Fields, FieldSessionID, "session_id"); ok {
			core.AssertEqual(t, "session-123", sid, "session_id")
		}
	}
}

func TestWithError(t *testing.T) {
	mockLog := testutils.NewMockFieldLogger()
	testErr := errors.New("test error")

	result := WithError(mockLog, testErr)
	core.AssertNotNil(t, result, "logger")

	if ml, ok := result.(*testutils.MockFieldLogger); ok {
		if err, ok := testutils.AssertFieldTypeIs[error](t, ml.Fields, FieldError, "error"); ok {
			core.AssertEqual(t, testErr, err, "error")
		}
	}

	// Test with nil error
	result = WithError(mockLog, nil)
	core.AssertEqual[slog.Logger](t, mockLog, result, "original logger")

	// Test with nil logger
	result = WithError(nil, testErr)
	core.AssertNil(t, result, "nil logger result")
}
