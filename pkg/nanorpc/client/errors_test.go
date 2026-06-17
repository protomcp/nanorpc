package client

import (
	"errors"
	"os"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = isInvalidTestCase{}

// isInvalidTestCase exercises IsInvalid against both invalid-argument
// bases, a package sentinel, a context-wrapped sentinel, and unrelated
// errors.
type isInvalidTestCase struct {
	err  error
	name string
	want bool
}

func (tc isInvalidTestCase) Name() string { return tc.name }

func (tc isInvalidTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, IsInvalid(tc.err), "IsInvalid")
}

func newIsInvalidTestCase(name string, err error, want bool) isInvalidTestCase {
	return isInvalidTestCase{err: err, name: name, want: want}
}

func isInvalidTestCases() []isInvalidTestCase {
	return []isInvalidTestCase{
		newIsInvalidTestCase("nil", nil, false),
		newIsInvalidTestCase("core_base", core.ErrInvalid, true),
		newIsInvalidTestCase("os_base", os.ErrInvalid, true),
		newIsInvalidTestCase("sentinel", ErrNilRequest, true),
		newIsInvalidTestCase("wrapped_sentinel",
			core.QuietWrap(ErrNoSubscription, "request_id %d", 5), true),
		newIsInvalidTestCase("unrelated_sentinel", core.ErrUnknown, false),
		newIsInvalidTestCase("plain", errors.New("boom"), false),
	}
}

// TestIsInvalid exercises the invalid-argument family predicate.
func TestIsInvalid(t *testing.T) {
	core.RunTestCases(t, isInvalidTestCases())
}
