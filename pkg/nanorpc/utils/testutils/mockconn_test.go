package testutils

import (
	"net"
	"testing"
	"time"

	"darvaza.org/core"
)

// TestMockConn_Read tests the Read method
func TestMockConn_Read(t *testing.T) {
	// Test normal read operation
	conn := &MockConn{
		Data: []byte("hello world"),
	}

	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	core.AssertNoError(t, err, "Read should not fail")
	core.AssertEqual(t, 5, n, "bytes read")
	core.AssertEqual(t, "hello", string(buf), "data")
	core.AssertEqual(t, 5, conn.ReadPos, "ReadPos should advance")

	// Test reading remaining data
	buf = make([]byte, 10)
	n, err = conn.Read(buf)
	core.AssertNoError(t, err, "Read should not fail")
	core.AssertEqual(t, 6, n, "remaining bytes")
	core.AssertEqual(t, " world", string(buf[:n]), "remaining data")
	core.AssertEqual(t, 11, conn.ReadPos, "ReadPos should be at end")

	// Test EOF behaviour
	n, err = conn.Read(buf)
	core.AssertNoError(t, err, "Read at EOF should not error")
	core.AssertEqual(t, 0, n, "EOF bytes")
}

// TestMockConn_ReadClosed tests reading from closed connection
func TestMockConn_ReadClosed(t *testing.T) {
	conn := &MockConn{
		Data:   []byte("test"),
		Closed: true,
	}

	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	core.AssertEqual(t, 0, n, "closed conn bytes")
	core.AssertError(t, err, "closed conn error")
	core.AssertEqual(t, net.ErrClosed, err, "error type")
}

// TestMockConn_Write tests the Write method
func TestMockConn_Write(t *testing.T) {
	conn := &MockConn{}

	// Test first write
	data1 := []byte("hello")
	n, err := conn.Write(data1)
	core.AssertNoError(t, err, "Write should not fail")
	core.AssertEqual(t, 5, n, "bytes written")
	core.AssertSliceEqual(t, []byte("hello"), conn.WriteData, "WriteData should contain written data")

	// Test second write (should append)
	data2 := []byte(" world")
	n, err = conn.Write(data2)
	core.AssertNoError(t, err, "Write should not fail")
	core.AssertEqual(t, 6, n, "bytes written")
	core.AssertSliceEqual(t, []byte("hello world"), conn.WriteData, "WriteData should contain all written data")
}

// TestMockConn_WriteClosed tests writing to closed connection
func TestMockConn_WriteClosed(t *testing.T) {
	conn := &MockConn{
		Closed: true,
	}

	data := []byte("test")
	n, err := conn.Write(data)
	core.AssertEqual(t, 0, n, "closed conn bytes")
	core.AssertError(t, err, "closed conn error")
	core.AssertEqual(t, net.ErrClosed, err, "error type")
}

// TestMockConn_Close tests the Close method
func TestMockConn_Close(t *testing.T) {
	conn := &MockConn{}
	core.AssertFalse(t, conn.Closed, "connection should start open")

	err := conn.Close()
	core.AssertNoError(t, err, "Close should not fail")
	core.AssertTrue(t, conn.Closed, "connection should be closed")

	// Test double close
	err = conn.Close()
	core.AssertNoError(t, err, "double Close should not fail")
	core.AssertTrue(t, conn.Closed, "connection should remain closed")
}

// TestMockConn_Addresses tests LocalAddr and RemoteAddr
func TestMockConn_Addresses(t *testing.T) {
	conn := &MockConn{
		Local:  "127.0.0.1:8080",
		Remote: "192.168.1.1:12345",
	}

	localAddr := conn.LocalAddr()
	core.AssertNotNil(t, localAddr, "LocalAddr should not be nil")
	core.AssertEqual(t, "127.0.0.1:8080", localAddr.String(), "LocalAddr should match")
	core.AssertEqual(t, "tcp", localAddr.Network(), "LocalAddr should use tcp network")

	remoteAddr := conn.RemoteAddr()
	core.AssertNotNil(t, remoteAddr, "RemoteAddr should not be nil")
	core.AssertEqual(t, "192.168.1.1:12345", remoteAddr.String(), "RemoteAddr should match")
	core.AssertEqual(t, "tcp", remoteAddr.Network(), "RemoteAddr should use tcp network")
}

// TestMockConn_Deadlines tests deadline methods (should be no-ops)
func TestMockConn_Deadlines(t *testing.T) {
	conn := &MockConn{}
	deadline := time.Now().Add(time.Hour)

	err := conn.SetDeadline(deadline)
	core.AssertNoError(t, err, "SetDeadline should not fail")

	err = conn.SetReadDeadline(deadline)
	core.AssertNoError(t, err, "SetReadDeadline should not fail")

	err = conn.SetWriteDeadline(deadline)
	core.AssertNoError(t, err, "SetWriteDeadline should not fail")
}

// TestMockConn_Interface tests that MockConn implements net.Conn
func TestMockConn_Interface(t *testing.T) {
	var conn net.Conn = &MockConn{}
	core.AssertNotNil(t, conn, "MockConn should implement net.Conn")
}

// TestMockAddr tests MockAddr functionality
func TestMockAddr(t *testing.T) {
	addr := &MockAddr{Addr: "127.0.0.1:8080"}

	core.AssertEqual(t, "127.0.0.1:8080", addr.String(), "String should return address")
	core.AssertEqual(t, "tcp", addr.Network(), "Network should return tcp")

	// Test that MockAddr implements net.Addr
	var netAddr net.Addr = addr
	core.AssertNotNil(t, netAddr, "MockAddr should implement net.Addr")
}

// TestMockConn_ReadWriteIntegration tests read/write integration
func TestMockConn_ReadWriteIntegration(t *testing.T) {
	conn := &MockConn{}

	// Write some data
	writeData := []byte("integration test data")
	n, err := conn.Write(writeData)
	core.AssertNoError(t, err, "Write should not fail")
	core.AssertEqual(t, len(writeData), n, "all data written")

	// Verify written data is in WriteData buffer
	core.AssertSliceEqual(t, writeData, conn.WriteData, "WriteData should contain written data")

	// Set up read data separately
	conn.Data = []byte("read test data")
	conn.ReadPos = 0

	// Read some data
	readBuf := make([]byte, 14)
	n, err = conn.Read(readBuf)
	core.AssertNoError(t, err, "Read should not fail")
	core.AssertEqual(t, 14, n, "bytes read")
	core.AssertEqual(t, "read test data", string(readBuf), "data")

	// Verify read and write operations are independent
	core.AssertSliceEqual(t, writeData, conn.WriteData, "WriteData should be unchanged by read")
	core.AssertEqual(t, 14, conn.ReadPos, "ReadPos should advance correctly")
}
