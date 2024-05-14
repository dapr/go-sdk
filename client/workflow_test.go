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
package client

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestMarshalInput(t *testing.T) {
	var input any
	t.Run("string", func(t *testing.T) {
		input = "testString"
		data, err := marshalInput(input)
		require.NoError(t, err)
		assert.Equal(t, []byte{0x22, 0x74, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22}, data)
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
		require.NoError(t, err)
		assert.NotNil(t, resp.InstanceID)
	})
	t.Run("start workflow - valid (with id)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		require.NoError(t, err)
		assert.Equal(t, "TestID", resp.InstanceID)
	})
	t.Run("start workflow - valid (without component name)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
			WorkflowName:      "TestWorkflow",
		})
		require.NoError(t, err)
		assert.Equal(t, "TestID", resp.InstanceID)
	})
	t.Run("start workflow - rpc failure", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
		})
		require.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("start workflow - grpc failure", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "",
		})
		require.Error(t, err)
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
		require.Error(t, err)
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
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("start workflow - raw input (invalid)", func(t *testing.T) {
		resp, err := testClient.StartWorkflowBeta1(ctx, &StartWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			WorkflowName:      "TestWorkflow",
			Input:             "test string",
			SendRawInput:      true,
		})
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	// 2: GetWorkflow
	t.Run("get workflow", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - valid", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - valid (without component)", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("get workflow - invalid id", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("get workflow - grpc fail", func(t *testing.T) {
		resp, err := testClient.GetWorkflowBeta1(ctx, &GetWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	// 3: PauseWorkflow
	t.Run("pause workflow", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		require.NoError(t, err)
	})

	t.Run("pause workflow - valid (without component)", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		require.NoError(t, err)
	})

	t.Run("pause workflow invalid instanceid", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	t.Run("pause workflow", func(t *testing.T) {
		err := testClient.PauseWorkflowBeta1(ctx, &PauseWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	// 4: ResumeWorkflow
	t.Run("resume workflow", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		require.NoError(t, err)
	})

	t.Run("resume workflow - valid (without component)", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		require.NoError(t, err)
	})

	t.Run("resume workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	t.Run("resume workflow - grpc fail", func(t *testing.T) {
		err := testClient.ResumeWorkflowBeta1(ctx, &ResumeWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	// 5: TerminateWorkflow
	t.Run("terminate workflow", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		require.NoError(t, err)
	})

	t.Run("terminate workflow - valid (without component)", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		require.NoError(t, err)
	})

	t.Run("terminate workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	t.Run("terminate workflow - grpc failure", func(t *testing.T) {
		err := testClient.TerminateWorkflowBeta1(ctx, &TerminateWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	// 6: RaiseEventWorkflow
	t.Run("raise event workflow", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		require.NoError(t, err)
	})

	t.Run("raise event workflow - valid (without component)", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
			EventName:         "TestEvent",
		})
		require.NoError(t, err)
	})

	t.Run("raise event workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		require.Error(t, err)
	})

	t.Run("raise event workflow - invalid eventname", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "",
		})
		require.Error(t, err)
	})

	t.Run("raise event workflow - grpc failure", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
		})
		require.Error(t, err)
	})
	t.Run("raise event workflow - cannot serialize input", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         math.NaN(),
			SendRawData:       false,
		})
		require.Error(t, err)
	})
	t.Run("raise event workflow - raw input", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         []byte("teststring"),
			SendRawData:       true,
		})
		require.NoError(t, err)
	})

	t.Run("raise event workflow - raw input (invalid)", func(t *testing.T) {
		err := testClient.RaiseEventWorkflowBeta1(ctx, &RaiseEventWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
			EventName:         "TestEvent",
			EventData:         "test string",
			SendRawData:       true,
		})
		require.Error(t, err)
	})

	// 7: PurgeWorkflow
	t.Run("purge workflow", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "dapr",
		})
		require.NoError(t, err)
	})

	t.Run("purge workflow - valid (without component)", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "TestID",
			WorkflowComponent: "",
		})
		require.NoError(t, err)
	})

	t.Run("purge workflow - invalid instanceid", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        "",
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})

	t.Run("purge workflow - grpc failure", func(t *testing.T) {
		err := testClient.PurgeWorkflowBeta1(ctx, &PurgeWorkflowRequest{
			InstanceID:        testWorkflowFailureID,
			WorkflowComponent: "dapr",
		})
		require.Error(t, err)
	})
}
