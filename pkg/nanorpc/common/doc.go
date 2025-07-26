// Package common provides shared constants, types, and utilities used by both
// the nanorpc client and server implementations.
//
// # Field Constants
//
// The package defines standard field names for structured logging to ensure
// consistency across client and server logs:
//
//	logger.WithField(common.FieldComponent, common.ComponentClient).
//		WithField(common.FieldSessionID, sessionID).
//		Info("session created")
//
// # Component Names
//
// Standard component names help with log filtering and monitoring:
//
//	// Filter logs by component
//	logger.WithField(common.FieldComponent, common.ComponentServer)
//
// # State Constants
//
// Connection state constants provide consistent state tracking:
//
//	logger.WithField(common.FieldState, common.StateConnected)
//
// # Sub-packages
//
// The testutils subpackage provides testing utilities including mock loggers
// and helper functions for testing nanorpc components.
package common
