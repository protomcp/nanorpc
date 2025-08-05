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
	core.AssertNil(t, err, "Hash should not error for path %s", tc.path)
	core.AssertNotEqual(t, uint32(0), hash1, "Hash should not be 0 for path %s", tc.path)

	// Test cached hash retrieval (should be same)
	hash2, err := hc.Hash(tc.path)
	core.AssertNil(t, err, "Hash should not error on second call")
	core.AssertEqual(t, hash1, hash2, "Hash should be consistent")

	// Test reverse lookup
	retrievedPath, ok := hc.Path(hash1)
	core.AssertTrue(t, ok, "Should be able to retrieve path from hash")
	core.AssertEqual(t, tc.path, retrievedPath, "Path mismatch")
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
	core.AssertFalse(t, ok, "Should not find path for unknown hash")
	core.AssertEqual(t, "", path, "Should return empty path for unknown hash")

	// Test known hash
	originalPath := "/test/path"
	hash, err := hc.Hash(originalPath)
	core.AssertNil(t, err, "Hash should not error")

	retrievedPath, ok := hc.Path(hash)
	core.AssertTrue(t, ok, "Should find path for known hash")
	core.AssertEqual(t, originalPath, retrievedPath, "Path mismatch")
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
	core.AssertEqual(t, tc.expectOK, ok, "DehashRequest ok result")

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
	core.AssertNotNil(t, result, "Expected result to be non-nil")
	pathOneof, ok := core.AssertTypeIs[*NanoRPCRequest_Path](t, result.PathOneof,
		"PathOneof should be *NanoRPCRequest_Path")
	if ok {
		core.AssertEqual(t, tc.expectPath, pathOneof.Path, "Path mismatch")
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
	core.AssertNil(t, err, "Hash should not error")

	// Compute expected hash manually
	h := fnv.New32a()
	n, err := h.Write([]byte(testPath))
	core.AssertNoError(t, err, "Failed to write to fnv hasher")
	core.AssertEqual(t, len(testPath), n, "Expected to write all bytes")
	expectedHash := h.Sum32()

	core.AssertEqual(t, expectedHash, cacheHash, "Hash should match fnv.New32a output")
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
	core.AssertTrue(t, ok, "Failed to get first result as uint32")
	for i := range results {
		hash, ok := GetResult[uint32](results, i)
		core.AssertTrue(t, ok, "Failed to get result %d as uint32", i)
		core.AssertEqual(t, expectedHash, hash, "Hash mismatch at index %d", i)
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
		core.AssertNil(t, err, "Hash should not error for path %s", path)
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
	core.AssertNil(t, err, "Hash should not error for empty path")
	core.AssertNotEqual(t, uint32(0), emptyHash, "Empty path should have non-zero hash")

	// Test very long path
	longPath := "/very/long/path/that/goes/on/and/on/and/on/with/many/segments/to/test/handling/of/longer/paths"
	longHash, err := hc.Hash(longPath)
	core.AssertNil(t, err, "Hash should not error for long path")
	core.AssertNotEqual(t, uint32(0), longHash, "Long path should have non-zero hash")

	// Test special characters
	specialPath := "/path/with/unicode/characters"
	specialHash, err := hc.Hash(specialPath)
	core.AssertNil(t, err, "Hash should not error for special path")
	core.AssertNotEqual(t, uint32(0), specialHash, "Special character path should have non-zero hash")

	// All should be retrievable
	retrievedPath, ok := hc.Path(emptyHash)
	core.AssertTrue(t, ok, "Empty path should be retrievable")
	core.AssertEqual(t, "", retrievedPath, "Empty path retrieval failed")

	retrievedPath, ok = hc.Path(longHash)
	core.AssertTrue(t, ok, "Long path should be retrievable")
	core.AssertEqual(t, longPath, retrievedPath, "Long path retrieval failed")

	retrievedPath, ok = hc.Path(specialHash)
	core.AssertTrue(t, ok, "Special character path should be retrievable")
	core.AssertEqual(t, specialPath, retrievedPath, "Special character path retrieval failed")
}

func TestHashCache_ResolvePath(t *testing.T) {
	hc := &HashCache{}

	t.Run("nil_request", func(t *testing.T) {
		path, hash, err := hc.ResolvePath(nil)
		core.AssertNil(t, err, "ResolvePath should not error for nil request")
		core.AssertEqual(t, "", path, "Path should be empty for nil request")
		core.AssertEqual(t, uint32(0), hash, "Hash should be 0 for nil request")
	})

	t.Run("string_path", func(t *testing.T) {
		testPath := "/test/resolve/path"
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfString(testPath),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "ResolvePath should not error for string path")
		core.AssertEqual(t, testPath, path, "Path should match input")
		core.AssertNotEqual(t, uint32(0), hash, "Hash should not be 0")

		// Verify hash is cached
		cachedPath, ok := hc.Path(hash)
		core.AssertTrue(t, ok, "Hash should be cached")
		core.AssertEqual(t, testPath, cachedPath, "Cached path should match")
	})

	t.Run("known_hash_path", func(t *testing.T) {
		// First, cache a path
		testPath := "/test/known/hash"
		expectedHash, err := hc.Hash(testPath)
		core.AssertNil(t, err, "Hash should not error")

		// Now resolve using hash
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfHash(expectedHash),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "ResolvePath should not error for known hash")
		core.AssertEqual(t, testPath, path, "Path should be resolved from hash")
		core.AssertEqual(t, expectedHash, hash, "Hash should match input")
	})

	t.Run("unknown_hash_path", func(t *testing.T) {
		unknownHash := uint32(99999999)
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfHash(unknownHash),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "ResolvePath should not error for unknown hash")
		core.AssertEqual(t, "", path, "Path should be empty for unknown hash")
		core.AssertEqual(t, unknownHash, hash, "Hash should match input")
	})

	t.Run("no_path_specified", func(t *testing.T) {
		req := &NanoRPCRequest{
			PathOneof: nil,
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "ResolvePath should not error when no path specified")
		core.AssertEqual(t, "", path, "Path should be empty")
		core.AssertEqual(t, uint32(0), hash, "Hash should be 0")
	})

	t.Run("empty_string_path", func(t *testing.T) {
		req := &NanoRPCRequest{
			PathOneof: GetPathOneOfString(""),
		}

		path, hash, err := hc.ResolvePath(req)
		core.AssertNil(t, err, "ResolvePath should not error for empty string")
		core.AssertEqual(t, "", path, "Path should be empty")

		// Verify the actual behaviour of AsPathOneOfString for empty strings
		if _, ok := AsPathOneOfString(req.PathOneof); ok {
			// If AsPathOneOfString returns true, hash should be computed
			core.AssertNotEqual(t, uint32(0), hash, "Hash should be computed for empty string")
		} else {
			// If AsPathOneOfString returns false, hash should be 0
			core.AssertEqual(t, uint32(0), hash, "Hash should be 0 when AsPathOneOfString returns false")
		}
	})
}

// setupCollisionScenario sets up a simulated hash collision scenario
func setupCollisionScenario(t *testing.T, hc *HashCache, path1, path2 string) {
	t.Helper()

	// First, cache path1 to get its hash
	cachedHash, err := hc.Hash(path1)
	core.AssertNil(t, err, "Hash should not error for first path")

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
	core.AssertNotNil(t, err, "Hash should error when collision detected")
	core.AssertTrue(t, errors.Is(err, ErrHashCollision), "Error should be ErrHashCollision")

	// Verify error message contains both conflicting paths
	errMsg := err.Error()
	core.AssertTrue(t, strings.Contains(errMsg, path1), "Error should mention path1: %s", errMsg)
	core.AssertTrue(t, strings.Contains(errMsg, path2), "Error should mention path2: %s", errMsg)
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
	core.AssertNotNil(t, err, "ResolvePath should error on collision")
	core.AssertTrue(t, errors.Is(err, ErrHashCollision), "ResolvePath error should be ErrHashCollision")
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
