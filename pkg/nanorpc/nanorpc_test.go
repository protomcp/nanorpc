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
	core.AssertNoError(t, err, "EncodeRequest")

	b2 := core.SliceCopy(b1)
	req2, n, err := DecodeRequest(b2)
	core.AssertNoError(t, err, "DecodeRequest")
	core.AssertEqual(t, len(b1), n, "DecodeRequest length")

	j1, err := json.Marshal(tc.request)
	core.AssertNoError(t, err, "json.Marshal original")

	j2, err := json.Marshal(req2)
	core.AssertNoError(t, err, "json.Marshal decoded")

	core.AssertTrue(t, bytes.Equal(j1, j2), "Request match", "original %q != decoded %q", j1, j2)

	t.Logf("Encoded: %q", b1)
}

func newBasicRequestTestCase(name string, request *NanoRPCRequest) basicRequestTestCase {
	return basicRequestTestCase{
		name:    name,
		request: request,
	}
}

func basicRequestTestCases() []basicRequestTestCase {
	return core.S(
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
