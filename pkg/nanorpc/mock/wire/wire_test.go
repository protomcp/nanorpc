package wire_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc/mock/wire"
)

// lineConfig builds a newline-delimited string codec, exercising the generic
// Peer without coupling the engine's test to the nanorpc codec.
func lineConfig(conn net.Conn) wire.Config[string, string] {
	return wire.Config[string, string]{
		Conn:  conn,
		Split: bufio.ScanLines,
		Encode: func(w io.Writer, msg string) error {
			_, err := io.WriteString(w, msg+"\n")
			return err
		},
		Decode: func(data []byte) (string, error) {
			return string(data), nil
		},
	}
}

// Compile-time check that the type implements core.TestCase.
var _ core.TestCase = nilConfigTestCase{}

// nilConfigTestCase mutates one required field of an otherwise valid Config
// to nil and asserts New rejects it.
type nilConfigTestCase struct {
	mutate func(*wire.Config[string, string])
	name   string
}

func (tc nilConfigTestCase) Name() string { return tc.name }

func (tc nilConfigTestCase) Test(t *testing.T) {
	t.Helper()

	a, b := net.Pipe()
	defer func() { _ = a.Close() }()
	defer func() { _ = b.Close() }()

	cfg := lineConfig(a)
	tc.mutate(&cfg)

	err := core.Catch(func() error {
		wire.New(cfg)
		return nil
	})
	core.AssertErrorIs(t, err, core.ErrInvalid, "%s rejected", tc.name)
}

func newNilConfigTestCase(name string,
	mutate func(*wire.Config[string, string])) nilConfigTestCase {
	return nilConfigTestCase{name: name, mutate: mutate}
}

func nilConfigTestCases() []nilConfigTestCase {
	return core.S(
		newNilConfigTestCase("nil Conn",
			func(c *wire.Config[string, string]) { c.Conn = nil }),
		newNilConfigTestCase("nil Split",
			func(c *wire.Config[string, string]) { c.Split = nil }),
		newNilConfigTestCase("nil Encode",
			func(c *wire.Config[string, string]) { c.Encode = nil }),
		newNilConfigTestCase("nil Decode",
			func(c *wire.Config[string, string]) { c.Decode = nil }),
	)
}

// TestNew_nilFieldRejected confirms New rejects a Config missing a required
// field synchronously, panicking with a core.ErrInvalid-wrapped payload a
// recovering caller can match, rather than deferring the failure to an async
// reader panic or a later Send.
func TestNew_nilFieldRejected(t *testing.T) {
	core.RunTestCases(t, nilConfigTestCases())
}

// TestPeer_roundTrip sends a message from one peer to another over a
// net.Pipe and confirms a clean shutdown reports no error.
func TestPeer_roundTrip(t *testing.T) {
	a, b := net.Pipe()
	pa := wire.New(lineConfig(a))
	pb := wire.New(lineConfig(b))

	core.AssertMustNoError(t, pa.Send("hello"), "send")
	core.AssertEqual(t, "hello", <-pb.Recv(), "received")

	core.AssertNoError(t, pa.Close(), "close a")
	core.AssertNoError(t, pb.Close(), "close b")
}

// TestPeer_closeIsIdempotent confirms a second Close is a no-op that still
// reports the clean outcome.
func TestPeer_closeIsIdempotent(t *testing.T) {
	a, b := net.Pipe()
	defer func() { _ = b.Close() }()
	pa := wire.New(lineConfig(a))

	core.AssertNoError(t, pa.Close(), "first close")
	core.AssertNoError(t, pa.Close(), "second close")
}

