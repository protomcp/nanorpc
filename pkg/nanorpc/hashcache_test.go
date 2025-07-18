package nanorpc

import (
	"hash/fnv"
	"testing"
)

// HashTestCase represents a test case for hash computation
type HashTestCase struct {
	name string
	path string
}

func (tc HashTestCase) GetName() string {
	return tc.name
}

func (tc HashTestCase) test(t *testing.T) {
	t.Helper()
	hc := &HashCache{}

	// Test first hash computation
	hash1 := hc.Hash(tc.path)
	AssertNotEqual(t, uint32(0), hash1, "Hash should not be 0 for path %s", tc.path)

	// Test cached hash retrieval (should be same)
	hash2 := hc.Hash(tc.path)
	AssertEqual(t, hash1, hash2, "Hash should be consistent")

	// Test reverse lookup
	retrievedPath, ok := hc.Path(hash1)
	AssertTrue(t, ok, "Should be able to retrieve path from hash")
	AssertEqual(t, tc.path, retrievedPath, "Path mismatch")
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
		t.Run(tc.name, tc.test)
	}
}

func TestHashCache_Path(t *testing.T) {
	hc := &HashCache{}

	// Test unknown hash
	path, ok := hc.Path(12345)
	AssertFalse(t, ok, "Should not find path for unknown hash")
	AssertEqual(t, "", path, "Should return empty path for unknown hash")

	// Test known hash
	originalPath := "/test/path"
	hash := hc.Hash(originalPath)

	retrievedPath, ok := hc.Path(hash)
	AssertTrue(t, ok, "Should find path for known hash")
	AssertEqual(t, originalPath, retrievedPath, "Path mismatch")
}

// DehashRequestTestCase represents a test case for DehashRequest
type DehashRequestTestCase struct {
	request    *NanoRPCRequest
	name       string
	expectPath string
	expectOK   bool
}

func (tc DehashRequestTestCase) GetName() string {
	return tc.name
}

func (tc DehashRequestTestCase) test(t *testing.T) {
	t.Helper()
	hc := &HashCache{}

	// Setup known hash if needed
	if tc.expectOK && tc.request != nil {
		if hashOneof, ok := tc.request.PathOneof.(*NanoRPCRequest_PathHash); ok {
			// Make sure this hash is known
			hc.Hash(tc.expectPath)
			hashOneof.PathHash = hc.Hash(tc.expectPath)
		}
	}

	result, ok := hc.DehashRequest(tc.request)
	AssertEqual(t, tc.expectOK, ok, "DehashRequest ok result")

	if tc.expectOK && tc.request != nil {
		AssertNotNil(t, result, "Expected result to be non-nil")
		pathOneof := AssertTypeIs[*NanoRPCRequest_Path](t, result.PathOneof, "PathOneof should be *NanoRPCRequest_Path")
		AssertEqual(t, tc.expectPath, pathOneof.Path, "Path mismatch")
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
		t.Run(tc.name, tc.test)
	}
}

func TestHashCache_Consistency(t *testing.T) {
	hc := &HashCache{}

	// Test hash function consistency (should match fnv.New32a)
	testPath := "/test/consistency"
	cacheHash := hc.Hash(testPath)

	// Compute expected hash manually
	h := fnv.New32a()
	n, err := h.Write([]byte(testPath))
	AssertNoError(t, err, "Failed to write to fnv hasher")
	AssertEqual(t, len(testPath), n, "Expected to write all bytes")
	expectedHash := h.Sum32()

	AssertEqual(t, expectedHash, cacheHash, "Hash should match fnv.New32a output")
}

func TestHashCache_Concurrency(t *testing.T) {
	hc := &HashCache{}

	// Test concurrent access to same path
	path := "/test/concurrent"
	numGoroutines := 50

	helper := NewConcurrentTestHelper(t, numGoroutines)
	helper.Run(func(_ int) (any, error) {
		return hc.Hash(path), nil
	})

	// All results should be the same
	helper.AssertNoErrors()
	results, _ := helper.GetResults()
	expectedHash, ok := GetResult[uint32](results, 0)
	AssertTrue(t, ok, "Failed to get first result as uint32")
	for i := range results {
		hash, ok := GetResult[uint32](results, i)
		AssertTrue(t, ok, "Failed to get result %d as uint32", i)
		AssertEqual(t, expectedHash, hash, "Hash mismatch at index %d", i)
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
		hash := hc.Hash(path)
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
	emptyHash := hc.Hash("")
	AssertNotEqual(t, uint32(0), emptyHash, "Empty path should have non-zero hash")

	// Test very long path
	longPath := "/very/long/path/that/goes/on/and/on/and/on/with/many/segments/to/test/handling/of/longer/paths"
	longHash := hc.Hash(longPath)
	AssertNotEqual(t, uint32(0), longHash, "Long path should have non-zero hash")

	// Test special characters
	specialPath := "/path/with/unicode/characters"
	specialHash := hc.Hash(specialPath)
	AssertNotEqual(t, uint32(0), specialHash, "Special character path should have non-zero hash")

	// All should be retrievable
	retrievedPath, ok := hc.Path(emptyHash)
	AssertTrue(t, ok, "Empty path should be retrievable")
	AssertEqual(t, "", retrievedPath, "Empty path retrieval failed")

	retrievedPath, ok = hc.Path(longHash)
	AssertTrue(t, ok, "Long path should be retrievable")
	AssertEqual(t, longPath, retrievedPath, "Long path retrieval failed")

	retrievedPath, ok = hc.Path(specialHash)
	AssertTrue(t, ok, "Special character path should be retrievable")
	AssertEqual(t, specialPath, retrievedPath, "Special character path retrieval failed")
}
