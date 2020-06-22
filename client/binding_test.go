package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeBinding(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx, t)
	defer closer()

	mIn := make(map[string]string, 0)
	mIn["test"] = "value"

	out, mOut, err := client.InvokeBinding(ctx, "serving", "EchoMethod", []byte("ping"), mIn)
	assert.Nil(t, err)
	assert.NotNil(t, mOut)
	assert.NotNil(t, out)
}

func TestInvokeOutputBinding(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx, t)
	defer closer()

	err := client.InvokeOutputBinding(ctx, "serving", "EchoMethod", []byte("ping"))
	assert.Nil(t, err)
}
