package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
