package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -timeout 30s ./client -count 1 -run ^TestInvokeServiceWithContent$

func TestInvokeServiceWithContent(t *testing.T) {
	ctx := context.Background()
	data := "ping"

	t.Run("with content", func(t *testing.T) {
		content := &ServiceContent{
			ContentType: "text/plain",
			Data:        []byte(data),
		}
		resp, err := testClient.InvokeServiceWithContent(ctx, "test", "fn", content)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, string(resp), data)

	})

	t.Run("without content", func(t *testing.T) {
		resp, err := testClient.InvokeService(ctx, "test", "fn")
		assert.Nil(t, err)
		assert.Nil(t, resp)

	})
}
