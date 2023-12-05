package client

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowAlpha1(t *testing.T) {
	ctx := context.Background()

	// 1: StartWorkflow
	t.Run("start workflow - valid (without id)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp.InstanceID)
	})
	t.Run("start workflow - valid (with id)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		assert.NoError(t, err)
		assert.Equal(t, "TestID", resp.InstanceID)
	})
	t.Run("start workflow - rpc failure", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - invalid WorkflowComponent", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "",
			WorkflowName:      "TestWorkflow",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - grpc failure", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - cannot serialize input", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
			Input:             math.NaN(),
			SendRawInput:      false,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - raw input", func(t *testing.T) {
		resp, err := testClient.StartWorkflowAlpha1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
			Input:             []byte("stringtest"),
			SendRawInput:      true,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	// 2: GetWorkflow
	t.Run("get workflow", func(t *testing.T) {
		resp, err := testClient.GetWorkflowAlpha1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - valid", func(t *testing.T) {
		resp, err := testClient.GetWorkflowAlpha1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - invalid id", func(t *testing.T) {
		resp, err := testClient.GetWorkflowAlpha1(ctx, &GetWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("get workflow - invalid workflowcomponent", func(t *testing.T) {
		resp, err := testClient.GetWorkflowAlpha1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("get workflow - grpc fail", func(t *testing.T) {
		resp, err := testClient.GetWorkflowAlpha1(ctx, &GetWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	// 3: PauseWorkflow
	t.Run("pause workflow", func(t *testing.T) {
		err := testClient.PauseWorkflowAlpha1(ctx, &PauseWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("pause workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.PauseWorkflowAlpha1(ctx, &PauseWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("pause workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.PauseWorkflowAlpha1(ctx, &PauseWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("pause workflow", func(t *testing.T) {
		err := testClient.PauseWorkflowAlpha1(ctx, &PauseWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	// 4: ResumeWorkflow
	t.Run("resume workflow", func(t *testing.T) {
		err := testClient.ResumeWorkflowAlpha1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("resume workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.ResumeWorkflowAlpha1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("resume workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.ResumeWorkflowAlpha1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("resume workflow - grpc fail", func(t *testing.T) {
		err := testClient.ResumeWorkflowAlpha1(ctx, &ResumeWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	// 5: TerminateWorkflow
	t.Run("terminate workflow", func(t *testing.T) {
		err := testClient.TerminateWorkflowAlpha1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("terminate workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.TerminateWorkflowAlpha1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("terminate workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.TerminateWorkflowAlpha1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("terminate workflow - grpc failure", func(t *testing.T) {
		err := testClient.TerminateWorkflowAlpha1(ctx, &TerminateWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	// 6: RaiseEventWorkflow
	t.Run("raise event workflow", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		assert.NoError(t, err)
	})

	t.Run("raise event workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		assert.Error(t, err)
	})

	t.Run("raise event workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
			EventName:         "TestEvent",
		})
		assert.Error(t, err)
	})

	t.Run("raise event workflow - invalid eventname", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "",
		})
		assert.Error(t, err)
	})

	t.Run("raise event workflow - grpc failure", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		assert.Error(t, err)
	})
	t.Run("raise event workflow - cannot serialize input", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         math.NaN(),
			SendRawData:       false,
		})
		assert.Error(t, err)
	})
	t.Run("raise event workflow - raw input", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowAlpha1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         []byte("teststring"),
			SendRawData:       true,
		})
		assert.Error(t, err)
	})

	// 7: PurgeWorkflow
	t.Run("purge workflow", func(t *testing.T) {
		err := testClient.PurgeWorkflowAlpha1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("purge workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.PurgeWorkflowAlpha1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("purge workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.PurgeWorkflowAlpha1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("purge workflow - grpc failure", func(t *testing.T) {
		err := testClient.PurgeWorkflowAlpha1(ctx, &PurgeWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})
}

func TestWorkflowBeta1(t *testing.T) {
	ctx := context.Background()

	// 1: StartWorkflow
	t.Run("start workflow - valid (without id)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp.InstanceID)
	})
	t.Run("start workflow - valid (with id)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		assert.NoError(t, err)
		assert.Equal(t, "TestID", resp.InstanceID)
	})
	t.Run("start workflow - rpc failure", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - invalid WorkflowComponent", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "",
			WorkflowName:      "TestWorkflow",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - grpc failure", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - cannot serialize input", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
			Input:             math.NaN(),
			SendRawInput:      false,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - raw input", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
			Input:             []byte("stringtest"),
			SendRawInput:      true,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	// 2: GetWorkflow
	t.Run("get workflow", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - valid", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - invalid id", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("get workflow - invalid workflowcomponent", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("get workflow - grpc fail", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// 3: PauseWorkflow
	t.Run("pause workflow", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("pause workflow invalid instanceid", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("pause workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("pause workflow", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	// 4: ResumeWorkflow
	t.Run("resume workflow", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("resume workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("resume workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("resume workflow - grpc fail", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	// 5: TerminateWorkflow
	t.Run("terminate workflow", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("terminate workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("terminate workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("terminate workflow - grpc failure", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	// 6: RaiseEventWorkflow
	t.Run("raise event workflow", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		assert.NoError(t, err)
	})

	t.Run("raise event workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		assert.Error(t, err)
	})

	t.Run("raise event workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
			EventName:         "TestEvent",
		})
		assert.Error(t, err)
	})

	t.Run("raise event workflow - invalid eventname", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "",
		})
		assert.Error(t, err)
	})

	t.Run("raise event workflow - grpc failure", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		assert.Error(t, err)
	})
	t.Run("raise event workflow - cannot serialize input", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         math.NaN(),
			SendRawData:       false,
		})
		assert.Error(t, err)
	})
	t.Run("raise event workflow - raw input", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         []byte("teststring"),
			SendRawData:       true,
		})
		assert.Error(t, err)
	})

	// 7: PurgeWorkflow
	t.Run("purge workflow", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		assert.NoError(t, err)
	})

	t.Run("purge workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})

	t.Run("purge workflow - invalid workflowcomponent", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		assert.Error(t, err)
	})

	t.Run("purge workflow - grpc failure", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		assert.Error(t, err)
	})
}
