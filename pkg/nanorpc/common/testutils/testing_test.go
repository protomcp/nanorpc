package testutils

import (
	"testing"
)

// TestTInterface tests that our T interface works with *testing.T
func TestTInterface(t *testing.T) {
	// This is just a compile-time check that testing.T implements T
	var _ T = t

	// Test Helper method (safe to call)
	t.Helper()

	// We can't test Errorf/Fatalf/Fatal without failing the test,
	// but the compile-time check above ensures they're available
}

// Example test case type for testing RunTestCases
type exampleTestCase struct {
	name   string
	input  int
	output int
}

func (tc exampleTestCase) Name() string {
	return tc.name
}

func (tc exampleTestCase) Test(t *testing.T) {
	result := tc.input * 2
	if result != tc.output {
		t.Errorf("expected %d * 2 = %d, got %d", tc.input, tc.output, result)
	}
}

// TestRunTestCases tests the RunTestCases function
func TestRunTestCases(t *testing.T) {
	tests := []exampleTestCase{
		{name: "zero", input: 0, output: 0},
		{name: "positive", input: 5, output: 10},
		{name: "negative", input: -3, output: -6},
	}

	RunTestCases(t, tests)
}
