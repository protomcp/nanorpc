// Package nanorpc provides the core types and utilities for the NanoRPC protocol.
//
// This package contains:
//   - Protocol buffer definitions and generated types
//   - HashCache for efficient path hashing using FNV-1a
//   - Request/response encoding and decoding utilities
//   - Type aliases and helpers for working with protocol internals
//   - Error handling utilities
//
// The actual client and server implementations are in separate packages:
//   - Client: github.com/amery/nanorpc/pkg/nanorpc/client
//   - Server: github.com/amery/nanorpc/pkg/nanorpc/server
package nanorpc
