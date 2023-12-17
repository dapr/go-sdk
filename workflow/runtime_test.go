package workflow

import (
	"sync"
	"testing"

	"github.com/microsoft/durabletask-go/task"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntime(t *testing.T) {
	t.Run("failure to create newruntime without dapr", func(t *testing.T) {
		wr, err := NewRuntime("localhost", "50001")
		require.Error(t, err)
		assert.Equal(t, &WorkflowRuntime{}, wr)
	})
}

func TestWorkflowRuntime(t *testing.T) {
	testRuntime := WorkflowRuntime{
		tasks:  task.NewTaskRegistry(),
		client: nil,
		mutex:  sync.Mutex{},
		quit:   nil,
		cancel: nil,
	}

	// TODO: Mock grpc conn - currently requires dapr to be available
	t.Run("register workflow", func(t *testing.T) {
		err := testRuntime.RegisterWorkflow(testWorkflow)
		require.NoError(t, err)
	})
	t.Run("register workflow - anonymous func", func(t *testing.T) {
		err := testRuntime.RegisterWorkflow(func(ctx *Context) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
	})
	t.Run("register activity", func(t *testing.T) {
		err := testRuntime.RegisterActivity(testActivity)
		require.NoError(t, err)
	})
	t.Run("register activity - anonymous func", func(t *testing.T) {
		err := testRuntime.RegisterActivity(func(ctx ActivityContext) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
	})
}

func TestWrapWorkflow(t *testing.T) {
	t.Run("wrap workflow", func(t *testing.T) {
		orchestrator := wrapWorkflow(testWorkflow)
		assert.NotNil(t, orchestrator)
	})
}

func TestWrapActivity(t *testing.T) {
	t.Run("wrap activity", func(t *testing.T) {
		activity := wrapActivity(testActivity)
		assert.NotNil(t, activity)
	})
}

func TestGetDecorator(t *testing.T) {
	t.Run("get decorator", func(t *testing.T) {
		name, err := getDecorator(testWorkflow)
		require.NoError(t, err)
		assert.Equal(t, "testWorkflow", name)
	})
	t.Run("get decorator - nil", func(t *testing.T) {
		name, err := getDecorator(nil)
		require.Error(t, err)
		assert.Equal(t, "", name)
	})
}

func testWorkflow(ctx *Context) (any, error) {
	_ = ctx
	return nil, nil
}

func testActivity(ctx ActivityContext) (any, error) {
	_ = ctx
	return nil, nil
}
