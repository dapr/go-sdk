package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeBinding(t *testing.T) {
	ctx := context.Background()

	mIn := make(map[string]string)
	mIn["test"] = "value"

	out, mOut, err := testClient.InvokeBinding(ctx, "serving", "EchoMethod", []byte("ping"), mIn)
	assert.Nil(t, err)
	assert.NotNil(t, mOut)
	assert.NotNil(t, out)
}

func TestInvokeOutputBinding(t *testing.T) {
	ctx := context.Background()
	err := testClient.InvokeOutputBinding(ctx, "serving", "EchoMethod", []byte("ping"))
	assert.Nil(t, err)
}
