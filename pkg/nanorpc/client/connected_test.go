package client

import (
	"context"
	"testing"
	"time"

	"darvaza.org/core"
)

// newClientForTest builds a [Client] without starting reconnect. Tests use
// it for direct manipulation of the session lifecycle via setSession and
// endSession.
func newClientForTest(t *testing.T) *Client {
	t.Helper()
	cfg := Config{
		Context: context.Background(),
		Remote:  "127.0.0.1:1",
	}
	c, err := cfg.New()
	core.AssertMustNoError(t, err, "cfg.New")
	return c
}

// TestClient_IsConnected_initiallyFalse confirms a freshly constructed
// [Client] reports disconnected before any session has been attached.
func TestClient_IsConnected_initiallyFalse(t *testing.T) {
	c := newClientForTest(t)
	core.AssertFalse(t, c.IsConnected(), "IsConnected at construction")
}

// TestClient_Shutdown_userInitiatedReportsNoError verifies the documented
// clean-stop contract: a caller-initiated Shutdown returns nil and leaves
// Err nil, since the reconnect cause is only non-nil when the loop itself
// failed. A freshly constructed client has no workers running, so Shutdown
// drains at once — the bounded ctx guards against a regression that hangs.
func TestClient_Shutdown_userInitiatedReportsNoError(t *testing.T) {
	c := newClientForTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	core.AssertNoError(t, c.Shutdown(ctx), "Shutdown")
	core.AssertNoError(t, c.Err(), "Err after user-initiated Shutdown")
}

// TestClient_Connected_opensThenCloses verifies that the readiness
// channel is open before a session is attached and closes once
// setSession succeeds.
func TestClient_Connected_opensThenCloses(t *testing.T) {
	c := newClientForTest(t)
	ch := c.Connected()

	select {
	case <-ch:
		t.Fatal("Connected channel was closed without a session")
	default:
	}

	core.AssertMustNoError(t, c.setSession(&Session{}), "setSession")
	core.AssertTrue(t, c.IsConnected(), "IsConnected after setSession")

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("Connected channel did not close after setSession")
	}
}

// TestClient_Connected_swapsAfterEndSession verifies that endSession
// installs a fresh open readiness channel, so the next caller of
// Connected sees an open channel even though a prior session had
// closed the previous one.
func TestClient_Connected_swapsAfterEndSession(t *testing.T) {
	c := newClientForTest(t)
	core.AssertMustNoError(t, c.setSession(&Session{}), "setSession")
	prev := c.Connected()
	c.endSession(nil)

	core.AssertFalse(t, c.IsConnected(), "IsConnected after endSession")
	core.AssertNotSame(t, prev, c.Connected(),
		"Connected channel after endSession")
	select {
	case <-c.Connected():
		t.Fatal("Connected channel was closed after endSession")
	default:
	}
}

// TestClient_endSession_preservesChannelWhenDisconnected verifies that
// calling endSession while already disconnected keeps the live open
// channel, so a waiter holding it is not orphaned by a spurious reset.
func TestClient_endSession_preservesChannelWhenDisconnected(t *testing.T) {
	c := newClientForTest(t)
	prev := c.Connected()

	c.endSession(nil)

	core.AssertFalse(t, c.IsConnected(), "IsConnected after spurious endSession")
	core.AssertSame(t, prev, c.Connected(),
		"Connected channel preserved across spurious endSession")
	select {
	case <-c.Connected():
		t.Fatal("Connected channel was closed by spurious endSession")
	default:
	}
}

// TestClient_endSession_ignoresStaleSession verifies the defence-in-depth
// identity check: ending a session other than the one currently attached
// is a no-op, leaving the live session and its readiness channel intact.
func TestClient_endSession_ignoresStaleSession(t *testing.T) {
	c := newClientForTest(t)
	live := &Session{}
	core.AssertMustNoError(t, c.setSession(live), "setSession")
	ch := c.Connected()

	c.endSession(&Session{}) // a different, already-replaced session

	core.AssertTrue(t, c.IsConnected(), "IsConnected after stale endSession")
	core.AssertSame(t, live, c.cs, "attached session unchanged")
	core.AssertSame(t, ch, c.Connected(),
		"Connected channel preserved across stale endSession")
}

// TestClient_WaitConnected_connectedBeatsCancelledCtx verifies the fast
// path wins deterministically: an already-connected client returns nil
// even when ctx is already cancelled, where a lone select could otherwise
// choose the ctx arm pseudo-randomly.
func TestClient_WaitConnected_connectedBeatsCancelledCtx(t *testing.T) {
	c := newClientForTest(t)
	core.AssertMustNoError(t, c.setSession(&Session{}), "setSession")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // both the readiness channel and ctx are now ready

	core.AssertNoError(t, c.WaitConnected(ctx),
		"WaitConnected connected with cancelled ctx")
}

// TestClient_WaitConnected_alreadyConnected exercises the fast path:
// if a session is already attached, WaitConnected returns nil without
// blocking.
func TestClient_WaitConnected_alreadyConnected(t *testing.T) {
	c := newClientForTest(t)
	core.AssertMustNoError(t, c.setSession(&Session{}), "setSession")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	core.AssertNoError(t, c.WaitConnected(ctx),
		"WaitConnected when already connected")
}

// TestClient_WaitConnected_blocksThenSucceeds verifies WaitConnected
// blocks until a session is established, then returns nil.
func TestClient_WaitConnected_blocksThenSucceeds(t *testing.T) {
	c := newClientForTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- c.WaitConnected(ctx) }()

	// Give the goroutine time to enter the select before transitioning.
	time.Sleep(50 * time.Millisecond)

	// WaitConnected must still be blocked: no session is attached yet.
	select {
	case <-done:
		t.Fatal("WaitConnected returned before a session was attached")
	default:
	}

	core.AssertMustNoError(t, c.setSession(&Session{}), "setSession")

	select {
	case err := <-done:
		core.AssertNoError(t, err, "WaitConnected after setSession")
	case <-time.After(time.Second):
		t.Fatal("WaitConnected did not return after setSession")
	}
}

// TestClient_WaitConnected_ctxCancellation verifies WaitConnected
// surfaces the caller's context error when ctx fires before any
// session is established.
func TestClient_WaitConnected_ctxCancellation(t *testing.T) {
	c := newClientForTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := c.WaitConnected(ctx)
	core.AssertErrorIs(t, err, context.DeadlineExceeded,
		"WaitConnected ctx error")
}
