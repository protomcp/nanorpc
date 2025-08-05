package client

import (
	"context"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog/handlers/discard"
	"darvaza.org/x/net/reconnect"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// ClientConfigTestCase represents a test case for client configuration
type ClientConfigTestCase struct {
	config *Config
	name   string
}

func (tc ClientConfigTestCase) GetName() string {
	return tc.name
}

func (tc ClientConfigTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.config.SetDefaults()
	core.AssertNoError(t, err, "SetDefaults failed")

	// Verify all required fields are set
	core.AssertNotNil(t, tc.config.Context, "Context should not be nil after SetDefaults")
	core.AssertNotNil(t, tc.config.Logger, "Logger should not be nil after SetDefaults")
	core.AssertNotNil(t, tc.config.HashCache, "nanorpc.HashCache should not be nil after SetDefaults")
	core.AssertNotNil(t, tc.config.WaitReconnect, "WaitReconnect should not be nil after SetDefaults")
}

var clientConfigTestCases = []ClientConfigTestCase{
	{
		name:   "empty_config",
		config: &Config{},
	},
	{
		name: "partial_config",
		config: &Config{
			Remote:      "localhost:8080",
			DialTimeout: 1 * time.Second,
		},
	},
}

func TestClientConfig_SetDefaults(t *testing.T) {
	for _, tc := range clientConfigTestCases {
		t.Run(tc.name, tc.test)
	}
}

// TestClientConfig_DefaultValues tests that default values are set correctly
func TestClientConfig_DefaultValues(t *testing.T) {
	cfg := &Config{}
	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "SetDefaults failed")

	// Check specific default values
	core.AssertEqual(t, 2*time.Second, cfg.DialTimeout, "DialTimeout default mismatch")
	core.AssertEqual(t, 2*time.Second, cfg.ReadTimeout, "ReadTimeout default mismatch")
	core.AssertEqual(t, 2*time.Second, cfg.WriteTimeout, "WriteTimeout default mismatch")
	core.AssertEqual(t, 10*time.Second, cfg.IdleTimeout, "IdleTimeout default mismatch")
	core.AssertEqual(t, 5*time.Second, cfg.ReconnectDelay, "ReconnectDelay default mismatch")
	core.AssertEqual(t, 5*time.Second, cfg.KeepAlive, "KeepAlive default mismatch")
	core.AssertEqual(t, hashCache, cfg.HashCache, "nanorpc.HashCache should be global instance")
}

// TestClientConfig_PreserveExisting tests that existing values are preserved
func TestClientConfig_PreserveExisting(t *testing.T) {
	type contextKey string
	customCtx := context.WithValue(context.Background(), contextKey("test"), "value")
	customLogger := discard.New()
	customHashCache := &nanorpc.HashCache{}
	customWaiter := reconnect.NewConstantWaiter(1 * time.Second)

	cfg := &Config{
		Context:         customCtx,
		Logger:          customLogger,
		HashCache:       customHashCache,
		WaitReconnect:   customWaiter,
		DialTimeout:     3 * time.Second,
		ReadTimeout:     4 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     15 * time.Second,
		ReconnectDelay:  2 * time.Second,
		KeepAlive:       8 * time.Second,
		QueueSize:       100,
		AlwaysHashPaths: true,
	}

	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "SetDefaults failed")

	// Check that custom values are preserved
	core.AssertEqual(t, customCtx, cfg.Context, "Context was not preserved")
	core.AssertEqual(t, customLogger, cfg.Logger, "Logger was not preserved")
	core.AssertEqual(t, customHashCache, cfg.HashCache, "nanorpc.HashCache was not preserved")
	core.AssertNotNil(t, cfg.WaitReconnect, "WaitReconnect should not be nil")
	core.AssertEqual(t, 3*time.Second, cfg.DialTimeout, "DialTimeout was not preserved")
	core.AssertEqual(t, 4*time.Second, cfg.ReadTimeout, "ReadTimeout was not preserved")
	core.AssertEqual(t, 5*time.Second, cfg.WriteTimeout, "WriteTimeout was not preserved")
	core.AssertEqual(t, 15*time.Second, cfg.IdleTimeout, "IdleTimeout was not preserved")
	core.AssertEqual(t, 2*time.Second, cfg.ReconnectDelay, "ReconnectDelay was not preserved")
	core.AssertEqual(t, 8*time.Second, cfg.KeepAlive, "KeepAlive was not preserved")
	core.AssertEqual(t, 100, cfg.QueueSize, "QueueSize was not preserved")
	core.AssertTrue(t, cfg.AlwaysHashPaths, "AlwaysHashPaths was not preserved")
}

