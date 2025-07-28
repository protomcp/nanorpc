// Package common provides shared utilities for the nanorpc package.
package common

// ClearSlice zeros all elements in a slice and returns an empty slice
// that reuses the same underlying array. This prevents memory leaks
// when truncating slices containing pointers or other reference types.
//
// Example:
//
//	responses := []Response{{...}, {...}, {...}}
//	responses = ClearSlice(responses)  // All elements zeroed, length is 0
func ClearSlice[T any](s []T) []T {
	var zero T
	for i := range s {
		s[i] = zero
	}
	return s[:0]
}

// ClearAndNilSlice zeros all elements in a slice and returns nil.
// This completely releases the underlying array for garbage collection.
//
// Example:
//
//	responses := []Response{{...}, {...}, {...}}
//	responses = ClearAndNilSlice(responses)  // All elements zeroed, slice is nil
func ClearAndNilSlice[T any](s []T) []T {
	var zero T
	for i := range s {
		s[i] = zero
	}
	return nil
}
