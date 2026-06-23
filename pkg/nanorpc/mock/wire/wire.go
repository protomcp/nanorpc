// Package wire provides a length-delimited framing peer over a net.Conn —
// the shared engine behind the mock server and mock client.
//
// A [Peer] reads framed messages of type In from the connection and writes
// framed messages of type Out back. The two mock packages instantiate it
// with opposite type parameters: the server reads requests and writes
// responses, the client the reverse.
package wire

import (
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"syscall"

	"darvaza.org/core"
	"darvaza.org/x/sync/workgroup"
)

// DefaultQueueSize is the inbound-channel buffer the mock packages pass as
// [Config.QueueSize].
const DefaultQueueSize = 16

// maxFrameSize caps a single inbound frame. bufio.Scanner rejects tokens
// above 64 KiB by default; this lifts the ceiling well past nanorpc's
// small-message domain so an unusually large frame surfaces as data rather
// than a scanner error.
const maxFrameSize = 1 << 20

// Config configures a [Peer]. Split, Encode and Decode are required.
type Config[Out, In any] struct {
	// Context is the parent context. Cancelling it does not interrupt a
	// blocked Read — only [Peer.Close], which closes the connection, stops
	// the reader — but it abandons a pending delivery and becomes the
	// shutdown cause. A nil Context defaults to [context.Background].
	Context context.Context
	// Conn is the connection the peer reads from and writes to.
	Conn net.Conn
	// Split frames the inbound byte stream, e.g. nanorpc.Split.
	Split bufio.SplitFunc
	// Encode writes one Out message to w as a framed message.
	Encode func(w io.Writer, msg Out) error
	// Decode parses one framed In message from a frame's bytes.
	Decode func(data []byte) (In, error)
	// QueueSize bounds the buffered inbound channel; values below one are
	// treated as one.
	QueueSize int
}

// Peer drives one end of a framed connection over conn. A reader goroutine,
// supervised by a workgroup, decodes inbound frames to In and delivers them
// on the channel returned by [Peer.Recv]; [Peer.Send] encodes an Out frame
// back. The workgroup stores the reader's first error as the cancellation
// cause, so a peer shut down cleanly reports nil from [Peer.Wait] while a
// genuine framing fault surfaces. nanorpc framing reports a boundary close as
// [io.ErrUnexpectedEOF], indistinguishable from a real truncation by value
// alone, so two filters separate a clean shutdown from a fault: StopErr drops
// the error when this end initiated the shutdown (intent — was the peer
// closed?), and stopCatch drops it when the value names a connection ending
// rather than a frame (identity — did the other end hang up?). Anything that
// passes both is a true fault and becomes the cause.
type Peer[Out, In any] struct {
	conn   net.Conn
	wg     *workgroup.Group
	encode func(io.Writer, Out) error
	decode func([]byte) (In, error)
	split  bufio.SplitFunc
	in     chan In

	// non-pointer field kept last so the GC pointer scan stops early
	closeOnce sync.Once
}

// New starts a [Peer] over cfg.Conn, spawning its reader. The reader runs
// until the connection ends or the peer is closed.
func New[Out, In any](cfg Config[Out, In]) *Peer[Out, In] {
	ctx := cfg.Context
	if ctx == nil {
		ctx = context.Background()
	}

	qs := max(cfg.QueueSize, 1)

	p := &Peer[Out, In]{
		conn:   cfg.Conn,
		wg:     workgroup.New(ctx),
		encode: cfg.Encode,
		decode: cfg.Decode,
		split:  cfg.Split,
		in:     make(chan In, qs),
	}

	// Start the reader unless the parent context is already cancelled: a
	// blocking Read cannot be interrupted by cancellation, so a reader
	// started against a dead context would hang until Close. When it does
	// not start, close in so Recv unblocks and Wait reports the cause.
	started := false
	if ctx.Err() == nil {
		started = p.wg.GoCatch(p.read, stopCatch) == nil
	}
	if !started {
		close(p.in)
	}

	return p
}

