package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowRuntime(t *testing.T) {
	// TODO: Mock grpc conn - currently requires dapr to be available
	t.Run("test workflow name is correct", func(t *testing.T) {
		wr, err := NewRuntime("localhost", "50001")
		require.NoError(t, err)
		err = wr.RegisterWorkflow(testOrchestrator)
		require.NoError(t, err)
	})
}

func TestGetDecorator(t *testing.T) {
	name, err := getDecorator(testOrchestrator)
	require.NoError(t, err)
	assert.Equal(t, "testOrchestrator", name)
}

func testOrchestrator(ctx *Context) (any, error) {
	return nil, nil
}
