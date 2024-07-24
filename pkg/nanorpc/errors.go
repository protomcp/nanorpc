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
