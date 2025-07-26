package testutils

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// mockT is a mock T interface for testing assertion functions
type mockT struct {
	buf       strings.Builder
	fatalArgs []any
	failed    bool
}

func (*mockT) Helper() {}

func (m *mockT) Fatal(args ...any) {
	m.failed = true
	m.fatalArgs = args
	for i, arg := range args {
		if i > 0 {
			_, _ = m.buf.WriteString(" ")
		}
		_, _ = m.buf.WriteString(fmt.Sprint(arg))
	}
}

func (m *mockT) Failed() bool {
	return m.failed
}

// TestS tests the S slice constructor helper
func TestS(t *testing.T) {
	// Test with integers
	intSlice := S(1, 2, 3)
	if len(intSlice) != 3 || intSlice[0] != 1 || intSlice[1] != 2 || intSlice[2] != 3 {
		t.Errorf("S(1, 2, 3) = %v, want [1 2 3]", intSlice)
	}

	// Test with strings
	strSlice := S("a", "b", "c")
	if len(strSlice) != 3 || strSlice[0] != "a" || strSlice[1] != "b" || strSlice[2] != "c" {
		t.Errorf("S(\"a\", \"b\", \"c\") = %v, want [a b c]", strSlice)
	}

	// Test with empty slice
	emptySlice := S[int]()
	if len(emptySlice) != 0 {
		t.Errorf("S[int]() = %v, want []", emptySlice)
	}
}

// TestAssertNotNil tests the AssertNotNil function
func TestAssertNotNil(t *testing.T) {
	// Test with non-nil value - should not fail
	mt := &mockT{}
	AssertNotNil(mt, "not nil", "")
	if mt.Failed() {
		t.Error("AssertNotNil failed with non-nil value")
	}

	// Test with nil value - should fail
	mt = &mockT{}
	AssertNotNil(mt, nil, "")
	if !mt.Failed() {
		t.Error("AssertNotNil didn't fail with nil value")
	}
	AssertContains(t, mt.buf.String(), "expected non-nil value", "")

	// Test with custom message
	mt = &mockT{}
	AssertNotNil(mt, nil, "custom message: %s", "test")
	AssertContains(t, mt.buf.String(), "custom message: test", "")
}

// TestAssertNoError tests the AssertNoError function
func TestAssertNoError(t *testing.T) {
	// Test with nil error - should not fail
	mt := &mockT{}
	AssertNoError(mt, nil, "")
	if mt.Failed() {
		t.Error("AssertNoError failed with nil error")
	}

	// Test with error - should fail
	mt = &mockT{}
	testErr := errors.New("test error")
	AssertNoError(mt, testErr, "")
	if !mt.Failed() {
		t.Error("AssertNoError didn't fail with error")
	}
	AssertContains(t, mt.buf.String(), "unexpected error: test error", "")

	// Test with custom message
	mt = &mockT{}
	AssertNoError(mt, testErr, "custom error: %s", "message")
	AssertContains(t, mt.buf.String(), "custom error: message: unexpected error: test error", "")
}

// TestAssertError tests the AssertError function
func TestAssertError(t *testing.T) {
	// Test with error - should not fail
	mt := &mockT{}
	AssertError(mt, errors.New("test"), "")
	if mt.Failed() {
		t.Error("AssertError failed with error")
	}

	// Test with nil error - should fail
	mt = &mockT{}
	AssertError(mt, nil, "")
	if !mt.Failed() {
		t.Error("AssertError didn't fail with nil error")
	}
	AssertContains(t, mt.buf.String(), "expected an error but got nil", "")

	// Test with custom message
	mt = &mockT{}
	AssertError(mt, nil, "custom message: %s", "test")
	AssertContains(t, mt.buf.String(), "custom message: test", "")
}

