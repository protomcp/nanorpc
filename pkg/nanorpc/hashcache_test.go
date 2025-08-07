package nanorpc

import (
	"errors"
	"hash/fnv"
	"strings"
	"testing"

	"darvaza.org/core"
)

// HashTestCase represents a test case for hash computation
type HashTestCase struct {
	name string
	path string
}

func (tc HashTestCase) Name() string {
	return tc.name
}

func (tc HashTestCase) Test(t *testing.T) {
	t.Helper()
	hc := &HashCache{}

	// Test first hash computation
	hash1, err := hc.Hash(tc.path)
	core.AssertNil(t, err, "hash for path %s", tc.path)
	core.AssertNotEqual(t, uint32(0), hash1, "hash non-zero for path %s", tc.path)

	// Test cached hash retrieval (should be same)
	hash2, err := hc.Hash(tc.path)
	core.AssertNil(t, err, "cached hash")
	core.AssertEqual(t, hash1, hash2, "hash consistency")

	// Test reverse lookup
	retrievedPath, ok := hc.Path(hash1)
	core.AssertTrue(t, ok, "path retrieval")
	core.AssertEqual(t, tc.path, retrievedPath, "path")
}

var hashTestCases = []HashTestCase{
	{name: "simple_path", path: "/test"},
	{name: "nested_path", path: "/test/nested/path"},
	{name: "empty_path", path: ""},
	{name: "path_with_params", path: "/api/v1/users?id=123"},
	{name: "unicode_path", path: "/測試/naïve"},
}

func TestHashCache_Hash(t *testing.T) {
	for _, tc := range hashTestCases {
		t.Run(tc.Name(), tc.Test)
	}
}

func TestHashCache_Path(t *testing.T) {
	hc := &HashCache{}

	// Test unknown hash
	path, ok := hc.Path(12345)
	core.AssertFalse(t, ok, "unknown hash found")
	core.AssertEqual(t, "", path, "unknown hash path")

	// Test known hash
	originalPath := "/test/path"
	hash, err := hc.Hash(originalPath)
	core.AssertNil(t, err, "hash")

	retrievedPath, ok := hc.Path(hash)
	core.AssertTrue(t, ok, "known hash found")
	core.AssertEqual(t, originalPath, retrievedPath, "path")
}

// DehashRequestTestCase represents a test case for DehashRequest
type DehashRequestTestCase struct {
	request    *NanoRPCRequest
	name       string
	expectPath string
	expectOK   bool
}

func (tc DehashRequestTestCase) Name() string {
	return tc.name
}

func (tc DehashRequestTestCase) Test(t *testing.T) {
	t.Helper()
	hc := &HashCache{}

	tc.setupKnownHash(hc)
	result, ok := hc.DehashRequest(tc.request)
	core.AssertEqual(t, tc.expectOK, ok, "dehash result")

	if tc.expectOK && tc.request != nil {
		tc.verifyResult(t, result)
	}
}

func (tc DehashRequestTestCase) setupKnownHash(hc *HashCache) {
	if !tc.expectOK || tc.request == nil {
		return
	}
	if hashOneof, ok := tc.request.PathOneof.(*NanoRPCRequest_PathHash); ok {
		hash, _ := hc.Hash(tc.expectPath)
		hashOneof.PathHash = hash
	}
}

func (tc DehashRequestTestCase) verifyResult(t *testing.T, result *NanoRPCRequest) {
	core.AssertNotNil(t, result, "result")
	pathOneof, ok := core.AssertTypeIs[*NanoRPCRequest_Path](t, result.PathOneof,
		"path oneof type")
	if ok {
		core.AssertEqual(t, tc.expectPath, pathOneof.Path, "path")
	}
}

var dehashRequestTestCases = []DehashRequestTestCase{
	{
		name:       "nil_request",
		request:    nil,
		expectOK:   false,
		expectPath: "",
	},
	{
		name: "string_path_request",
		request: &NanoRPCRequest{
			PathOneof: &NanoRPCRequest_Path{
				Path: "/test/path",
			},
		},
		expectOK:   true,
		expectPath: "/test/path",
	},
	{
		name: "unknown_hash_request",
		request: &NanoRPCRequest{
			PathOneof: &NanoRPCRequest_PathHash{
				PathHash: 99999,
			},
		},
		expectOK:   false,
		expectPath: "",
	},
	{
		name: "known_hash_request",
		request: &NanoRPCRequest{
			PathOneof: &NanoRPCRequest_PathHash{
				PathHash: 0, // Will be set in test
			},
		},
		expectOK:   true,
		expectPath: "/test/known",
	},
}

