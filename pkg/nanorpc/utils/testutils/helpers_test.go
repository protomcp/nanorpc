package testutils

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
)

// TestCase interface validations - MANDATORY
var _ core.TestCase = getFieldTestCase{}
var _ core.TestCase = concurrentTestHelperTestCase{}
var _ core.TestCase = waitForConditionTestCase{}
var _ core.TestCase = assertWaitForConditionTestCase{}
var _ core.TestCase = concurrentTestHelperDefaultsTestCase{}
var _ core.TestCase = concurrentTestHelperSetDefaultsTestCase{}
var _ core.TestCase = waitForConditionLoopTestCase{}
var _ core.TestCase = assertWaitForConditionErrorMessageTestCase{}
var _ core.TestCase = assertWaitForConditionEmptyNameTestCase{}
var _ core.TestCase = assertFieldTypeIsTestCase{}

// getFieldTestCase tests GetField function
// Fields ordered for memory efficiency (large to small)
type getFieldTestCase struct {
	// 16+ bytes (map, interface)
	input    map[string]any
	expected any
	// 16 bytes (string header)
	key  string
	name string
	// 1 byte (bool)
	wantOK bool
}

func (tc getFieldTestCase) Name() string {
	return tc.name
}

func (tc getFieldTestCase) Test(t *testing.T) {
	t.Helper()

	result, ok := GetField[string, string](tc.input, tc.key)

	core.AssertEqual(t, tc.wantOK, ok, "key found")
	if tc.wantOK {
		if expectedStr, ok := tc.expected.(string); ok {
			core.AssertEqual(t, expectedStr, result, "value")
		}
	}
}

// Factory function for getFieldTestCase
func newGetFieldTestCase(name, key string, input map[string]any, expected any, wantOK bool) getFieldTestCase {
	return getFieldTestCase{
		name:     name,
		key:      key,
		input:    input,
		expected: expected,
		wantOK:   wantOK,
	}
}

func TestGetField(t *testing.T) {
	testCases := []getFieldTestCase{
		newGetFieldTestCase("existing string key", "name",
			map[string]any{"name": "test", "age": 25}, "test", true),
		newGetFieldTestCase("missing key", "missing",
			map[string]any{"name": "test"}, "", false),
		newGetFieldTestCase("wrong type", "age",
			map[string]any{"age": 25}, "", false),
		newGetFieldTestCase("empty map", "key",
			map[string]any{}, "", false),
		newGetFieldTestCase("nil value", "nil_key",
			map[string]any{"nil_key": nil}, "", false),
	}

	core.RunTestCases(t, testCases)
}

// concurrentTestHelperTestCase tests ConcurrentTestHelper
// Fields ordered for memory efficiency (large to small)
type concurrentTestHelperTestCase struct {
	// 16 bytes (string header)
	name string
	// ~24 bytes (struct with function pointer)
	helper ConcurrentTestHelper
	// 1 byte (bool)
	expectedErr bool
}

func (tc concurrentTestHelperTestCase) Name() string {
	return tc.name
}

func (tc concurrentTestHelperTestCase) Test(t *testing.T) {
	t.Helper()

	errs := tc.helper.Run()

	core.AssertEqual(t, tc.helper.NumGoroutines, len(errs), "error slice length")

	hasError := false
	for _, err := range errs {
		if err != nil {
			hasError = true
			break
		}
	}

	core.AssertEqual(t, tc.expectedErr, hasError, "error presence")
}

// Factory function for concurrentTestHelperTestCase
func newConcurrentTestHelperTestCase(name string, helper ConcurrentTestHelper,
	expectedErr bool) concurrentTestHelperTestCase {
	return concurrentTestHelperTestCase{
		name:        name,
		helper:      helper,
		expectedErr: expectedErr,
	}
}

func TestConcurrentTestHelper(t *testing.T) {
	testCases := []concurrentTestHelperTestCase{
		newConcurrentTestHelperTestCase("successful test",
			ConcurrentTestHelper{
				TestFunc:      func(_ int) error { return nil },
				NumGoroutines: 5,
				Timeout:       1 * time.Second,
			}, false),
		newConcurrentTestHelperTestCase("test with errors",
			ConcurrentTestHelper{
				TestFunc: func(id int) error {
					if id%2 == 0 {
						return errors.New("test error")
					}
					return nil
				},
				NumGoroutines: 4,
				Timeout:       1 * time.Second,
			}, true),
		newConcurrentTestHelperTestCase("nil test function",
			ConcurrentTestHelper{
				TestFunc:      nil,
				NumGoroutines: 2,
				Timeout:       1 * time.Second,
			}, false),
	}

	core.RunTestCases(t, testCases)
}

