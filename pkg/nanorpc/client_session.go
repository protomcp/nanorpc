package nanorpc

import (
	"context"
	"net"

	"darvaza.org/x/net/reconnect"
	"google.golang.org/protobuf/proto"
)

// Client is a reconnecting NanoRPC client.
type clientRequest struct {
	r *NanoRPCRequest
	d proto.Message
}

// ClientSession represents a connection to a NanoRPC server.
type ClientSession struct {
	reconnect.WorkGroup

	c  *Client
	rc *reconnect.Client
	ra net.Addr

	ss *reconnect.StreamSession[*NanoRPCResponse, clientRequest]
}

// Spawn starts the required workers to handle the session
func (cs *ClientSession) Spawn() error {
	return cs.ss.Spawn()
}

//
// StreamSession callbacks
//

func (cs *ClientSession) onSetReadDeadline() error {
	return cs.rc.ResetReadDeadline()
}

func (cs *ClientSession) onSetWriteDeadline() error {
	return cs.rc.ResetWriteDeadline()
}

func (cs *ClientSession) onUnsetReadDeadline() error {
	return cs.rc.SetReadDeadline(0)
}

func (cs *ClientSession) onUnsetWriteDeadline() error {
	return cs.rc.SetReadDeadline(0)
}

func (*ClientSession) onError(error) {}

//
// factory
//

func newClientSession(ctx context.Context, c *Client, queueSize uint, conn net.Conn) *ClientSession {
	ss := &reconnect.StreamSession[*NanoRPCResponse, clientRequest]{
		QueueSize: queueSize,
		Conn:      c.rc,
		Context:   ctx,

		Split: Split,
		Marshal: func(r clientRequest) ([]byte, error) {
			return EncodeRequest(r.r, r.d)
		},
		Unmarshal: func(data []byte) (*NanoRPCResponse, error) {
			resp, _, err := DecodeResponse(data)
			return resp, err
		},
	}

	cs := &ClientSession{
		c:  c,
		rc: c.rc,
		ra: conn.RemoteAddr(),

		ss:        ss,
		WorkGroup: ss,
	}

	cs.ss.SetReadDeadline = cs.onSetReadDeadline
	cs.ss.SetWriteDeadline = cs.onSetWriteDeadline
	cs.ss.UnsetReadDeadline = cs.onUnsetReadDeadline
	cs.ss.UnsetWriteDeadline = cs.onUnsetWriteDeadline
	cs.ss.OnError = cs.onError
	return cs
}