// TestAssertEqual tests the AssertEqual function
func TestAssertEqual(t *testing.T) {
	// Test with equal values - should not fail
	mt := &mockT{}
	AssertEqual(mt, 42, 42, "")
	if mt.Failed() {
		t.Error("AssertEqual failed with equal values")
	}

	// Test with unequal values - should fail
	mt = &mockT{}
	AssertEqual(mt, 42, 43, "")
	if !mt.Failed() {
		t.Error("AssertEqual didn't fail with unequal values")
	}
	AssertContains(t, mt.buf.String(), "expected 42, got 43", "error message should contain comparison")

	// Test with custom message
	mt = &mockT{}
	AssertEqual(mt, "foo", "bar", "custom message: %s", "test")
	AssertContains(t, mt.buf.String(),
		"custom message: test: expected foo, got bar", "custom message should be in output")

	// Test with slices
	mt = &mockT{}
	AssertEqual(mt, []int{1, 2, 3}, []int{1, 2, 3}, "")
	if mt.Failed() {
		t.Error("AssertEqual failed with equal slices")
	}

	mt = &mockT{}
	AssertEqual(mt, []int{1, 2, 3}, []int{1, 2, 4}, "")
	if !mt.Failed() {
		t.Error("AssertEqual didn't fail with unequal slices")
	}
}

// TestAssertTrue tests the AssertTrue function
func TestAssertTrue(t *testing.T) {
	// Test with true value - should not fail
	mt := &mockT{}
	AssertTrue(mt, true, "")
	if mt.Failed() {
		t.Error("AssertTrue failed with true value")
	}

	// Test with false value - should fail
	mt = &mockT{}
	AssertTrue(mt, false, "")
	if !mt.Failed() {
		t.Error("AssertTrue didn't fail with false value")
	}
	AssertContains(t, mt.buf.String(), "expected true, got false", "should show boolean comparison")

	// Test with custom message
	mt = &mockT{}
	AssertTrue(mt, false, "custom message: %s", "test")
	AssertContains(t, mt.buf.String(), "custom message: test", "")
}

// TestAssertFalse tests the AssertFalse function
func TestAssertFalse(t *testing.T) {
	// Test with false value - should not fail
	mt := &mockT{}
	AssertFalse(mt, false, "")
	if mt.Failed() {
		t.Error("AssertFalse failed with false value")
	}

	// Test with true value - should fail
	mt = &mockT{}
	AssertFalse(mt, true, "")
	if !mt.Failed() {
		t.Error("AssertFalse didn't fail with true value")
	}
	AssertContains(t, mt.buf.String(), "expected false, got true", "should show boolean comparison")

	// Test with custom message
	mt = &mockT{}
	AssertFalse(mt, true, "custom message: %s", "test")
	AssertContains(t, mt.buf.String(), "custom message: test", "")
}

// TestAssertNil tests the AssertNil function
func TestAssertNil(t *testing.T) {
	// Test with nil value - should not fail
	mt := &mockT{}
	AssertNil(mt, nil, "")
	if mt.Failed() {
		t.Error("AssertNil failed with nil value")
	}

	// Test with non-nil value - should fail
	mt = &mockT{}
	AssertNil(mt, "not nil", "")
	if !mt.Failed() {
		t.Error("AssertNil didn't fail with non-nil value")
	}
	AssertContains(t, mt.buf.String(), "expected nil, got not nil", "should show nil comparison")

	// Test with custom message
	mt = &mockT{}
	AssertNil(mt, 42, "custom message: %s", "test")
	AssertContains(t, mt.buf.String(),
		"custom message: test: expected nil, got 42", "custom message should be in output")
}

// TestAssertNotEqual tests the AssertNotEqual function
func TestAssertNotEqual(t *testing.T) {
	// Test with unequal values - should not fail
	mt := &mockT{}
	AssertNotEqual(mt, 42, 43, "")
	if mt.Failed() {
		t.Error("AssertNotEqual failed with unequal values")
	}

	// Test with equal values - should fail
	mt = &mockT{}
	AssertNotEqual(mt, 42, 42, "")
	if !mt.Failed() {
		t.Error("AssertNotEqual didn't fail with equal values")
	}
	AssertContains(t, mt.buf.String(), "expected values to be different, both were 42", "should show equality message")

	// Test with custom message
	mt = &mockT{}
	AssertNotEqual(mt, "foo", "foo", "custom message: %s", "test")
	AssertContains(t, mt.buf.String(),
		"custom message: test: expected values to be different, both were foo",
		"custom message should be in output")
}