// concurrentTestHelperDefaultsTestCase tests ConcurrentTestHelper defaults
// Fields ordered for memory efficiency (large to small)
type concurrentTestHelperDefaultsTestCase struct {
	name string
}

func (tc concurrentTestHelperDefaultsTestCase) Name() string {
	return tc.name
}

func (concurrentTestHelperDefaultsTestCase) Test(t *testing.T) {
	t.Helper()

	helper := &ConcurrentTestHelper{}
	helper.setDefaults()

	core.AssertEqual(t, 10, helper.NumGoroutines, "default goroutines")
	core.AssertEqual(t, 5*time.Second, helper.Timeout, "default timeout")
}

// Factory function for concurrentTestHelperDefaultsTestCase
func newConcurrentTestHelperDefaultsTestCase(name string) concurrentTestHelperDefaultsTestCase {
	return concurrentTestHelperDefaultsTestCase{
		name: name,
	}
}

func TestConcurrentTestHelperDefaults(t *testing.T) {
	testCases := []concurrentTestHelperDefaultsTestCase{
		newConcurrentTestHelperDefaultsTestCase("test setDefaults with empty helper"),
	}

	core.RunTestCases(t, testCases)
}

// concurrentTestHelperSetDefaultsTestCase tests ConcurrentTestHelper setDefaults
// Fields ordered for memory efficiency (large to small)
type concurrentTestHelperSetDefaultsTestCase struct {
	// 16 bytes (string header)
	name string
	// ~24 bytes (struct with function pointer)
	helper ConcurrentTestHelper
	// 8 bytes (time.Duration)
	expectedTimeout time.Duration
	// 4 bytes (int)
	expectedGoroutines int
}

func (tc concurrentTestHelperSetDefaultsTestCase) Name() string {
	return tc.name
}

func (tc concurrentTestHelperSetDefaultsTestCase) Test(t *testing.T) {
	t.Helper()

	helper := tc.helper
	helper.setDefaults()

	core.AssertEqual(t, tc.expectedGoroutines, helper.NumGoroutines, "goroutines")
	core.AssertEqual(t, tc.expectedTimeout, helper.Timeout, "timeout")
}

// Factory function for concurrentTestHelperSetDefaultsTestCase
func newConcurrentTestHelperSetDefaultsTestCase(name string, helper ConcurrentTestHelper,
	expectedGoroutines int, expectedTimeout time.Duration) concurrentTestHelperSetDefaultsTestCase {
	return concurrentTestHelperSetDefaultsTestCase{
		name:               name,
		helper:             helper,
		expectedGoroutines: expectedGoroutines,
		expectedTimeout:    expectedTimeout,
	}
}

func TestConcurrentTestHelperSetDefaults(t *testing.T) {
	testCases := []concurrentTestHelperSetDefaultsTestCase{
		newConcurrentTestHelperSetDefaultsTestCase("zero values", ConcurrentTestHelper{}, 10, 5*time.Second),
		newConcurrentTestHelperSetDefaultsTestCase("negative values",
			ConcurrentTestHelper{NumGoroutines: -1, Timeout: -1 * time.Second}, 10, 5*time.Second),
		newConcurrentTestHelperSetDefaultsTestCase("positive values unchanged",
			ConcurrentTestHelper{NumGoroutines: 3, Timeout: 2 * time.Second}, 3, 2*time.Second),
	}

	core.RunTestCases(t, testCases)
}

// waitForConditionTestCase tests WaitForCondition function
// Fields ordered for memory efficiency (large to small)
type waitForConditionTestCase struct {
	// 8 bytes (function pointer)
	conditionFunc func() bool
	// 16 bytes (string header)
	name string
	// 8 bytes each (time.Duration)
	timeout  time.Duration
	interval time.Duration
	// 1 byte (bool)
	expected bool
}

func (tc waitForConditionTestCase) Name() string {
	return tc.name
}

func (tc waitForConditionTestCase) Test(t *testing.T) {
	t.Helper()

	result := WaitForCondition(tc.conditionFunc, tc.timeout, tc.interval)
	core.AssertEqual(t, tc.expected, result, "condition result")
}

// Factory function for waitForConditionTestCase
func newWaitForConditionTestCase(name string, conditionFunc func() bool,
	timeout, interval time.Duration, expected bool) waitForConditionTestCase {
	return waitForConditionTestCase{
		conditionFunc: conditionFunc,
		name:          name,
		timeout:       timeout,
		interval:      interval,
		expected:      expected,
	}
}

