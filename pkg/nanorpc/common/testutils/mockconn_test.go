package testutils

import (
	"net"
	"testing"
	"time"
)

// TestMockConn_Read tests the Read method
func TestMockConn_Read(t *testing.T) {
	// Test normal read operation
	conn := &MockConn{
		Data: []byte("hello world"),
	}

	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	AssertNoError(t, err, "Read should not fail")
	AssertEqual(t, 5, n, "should read 5 bytes")
	AssertEqual(t, "hello", string(buf), "should read correct data")
	AssertEqual(t, 5, conn.ReadPos, "ReadPos should advance")

	// Test reading remaining data
	buf = make([]byte, 10)
	n, err = conn.Read(buf)
	AssertNoError(t, err, "Read should not fail")
	AssertEqual(t, 6, n, "should read remaining 6 bytes")
	AssertEqual(t, " world", string(buf[:n]), "should read remaining data")
	AssertEqual(t, 11, conn.ReadPos, "ReadPos should be at end")

	// Test EOF behaviour
	n, err = conn.Read(buf)
	AssertNoError(t, err, "Read at EOF should not error")
	AssertEqual(t, 0, n, "should read 0 bytes at EOF")
}

// TestMockConn_ReadClosed tests reading from closed connection
func TestMockConn_ReadClosed(t *testing.T) {
	conn := &MockConn{
		Data:   []byte("test"),
		Closed: true,
	}

	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	AssertEqual(t, 0, n, "should read 0 bytes from closed conn")
	AssertError(t, err, "should return error for closed conn")
	AssertEqual(t, net.ErrClosed, err, "should return net.ErrClosed")
}

// TestMockConn_Write tests the Write method
func TestMockConn_Write(t *testing.T) {
	conn := &MockConn{}

	// Test first write
	data1 := []byte("hello")
	n, err := conn.Write(data1)
	AssertNoError(t, err, "Write should not fail")
	AssertEqual(t, 5, n, "should write 5 bytes")
	AssertEqual(t, []byte("hello"), conn.WriteData, "WriteData should contain written data")

	// Test second write (should append)
	data2 := []byte(" world")
	n, err = conn.Write(data2)
	AssertNoError(t, err, "Write should not fail")
	AssertEqual(t, 6, n, "should write 6 bytes")
	AssertEqual(t, []byte("hello world"), conn.WriteData, "WriteData should contain all written data")
}

// TestMockConn_WriteClosed tests writing to closed connection
func TestMockConn_WriteClosed(t *testing.T) {
	conn := &MockConn{
		Closed: true,
	}

	data := []byte("test")
	n, err := conn.Write(data)
	AssertEqual(t, 0, n, "should write 0 bytes to closed conn")
	AssertError(t, err, "should return error for closed conn")
	AssertEqual(t, net.ErrClosed, err, "should return net.ErrClosed")
}

// TestMockConn_Close tests the Close method
func TestMockConn_Close(t *testing.T) {
	conn := &MockConn{}
	AssertFalse(t, conn.Closed, "connection should start open")

	err := conn.Close()
	AssertNoError(t, err, "Close should not fail")
	AssertTrue(t, conn.Closed, "connection should be closed")

	// Test double close
	err = conn.Close()
	AssertNoError(t, err, "double Close should not fail")
	AssertTrue(t, conn.Closed, "connection should remain closed")
}

// TestMockConn_Addresses tests LocalAddr and RemoteAddr
func TestMockConn_Addresses(t *testing.T) {
	conn := &MockConn{
		Local:  "127.0.0.1:8080",
		Remote: "192.168.1.1:12345",
	}

	localAddr := conn.LocalAddr()
	AssertNotNil(t, localAddr, "LocalAddr should not be nil")
	AssertEqual(t, "127.0.0.1:8080", localAddr.String(), "LocalAddr should match")
	AssertEqual(t, "tcp", localAddr.Network(), "LocalAddr should use tcp network")

	remoteAddr := conn.RemoteAddr()
	AssertNotNil(t, remoteAddr, "RemoteAddr should not be nil")
	AssertEqual(t, "192.168.1.1:12345", remoteAddr.String(), "RemoteAddr should match")
	AssertEqual(t, "tcp", remoteAddr.Network(), "RemoteAddr should use tcp network")
}

// TestMockConn_Deadlines tests deadline methods (should be no-ops)
func TestMockConn_Deadlines(t *testing.T) {
	conn := &MockConn{}
	deadline := time.Now().Add(time.Hour)

	err := conn.SetDeadline(deadline)
	AssertNoError(t, err, "SetDeadline should not fail")

	err = conn.SetReadDeadline(deadline)
	AssertNoError(t, err, "SetReadDeadline should not fail")

	err = conn.SetWriteDeadline(deadline)
	AssertNoError(t, err, "SetWriteDeadline should not fail")
}

// TestMockConn_Interface tests that MockConn implements net.Conn
func TestMockConn_Interface(t *testing.T) {
	var conn net.Conn = &MockConn{}
	AssertNotNil(t, conn, "MockConn should implement net.Conn")
}

// TestMockAddr tests MockAddr functionality
func TestMockAddr(t *testing.T) {
	addr := &MockAddr{Addr: "127.0.0.1:8080"}

	AssertEqual(t, "127.0.0.1:8080", addr.String(), "String should return address")
	AssertEqual(t, "tcp", addr.Network(), "Network should return tcp")

	// Test that MockAddr implements net.Addr
	var netAddr net.Addr = addr
	AssertNotNil(t, netAddr, "MockAddr should implement net.Addr")
}

// TestMockConn_ReadWriteIntegration tests read/write integration
func TestMockConn_ReadWriteIntegration(t *testing.T) {
	conn := &MockConn{}

	// Write some data
	writeData := []byte("integration test data")
	n, err := conn.Write(writeData)
	AssertNoError(t, err, "Write should not fail")
	AssertEqual(t, len(writeData), n, "should write all data")

	// Verify written data is in WriteData buffer
	AssertEqual(t, writeData, conn.WriteData, "WriteData should contain written data")

	// Set up read data separately
	conn.Data = []byte("read test data")
	conn.ReadPos = 0

	// Read some data
	readBuf := make([]byte, 14)
	n, err = conn.Read(readBuf)
	AssertNoError(t, err, "Read should not fail")
	AssertEqual(t, 14, n, "should read 14 bytes")
	AssertEqual(t, "read test data", string(readBuf), "should read correct data")

	// Verify read and write operations are independent
	AssertEqual(t, writeData, conn.WriteData, "WriteData should be unchanged by read")
	AssertEqual(t, 14, conn.ReadPos, "ReadPos should advance correctly")
}
