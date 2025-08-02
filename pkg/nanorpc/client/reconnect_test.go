package client

import (
	"context"
	"errors"
	"testing"

	"protomcp.org/nanorpc/pkg/nanorpc/common/testutils"
)

func testWithValidConn(t *testing.T, client *Client, testErr error) {
	ctx := context.Background()
	conn := &testutils.MockConn{
		Remote: "127.0.0.1:8080",
		Local:  "127.0.0.1:12345",
	}

	err := client.onReconnectError(ctx, conn, testErr)
	if err != testErr {
		t.Fatalf("Expected error %v, got %v", testErr, err)
	}
}

func testWithNilConn(t *testing.T, client *Client, testErr error) {
	ctx := context.Background()
	err := client.onReconnectError(ctx, nil, testErr)
	if err != testErr {
		t.Fatalf("Expected error %v, got %v", testErr, err)
	}
}

func testWithCustomErrorHandler(t *testing.T, testErr error) {
	ctx := context.Background()
	customErr := errors.New("custom handled error")

	clientWithHandler := &Client{
		logger: testutils.NewMockFieldLogger(),
		callOnError: func(_ context.Context, _ error) error {
			return customErr
		},
	}

	conn := &testutils.MockConn{
		Remote: "127.0.0.1:8080",
		Local:  "127.0.0.1:12345",
	}

	err := clientWithHandler.onReconnectError(ctx, conn, testErr)
	if err != customErr {
		t.Fatalf("Expected custom error %v, got %v", customErr, err)
	}

	err = clientWithHandler.onReconnectError(ctx, nil, testErr)
	if err != customErr {
		t.Fatalf("Expected custom error %v, got %v", customErr, err)
	}
}

func TestClient_onReconnectError(t *testing.T) {
	client := &Client{
		logger: testutils.NewMockFieldLogger(),
	}

	testErr := errors.New("test connection error")

	t.Run("with_conn", func(t *testing.T) { testWithValidConn(t, client, testErr) })
	t.Run("with_nil_conn", func(t *testing.T) { testWithNilConn(t, client, testErr) })
	t.Run("with_custom_error_handler", func(t *testing.T) { testWithCustomErrorHandler(t, testErr) })
}
