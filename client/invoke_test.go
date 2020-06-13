package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeServiceWithContent(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx)
	assert.NotNil(t, closer)
	defer closer()
	assert.NotNil(t, client)

	// TODO: fails with rpc error: code = Unimplemented desc = unknown service dapr.proto.runtime.v1.Dapr
	// resp, err := client.InvokeServiceWithContent(ctx, "serving", "EchoMethod",
	// 	"text/plain; charset=UTF-8", []byte("ping"))
	// assert.Nil(t, err)
	// assert.NotNil(t, resp)
}
