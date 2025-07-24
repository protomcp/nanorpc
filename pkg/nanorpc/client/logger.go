package client

import (
	"net"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"

	"github.com/amery/nanorpc/pkg/nanorpc/common"
)

// getLogger returns the base logger for the client, creating one if needed
func (c *Client) getLogger() slog.Logger {
	if c.logger == nil {
		// Create a discard logger if none provided
		c.logger = discard.New().WithField(common.FieldComponent, common.ComponentClient)
	}
	return c.logger
}

// WithDebug returns an annotated debug-level logger
func (c *Client) WithDebug(addr net.Addr) (slog.Logger, bool) {
	logger := c.getLogger()
	if debug, ok := logger.Debug().WithEnabled(); ok {
		return debug.WithField(common.FieldRemoteAddr, addr.String()), true
	}
	return nil, false
}

// LogDebug writes a log entry at debug-level.
func (c *Client) LogDebug(addr net.Addr, msg string) {
	if l, ok := c.WithDebug(addr); ok {
		l.Print(msg)
	}
}

// WithInfo returns an annotated info-level logger
func (c *Client) WithInfo(addr net.Addr) (slog.Logger, bool) {
	logger := c.getLogger()
	if info, ok := logger.Info().WithEnabled(); ok {
		return info.WithField(common.FieldRemoteAddr, addr.String()), true
	}
	return nil, false
}

// LogInfo writes a log entry at info-level.
func (c *Client) LogInfo(addr net.Addr, msg string) {
	if l, ok := c.WithInfo(addr); ok {
		l.Print(msg)
	}
}

// WithError returns an annotated error-level logger
func (c *Client) WithError(addr net.Addr, err error) (slog.Logger, bool) {
	logger := c.getLogger()
	if errorLog, ok := logger.Error().WithEnabled(); ok {
		return errorLog.
			WithField(common.FieldRemoteAddr, addr.String()).
			WithField(common.FieldError, err), true
	}
	return nil, false
}

// getErrorLogger returns an error-level logger without address (for internal use)
func (c *Client) getErrorLogger(err error) (slog.Logger, bool) {
	logger := c.getLogger()
	if errorLog, ok := logger.Error().WithEnabled(); ok {
		return errorLog.WithField(common.FieldError, err), true
	}
	return nil, false
}

// LogError writes a log entry at error-level.
func (c *Client) LogError(addr net.Addr, err error, msg string) {
	if l, ok := c.WithError(addr, err); ok {
		l.Print(msg)
	}
}

// getLogger returns the session logger with session-specific fields
func (cs *Session) getLogger() slog.Logger {
	logger := cs.c.getLogger()
	return logger.
		WithField(common.FieldComponent, common.ComponentSession).
		WithField(common.FieldRemoteAddr, cs.ra.String())
}

// WithDebug returns an annotated debug-level logger
func (cs *Session) WithDebug() (slog.Logger, bool) {
	logger := cs.getLogger()
	if debug, ok := logger.Debug().WithEnabled(); ok {
		return debug, true
	}
	return nil, false
}

// LogDebug writes a log entry at debug-level.
func (cs *Session) LogDebug(msg string) {
	if l, ok := cs.WithDebug(); ok {
		l.Print(msg)
	}
}

// WithInfo returns an annotated info-level logger
func (cs *Session) WithInfo() (slog.Logger, bool) {
	logger := cs.getLogger()
	if info, ok := logger.Info().WithEnabled(); ok {
		return info, true
	}
	return nil, false
}

// LogInfo writes a log entry at info-level.
func (cs *Session) LogInfo(msg string) {
	if l, ok := cs.WithInfo(); ok {
		l.Print(msg)
	}
}

// WithError returns an annotated error-level logger
func (cs *Session) WithError(err error) (slog.Logger, bool) {
	logger := cs.getLogger()
	if errorLog, ok := logger.Error().WithEnabled(); ok {
		return errorLog.WithField(common.FieldError, err), true
	}
	return nil, false
}

// LogError writes a log entry at error-level.
func (cs *Session) LogError(err error, msg string) {
	if l, ok := cs.WithError(err); ok {
		l.Print(msg)
	}
}
