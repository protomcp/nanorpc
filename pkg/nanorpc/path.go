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
	mu   sync.RWMutex
	path map[uint32]string
	hash map[string]uint32
}

// Hash returns the path_hash for a given path,
// and stores it if new.
func (hc *HashCache) Hash(path string) uint32 {
	if v, ok := hc.getHash(path); ok {
		return v
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

func (hc *HashCache) computeHash(path string) uint32 {
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
		hc.hash[path] = value
		hc.path[value] = path
		return value
	case err == nil:
		err = errors.New("failed to write to fnv-1a hasher")
	}

	panic(core.NewPanicError(1, err)) // reference hc.Hash
}

// RegisterPath pre-computes the path_hash for a given path
// into a the global cache.
func RegisterPath(path string) {
	hashCache.Hash(path)
}

var hashCache = new(HashCache)
