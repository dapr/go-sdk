package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -timeout 30s ./client -count 1 -run ^TestInvokeBinding$

func TestInvokeBinding(t *testing.T) {
	ctx := context.Background()
	data := "ping"

	t.Run("output binding", func(t *testing.T) {
		err := testClient.InvokeOutputBinding(ctx, "test", "fn", []byte(data))
		assert.Nil(t, err)
	})

	t.Run("output binding without data", func(t *testing.T) {
		err := testClient.InvokeOutputBinding(ctx, "test", "fn", []byte(data))
		assert.Nil(t, err)
	})

	t.Run("binding without data", func(t *testing.T) {
		out, mOut, err := testClient.InvokeBinding(ctx, "test", "fn", nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, mOut)
		assert.NotNil(t, out)
	})

	t.Run("binding with data and meta", func(t *testing.T) {
		mIn := map[string]string{"k1": "v1", "k2": "v2"}
		out, mOut, err := testClient.InvokeBinding(ctx, "test", "fn", []byte(data), mIn)
		assert.Nil(t, err)
		assert.NotNil(t, mOut)
		assert.NotNil(t, out)
		assert.Equal(t, data, string(out))
	})

}
