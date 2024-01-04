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
		assert.Equal(t, "RUNNING", s.String())
	})

	wfState.Metadata.RuntimeStatus = 1

	t.Run("test completed", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "COMPLETED", s.String())
	})

	wfState.Metadata.RuntimeStatus = 2

	t.Run("test continued_as_new", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "CONTINUED_AS_NEW", s.String())
	})

	wfState.Metadata.RuntimeStatus = 3

	t.Run("test failed", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "FAILED", s.String())
	})

	wfState.Metadata.RuntimeStatus = 4

	t.Run("test canceled", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "CANCELED", s.String())
	})

	wfState.Metadata.RuntimeStatus = 5

	t.Run("test terminated", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "TERMINATED", s.String())
	})

	wfState.Metadata.RuntimeStatus = 6

	t.Run("test pending", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "PENDING", s.String())
	})

	wfState.Metadata.RuntimeStatus = 7

	t.Run("test suspended", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "SUSPENDED", s.String())
	})

	wfState.Metadata.RuntimeStatus = 8

	t.Run("test unknown", func(t *testing.T) {
		s := wfState.RuntimeStatus()
		assert.Equal(t, "UNKNOWN", s.String())
	})
}