// ExportTestCase represents a test case for Export method
type ExportTestCase struct {
	config      *Config
	name        string
	expectError bool
}

func (tc ExportTestCase) GetName() string {
	return tc.name
}

func (tc ExportTestCase) test(t *testing.T) {
	t.Helper()
	result, err := tc.config.Export()

	if tc.expectError {
		core.AssertError(t, err, "Expected error")
		core.AssertNil(t, result, "Expected nil result on error")
	} else {
		core.AssertNoError(t, err, "Expected no error")
		core.AssertNotNil(t, result, "Expected result")

		// Verify result fields
		core.AssertEqual(t, tc.config.Remote, result.Remote, "Remote mismatch")
		core.AssertNotNil(t, result.Context, "Context should not be nil in result")
		core.AssertNotNil(t, result.Logger, "Logger should not be nil in result")
		core.AssertNotNil(t, result.WaitReconnect, "WaitReconnect should not be nil in result")
	}
}

var exportTestCases = []ExportTestCase{
	{
		name:        "missing_remote",
		config:      &Config{},
		expectError: true,
	},
	{
		name: "no_port",
		config: &Config{
			Remote: "localhost",
		},
		expectError: true,
	},
	{
		name: "port_zero",
		config: &Config{
			Remote: "localhost:0",
		},
		expectError: true,
	},
	{
		name: "valid_config",
		config: &Config{
			Remote: "localhost:8080",
		},
		expectError: false,
	},
	{
		name: "ipv6_config",
		config: &Config{
			Remote: "[::1]:8080",
		},
		expectError: false,
	},
}

func TestClientConfig_Export(t *testing.T) {
	for _, tc := range exportTestCases {
		t.Run(tc.name, tc.test)
	}
}

// TestClientConfig_getHashCache tests the getHashCache method
func TestClientConfig_getHashCache(t *testing.T) {
	// Test with nil HashCache
	cfg := &Config{}
	hc := cfg.getHashCache()
	core.AssertNotNil(t, hc, "getHashCache should not return nil")
	core.AssertEqual(t, hashCache, hc, "Expected global hashCache")

	// Test with custom HashCache
	customHC := &nanorpc.HashCache{}
	cfg.HashCache = customHC
	hc = cfg.getHashCache()
	core.AssertEqual(t, customHC, hc, "Expected custom nanorpc.HashCache")
}

// GetPathOneOfTestCase represents a test case for newGetPathOneOf
type GetPathOneOfTestCase struct {
	hc       *nanorpc.HashCache
	testFunc func(t *testing.T, hc *nanorpc.HashCache)
	name     string
}

func (tc GetPathOneOfTestCase) GetName() string {
	return tc.name
}

func (tc GetPathOneOfTestCase) test(t *testing.T) {
	t.Helper()
	tc.testFunc(t, tc.hc)
}

// newGetPathOneOfTestCase creates a new test case
func newGetPathOneOfTestCase(
	name string,
	hc *nanorpc.HashCache,
	testFunc func(t *testing.T, hc *nanorpc.HashCache),
) GetPathOneOfTestCase {
	return GetPathOneOfTestCase{
		name:     name,
		hc:       hc,
		testFunc: testFunc,
	}
}

// testAlwaysHashPathsFalse tests AlwaysHashPaths=false behaviour
func testAlwaysHashPathsFalse(t *testing.T, hc *nanorpc.HashCache) {
	cfg := &Config{
		AlwaysHashPaths: false,
	}

	getPathOneOf := cfg.newGetPathOneOf(hc)
	result := getPathOneOf("/test/path")

	pathOneof, ok := core.AssertTypeIs[*nanorpc.NanoRPCRequest_Path](
		t, result, "Expected *nanorpc.NanoRPCRequest_Path")
	if ok {
		core.AssertEqual(t, "/test/path", pathOneof.Path, "Path mismatch")
	}
}

