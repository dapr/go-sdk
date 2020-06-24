package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeServiceWithContent(t *testing.T) {
	ctx := context.Background()
	resp, err := testClient.InvokeServiceWithContent(ctx, "serving", "EchoMethod",
		"text/plain; charset=UTF-8", []byte("ping"))
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, string(resp), "ping")
}

func TestInvokeService(t *testing.T) {
	ctx := context.Background()
	resp, err := testClient.InvokeService(ctx, "serving", "EchoMethod")
	assert.Nil(t, err)
	assert.Nil(t, resp)
}
