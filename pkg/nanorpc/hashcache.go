package nanorpc

import (
	"errors"
	"hash/fnv"
	"sync"

	"darvaza.org/core"
)

// HashCache stores and computes path_hash values
// for [NanoRPCRequest]s.
type HashCache struct {
	path map[uint32]string
	hash map[string]uint32
	mu   sync.RWMutex
}

// Hash returns the path_hash for a given path,
// and stores it if new. Returns an error if a hash collision is detected.
func (hc *HashCache) Hash(path string) (uint32, error) {
	if v, ok := hc.getHash(path); ok {
		return v, nil
	}
	return hc.computeHash(path)
}

// Path returns a known path for a given path_hash.
func (hc *HashCache) Path(value uint32) (string, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	s, ok := hc.path[value]
	return s, ok
}

func (hc *HashCache) getHash(path string) (uint32, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	v, ok := hc.hash[path]
	return v, ok
}

func (hc *HashCache) computeHash(path string) (uint32, error) {
	h := fnv.New32a()
	n, err := h.Write([]byte(path))

	switch {
	case n == len(path):
		hc.mu.Lock()
		defer hc.mu.Unlock()

		if hc.hash == nil {
			hc.hash = make(map[string]uint32)
			hc.path = make(map[uint32]string)
		}

		value := h.Sum32()

		// Check for hash collision
		if existingPath, exists := hc.path[value]; exists && existingPath != path {
			// Hash collision detected
			return 0, core.Wrapf(ErrHashCollision,
				"paths %q and %q both hash to 0x%08x",
				existingPath, path, value)
		}

		hc.hash[path] = value
		hc.path[value] = path
		return value, nil
	case err == nil:
		err = errors.New("failed to write to fnv-1a hasher")
	}

	return 0, err
}

// DehashRequest attempts to convert path_hash in a [NanoRPCRequest]
// into a string path.
func (hc *HashCache) DehashRequest(r *NanoRPCRequest) (*NanoRPCRequest, bool) {
	if r != nil {
		switch p := r.PathOneof.(type) {
		case *NanoRPCRequest_PathHash:
			path, ok := hc.Path(p.PathHash)
			if ok {
				// known hash, replace with string
				r.PathOneof = &NanoRPCRequest_Path{
					Path: path,
				}
				return r, true
			}
		case *NanoRPCRequest_Path:
			// already string path
			return r, true
		}
	}

	return r, false
}
