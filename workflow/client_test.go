package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	// Currently will always fail if no dapr connection available
	client, err := NewClient()
	assert.Empty(t, client)
	require.Error(t, err)
}
