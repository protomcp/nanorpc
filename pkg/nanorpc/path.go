package nanorpc

import "darvaza.org/core"

// RegisterPath pre-computes the path_hash for a given path
// into a the global cache.
func RegisterPath(path string) {
	_, err := hashCache.Hash(path)
	if err != nil {
		panic(core.NewPanicWrapf(1, err, "Hash(%q)", path))
	}
}

// DehashRequest attempts to convert path_hash in a [NanoRPCRequest]
// into a string path.
func DehashRequest(r *NanoRPCRequest) (*NanoRPCRequest, bool) {
	return hashCache.DehashRequest(r)
}

var hashCache = new(HashCache)
