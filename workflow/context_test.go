package workflow

import (
	"testing"
	"time"

	"github.com/microsoft/durabletask-go/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	c := Context{
		orchestrationContext: &task.OrchestrationContext{
			ID:             "test-id",
			Name:           "test-workflow-context",
			IsReplaying:    false,
			CurrentTimeUtc: time.Date(2023, time.December, 17, 18, 44, 0, 0, time.UTC),
		},
	}
	t.Run("get input - empty", func(t *testing.T) {
		var input string
		err := c.GetInput(&input)
		require.NoError(t, err)
		assert.Equal(t, "", input)
	})
	t.Run("workflow name", func(t *testing.T) {
		name := c.Name()
		assert.Equal(t, "test-workflow-context", name)
	})
	t.Run("instance id", func(t *testing.T) {
		instanceID := c.InstanceID()
		assert.Equal(t, "test-id", instanceID)
	})
	t.Run("current utc date time", func(t *testing.T) {
		date := c.CurrentUTCDateTime()
		assert.Equal(t, time.Date(2023, time.December, 17, 18, 44, 0, 0, time.UTC), date)
	})
	t.Run("is replaying", func(t *testing.T) {
		replaying := c.IsReplaying()
		assert.False(t, replaying)
	})
}
