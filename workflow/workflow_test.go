package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/task"
)

func TestConvertMetadata(t *testing.T) {
	t.Run("convert metadata", func(t *testing.T) {
		rawMetadata := &protos.OrchestrationMetadata{
			InstanceId: "test",
		}
		metadata := convertMetadata(rawMetadata)
		assert.NotEmpty(t, metadata)
	})
}

func TestCallChildWorkflowOptions(t *testing.T) {
	t.Run("child workflow input - valid", func(t *testing.T) {
		opts := returnCallChildWorkflowOptions(ChildWorkflowInput("test"))
		assert.Equal(t, "\"test\"", opts.rawInput.GetValue())
	})

	t.Run("child workflow raw input - valid", func(t *testing.T) {
		opts := returnCallChildWorkflowOptions(ChildWorkflowRawInput("test"))
		assert.Equal(t, "test", opts.rawInput.GetValue())
	})

	t.Run("child workflow instance id - valid", func(t *testing.T) {
		opts := returnCallChildWorkflowOptions(ChildWorkflowInstanceID("test"))
		assert.Equal(t, "test", opts.instanceID)
	})

	t.Run("child workflow input - invalid", func(t *testing.T) {
		opts := returnCallChildWorkflowOptions(ChildWorkflowInput(make(chan int)))
		assert.Empty(t, opts.rawInput.GetValue())
	})

	t.Run("child workflow retry policy - set", func(t *testing.T) {
		opts := returnCallChildWorkflowOptions(ChildWorkflowRetryPolicy(RetryPolicy{
			MaxAttempts:          3,
			InitialRetryInterval: 100 * time.Millisecond,
			BackoffCoefficient:   2,
			MaxRetryInterval:     2 * time.Second,
		}))
		assert.Equal(t, &task.RetryPolicy{
			MaxAttempts:          3,
			InitialRetryInterval: 100 * time.Millisecond,
			BackoffCoefficient:   2,
			MaxRetryInterval:     2 * time.Second,
		}, opts.getRetryPolicy())
	})

	t.Run("child workflow retry policy - empty", func(t *testing.T) {
		opts := returnCallChildWorkflowOptions()
		assert.Empty(t, opts.getRetryPolicy())
	})
}

func returnCallChildWorkflowOptions(opts ...callChildWorkflowOption) callChildWorkflowOptions {
	options := new(callChildWorkflowOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return *options
		}
	}
	return *options
}

func TestNewTaskSlice(t *testing.T) {
	tasks := NewTaskSlice(10)
	assert.Len(t, tasks, 10)
}
