package nanorpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"darvaza.org/core"
)

type basicRequestTestCase struct {
	request *NanoRPCRequest
	name    string
}

func (tc basicRequestTestCase) test(t *testing.T) {
	helper := NewEncodeDecodeTestHelper(t)
	helper.TestRequestRoundTrip(tc.request, nil)

	// Also test the original JSON comparison for backward compatibility
	b1, err := EncodeRequest(tc.request, nil)
	AssertNoError(t, err, "EncodeRequest failed")

	b2 := core.SliceCopy(b1)
	req2, n, err := DecodeRequest(b2)
	AssertNoError(t, err, "DecodeRequest failed")
	AssertEqual(t, len(b1), n, "DecodeRequest length mismatch")

	j1, err := json.Marshal(tc.request)
	AssertNoError(t, err, "json.Marshal original failed")

	j2, err := json.Marshal(req2)
	AssertNoError(t, err, "json.Marshal decoded failed")

	AssertTrue(t, bytes.Equal(j1, j2), "Request mismatch: original %q != decoded %q", j1, j2)

	t.Logf("Encoded: %q", b1)
}

func newBasicRequestTestCase(name string, request *NanoRPCRequest) basicRequestTestCase {
	return basicRequestTestCase{
		name:    name,
		request: request,
	}
}

func basicRequestTestCases() []basicRequestTestCase {
	return S(
		newBasicRequestTestCase("ping_request", &NanoRPCRequest{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_PING,
		}),
	)
}

func TestEncodeDecodeRequest(t *testing.T) {
	for _, tc := range basicRequestTestCases() {
		t.Run(tc.name, tc.test)
	}
}