func TestWaitForCondition(t *testing.T) {
	// Test immediate success
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	// Test delayed success using atomic counter
	var counter atomic.Int32
	delayedCondition := func() bool {
		return counter.Add(1) >= 3 // Becomes true on 3rd call
	}

	testCases := []waitForConditionTestCase{
		newWaitForConditionTestCase("immediate true", trueCondition,
			100*time.Millisecond, 10*time.Millisecond, true),
		newWaitForConditionTestCase("always false timeout", falseCondition,
			50*time.Millisecond, 10*time.Millisecond, false),
		newWaitForConditionTestCase("delayed success", delayedCondition,
			200*time.Millisecond, 20*time.Millisecond, true),
		newWaitForConditionTestCase("zero interval uses default", trueCondition,
			100*time.Millisecond, 0, true),
	}

	core.RunTestCases(t, testCases)
}

// waitForConditionLoopTestCase tests waitForConditionLoop function directly
// Fields ordered for memory efficiency (large to small)
type waitForConditionLoopTestCase struct {
	// 8 bytes (function pointer)
	conditionFunc func() bool
	// 16 bytes (string header)
	name string
	// 8 bytes each (time.Duration)
	timeout  time.Duration
	interval time.Duration
	// 1 byte (bool)
	expected bool
}

func (tc waitForConditionLoopTestCase) Name() string {
	return tc.name
}

func (tc waitForConditionLoopTestCase) Test(t *testing.T) {
	t.Helper()

	result := waitForConditionLoop(tc.conditionFunc, tc.timeout, tc.interval)
	core.AssertEqual(t, tc.expected, result, "condition result")
}

// Factory function for waitForConditionLoopTestCase
func newWaitForConditionLoopTestCase(name string, conditionFunc func() bool,
	timeout, interval time.Duration, expected bool) waitForConditionLoopTestCase {
	return waitForConditionLoopTestCase{
		conditionFunc: conditionFunc,
		name:          name,
		timeout:       timeout,
		interval:      interval,
		expected:      expected,
	}
}

func waitForConditionLoopTestCases() []waitForConditionLoopTestCase {
	// Test delayed success using atomic counter
	var callCount atomic.Int32
	delayedCondition := func() bool {
		return callCount.Add(1) >= 2 // True on second call
	}

	neverTrue := func() bool { return false }

	return []waitForConditionLoopTestCase{
		newWaitForConditionLoopTestCase("delayed success", delayedCondition,
			100*time.Millisecond, 10*time.Millisecond, true),
		newWaitForConditionLoopTestCase("timeout failure", neverTrue,
			20*time.Millisecond, 5*time.Millisecond, false),
	}
}

func TestWaitForConditionLoop(t *testing.T) {
	core.RunTestCases(t, waitForConditionLoopTestCases())
}

// assertWaitForConditionTestCase tests AssertWaitForCondition function
// Fields ordered for memory efficiency (large to small)
type assertWaitForConditionTestCase struct {
	// 8 bytes (function pointer)
	conditionFunc func() bool
	// 16 bytes (string header)
	name string
	// 8 bytes (time.Duration)
	timeout time.Duration
	// 1 byte (bool)
	expectSuccess bool
}

func (tc assertWaitForConditionTestCase) Name() string {
	return tc.name
}

func (tc assertWaitForConditionTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &core.MockT{}
	result := AssertWaitForCondition(mock, tc.conditionFunc, tc.timeout, "test condition")

	core.AssertEqual(t, tc.expectSuccess, result, "assertion result")

	if tc.expectSuccess {
		core.AssertFalse(t, mock.HasErrors(), "should not have errors on success")
	} else {
		core.AssertTrue(t, mock.HasErrors(), "should have errors on failure")
	}
}

// Factory function for assertWaitForConditionTestCase
func newAssertWaitForConditionTestCase(name string, conditionFunc func() bool,
	timeout time.Duration, expectSuccess bool) assertWaitForConditionTestCase {
	return assertWaitForConditionTestCase{
		conditionFunc: conditionFunc,
		name:          name,
		timeout:       timeout,
		expectSuccess: expectSuccess,
	}
}

func TestAssertWaitForCondition(t *testing.T) {
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	testCases := []assertWaitForConditionTestCase{
		newAssertWaitForConditionTestCase("successful condition", trueCondition,
			100*time.Millisecond, true),
		newAssertWaitForConditionTestCase("failing condition", falseCondition,
			50*time.Millisecond, false),
	}

	core.RunTestCases(t, testCases)
}

