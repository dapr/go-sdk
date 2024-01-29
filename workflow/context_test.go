/*
Copyright 2024 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package workflow

import (
	"testing"
	"time"

	"github.com/microsoft/durabletask-go/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	c := WorkflowContext{
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

	t.Run("waitforexternalevent - empty ids", func(t *testing.T) {
		completableTask := c.WaitForExternalEvent("", time.Second)
		assert.Nil(t, completableTask)
	})

	t.Run("continueasnew", func(t *testing.T) {
		c.ContinueAsNew("test", true)
		c.ContinueAsNew("test", false)
	})
}
