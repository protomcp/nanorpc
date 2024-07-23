package nanorpc

import (
	"net"

	"darvaza.org/slog"
)

const (
	// Subsystem is the label used when logging the name of
	// the component making the log entries
	Subsystem = "subsystem"

	// ClientSubsystem is the value used for the [Subsystem] field
	// when logging happens from the [Client].
	ClientSubsystem = "nanorpc-client"
	// ClientSessionSubsystem is the value used for the [Subsystem] field
	// when logging happens from the [ClientSession].
	ClientSessionSubsystem = "nanorpc-client-session"
)

// WithDebug returns an annotated debug-level logger
func (c *Client) WithDebug(addr net.Addr) (slog.Logger, bool) {
	if l, ok := c.rc.WithDebug(addr); ok {
		l = l.WithField(Subsystem, ClientSubsystem)
		return l, true
	}

	return nil, false
}

// LogDebug writes a formatted log entry at debug-level.
func (c *Client) LogDebug(addr net.Addr, format string, args ...any) {
	if l, ok := c.WithDebug(addr); ok {
		l.Printf(format, args...)
	}
}

// WithInfo returns an annotated info-level logger
func (c *Client) WithInfo(addr net.Addr) (slog.Logger, bool) {
	if l, ok := c.rc.WithInfo(addr); ok {
		l = l.WithField(Subsystem, ClientSubsystem)
		return l, true
	}

	return nil, false
}

// LogInfo writes a formatted log entry at info-level.
func (c *Client) LogInfo(addr net.Addr, format string, args ...any) {
	if l, ok := c.WithInfo(addr); ok {
		l.Printf(format, args...)
	}
}

// WithError returns an annotated error-level logger
func (c *Client) WithError(addr net.Addr, err error) (slog.Logger, bool) {
	if l, ok := c.rc.WithError(addr, err); ok {
		l = l.WithField(Subsystem, ClientSubsystem)
		return l, true
	}

	return nil, false
}

// LogError writes a formatted log entry at error-level.
func (c *Client) LogError(addr net.Addr, err error, format string, args ...any) {
	if l, ok := c.WithError(addr, err); ok {
		l.Printf(format, args...)
	}
}

// WithDebug returns an annotated debug-level logger
func (cs *ClientSession) WithDebug() (slog.Logger, bool) {
	if l, ok := cs.rc.WithDebug(cs.ra); ok {
		l = l.WithField(Subsystem, ClientSessionSubsystem)
		return l, true
	}

	return nil, false
}

// LogDebug writes a formatted log entry at debug-level.
func (cs *ClientSession) LogDebug(format string, args ...any) {
	if l, ok := cs.WithDebug(); ok {
		l.Printf(format, args...)
	}
}

// WithInfo returns an annotated info-level logger
func (cs *ClientSession) WithInfo() (slog.Logger, bool) {
	if l, ok := cs.rc.WithInfo(cs.ra); ok {
		l = l.WithField(Subsystem, ClientSessionSubsystem)
		return l, true
	}

	return nil, false
}

// LogInfo writes a formatted log entry at info-level.
func (cs *ClientSession) LogInfo(format string, args ...any) {
	if l, ok := cs.WithInfo(); ok {
		l.Printf(format, args...)
	}
}

// WithError returns an annotated error-level logger
func (cs *ClientSession) WithError(err error) (slog.Logger, bool) {
	if l, ok := cs.rc.WithError(cs.ra, err); ok {
		l = l.WithField(Subsystem, ClientSubsystem)
		return l, true
	}

	return nil, false
}

// LogError writes a formatted log entry at error-level.
func (cs *ClientSession) LogError(err error, format string, args ...any) {
	if l, ok := cs.WithError(err); ok {
		l.Printf(format, args...)
	}
}
