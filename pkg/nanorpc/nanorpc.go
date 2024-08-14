package nanorpc

import (
	"bytes"
	"io"
	"math"
	"os"

	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"darvaza.org/core"
)

// DecodeResponse attempts to decode a wrapped NanoRPC response
// from a buffer.
func DecodeResponse(data []byte) (*NanoRPCResponse, int, error) {
	from, to, err := DecodeSplit(data)
	if err != nil {
		return nil, 0, err
	}

	msg := data[from:to]
	out := new(NanoRPCResponse)
	if err = proto.Unmarshal(msg, out); err != nil {
		return nil, to, err
	}

	return out, to, nil
}

// DecodeResponseData attempts to decode the payload of a NanoRPC response.
func DecodeResponseData[T proto.Message](res *NanoRPCResponse, out T) (T, bool, error) {
	err := ResponseAsError(res)
	switch {
	case err != nil:
		return out, false, err
	case len(res.Data) == 0:
		return out, false, nil
	default:
		err = proto.Unmarshal(res.Data, out)
		return out, true, err
	}
}

// DecodeRequest attempts to decode a wrapped NanoRPC request
// from a buffer
func DecodeRequest(data []byte) (*NanoRPCRequest, int, error) {
	from, to, err := DecodeSplit(data)
	if err != nil {
		return nil, 0, err
	}

	msg := data[from:to]
	out := new(NanoRPCRequest)
	if err = proto.Unmarshal(msg, out); err != nil {
		return nil, to, err
	}

	return out, to, nil
}

// EncodeRequestTo encodes a wrapped NanoRPC request.
// If request data is provided, it will be encoded into the
// [NanoRPCRequest], otherwise the request will be used as-is.
func EncodeRequestTo(w io.Writer, req *NanoRPCRequest, data proto.Message) (int, error) {
	if data != nil {
		b, err := proto.Marshal(data)
		switch {
		case err != nil:
			return 0, err
		case len(b) == 0:
			req.Data = nil
		default:
			req.Data = b
		}
	}

	return protodelim.MarshalTo(w, req)
}

// EncodeRequest encodes a wrapped NanoRPC request.
// If request data is provided, it will be encoded into the
// [NanoRPCRequest], otherwise the request will be used as-is.
func EncodeRequest(req *NanoRPCRequest, data proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	_, err := EncodeRequestTo(&buf, req, data)
	return buf.Bytes(), err
}

// EncodeResponseTo encodes a wrapped NanoRPC response.
// If response data is provided, it will be encoded into the
// [NanoRPCResponse], otherwise the response will be used as-is.
func EncodeResponseTo(w io.Writer, res *NanoRPCResponse, data proto.Message) (int, error) {
	if data != nil {
		b, err := proto.Marshal(data)
		switch {
		case err != nil:
			return 0, err
		case len(b) == 0:
			res.Data = nil
		default:
			res.Data = b
		}
	}

	return protodelim.MarshalTo(w, res)
}

// EncodeResponse encodes a wrapped NanoRPC response.
// If response data is provided, it will be encoded into the
// [NanoRPCResponse], otherwise the response will be used as-is.
func EncodeResponse(res *NanoRPCResponse, data proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	_, err := EncodeResponseTo(&buf, res, data)
	return buf.Bytes(), err
}

// Split identifies a NanoRPC wrapped message from a buffer.
func Split(data []byte, atEOF bool) (advance int, msg []byte, err error) {
	_, n, err := DecodeSplit(data)

	switch {
	case err == io.ErrUnexpectedEOF && !atEOF:
		// more data needed
		return 0, nil, nil
	case err != nil:
		// bad data
		return 0, nil, err
	}

	return n, data[:n], nil
}

// DecodeSplit identifies the size of the wrapped message
// and if enough data is already buffered.
func DecodeSplit(data []byte) (prefixLen, totalLen int, err error) {
	size, prefixLen := protowire.ConsumeVarint(data)
	if err = protowire.ParseError(prefixLen); err != nil {
		return 0, 0, err
	}

	if size > math.MaxInt32 {
		err = core.Wrap(os.ErrInvalid, "size out of range: %v", size)
		return prefixLen, 0, err
	}

	totalLen = prefixLen + int(size)
	if len(data) < totalLen {
		err = io.ErrUnexpectedEOF
	}

	return prefixLen, totalLen, err
}
