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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	daprClient "github.com/dapr/go-sdk/client"
)

func TestNewClient(t *testing.T) {
	// Currently will always fail if no dapr connection available
	testClient, err := NewClient()
	assert.Empty(t, testClient)
	require.Error(t, err)
}

func TestClientOptions(t *testing.T) {
	t.Run("with client", func(t *testing.T) {
		opts := returnClientOptions(WithDaprClient(&daprClient.GRPCClient{}))
		assert.NotNil(t, opts.daprClient)
	})
}

func returnClientOptions(opts ...clientOption) clientOptions {
	options := new(clientOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return *options
		}
	}
	return *options
}

func TestClientMethods(t *testing.T) {
	testClient := client{
		taskHubClient: nil,
	}
	ctx := context.Background()
	t.Run("ScheduleNewWorkflow - empty wf name", func(t *testing.T) {
		id, err := testClient.ScheduleNewWorkflow(ctx, "", nil)
		require.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("FetchWorkflowMetadata - empty id", func(t *testing.T) {
		metadata, err := testClient.FetchWorkflowMetadata(ctx, "")
		require.Error(t, err)
		assert.Nil(t, metadata)
	})

	t.Run("WaitForWorkflowStart - empty id", func(t *testing.T) {
		metadata, err := testClient.WaitForWorkflowStart(ctx, "")
		require.Error(t, err)
		assert.Nil(t, metadata)
	})

	t.Run("WaitForWorkflowCompletion - empty id", func(t *testing.T) {
		metadata, err := testClient.WaitForWorkflowCompletion(ctx, "")
		require.Error(t, err)
		assert.Nil(t, metadata)
	})

	t.Run("TerminateWorkflow - empty id", func(t *testing.T) {
		err := testClient.TerminateWorkflow(ctx, "")
		require.Error(t, err)
	})

	t.Run("RaiseEvent - empty id", func(t *testing.T) {
		err := testClient.RaiseEvent(ctx, "", "EventName")
		require.Error(t, err)
	})

	t.Run("RaiseEvent - empty eventName", func(t *testing.T) {
		err := testClient.RaiseEvent(ctx, "testID", "")
		require.Error(t, err)
	})

	t.Run("SuspendWorkflow - empty id", func(t *testing.T) {
		err := testClient.SuspendWorkflow(ctx, "", "reason")
		require.Error(t, err)
	})

	t.Run("ResumeWorkflow - empty id", func(t *testing.T) {
		err := testClient.ResumeWorkflow(ctx, "", "reason")
		require.Error(t, err)
	})

	t.Run("PurgeWorkflow - empty id", func(t *testing.T) {
		err := testClient.PurgeWorkflow(ctx, "")
		require.Error(t, err)
	})
}