// assertWaitForConditionErrorMessageTestCase tests AssertWaitForCondition error formatting
// Fields ordered for memory efficiency (large to small)
type assertWaitForConditionErrorMessageTestCase struct {
	// 16 bytes (string header)
	name                   string
	expectedPrefixContent  string
	expectedTimeoutContent string
	// 8 bytes (time.Duration)
	timeout time.Duration
}

func (tc assertWaitForConditionErrorMessageTestCase) Name() string {
	return tc.name
}

func (tc assertWaitForConditionErrorMessageTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &core.MockT{}
	falseCondition := func() bool { return false }

	// Test with formatted name
	AssertWaitForCondition(mock, falseCondition, tc.timeout, "operation %s", "test")

	core.AssertTrue(t, mock.HasErrors(), "should have error")
	errs := mock.Errors
	core.AssertTrue(t, len(errs) > 0, "should have error messages")

	errorMsg := errs[0]
	core.AssertContains(t, errorMsg, tc.expectedPrefixContent, "should contain formatted prefix")
	core.AssertContains(t, errorMsg, tc.expectedTimeoutContent, "should contain timeout message")
}

// Factory function for assertWaitForConditionErrorMessageTestCase
func newAssertWaitForConditionErrorMessageTestCase(name, expectedPrefixContent, expectedTimeoutContent string,
	timeout time.Duration) assertWaitForConditionErrorMessageTestCase {
	return assertWaitForConditionErrorMessageTestCase{
		name:                   name,
		expectedPrefixContent:  expectedPrefixContent,
		expectedTimeoutContent: expectedTimeoutContent,
		timeout:                timeout,
	}
}

func TestAssertWaitForConditionErrorMessage(t *testing.T) {
	testCases := []assertWaitForConditionErrorMessageTestCase{
		newAssertWaitForConditionErrorMessageTestCase("test formatted error message",
			"operation test", "did not become true", 10*time.Millisecond),
	}

	core.RunTestCases(t, testCases)
}

// assertWaitForConditionEmptyNameTestCase tests AssertWaitForCondition with empty name
// Fields ordered for memory efficiency (large to small)
type assertWaitForConditionEmptyNameTestCase struct {
	// 16 bytes (string header)
	name                   string
	expectedTimeoutContent string
	excludedContent        string
	// 8 bytes (time.Duration)
	timeout time.Duration
}

func (tc assertWaitForConditionEmptyNameTestCase) Name() string {
	return tc.name
}

func (tc assertWaitForConditionEmptyNameTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &core.MockT{}
	falseCondition := func() bool { return false }

	// Test with empty name
	AssertWaitForCondition(mock, falseCondition, tc.timeout, "")

	core.AssertTrue(t, mock.HasErrors(), "should have error")
	errs := mock.Errors
	core.AssertTrue(t, len(errs) > 0, "should have error messages")

	errorMsg := errs[0]
	core.AssertContains(t, errorMsg, tc.expectedTimeoutContent, "should contain timeout message")
	core.AssertFalse(t, strings.Contains(errorMsg, tc.excludedContent), "should not have colon prefix")
}

// Factory function for assertWaitForConditionEmptyNameTestCase
func newAssertWaitForConditionEmptyNameTestCase(name, expectedTimeoutContent, excludedContent string,
	timeout time.Duration) assertWaitForConditionEmptyNameTestCase {
	return assertWaitForConditionEmptyNameTestCase{
		name:                   name,
		expectedTimeoutContent: expectedTimeoutContent,
		excludedContent:        excludedContent,
		timeout:                timeout,
	}
}

func TestAssertWaitForConditionEmptyName(t *testing.T) {
	testCases := []assertWaitForConditionEmptyNameTestCase{
		newAssertWaitForConditionEmptyNameTestCase("test empty name formatting",
			"did not become true", ": condition", 10*time.Millisecond),
	}

	core.RunTestCases(t, testCases)
}

// assertFieldTypeIsTestCase tests AssertFieldTypeIs function
// Fields ordered for memory efficiency (large to small)
type assertFieldTypeIsTestCase struct {
	// 16+ bytes (map, interface)
	input    map[string]any
	expected any
	// 16 bytes (string headers)
	name     string
	field    string
	testName string
	// 1 byte (bool)
	wantOK bool
}

func (tc assertFieldTypeIsTestCase) Name() string {
	return tc.testName
}

func (tc assertFieldTypeIsTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &core.MockT{}
	result, ok := AssertFieldTypeIs[string](mock, tc.input, tc.field, tc.name)

	core.AssertEqual(t, tc.wantOK, ok, "field type assertion")

	if tc.wantOK {
		tc.validateSuccessCase(t, mock, result)
	} else {
		tc.validateFailureCase(t, mock)
	}
}

