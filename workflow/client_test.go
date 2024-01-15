package workflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	// Currently will always fail if no dapr connection available
	testClient, err := NewClient()
	assert.Empty(t, testClient)
	require.Error(t, err)
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

	t.Run("SuspendWorkflow - empty reason", func(t *testing.T) {
		err := testClient.SuspendWorkflow(ctx, "testID", "")
		require.Error(t, err)
	})

	t.Run("ResumeWorkflow - empty id", func(t *testing.T) {
		err := testClient.ResumeWorkflow(ctx, "", "reason")
		require.Error(t, err)
	})

	t.Run("ResumeWorkflow - empty reason", func(t *testing.T) {
		err := testClient.ResumeWorkflow(ctx, "testID", "")
		require.Error(t, err)
	})

	t.Run("PurgeWorkflow - empty id", func(t *testing.T) {
		err := testClient.PurgeWorkflow(ctx, "")
		require.Error(t, err)
	})
}
