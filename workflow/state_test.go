package workflow

import (
	"testing"

	"github.com/microsoft/durabletask-go/api"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	wfState := WorkflowState{Metadata: api.OrchestrationMetadata{RuntimeStatus: 0}}

	t.Run("test running", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "running", s.String())
	})

	wfState.Metadata.RuntimeStatus = 1

	t.Run("test completed", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "completed", s.String())
	})

	wfState.Metadata.RuntimeStatus = 2

	t.Run("test continued_as_new", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "continued_as_new", s.String())
	})

	wfState.Metadata.RuntimeStatus = 3

	t.Run("test failed", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "failed", s.String())
	})

	wfState.Metadata.RuntimeStatus = 4

	t.Run("test canceled", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "canceled", s.String())
	})

	wfState.Metadata.RuntimeStatus = 5

	t.Run("test terminated", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "terminated", s.String())
	})

	wfState.Metadata.RuntimeStatus = 6

	t.Run("test pending", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "pending", s.String())
	})

	wfState.Metadata.RuntimeStatus = 7

	t.Run("test suspended", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "suspended", s.String())
	})
}
