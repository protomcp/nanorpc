package nanorpc

import (
	"testing"
)

// testBasicCounter tests basic counter functionality
func testBasicCounter(t *testing.T) {
	t.Helper()
	counter, err := NewRequestCounter()
	AssertNoError(t, err, "NewRequestCounter failed")

	// Test first ID
	id1 := counter.Next()
	AssertTrue(t, id1 > 0, "First ID should be positive")

	// Test second ID (should be different)
	id2 := counter.Next()
	AssertTrue(t, id2 > 0, "Second ID should be positive")
	AssertNotEqual(t, id1, id2, "IDs should be different")

	// Test sequential nature
	id3 := counter.Next()
	AssertEqual(t, id2+1, id3, "Expected sequential increment")
}

func TestRequestCounter_Basic(t *testing.T) {
	t.Run("basic_counter", testBasicCounter)
}

// testNilReceiver tests nil receiver behaviour
func testNilReceiver(t *testing.T) {
	t.Helper()
	var counter *RequestCounter
	id := counter.Next()
	AssertTrue(t, id > 0, "Nil receiver should return positive random ID")
}

func TestRequestCounter_NilReceiver(t *testing.T) {
	t.Run("nil_receiver", testNilReceiver)
}

// testSkipZero tests that zero is never returned
func testSkipZero(t *testing.T) {
	t.Helper()
	counter, err := NewRequestCounter()
	AssertNoError(t, err, "NewRequestCounter failed")

	// Generate many IDs to test that zero is never returned
	for i := range 1000 {
		id := counter.Next()
		AssertNotEqual(t, int32(0), id, "Counter should never return 0 at iteration %d", i)
		AssertTrue(t, id > 0, "Counter should never return negative at iteration %d", i)
	}
}

func TestRequestCounter_SkipZero(t *testing.T) {
	t.Run("skip_zero", testSkipZero)
}

// testConcurrency tests concurrent access to counter
func testConcurrency(t *testing.T) {
	t.Helper()
	counter, err := NewRequestCounter()
	AssertNoError(t, err, "NewRequestCounter failed")

	numGoroutines := 100
	numIDsPerGoroutine := 10

	helper := NewConcurrentTestHelper(t, numGoroutines)
	helper.Run(func(_ int) (any, error) {
		results := make([]int32, numIDsPerGoroutine)
		for j := range numIDsPerGoroutine {
			results[j] = counter.Next()
		}
		return results, nil
	})

	helper.AssertNoErrors()

	// Check all IDs are unique and positive
	seen := make(map[int32]bool)
	results, _ := helper.GetResults()
	for i := range results {
		ids, ok := GetResult[[]int32](results, i)
		AssertTrue(t, ok, "Failed to get result %d as []int32", i)
		for j, id := range ids {
			AssertTrue(t, id > 0, "ID at goroutine %d, index %d should be positive", i, j)
			AssertFalse(t, seen[id], "Duplicate ID found: %d", id)
			seen[id] = true
		}
	}
}

func TestRequestCounter_Concurrency(t *testing.T) {
	t.Run("concurrency", testConcurrency)
}

// testWraparound tests counter wraparound behaviour
func testWraparound(t *testing.T) {
	t.Helper()
	counter, err := NewRequestCounter()
	AssertNoError(t, err, "NewRequestCounter failed")

	// Simulate wraparound by setting counter to near max
	counter.counter.Store(2147483646) // Just below MaxInt32

	id1 := counter.Next()
	AssertEqual(t, int32(2147483647), id1, "Expected MaxInt32")

	id2 := counter.Next()
	AssertEqual(t, int32(1), id2, "Expected wraparound to 1 (skipping 0)")

	id3 := counter.Next()
	AssertEqual(t, int32(2), id3, "Expected 2 after wraparound")
}

func TestRequestCounter_Wraparound(t *testing.T) {
	t.Run("wraparound", testWraparound)
}

// testRandomRequestID tests random ID generation
func testRandomRequestID(t *testing.T) {
	t.Helper()
	// Test that NewRandomRequestID generates valid IDs
	for range 100 {
		id, err := NewRandomRequestID()
		AssertNoError(t, err, "NewRandomRequestID failed")
		AssertTrue(t, id > 0, "NewRandomRequestID should return positive ID")
	}
}

func TestNewRandomRequestID(t *testing.T) {
	t.Run("random_request_id", testRandomRequestID)
}

// testRandomRequestIDUniqueness tests random ID uniqueness
func testRandomRequestIDUniqueness(t *testing.T) {
	t.Helper()
	// Test that NewRandomRequestID generates reasonably unique IDs
	numIDs := 1000
	seen := make(map[int32]bool)

	for range numIDs {
		id, err := NewRandomRequestID()
		AssertNoError(t, err, "NewRandomRequestID failed")
		seen[id] = true
	}

	// We should have close to numIDs unique values
	// Allow for some duplicates due to randomness
	minExpected := numIDs * 95 / 100
	AssertTrue(t, len(seen) > minExpected, "Expected at least %d unique IDs, got %d", minExpected, len(seen))
}

func TestNewRandomRequestID_Uniqueness(t *testing.T) {
	t.Run("uniqueness", testRandomRequestIDUniqueness)
}

// testStartingPoint tests that different counters start at different points
func testStartingPoint(t *testing.T) {
	t.Helper()
	// Test that different counters start at different points
	counter1, err := NewRequestCounter()
	AssertNoError(t, err, "NewRequestCounter failed")

	counter2, err := NewRequestCounter()
	AssertNoError(t, err, "NewRequestCounter failed")

	id1 := counter1.Next()
	id2 := counter2.Next()

	// They should very likely be different (random starting points)
	if id1 == id2 {
		// This could happen by chance, so let's check a few more
		sameCount := 0
		for range 10 {
			if counter1.Next() == counter2.Next() {
				sameCount++
			}
		}
		AssertTrue(t, sameCount < 6,
			"Counters appear to be synchronized, expected different starting points. Same count: %d", sameCount)
	}
}

func TestRequestCounter_StartingPoint(t *testing.T) {
	t.Run("starting_point", testStartingPoint)
}
