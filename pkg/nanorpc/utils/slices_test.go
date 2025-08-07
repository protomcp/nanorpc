package utils

import (
	"testing"
	"unsafe"

	"darvaza.org/core"
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
	core.AssertEqual(t, 0, len(result), "length")
	core.AssertEqual(t, originalCap, cap(result), "capacity")

	// Verify same underlying array
	if len(result) > 0 {
		resultArrayPtr := unsafe.Pointer(&result[0])
		core.AssertEqual(t, originalArrayPtr, resultArrayPtr, "array pointer")
	}

	// Verify original slice elements are zeroed
	for i := 0; i < len(ints); i++ {
		core.AssertEqual(t, 0, ints[i], "element %d", i)
	}
}

func testClearSliceStrings(t *testing.T) {
	strings := []string{"hello", "world", "test"}
	originalCap := cap(strings)

	result := ClearSlice(strings)

	core.AssertEqual(t, 0, len(result), "length")
	core.AssertEqual(t, originalCap, cap(result), "capacity")

	// Verify original slice elements are zeroed
	for i := 0; i < len(strings); i++ {
		core.AssertEqual(t, "", strings[i], "string %d", i)
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

	core.AssertEqual(t, 0, len(result), "length")
	core.AssertEqual(t, originalCap, cap(result), "capacity")

	// Verify all fields are zeroed
	for i := 0; i < len(structs); i++ {
		core.AssertEqual(t, 0, structs[i].ID, "ID %d", i)
		core.AssertEqual(t, "", structs[i].Name, "Name %d", i)
		core.AssertNil(t, structs[i].Data, "Data %d", i)
		core.AssertNil(t, structs[i].Ptr, "Ptr %d", i)
	}
}

func testClearSliceEmpty(t *testing.T) {
	empty := []int{}
	result := ClearSlice(empty)

	core.AssertEqual(t, 0, len(result), "length")
	core.AssertEqual(t, 0, cap(result), "capacity")
}

func testClearSliceNil(t *testing.T) {
	var nilSlice []int
	result := ClearSlice(nilSlice)

	core.AssertNil(t, result, "result")
}

func testClearSliceCapacity(t *testing.T) {
	// Create slice with extra capacity
	slice := make([]int, 3, 10)
	slice[0], slice[1], slice[2] = 1, 2, 3

	result := ClearSlice(slice)

	core.AssertEqual(t, 0, len(result), "length")
	core.AssertEqual(t, 10, cap(result), "capacity")

	// Can reuse the slice
	result = append(result, 10, 20, 30, 40, 50)
	core.AssertEqual(t, 5, len(result), "length after append")
	core.AssertEqual(t, 10, cap(result), "capacity after append")
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
	core.AssertNil(t, result, "result")

	// Original slice should have zeroed elements
	for i := 0; i < len(ints); i++ {
		core.AssertEqual(t, 0, ints[i], "element %d", i)
	}
}

func testClearAndNilStructs(t *testing.T) {
	val := 42
	structs := []testStruct{
		{ID: 1, Name: "test", Data: []byte{1, 2, 3}, Ptr: &val},
	}

	result := ClearAndNilSlice(structs)

	// Result should be nil
	core.AssertNil(t, result, "result")

	// Original struct should be zeroed
	core.AssertEqual(t, 0, structs[0].ID, "ID")
	core.AssertEqual(t, "", structs[0].Name, "Name")
	core.AssertNil(t, structs[0].Data, "Data")
	core.AssertNil(t, structs[0].Ptr, "Ptr")
}

func testClearAndNilNil(t *testing.T) {
	var nilSlice []int
	result := ClearAndNilSlice(nilSlice)

	core.AssertNil(t, result, "result")
}
