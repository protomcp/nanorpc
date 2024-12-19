package nanorpc

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"darvaza.org/core"
)

var (
	// ErrNoResponse indicates the server didn't answer before disconnection
	ErrNoResponse = core.NewTimeoutError(errors.New("no response"))

	// ErrInternalServerError indicates the server reported an internal error
	ErrInternalServerError = errors.New("internal server error")
)

var (
	_ error            = (*ResponseError)(nil)
	_ fmt.Stringer     = (*ResponseError)(nil)
	_ core.Unwrappable = (*ResponseError)(nil)
)

// ResponseError represents a NanoRPC error response.
type ResponseError struct {
	Status NanoRPCResponse_Status
	Err    error
	Msg    string
}

func (e ResponseError) Error() string {
	return e.String()
}

func (e ResponseError) Unwrap() error {
	return e.Err
}

func (e ResponseError) String() string {
	status, ok := strings.CutPrefix(e.Status.String(), "STATUS_")
	switch {
	case !ok, e.Err == core.ErrUnknown:
		status = fmt.Sprintf("unknown status %d", e.Status)
	case e.Status == NanoRPCResponse_STATUS_UNSPECIFIED:
		status = fmt.Sprintf("%s: invalid status", status)
	}

	if e.Msg == "" {
		return fmt.Sprintf("nanorpc: %s", status)
	}

	return fmt.Sprintf("nanorpc: %s: %s", status, e.Msg)
}

// ResponseAsError extracts an error from the
// status of a response.
func ResponseAsError(res *NanoRPCResponse) error {
	var err error

	if res == nil {
		return ErrNoResponse
	}

	switch res.ResponseStatus {
	case NanoRPCResponse_STATUS_OK:
		return nil
	case NanoRPCResponse_STATUS_NOT_FOUND:
		err = fs.ErrNotExist
	case NanoRPCResponse_STATUS_NOT_AUTHORIZED:
		err = fs.ErrPermission
	case NanoRPCResponse_STATUS_INTERNAL_ERROR:
		err = ErrInternalServerError
	case NanoRPCResponse_STATUS_UNSPECIFIED:
		err = core.ErrInvalid
	default:
		err = core.ErrUnknown
	}

	return &ResponseError{
		Status: res.ResponseStatus,
		Msg:    res.ResponseMessage,
		Err:    err,
	}
}

// IsNotFound checks if the error represents a STATUS_NOT_FOUND response.
func IsNotFound(err error) bool {
	return core.IsError(err, fs.ErrNotExist)
}

// IsNotAuthorized checks if the error represents a STATUS_NOT_AUTHORIZED response.
func IsNotAuthorized(err error) bool {
	return core.IsError(err, fs.ErrPermission)
}

// IsNoResponse checks if the error represents no response being received.
// This error is also used to notify the connection was closed.
func IsNoResponse(err error) bool {
	return core.IsError(err, ErrNoResponse)
}
