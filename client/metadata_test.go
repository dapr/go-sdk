package client

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

// Test GetMetadata returns
func TestGetMetadata(t *testing.T) {
	ctx := t.Context()
	t.Run("get meta", func(t *testing.T) {
		metadata, err := testClient.GetMetadata(ctx)
		require.NoError(t, err)
		assert.NotNil(t, metadata)
	})
}

func TestSetMetadata(t *testing.T) {
	ctx := t.Context()
	t.Run("set meta", func(t *testing.T) {
		err := testClient.SetMetadata(ctx, "test_key", "test_value")
		require.NoError(t, err)
		metadata, err := testClient.GetMetadata(ctx)
		require.NoError(t, err)
		assert.Equal(t, "test_value", metadata.ExtendedMetadata["test_key"])
	})
}
