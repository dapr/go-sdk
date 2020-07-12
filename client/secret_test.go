package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -timeout 30s ./client -count 1 -run ^TestGetSecret$
func TestGetSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("without meta", func(t *testing.T) {
		out, err := testClient.GetSecret(ctx, "store", "key1", nil)
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})

	t.Run("with meta", func(t *testing.T) {
		in := map[string]string{"k1": "v1", "k2": "v2"}
		out, err := testClient.GetSecret(ctx, "store", "key1", in)
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})
}