// TestPeer_closeAfterContextCancel pins the deadlock fix: with a live
// Config.Context cancelled while the reader is blocked on Read, Close must
// still close the connection and return rather than block on Wait. A
// cancelled parent context flips the workgroup to its cancelled state on its
// own, so Close's Cancel loses the race; gating the connection close on
// winning that race — as the original code did — left Wait blocked on a
// reader nothing could unblock.
func TestPeer_closeAfterContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	a, b := net.Pipe()
	defer func() { _ = b.Close() }()

	cfg := lineConfig(a)
	cfg.Context = ctx
	p := wire.New(cfg)

	// Cancel while the reader is blocked on Read; cancellation cannot
	// interrupt it, so only Close's connection close can.
	cancel()

	done := make(chan error, 1)
	go func() { done <- p.Close() }()

	select {
	case err := <-done:
		core.AssertNoError(t, err, "close after context cancel")
	case <-time.After(time.Second):
		t.Fatal("Close deadlocked after context cancellation")
	}
}

// TestPeer_cancelledContext starts a peer whose parent context is already
// cancelled: the reader never starts, so Recv is closed immediately rather
// than blocking on a Read that cancellation cannot interrupt.
func TestPeer_cancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a, b := net.Pipe()
	defer func() { _ = b.Close() }()

	cfg := lineConfig(a)
	cfg.Context = ctx
	p := wire.New(cfg)

	_, ok := <-p.Recv()
	core.AssertFalse(t, ok, "Recv closed for a cancelled context")
	core.AssertNoError(t, p.Close(), "close")
}

// TestPeer_connEndedSwallowed confirms a connection-ended error from the
// reader is treated as a clean stop, not a fault. Closing the conn under the
// reader fails its Read with io.ErrClosedPipe — one of the connection-ended
// errors the catcher must swallow — so Wait and Close report nil. (A peer
// hang-up cannot stand in here: bufio.ScanLines reports that as a clean EOF;
// the nanorpc codec, which turns it into io.ErrUnexpectedEOF, is covered by
// the mock client and server tests.)
func TestPeer_connEndedSwallowed(t *testing.T) {
	a, b := net.Pipe()
	defer func() { _ = b.Close() }()

	p := wire.New(lineConfig(a))

	core.AssertMustNoError(t, a.Close(), "close conn under reader")

	core.AssertNoError(t, p.Wait(), "connection-ended error is not a fault")
	core.AssertNoError(t, p.Close(), "close after the connection ended")
}

// TestPeer_decodeFaultSurfaces confirms a genuine fault — a frame that splits
// cleanly but Decode rejects, distinct from a connection ending — is not
// swallowed: it becomes the cancellation cause that both Wait and Close
// report. Both are asserted via Wait first, which the workgroup completes only
// after the catcher has stored the cause, so the assertion does not race
// Close's own Cancel.
func TestPeer_decodeFaultSurfaces(t *testing.T) {
	errBadFrame := errors.New("bad frame")

	a, b := net.Pipe()
	defer func() { _ = b.Close() }()

	cfg := lineConfig(a)
	cfg.Decode = func([]byte) (string, error) { return "", errBadFrame }
	p := wire.New(cfg)

	_, err := io.WriteString(b, "boom\n")
	core.AssertMustNoError(t, err, "write frame")

	core.AssertErrorIs(t, p.Wait(), errBadFrame, "fault surfaces from Wait")
	core.AssertErrorIs(t, p.Close(), errBadFrame, "fault surfaces from Close")
}

// TestPeer_deliveryAbandonedOnCancel covers read's branch where a decoded
// message cannot be delivered because the inbound channel is full and the
// group is cancelled: the reader abandons the pending message and stops
// cleanly rather than blocking forever. QueueSize 1 fills after one
// undelivered message, so the second blocks the reader in the delivery select
// while nothing drains Recv; cancelling the parent context then takes the
// ctx.Done arm.
func TestPeer_deliveryAbandonedOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	a, b := net.Pipe()
	defer func() { _ = b.Close() }()

	cfg := lineConfig(a)
	cfg.Context = ctx
	cfg.QueueSize = 1
	p := wire.New(cfg)

	// "first" fills the one-slot buffer; "second" blocks the reader on
	// delivery since Recv is never drained.
	_, err := io.WriteString(b, "first\nsecond\n")
	core.AssertMustNoError(t, err, "write frames")

	cancel()

	core.AssertNoError(t, p.Wait(), "delivery abandoned on cancel is clean")
	core.AssertNoError(t, p.Close(), "close after abandoned delivery")
}
