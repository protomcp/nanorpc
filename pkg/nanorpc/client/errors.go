package client

import "darvaza.org/core"

// Invalid-argument sentinels for the client package. Each wraps
// [core.ErrInvalid], so a caller can match a specific cause or the whole
// family via [IsInvalid]. Call sites add dynamic context by wrapping the
// sentinel, e.g. core.QuietWrap(ErrNoSubscription, "request_id %d", id).
var (
	// ErrNilRequest indicates Send was called with a nil request.
	ErrNilRequest = core.QuietWrap(core.ErrInvalid, "nil request")

	// ErrMissingCallback indicates a request type that requires a callback
	// was sent without one.
	ErrMissingCallback = core.QuietWrap(core.ErrInvalid, "missing callback")

	// ErrInvalidRequestType indicates an unsupported request type.
	ErrInvalidRequestType = core.QuietWrap(core.ErrInvalid, "invalid request type")

	// ErrNoSubscription indicates an unsubscribe targeted a request_id with
	// no registered subscription.
	ErrNoSubscription = core.QuietWrap(core.ErrInvalid, "no matching subscription")

	// ErrSubscriptionPending indicates an unsubscribe targeted a
	// subscription that has not yet been acknowledged.
	ErrSubscriptionPending = core.QuietWrap(core.ErrInvalid, "subscription not yet active")

	// ErrNoSession indicates no session is currently attached.
	ErrNoSession = core.QuietWrap(core.ErrInvalid, "missing session")

	// ErrSessionAttached indicates a session is already attached.
	ErrSessionAttached = core.QuietWrap(core.ErrInvalid, "session already attached")

	// ErrMissingClient indicates a nil client was passed to a helper.
	ErrMissingClient = core.QuietWrap(core.ErrInvalid, "client missing")

	// ErrMissingOut indicates a nil out message was passed to a helper.
	ErrMissingOut = core.QuietWrap(core.ErrInvalid, "out missing")

	// ErrMissingNewOut indicates a nil newOut factory was passed to Subscribe.
	ErrMissingNewOut = core.QuietWrap(core.ErrInvalid, "newOut missing")

	// ErrNilOut indicates the newOut factory returned a nil message.
	ErrNilOut = core.QuietWrap(core.ErrInvalid, "newOut returned nil")
)

// IsInvalid reports whether err is an invalid-argument error. It matches
// [core.ErrInvalid] — the base the package's sentinels wrap, and itself an
// alias of [fs.ErrInvalid] / [os.ErrInvalid] — anywhere in the chain.
func IsInvalid(err error) bool {
	return core.IsErrorFn(checkIsInvalid, err)
}

func checkIsInvalid(err error) bool {
	return err == core.ErrInvalid
}
