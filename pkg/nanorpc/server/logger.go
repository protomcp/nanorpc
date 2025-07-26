package server

import (
	"net"

	"darvaza.org/slog"

	"protomcp.org/nanorpc/pkg/nanorpc/common"
)

// Server logging methods

// WithDebug returns an annotated debug-level logger
func (s *Server) WithDebug() (slog.Logger, bool) {
	logger := s.getLogger()
	if debug, ok := logger.Debug().WithEnabled(); ok {
		return debug, true
	}
	return nil, false
}

// LogDebug writes a log entry at debug-level.
func (s *Server) LogDebug(msg string) {
	if l, ok := s.WithDebug(); ok {
		l.Print(msg)
	}
}

// WithInfo returns an annotated info-level logger
func (s *Server) WithInfo() (slog.Logger, bool) {
	logger := s.getLogger()
	if info, ok := logger.Info().WithEnabled(); ok {
		return info, true
	}
	return nil, false
}

// LogInfo writes a log entry at info-level.
func (s *Server) LogInfo(msg string) {
	if l, ok := s.WithInfo(); ok {
		l.Print(msg)
	}
}

// WithWarn returns an annotated warn-level logger
func (s *Server) WithWarn(err error) (slog.Logger, bool) {
	logger := s.getLogger()
	if warn, ok := logger.Warn().WithEnabled(); ok {
		return common.WithError(warn, err), true
	}
	return nil, false
}

// LogWarn writes a log entry at warn-level.
func (s *Server) LogWarn(err error, msg string) {
	if l, ok := s.WithWarn(err); ok {
		l.Print(msg)
	}
}

// WithError returns an annotated error-level logger
func (s *Server) WithError(err error) (slog.Logger, bool) {
	logger := s.getLogger()
	if errorLog, ok := logger.Error().WithEnabled(); ok {
		return common.WithError(errorLog, err), true
	}
	return nil, false
}

// LogError writes a log entry at error-level.
func (s *Server) LogError(err error, msg string) {
	if l, ok := s.WithError(err); ok {
		l.Print(msg)
	}
}

// SessionManager logging methods

// WithDebug returns an annotated debug-level logger
func (sm *DefaultSessionManager) WithDebug() (slog.Logger, bool) {
	logger := sm.getLogger()
	if debug, ok := logger.Debug().WithEnabled(); ok {
		return debug, true
	}
	return nil, false
}

// LogDebug writes a log entry at debug-level.
func (sm *DefaultSessionManager) LogDebug(msg string) {
	if l, ok := sm.WithDebug(); ok {
		l.Print(msg)
	}
}

// WithInfo returns an annotated info-level logger
func (sm *DefaultSessionManager) WithInfo() (slog.Logger, bool) {
	logger := sm.getLogger()
	if info, ok := logger.Info().WithEnabled(); ok {
		return info, true
	}
	return nil, false
}

// LogInfo writes a log entry at info-level.
func (sm *DefaultSessionManager) LogInfo(msg string) {
	if l, ok := sm.WithInfo(); ok {
		l.Print(msg)
	}
}

// WithWarn returns an annotated warn-level logger
func (sm *DefaultSessionManager) WithWarn(err error) (slog.Logger, bool) {
	logger := sm.getLogger()
	if warn, ok := logger.Warn().WithEnabled(); ok {
		return common.WithError(warn, err), true
	}
	return nil, false
}

// LogWarn writes a log entry at warn-level.
func (sm *DefaultSessionManager) LogWarn(err error, msg string) {
	if l, ok := sm.WithWarn(err); ok {
		l.Print(msg)
	}
}

// WithError returns an annotated error-level logger
func (sm *DefaultSessionManager) WithError(err error) (slog.Logger, bool) {
	logger := sm.getLogger()
	if errorLog, ok := logger.Error().WithEnabled(); ok {
		return common.WithError(errorLog, err), true
	}
	return nil, false
}

// LogError writes a log entry at error-level.
func (sm *DefaultSessionManager) LogError(err error, msg string) {
	if l, ok := sm.WithError(err); ok {
		l.Print(msg)
	}
}

// Session logging methods

// WithDebug returns an annotated debug-level logger
func (s *DefaultSession) WithDebug() (slog.Logger, bool) {
	logger := s.getLogger()
	if debug, ok := logger.Debug().WithEnabled(); ok {
		return debug, true
	}
	return nil, false
}

// LogDebug writes a log entry at debug-level.
func (s *DefaultSession) LogDebug(msg string) {
	if l, ok := s.WithDebug(); ok {
		l.Print(msg)
	}
}

// WithInfo returns an annotated info-level logger
func (s *DefaultSession) WithInfo() (slog.Logger, bool) {
	logger := s.getLogger()
	if info, ok := logger.Info().WithEnabled(); ok {
		return info, true
	}
	return nil, false
}

// LogInfo writes a log entry at info-level.
func (s *DefaultSession) LogInfo(msg string) {
	if l, ok := s.WithInfo(); ok {
		l.Print(msg)
	}
}

// WithWarn returns an annotated warn-level logger
func (s *DefaultSession) WithWarn(err error) (slog.Logger, bool) {
	logger := s.getLogger()
	if warn, ok := logger.Warn().WithEnabled(); ok {
		return common.WithError(warn, err), true
	}
	return nil, false
}

// LogWarn writes a log entry at warn-level.
func (s *DefaultSession) LogWarn(err error, msg string) {
	if l, ok := s.WithWarn(err); ok {
		l.Print(msg)
	}
}

// WithError returns an annotated error-level logger
func (s *DefaultSession) WithError(err error) (slog.Logger, bool) {
	logger := s.getLogger()
	if errorLog, ok := logger.Error().WithEnabled(); ok {
		return common.WithError(errorLog, err), true
	}
	return nil, false
}

// LogError writes a log entry at error-level.
func (s *DefaultSession) LogError(err error, msg string) {
	if l, ok := s.WithError(err); ok {
		l.Print(msg)
	}
}

// acceptLoop logging helpers

// logAccept logs successful connection acceptance
func (s *Server) logAccept(conn net.Conn) {
	if l, ok := s.WithDebug(); ok {
		l = common.WithConnAddrs(l, conn)
		l.Print("connection accepted")
	}
}