func TestHashCache_DehashRequest(t *testing.T) {
	for _, tc := range dehashRequestTestCases {
		t.Run(tc.Name(), tc.Test)
	}
}

func TestHashCache_Consistency(t *testing.T) {
	hc := &HashCache{}

	// Test hash function consistency (should match fnv.New32a)
	testPath := "/test/consistency"
	cacheHash, err := hc.Hash(testPath)
	core.AssertNil(t, err, "hash")

	// Compute expected hash manually
	h := fnv.New32a()
	n, err := h.Write([]byte(testPath))
	core.AssertNoError(t, err, "fnv write")
	core.AssertEqual(t, len(testPath), n, "bytes written")
	expectedHash := h.Sum32()

	core.AssertEqual(t, expectedHash, cacheHash, "hash consistency")
}

func TestHashCache_Concurrency(t *testing.T) {
	hc := &HashCache{}

	// Test concurrent access to same path
	path := "/test/concurrent"
	numGoroutines := 50

	helper := NewConcurrentTestHelper(t, numGoroutines)
	helper.Run(func(_ int) (any, error) {
		return hc.Hash(path)
	})

	// All results should be the same
	helper.AssertNoErrors()
	results, _ := helper.GetResults()
	expectedHash, ok := GetResult[uint32](results, 0)
	core.AssertTrue(t, ok, "first result")
	for i := range results {
		hash, ok := GetResult[uint32](results, i)
		core.AssertTrue(t, ok, "result %d", i)
		core.AssertEqual(t, expectedHash, hash, "hash at index %d", i)
	}
}

func TestHashCache_DifferentPaths(t *testing.T) {
	hc := &HashCache{}

	// Test that different paths produce different hashes
	paths := []string{
		"/test/path1",
		"/test/path2",
		"/different/path",
		"/api/v1/endpoint",
		"/api/v2/endpoint",
	}

	hashes := make(map[uint32]string)
	for _, path := range paths {
		hash, err := hc.Hash(path)
		core.AssertNil(t, err, "hash for path %s", path)
		if existingPath, exists := hashes[hash]; exists {
			t.Errorf("Hash collision: path %s and %s both have hash %d", path, existingPath, hash)
		} else {
			hashes[hash] = path
		}
	}
}

func TestHashCache_EdgeCases(t *testing.T) {
	hc := &HashCache{}

	// Test empty path
	emptyHash, err := hc.Hash("")
	core.AssertNil(t, err, "empty path hash")
	core.AssertNotEqual(t, uint32(0), emptyHash, "empty path non-zero")

	// Test very long path
	longPath := "/very/long/path/that/goes/on/and/on/and/on/with/many/segments/to/test/handling/of/longer/paths"
	longHash, err := hc.Hash(longPath)
	core.AssertNil(t, err, "long path hash")
	core.AssertNotEqual(t, uint32(0), longHash, "long path non-zero")

	// Test special characters
	specialPath := "/path/with/unicode/characters"
	specialHash, err := hc.Hash(specialPath)
	core.AssertNil(t, err, "special path hash")
	core.AssertNotEqual(t, uint32(0), specialHash, "special path non-zero")

	// All should be retrievable
	retrievedPath, ok := hc.Path(emptyHash)
	core.AssertTrue(t, ok, "empty path retrieval")
	core.AssertEqual(t, "", retrievedPath, "empty path")

	retrievedPath, ok = hc.Path(longHash)
	core.AssertTrue(t, ok, "long path retrieval")
	core.AssertEqual(t, longPath, retrievedPath, "long path")

	retrievedPath, ok = hc.Path(specialHash)
	core.AssertTrue(t, ok, "special path retrieval")
	core.AssertEqual(t, specialPath, retrievedPath, "special path")
}

