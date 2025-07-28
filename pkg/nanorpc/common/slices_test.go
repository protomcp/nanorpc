package common

import (
	"testing"
	"unsafe"

	"protomcp.org/nanorpc/pkg/nanorpc/common/testutils"
)

// testStruct is used to verify that pointer fields are properly zeroed
type testStruct struct {
	Ptr  *int
	Name string
	Data []byte
	ID   int
}

func TestClearSlice(t *testing.T) {
	t.Run("clears primitive types", testClearSlicePrimitives)
	t.Run("clears strings", testClearSliceStrings)
	t.Run("clears structs with pointers", testClearSliceStructs)
	t.Run("handles empty slice", testClearSliceEmpty)
	t.Run("handles nil slice", testClearSliceNil)
	t.Run("preserves capacity for reuse", testClearSliceCapacity)
}

func testClearSlicePrimitives(t *testing.T) {
	ints := []int{1, 2, 3, 4, 5}
	originalCap := cap(ints)
	originalArrayPtr := unsafe.Pointer(&ints[0])

	result := ClearSlice(ints)

	// Verify length is 0 but capacity unchanged
	testutils.AssertEqual(t, 0, len(result), "length should be 0")
	testutils.AssertEqual(t, originalCap, cap(result), "capacity should be unchanged")

	// Verify same underlying array
	if len(result) > 0 {
		resultArrayPtr := unsafe.Pointer(&result[0])
		testutils.AssertEqual(t, originalArrayPtr, resultArrayPtr, "should reuse same array")
	}

	// Verify original slice elements are zeroed
	for i := 0; i < len(ints); i++ {
		testutils.AssertEqual(t, 0, ints[i], "element should be zeroed")
	}
}

func testClearSliceStrings(t *testing.T) {
	strings := []string{"hello", "world", "test"}
	originalCap := cap(strings)

	result := ClearSlice(strings)

	testutils.AssertEqual(t, 0, len(result), "length should be 0")
	testutils.AssertEqual(t, originalCap, cap(result), "capacity should be unchanged")

	// Verify original slice elements are zeroed
	for i := 0; i < len(strings); i++ {
		testutils.AssertEqual(t, "", strings[i], "string should be empty")
	}
}

func testClearSliceStructs(t *testing.T) {
	val1, val2 := 10, 20
	structs := []testStruct{
		{ID: 1, Name: "first", Data: []byte{1, 2, 3}, Ptr: &val1},
		{ID: 2, Name: "second", Data: []byte{4, 5, 6}, Ptr: &val2},
	}
	originalCap := cap(structs)

	result := ClearSlice(structs)

	testutils.AssertEqual(t, 0, len(result), "length should be 0")
	testutils.AssertEqual(t, originalCap, cap(result), "capacity should be unchanged")

	// Verify all fields are zeroed
	for i := 0; i < len(structs); i++ {
		testutils.AssertEqual(t, 0, structs[i].ID, "ID should be zero")
		testutils.AssertEqual(t, "", structs[i].Name, "Name should be empty")
		testutils.AssertNil(t, structs[i].Data, "Data should be nil")
		testutils.AssertNil(t, structs[i].Ptr, "Ptr should be nil")
	}
}

func testClearSliceEmpty(t *testing.T) {
	empty := []int{}
	result := ClearSlice(empty)

	testutils.AssertEqual(t, 0, len(result), "length should be 0")
	testutils.AssertEqual(t, 0, cap(result), "capacity should be 0")
}

func testClearSliceNil(t *testing.T) {
	var nilSlice []int
	result := ClearSlice(nilSlice)

	testutils.AssertNil(t, result, "result should be nil")
}

func testClearSliceCapacity(t *testing.T) {
	// Create slice with extra capacity
	slice := make([]int, 3, 10)
	slice[0], slice[1], slice[2] = 1, 2, 3

	result := ClearSlice(slice)

	testutils.AssertEqual(t, 0, len(result), "length should be 0")
	testutils.AssertEqual(t, 10, cap(result), "capacity should be preserved")

	// Can reuse the slice
	result = append(result, 10, 20, 30, 40, 50)
	testutils.AssertEqual(t, 5, len(result), "should be able to append")
	testutils.AssertEqual(t, 10, cap(result), "capacity still unchanged")
}

func TestClearAndNilSlice(t *testing.T) {
	t.Run("clears and nils primitive slice", testClearAndNilPrimitives)
	t.Run("clears and nils struct slice", testClearAndNilStructs)
	t.Run("handles nil slice", testClearAndNilNil)
}

func testClearAndNilPrimitives(t *testing.T) {
	ints := []int{1, 2, 3, 4, 5}
	intsCopy := make([]int, len(ints))
	copy(intsCopy, ints)

	result := ClearAndNilSlice(ints)

	// Result should be nil
	testutils.AssertNil(t, result, "result should be nil")

	// Original slice should have zeroed elements
	for i := 0; i < len(ints); i++ {
		testutils.AssertEqual(t, 0, ints[i], "element should be zeroed")
	}
}

func testClearAndNilStructs(t *testing.T) {
	val := 42
	structs := []testStruct{
		{ID: 1, Name: "test", Data: []byte{1, 2, 3}, Ptr: &val},
	}

	result := ClearAndNilSlice(structs)

	// Result should be nil
	testutils.AssertNil(t, result, "result should be nil")

	// Original struct should be zeroed
	testutils.AssertEqual(t, 0, structs[0].ID, "ID should be zero")
	testutils.AssertEqual(t, "", structs[0].Name, "Name should be empty")
	testutils.AssertNil(t, structs[0].Data, "Data should be nil")
	testutils.AssertNil(t, structs[0].Ptr, "Ptr should be nil")
}

func testClearAndNilNil(t *testing.T) {
	var nilSlice []int
	result := ClearAndNilSlice(nilSlice)

	testutils.AssertNil(t, result, "result should be nil")
}