func (tc assertFieldTypeIsTestCase) validateSuccessCase(t *testing.T, mock *core.MockT, result string) {
	t.Helper()
	core.AssertFalse(t, mock.HasErrors(), "should not have errors on success")
	if expectedStr, ok := tc.expected.(string); ok {
		core.AssertEqual(t, expectedStr, result, "field value")
	}
}

func (tc assertFieldTypeIsTestCase) validateFailureCase(t *testing.T, mock *core.MockT) {
	t.Helper()
	core.AssertTrue(t, mock.HasErrors(), "should have errors on failure")
	tc.validateErrorMessage(t, mock)
}

func (tc assertFieldTypeIsTestCase) validateErrorMessage(t *testing.T, mock *core.MockT) {
	t.Helper()
	if len(mock.Errors) == 0 || tc.field == "" || tc.input == nil {
		return
	}

	errorMsg := mock.Errors[0]
	if _, exists := tc.input[tc.field]; !exists {
		core.AssertContains(t, errorMsg, "not found", "missing field error")
	} else {
		core.AssertContains(t, errorMsg, "type", "type mismatch error")
	}
}

// Factory function for assertFieldTypeIsTestCase
//
//revive:disable-next-line:argument-limit
func newAssertFieldTypeIsTestCase(testName, name, field string, input map[string]any,
	expected any, wantOK bool) assertFieldTypeIsTestCase {
	return assertFieldTypeIsTestCase{
		testName: testName,
		name:     name,
		field:    field,
		input:    input,
		expected: expected,
		wantOK:   wantOK,
	}
}

// Convenience factory for successful string field assertions
func newAssertFieldTypeIsTestCaseSuccess(testName, field, expectedValue string) assertFieldTypeIsTestCase {
	return newAssertFieldTypeIsTestCase(
		testName,
		"field",
		field,
		map[string]any{field: expectedValue},
		expectedValue,
		true,
	)
}

// Convenience factory for missing field assertions
func newAssertFieldTypeIsTestCaseMissing(testName, field string) assertFieldTypeIsTestCase {
	return newAssertFieldTypeIsTestCase(
		testName,
		"field",
		field,
		map[string]any{"other": "value"},
		"",
		false,
	)
}

// Convenience factory for type mismatch assertions
func newAssertFieldTypeIsTestCaseTypeMismatch(testName, field string, wrongTypeValue any) assertFieldTypeIsTestCase {
	return newAssertFieldTypeIsTestCase(
		testName,
		"field",
		field,
		map[string]any{field: wrongTypeValue},
		"",
		false,
	)
}

func assertFieldTypeIsTestCases() []assertFieldTypeIsTestCase {
	return []assertFieldTypeIsTestCase{
		// Successful cases
		newAssertFieldTypeIsTestCaseSuccess("string field exists", "name", "test"),
		newAssertFieldTypeIsTestCaseSuccess("empty string field", "empty", ""),
		newAssertFieldTypeIsTestCase("string with custom name", "test field", "value",
			map[string]any{"value": "hello"}, "hello", true),

		// Missing field cases
		newAssertFieldTypeIsTestCaseMissing("missing field", "missing"),
		newAssertFieldTypeIsTestCase("nil map", "field", "key",
			nil, "", false),
		newAssertFieldTypeIsTestCase("empty map", "field", "key",
			map[string]any{}, "", false),

		// Type mismatch cases
		newAssertFieldTypeIsTestCaseTypeMismatch("int instead of string", "age", 25),
		newAssertFieldTypeIsTestCaseTypeMismatch("bool instead of string", "flag", true),
		newAssertFieldTypeIsTestCaseTypeMismatch("float instead of string", "price", 19.99),
		newAssertFieldTypeIsTestCaseTypeMismatch("slice instead of string", "list", []string{"a", "b"}),
		newAssertFieldTypeIsTestCaseTypeMismatch("map instead of string", "nested", map[string]int{"key": 1}),

		// Special cases
		newAssertFieldTypeIsTestCase("nil value in map", "field", "nil_key",
			map[string]any{"nil_key": nil}, "", false),
		newAssertFieldTypeIsTestCase("interface{} conversion", "field", "interface",
			map[string]any{"interface": "string_value"}, "string_value", true),
	}
}

func TestAssertFieldTypeIs(t *testing.T) {
	core.RunTestCases(t, assertFieldTypeIsTestCases())
}
