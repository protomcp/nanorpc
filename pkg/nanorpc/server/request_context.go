package server

import (
	"encoding/json"
	"errors"

	"darvaza.org/core"
	"google.golang.org/protobuf/proto"

	"github.com/amery/nanorpc/pkg/nanorpc"
)

// SendOK sends a successful response with optional data
func (rc *RequestContext) SendOK(data []byte) error {
	if rc == nil {
		return core.ErrNilReceiver
	}

	response := &nanorpc.NanoRPCResponse{
		RequestId:      rc.Request.RequestId,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
		Data:           data,
	}

	return rc.Session.SendResponse(rc.Request, response)
}

// SendError sends an error response with the specified status and message
func (rc *RequestContext) SendError(status nanorpc.NanoRPCResponse_Status, message string) error {
	if rc == nil {
		return core.ErrNilReceiver
	}

	// Ensure we don't use STATUS_OK for errors
	if status == nanorpc.NanoRPCResponse_STATUS_OK {
		status = nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR
	}

	response := &nanorpc.NanoRPCResponse{
		RequestId:       rc.Request.RequestId,
		ResponseType:    nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus:  status,
		ResponseMessage: message,
	}

	return rc.Session.SendResponse(rc.Request, response)
}

// SendNotFound sends a STATUS_NOT_FOUND response
func (rc *RequestContext) SendNotFound(message string) error {
	if message == "" {
		message = "resource not found"
	}
	return rc.SendError(nanorpc.NanoRPCResponse_STATUS_NOT_FOUND, message)
}

// SendBadRequest sends a STATUS_INTERNAL_ERROR response (treating as bad request)
// Note: The protocol doesn't have STATUS_BAD_REQUEST, so we use INTERNAL_ERROR
func (rc *RequestContext) SendBadRequest(message string) error {
	if message == "" {
		message = "bad request"
	}
	return rc.SendError(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR, message)
}

// SendUnauthorized sends a STATUS_NOT_AUTHORIZED response
func (rc *RequestContext) SendUnauthorized(message string) error {
	if message == "" {
		message = "not authorized"
	}
	return rc.SendError(nanorpc.NanoRPCResponse_STATUS_NOT_AUTHORIZED, message)
}

// SendInternalError sends a STATUS_INTERNAL_ERROR response
func (rc *RequestContext) SendInternalError(message string) error {
	if message == "" {
		message = "internal server error"
	}
	return rc.SendError(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR, message)
}

// SendJSON marshals the value as JSON and sends it as a successful response
func (rc *RequestContext) SendJSON(v any) error {
	if rc == nil {
		return core.ErrNilReceiver
	}

	data, err := json.Marshal(v)
	if err != nil {
		return core.Wrapf(err, "failed to marshal JSON response")
	}

	return rc.SendOK(data)
}

// SendProtobuf marshals the protobuf message and sends it as a successful response
func (rc *RequestContext) SendProtobuf(msg proto.Message) error {
	if rc == nil {
		return core.ErrNilReceiver
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return core.Wrapf(err, "failed to marshal protobuf response")
	}

	return rc.SendOK(data)
}

// UnmarshalRequestJSON decodes the request data as JSON into v
func (rc *RequestContext) UnmarshalRequestJSON(v any) error {
	if rc == nil {
		return core.ErrNilReceiver
	}

	if len(rc.Request.Data) == 0 {
		return errors.New("request has no data")
	}

	if err := json.Unmarshal(rc.Request.Data, v); err != nil {
		return core.Wrapf(err, "failed to unmarshal JSON request")
	}

	return nil
}

// UnmarshalRequestProtobuf decodes the request data as protobuf into msg
func (rc *RequestContext) UnmarshalRequestProtobuf(msg proto.Message) error {
	if rc == nil {
		return core.ErrNilReceiver
	}

	if len(rc.Request.Data) == 0 {
		return errors.New("request has no data")
	}

	if err := proto.Unmarshal(rc.Request.Data, msg); err != nil {
		return core.Wrapf(err, "failed to unmarshal protobuf request")
	}

	return nil
}

// GetRequestID returns the request ID
func (rc *RequestContext) GetRequestID() int32 {
	if rc == nil || rc.Request == nil {
		return 0
	}
	return rc.Request.RequestId
}

// GetData returns the request data
func (rc *RequestContext) GetData() []byte {
	if rc == nil || rc.Request == nil {
		return nil
	}
	return rc.Request.Data
}

// HasData returns true if the request has data
func (rc *RequestContext) HasData() bool {
	return rc != nil && rc.Request != nil && len(rc.Request.Data) > 0
}
