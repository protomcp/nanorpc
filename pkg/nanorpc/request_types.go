package nanorpc

// This file contains type aliases and helper functions for working with NanoRPC protocol types.
// These are shared between client and server packages to handle path variants in requests.

// PathOneOf is a type alias for the internal isNanoRPCRequest_PathOneof interface.
// This interface represents the oneof field in NanoRPCRequest that can contain
// either a string path or a uint32 path hash. It is used by both client and
// server packages for flexible path handling.
type PathOneOf = isNanoRPCRequest_PathOneof

// GetPathOneOfString creates a PathOneOf containing a string path.
// This is used when the full path string should be sent in the request.
func GetPathOneOfString(path string) PathOneOf {
	return &NanoRPCRequest_Path{Path: path}
}

// GetPathOneOfHash creates a PathOneOf containing a path hash.
// This is used when only the 32-bit FNV-1a hash of the path should be sent,
// which is more efficient for embedded systems with limited bandwidth.
func GetPathOneOfHash(hash uint32) PathOneOf {
	return &NanoRPCRequest_PathHash{PathHash: hash}
}

// AsPathOneOfString extracts the string path from a PathOneOf if it contains one.
// Returns the path string and true if the PathOneOf contains a string path,
// or an empty string and false if it contains a hash or is nil.
func AsPathOneOfString(p PathOneOf) (string, bool) {
	ps, ok := p.(*NanoRPCRequest_Path)
	if ok {
		return ps.Path, ps.Path != ""
	}
	return "", false
}

// AsPathOneOfHash extracts the path hash from a PathOneOf if it contains one.
// Returns the hash value and true if the PathOneOf contains a path hash,
// or 0 and false if it contains a string path or is nil.
func AsPathOneOfHash(p PathOneOf) (uint32, bool) {
	ph, ok := p.(*NanoRPCRequest_PathHash)
	if ok {
		return ph.PathHash, ph.PathHash != 0
	}
	return 0, ok
}
