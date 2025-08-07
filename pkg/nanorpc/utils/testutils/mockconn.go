package testutils

import (
	"net"
	"time"
)

// MockConn implements net.Conn for testing
type MockConn struct {
	// Connection addresses
	Remote string
	Local  string

	// Read/write data buffers
	Data      []byte
	WriteData []byte

	// State tracking
	ReadPos int
	Closed  bool
}

// Read implements net.Conn
func (m *MockConn) Read(b []byte) (int, error) {
	if m.Closed {
		return 0, net.ErrClosed
	}
	if m.ReadPos >= len(m.Data) {
		// Return 0 to simulate EOF for test
		return 0, nil
	}
	n := copy(b, m.Data[m.ReadPos:])
	m.ReadPos += n
	return n, nil
}

// Write implements net.Conn
func (m *MockConn) Write(b []byte) (int, error) {
	if m.Closed {
		return 0, net.ErrClosed
	}
	m.WriteData = append(m.WriteData, b...)
	return len(b), nil
}

// Close implements net.Conn
func (m *MockConn) Close() error {
	m.Closed = true
	return nil
}

// LocalAddr implements net.Conn
func (m *MockConn) LocalAddr() net.Addr {
	return &MockAddr{Addr: m.Local}
}

// RemoteAddr implements net.Conn
func (m *MockConn) RemoteAddr() net.Addr {
	return &MockAddr{Addr: m.Remote}
}

// SetDeadline implements net.Conn
func (*MockConn) SetDeadline(_ time.Time) error { return nil }

// SetReadDeadline implements net.Conn
func (*MockConn) SetReadDeadline(_ time.Time) error { return nil }

// SetWriteDeadline implements net.Conn
func (*MockConn) SetWriteDeadline(_ time.Time) error { return nil }

// MockAddr implements net.Addr for testing
type MockAddr struct {
	Addr string
}

// Network implements net.Addr
func (*MockAddr) Network() string { return "tcp" }

// String implements net.Addr
func (m *MockAddr) String() string { return m.Addr }