// TestAssertContains tests the AssertContains function
func TestAssertContains(t *testing.T) {
	// Test with substring present - should not fail
	mt := &mockT{}
	AssertContains(mt, "hello world", "world", "")
	if mt.Failed() {
		t.Error("AssertContains failed with substring present")
	}

	// Test with substring not present - should fail
	mt = &mockT{}
	AssertContains(mt, "hello world", "foo", "")
	if !mt.Failed() {
		t.Error("AssertContains didn't fail with substring not present")
	}
	if !strings.Contains(mt.buf.String(), `expected "hello world" to contain "foo"`) {
		t.Errorf("AssertContains error message = %q, want containing 'expected \"hello world\" to contain \"foo\"'",
			mt.buf.String())
	}

	// Test with empty substring - should not fail
	mt = &mockT{}
	AssertContains(mt, "hello", "", "")
	if mt.Failed() {
		t.Error("AssertContains failed with empty substring")
	}

	// Test with custom message
	mt = &mockT{}
	AssertContains(mt, "hello", "bye", "custom message: %s", "test")
	if !strings.Contains(mt.buf.String(), `custom message: test: expected "hello" to contain "bye"`) {
		t.Errorf("AssertContains custom message = %q, "+
			"want containing 'custom message: test: expected \"hello\" to contain \"bye\"'",
			mt.buf.String())
	}
}

// TestContains tests the internal contains function
func TestContains(t *testing.T) {
	tests := []struct {
		str      string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "o w", true},
		{"hello world", "foo", false},
		{"hello", "", true},
		{"", "", true},
		{"", "a", false},
		{"short", "longer string", false},
	}

	for _, tt := range tests {
		result := strings.Contains(tt.str, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
		}
	}
}

// TestAssertTypeIs tests the AssertTypeIs function
func TestAssertTypeIs(t *testing.T) {
	// Test successful type assertion
	mt := &mockT{}
	var value any = "hello"
	result := AssertTypeIs[string](mt, value, "")
	AssertFalse(t, mt.Failed(), "AssertTypeIs failed with correct type")
	AssertEqual(t, "hello", result, "AssertTypeIs should return the correct value")

	// Test failed type assertion
	mt = &mockT{}
	var intValue any = 42
	_ = AssertTypeIs[string](mt, intValue, "")
	AssertTrue(t, mt.Failed(), "AssertTypeIs should fail with incorrect type")
	AssertContains(t, mt.buf.String(), "expected type string but got int", "")

	// Test with custom message
	mt = &mockT{}
	_ = AssertTypeIs[string](mt, intValue, "custom message: %s", "test")
	AssertContains(t, mt.buf.String(), "custom message: test: expected type string but got int", "")

	// Test with pointer types
	mt = &mockT{}
	type MyStruct struct{ Value int }
	var structPtr any = &MyStruct{Value: 123}
	resultPtr := AssertTypeIs[*MyStruct](mt, structPtr, "")
	AssertFalse(t, mt.Failed(), "AssertTypeIs failed with correct pointer type")
	AssertEqual(t, 123, resultPtr.Value, "AssertTypeIs should return struct with correct value")

	// Test with interface types
	mt = &mockT{}
	var errValue any = errors.New("test error")
	resultErr := AssertTypeIs[error](mt, errValue, "")
	AssertFalse(t, mt.Failed(), "AssertTypeIs failed with interface type")
	AssertEqual(t, "test error", resultErr.Error(), "AssertTypeIs should return correct error")
}