// testAlwaysHashPathsTrue tests AlwaysHashPaths=true behaviour
func testAlwaysHashPathsTrue(t *testing.T, hc *nanorpc.HashCache) {
	cfg := &Config{
		AlwaysHashPaths: true,
	}

	getPathOneOf := cfg.newGetPathOneOf(hc)
	result := getPathOneOf("/test/path")

	pathHashOneof, ok := core.AssertTypeIs[*nanorpc.NanoRPCRequest_PathHash](t, result,
		"Expected *nanorpc.NanoRPCRequest_PathHash")
	if ok {
		core.AssertNotEqual(t, uint32(0), pathHashOneof.PathHash, "Expected non-zero hash")
	}

	// Test consistency
	result2 := getPathOneOf("/test/path")
	pathHashOneof2, ok := core.AssertTypeIs[*nanorpc.NanoRPCRequest_PathHash](t, result2,
		"Expected *nanorpc.NanoRPCRequest_PathHash")
	if ok {
		core.AssertEqual(t, pathHashOneof.PathHash, pathHashOneof2.PathHash, "Hash should be consistent")
	}
}

// testAlwaysHashPathsTrueNilCache tests AlwaysHashPaths=true with nil cache
func testAlwaysHashPathsTrueNilCache(t *testing.T, _ *nanorpc.HashCache) {
	cfg := &Config{
		AlwaysHashPaths: true,
		HashCache:       nil,
	}

	getPathOneOf := cfg.newGetPathOneOf(nil)
	result := getPathOneOf("/test/path")

	pathHashOneof, ok := core.AssertTypeIs[*nanorpc.NanoRPCRequest_PathHash](t, result,
		"Expected *nanorpc.NanoRPCRequest_PathHash")
	if ok {
		core.AssertNotEqual(t, uint32(0), pathHashOneof.PathHash, "Expected non-zero hash")
	}
}

func getPathOneOfTestCases(hc *nanorpc.HashCache) []GetPathOneOfTestCase {
	return []GetPathOneOfTestCase{
		newGetPathOneOfTestCase("AlwaysHashPaths_false", hc, testAlwaysHashPathsFalse),
		newGetPathOneOfTestCase("AlwaysHashPaths_true", hc, testAlwaysHashPathsTrue),
		newGetPathOneOfTestCase("AlwaysHashPaths_true_nil_cache", hc, testAlwaysHashPathsTrueNilCache),
	}
}

// TestClientConfig_newGetPathOneOf tests the newGetPathOneOf method
func TestClientConfig_newGetPathOneOf(t *testing.T) {
	hc := &nanorpc.HashCache{}
	for _, tc := range getPathOneOfTestCases(hc) {
		t.Run(tc.name, tc.test)
	}
}

// createConfigWithCallbacks creates a test config with callback tracking
func createConfigWithCallbacks() (cfgPtr *Config, onDisconnectPtr, onErrorPtr *bool) {
	onDisconnectCalled := false
	onErrorCalled := false

	cfg := &Config{
		Remote: "localhost:8080",
		OnConnect: func(context.Context, reconnect.WorkGroup) error {
			return nil
		},
		OnDisconnect: func(context.Context) error {
			onDisconnectCalled = true
			return nil
		},
		OnError: func(context.Context, error) error {
			onErrorCalled = true
			return nil
		},
	}

	return cfg, &onDisconnectCalled, &onErrorCalled
}

// TestClientConfig_CallbacksPreserved tests that callbacks are preserved
func TestClientConfig_CallbacksPreserved(t *testing.T) {
	cfg, onDisconnectCalled, onErrorCalled := createConfigWithCallbacks()

	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "SetDefaults failed")

	// Test that callbacks are preserved
	core.AssertNotNil(t, cfg.OnConnect, "OnConnect callback should not be nil")
	core.AssertNotNil(t, cfg.OnDisconnect, "OnDisconnect callback should not be nil")
	core.AssertNotNil(t, cfg.OnError, "OnError callback should not be nil")

	// Test that they work (skip OnConnect as it needs WorkGroup interface)
	err = cfg.OnDisconnect(context.Background())
	core.AssertNoError(t, err, "OnDisconnect failed")
	core.AssertTrue(t, *onDisconnectCalled, "OnDisconnect callback was not called")

	err = cfg.OnError(context.Background(), context.Canceled)
	core.AssertNoError(t, err, "OnError failed")
	core.AssertTrue(t, *onErrorCalled, "OnError callback was not called")
}
