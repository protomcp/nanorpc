package nanorpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"darvaza.org/core"
)

// BasicRequestTestCase represents a test case for basic request encoding/decoding
type BasicRequestTestCase struct {
	name    string
	request *NanoRPCRequest
}

func (tc BasicRequestTestCase) test(t *testing.T) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		b1, err := EncodeRequest(tc.request, nil)
		if err != nil {
			t.Errorf("EncodeRequest failed: %v", err)
			return
		}

		b2 := core.SliceCopy(b1)
		req2, n, err := DecodeRequest(b2)
		switch {
		case err != nil:
			t.Errorf("DecodeRequest(%q) failed: %v", b1, err)
			return
		case len(b1) != n:
			t.Errorf("DecodeRequest length mismatch: expected %d, got %d", len(b1), n)
			return
		}

		j1, err := json.Marshal(tc.request)
		if err != nil {
			t.Errorf("json.Marshal original failed: %v", err)
			return
		}

		j2, err := json.Marshal(req2)
		if err != nil {
			t.Errorf("json.Marshal decoded failed: %v", err)
			return
		}

		if !bytes.Equal(j1, j2) {
			t.Errorf("Request mismatch: original %q != decoded %q", j1, j2)
			return
		}

		t.Logf("Encoded: %q", b1)
	})
}

// Test cases for basic request encoding/decoding
var basicRequestTestCases = []BasicRequestTestCase{
	{
		name: "ping_request",
		request: &NanoRPCRequest{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_PING,
		},
	},
}

func TestEncodeDecodeRequest(t *testing.T) {
	for _, tc := range basicRequestTestCases {
		tc.test(t)
	}
}
