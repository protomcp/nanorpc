package nanorpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"darvaza.org/core"
)

func TestEncodeDecodeRequest(t *testing.T) {
	for i, req := range []*NanoRPCRequest{
		{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_PING,
		},
	} {
		doTestEncodeDecodeRequest(t, i, req)
	}
}

func doTestEncodeDecodeRequest(t *testing.T, i int, req1 *NanoRPCRequest) {
	b1, err := EncodeRequest(req1, nil)
	if err != nil {
		t.Errorf("[%v]: ERROR: EncodeRequest -> %v", i, err)
		return
	}

	b2 := core.SliceCopy(b1)
	req2, n, err := DecodeRequest(b2)
	switch {
	case err != nil:
		t.Errorf("[%v]: ERROR: DecodeRequest(%q) -> %v", i, b1, err)
		return
	case len(b1) != n:
		t.Errorf("[%v]: ERROR: DecodeRequest(%q) -> %v", i, b1, n)
		return
	}

	j1, err := json.Marshal(req1)
	if err != nil {
		t.Errorf("[%v]: ERROR: json.Marshal -> %v", i, err)
		return
	}

	j2, err := json.Marshal(req2)
	if err != nil {
		t.Errorf("[%v]: ERROR: json.Marshal -> %v", i, err)
		return
	}

	if !bytes.Equal(j1, j2) {
		t.Errorf("[%v]: ERROR: %q != %q", i, j1, j2)
		return
	}

	t.Logf("[%v]: %q", i, b1)
}