// read decodes the inbound frame stream onto the channel returned by Recv,
// stopping when the workgroup is cancelled or the stream ends. A stream error
// seen after cancellation is the shutdown the caller initiated and is reported
// as nil; before cancellation the error is genuine and becomes the
// workgroup's stored cause.
func (p *Peer[Out, In]) read(ctx context.Context) error {
	defer close(p.in)

	sc := bufio.NewScanner(p.conn)
	sc.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxFrameSize)
	sc.Split(p.split)
	for sc.Scan() {
		msg, err := p.decode(sc.Bytes())
		if err != nil {
			return StopErr(ctx, err)
		}
		select {
		case p.in <- msg:
		case <-ctx.Done():
			return nil
		}
	}

	return StopErr(ctx, sc.Err())
}

// StopErr drops a stream error once ctx is cancelled: nanorpc framing
// surfaces a boundary close as [io.ErrUnexpectedEOF], which a clean shutdown
// and a real truncation share by value, so intent — was the peer or listener
// closed? — decides. Before cancellation the error is genuine and is returned
// unchanged. It serves the same purpose for an accept loop's listener error.
func StopErr(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return nil
	}
	return err
}

// stopCatch is the reader's workgroup catcher: it drops a connection-ended
// error so a remote hang-up is not reported as a fault. StopErr inside read
// already silences a stream error once the local peer initiated shutdown; this
// covers the complementary case where the other end closed first, leaving the
// reader to observe the close as its own error before the group was cancelled.
// A genuine framing fault — anything isConnDone rejects — passes through and
// becomes the cancellation cause.
func stopCatch(_ context.Context, err error) error {
	if err != nil && core.IsErrorFn(isConnDone, err) {
		return nil
	}
	return err
}

// isConnDone reports whether err is a connection-ended signal rather than a
// framing fault. nanorpc framing surfaces a boundary close as
// [io.ErrUnexpectedEOF]; a pipe close adds [io.ErrClosedPipe] and
// [net.ErrClosed]; a TCP peer that hangs up with data in flight yields
// [syscall.ECONNRESET]. It is the check passed to [core.IsErrorFn], which
// unwraps the chain — including the [net.OpError] wrapping the reset — and
// calls this on each leaf.
func isConnDone(err error) bool {
	switch err {
	case io.EOF, io.ErrUnexpectedEOF, io.ErrClosedPipe,
		net.ErrClosed, syscall.ECONNRESET:
		return true
	default:
		return false
	}
}

// Recv returns the channel of inbound messages. The channel is closed once
// the reader stops, so a receiving select observes shutdown.
func (p *Peer[Out, In]) Recv() <-chan In {
	return p.in
}

// Send encodes msg as a framed message and writes it to the connection. It is
// not safe for concurrent use: concurrent Sends interleave their writes, and a
// Send racing [Peer.Close] writes to a closing connection.
func (p *Peer[Out, In]) Send(msg Out) error {
	return p.encode(p.conn, msg)
}

// Wait blocks until the reader has stopped and returns the cancellation
// cause: nil for a [Peer.Close]-initiated shutdown, the framing error that
// stopped the reader, or — when the parent context is cancelled with a cause
// other than context.Canceled — that cause.
func (p *Peer[Out, In]) Wait() error {
	return p.wg.Wait()
}

// Close stops the reader and closes the connection, returning the reader's
// outcome: nil for a clean shutdown, a framing fault that stopped the reader
// earlier, or the parent context's cancellation cause when it carried one.
// Calling Close more than once is safe, and so is closing a peer whose context
// was already cancelled; the connection is closed exactly once.
func (p *Peer[Out, In]) Close() error {
	// Close the connection regardless of who first cancelled the group:
	// closing it is the only thing that unblocks a blocked Read, and
	// cancellation alone — including parent-context cancellation, which the
	// workgroup propagates to the cancelled state on its own — never
	// interrupts it. Gating the close on winning Cancel would deadlock Wait
	// whenever the context was cancelled first. Cancel keeps a clean Close's
	// cause nil; a framing fault that cancelled the group earlier still wins.
	p.wg.Cancel(nil)

	var closeErr error
	p.closeOnce.Do(func() { closeErr = p.conn.Close() })

	if waitErr := p.wg.Wait(); waitErr != nil {
		return waitErr
	}
	return closeErr
}
