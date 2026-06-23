package server_test

import (
	"testing"

	"darvaza.org/core"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/mock/client"
	"protomcp.org/nanorpc/pkg/nanorpc/mock/server"
)

// TestServer_roundTrip dials a listening mock server with a mock client and
// exchanges a request and its response, covering the listener accept path and
// the framing in both directions.
func TestServer_roundTrip(t *testing.T) {
	srv := server.New(t)
	cli := client.New(t, srv.Addr())

	cli.Send(&nanorpc.NanoRPCRequest{
		RequestId:   7,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   nanorpc.GetPathOneOfString("/echo"),
	})

	conn := srv.Accept()
	req := conn.Recv()
	core.AssertEqual(t, int32(7), req.RequestId, "request_id")
	core.AssertEqual(t, "/echo", req.GetPath(), "path")

	conn.Reply(&nanorpc.NanoRPCResponse{
		RequestId:      req.RequestId,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	})

	res := cli.Recv()
	core.AssertEqual(t, int32(7), res.RequestId, "response request_id")
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, res.ResponseStatus,
		"response_status")
}