func TestHashCache_ResolvePath(t *testing.T) {
	hc := &HashCache{}

	t.Run("nil_request", func(t *testing.T) {
		path, hash, err := hc.ResolvePath(nil)
		core.AssertNil(t, err, "nil request")
		core.AssertEqual(t, "", path, "nil request path")
		core.AssertEqual(t, uint32(0), hash, "nil request hash")
	})

	t.Run("string_path", func(t *testing.T) {
		testPath := "/test/resolve/path"
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfString(testPath),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "string path")
		core.AssertEqual(t, testPath, path, "path")
		core.AssertNotEqual(t, uint32(0), hash, "hash non-zero")

		// Verify hash is cached
		cachedPath, ok := hc.Path(hash)
		core.AssertTrue(t, ok, "hash cached")
		core.AssertEqual(t, testPath, cachedPath, "cached path")
	})

	t.Run("known_hash_path", func(t *testing.T) {
		// First, cache a path
		testPath := "/test/known/hash"
		expectedHash, err := hc.Hash(testPath)
		core.AssertNil(t, err, "hash")

		// Now resolve using hash
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfHash(expectedHash),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "known hash")
		core.AssertEqual(t, testPath, path, "path")
		core.AssertEqual(t, expectedHash, hash, "hash")
	})

	t.Run("unknown_hash_path", func(t *testing.T) {
		unknownHash := uint32(99999999)
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfHash(unknownHash),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "unknown hash")
		core.AssertEqual(t, "", path, "unknown hash path")
		core.AssertEqual(t, unknownHash, hash, "hash")
	})

	t.Run("no_path_specified", func(t *testing.T) {
		req := &NanoRPCRequest{
			PathOneof: nil,
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "no path specified")
		core.AssertEqual(t, "", path, "path")
		core.AssertEqual(t, uint32(0), hash, "hash")
	})

	t.Run("empty_string_path", func(t *testing.T) {
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfString(""),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "empty string")
		core.AssertEqual(t, "", path, "path")

		// Verify the actual behaviour of AsPathOneOfString for empty strings
		if _, ok := AsPathOneOfString(req.PathOneof); ok {
			// If AsPathOneOfString returns true, hash should be computed
			core.AssertNotEqual(t, uint32(0), hash, "empty string hash")
		} else {
			// If AsPathOneOfString returns false, hash should be 0
			core.AssertEqual(t, uint32(0), hash, "hash zero")
		}
	})
}

// setupCollisionScenario sets up a simulated hash collision scenario
func setupCollisionScenario(t *testing.T, hc *HashCache, path1, path2 string) {
	t.Helper()

	// First, cache path1 to get its hash
	cachedHash, err := hc.Hash(path1)
	core.AssertNil(t, err, "first path hash")

	// Simulate the state where path2 would hash to the same value as path1
	// by manually setting up the collision condition
	hc.mu.Lock()
	// Replace path1 with path2 in the hash->path mapping
	hc.path[cachedHash] = path2
	// Remove path1 from path->hash mapping
	delete(hc.hash, path1)
	// Add path2 with the same hash
	hc.hash[path2] = cachedHash
	hc.mu.Unlock()
}

// testHashMethodCollision tests collision detection in Hash method
func testHashMethodCollision(t *testing.T, path1, path2 string) {
	t.Helper()

	hc := &HashCache{}
	setupCollisionScenario(t, hc, path1, path2)

	// Now when we try to hash path1, it should detect the collision
	_, err := hc.Hash(path1)
	core.AssertNotNil(t, err, "collision error")
	core.AssertTrue(t, errors.Is(err, ErrHashCollision), "error type")

	// Verify error message contains both conflicting paths
	errMsg := err.Error()
	core.AssertTrue(t, strings.Contains(errMsg, path1), "error mentions path1")
	core.AssertTrue(t, strings.Contains(errMsg, path2), "error mentions path2")
}

// testResolvePathCollision tests collision detection in ResolvePath method
func testResolvePathCollision(t *testing.T, path1, path2 string) {
	t.Helper()

	hc := &HashCache{}
	setupCollisionScenario(t, hc, path1, path2)

	// Test ResolvePath collision detection
	req := &NanoRPCRequest{
		PathOneof: GetPathOneOfString(path1),
	}
	_, _, err := hc.ResolvePath(req)
	core.AssertNotNil(t, err, "resolve collision error")
	core.AssertTrue(t, errors.Is(err, ErrHashCollision), "error type")
}

func TestHashCache_CollisionDetection(t *testing.T) {
	// Create test paths that we'll use for collision simulation
	path1 := "/test/path1"
	path2 := "/different/path"

	// Since real hash collisions are extremely rare with FNV-1a, we simulate
	// the collision scenario by setting up the cache state that would occur
	// if two different paths computed to the same hash value.
	//
	// This simulates what would happen if path2 computed to the same hash as path1:
	// 1. The hash would already exist in hc.path[hash] with path1
	// 2. When we try to store path2 with the same hash, collision detection triggers

	t.Run("hash_method_collision", func(t *testing.T) {
		testHashMethodCollision(t, path1, path2)
	})

	t.Run("resolve_path_collision", func(t *testing.T) {
		testResolvePathCollision(t, path1, path2)
	})
}
