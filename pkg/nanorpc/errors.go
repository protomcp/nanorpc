package nanorpc

import (
	"errors"
	"fmt"
	"io/fs"

	"darvaza.org/core"
)

var (
	// ErrNoResponse indicates the server didn't answer before disconnection
	ErrNoResponse = core.NewTimeoutError(errors.New("no response"))
)

// ResponseAsError extracts an error from the
// status of a response.
func ResponseAsError(res *NanoRPCResponse) error {
	if res != nil {
		switch res.ResponseStatus {
		case NanoRPCResponse_STATUS_OK:
			return nil
		case NanoRPCResponse_STATUS_NOT_FOUND:
			return fs.ErrNotExist
		case NanoRPCResponse_STATUS_NOT_AUTHORIZED:
			return fs.ErrPermission
		default:
			return fmt.Errorf("invalid state %v", int(res.ResponseStatus))
		}
	}
	return ErrNoResponse
}

// IsNotFound checks if the error represents a STATUS_NOT_FOUND response.
func IsNotFound(err error) bool {
	return core.IsError(err, fs.ErrNotExist)
}

// IsNotAuthorized checks if the error represents a STATUS_NOT_AUTHORIZED response.
func IsNotAuthorized(err error) bool {
	return core.IsError(err, fs.ErrPermission)
}

// IsNoResponse checks if the error is [ErrNoResponse].
func IsNoResponse(err error) bool {
	return core.IsError(err, ErrNoResponse)
}
