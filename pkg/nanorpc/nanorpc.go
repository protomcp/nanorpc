package nanorpc

import (
	"io"
	"math"
	"os"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"darvaza.org/core"
)

func unmarshal(data []byte, out proto.Message) (int, error) {
	prefixLen, payloadLen, err := DecodeSplit(data)
	if err != nil {
		return 0, err
	}

	begin := prefixLen
	end := prefixLen + payloadLen

	err = proto.Unmarshal(data[begin:end], out)
	return end, err
}

// DecodeResponse attempts to decode a wrapped NanoRPC response
// from a buffer.
func DecodeResponse(data []byte) (*NanoRPCResponse, int, error) {
	out := new(NanoRPCResponse)
	size, err := unmarshal(data, out)
	if err != nil {
		return nil, 0, err
	}
	return out, size, nil
}

// DecodeRequest attempts to decode a wrapped NanoRPC request
// from a buffer
func DecodeRequest(data []byte) (*NanoRPCRequest, error) {
	prefixLen, _, err := DecodeSplit(data)
	if err != nil {
		return nil, err
	}

	msg := data[prefixLen:]
	out := new(NanoRPCRequest)
	if err = proto.Unmarshal(msg, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Split identifies a NanoRPC wrapped message from a buffer.
func Split(data []byte, atEOF bool) (advance int, msg []byte, err error) {
	prefixLen, payloadLen, err := DecodeSplit(data)

	switch {
	case err == io.ErrUnexpectedEOF && !atEOF:
		// more data needed
		return 0, nil, nil
	case err != nil:
		// bad data
		return 0, nil, err
	}

	n := prefixLen + payloadLen
	return n, data[:n], nil
}

// DecodeSplit identifies the size of the wrapping and payload of a message,
// and if enough data is already buffered.
// If
func DecodeSplit(data []byte) (prefixLen, payloadLen int, err error) {
	// <0,LEN>
	//
	tag, tagLen := protowire.ConsumeVarint(data)
	if err = protowire.ParseError(tagLen); err != nil {
		return 0, 0, err
	}
	prefixLen += tagLen

	tagNum, tagType := protowire.DecodeTag(tag)
	if tagNum != 0 || tagType != protowire.BytesType {
		err := core.Wrap(os.ErrInvalid, "unexpected tag: <%v,%v>", tagNum, tagType)
		return prefixLen, 0, err
	}

	// payload size
	//
	size, sizeLen := protowire.ConsumeVarint(data[tagLen:])
	if err = protowire.ParseError(sizeLen); err != nil {
		return prefixLen, 0, err
	}
	prefixLen += sizeLen

	if size > math.MaxInt32 {
		err = core.Wrap(os.ErrInvalid, "size out of range: %v", size)
		return prefixLen, 0, err
	}

	// payload
	//
	payloadLen = int(size)

	if len(data) < (prefixLen + payloadLen) {
		err = io.ErrUnexpectedEOF
	}

	return prefixLen, payloadLen, err
}
