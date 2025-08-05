// Package utils provides shared constants, types, and utilities used by both
// the nanorpc client and server implementations.
//
// # Field Constants
//
// The package defines standard field names for structured logging to ensure
// consistency across client and server logs:
//
//	logger.WithField(utils.FieldComponent, utils.ComponentClient).
//		WithField(utils.FieldSessionID, sessionID).
//		Info("session created")
//
// # Component Names
//
// Standard component names help with log filtering and monitoring:
//
//	// Filter logs by component
//	logger.WithField(utils.FieldComponent, utils.ComponentServer)
//
// # State Constants
//
// Connection state constants provide consistent state tracking:
//
//	logger.WithField(utils.FieldState, utils.StateConnected)
//
// # Slice Utilities
//
// The package provides generic slice manipulation utilities for preventing
// memory leaks when working with slices containing reference types:
//
//	// Clear a slice for reuse, preventing memory leaks
//	responses = utils.ClearSlice(responses)
//
//	// Clear and release the underlying array
//	responses = utils.ClearAndNilSlice(responses)
//
// # Sub-packages
//
// The testutils subpackage provides testing utilities including mock loggers
// and helper functions for testing nanorpc components.
package utils
