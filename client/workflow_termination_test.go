/*
Copyright 2026 The Dapr Authors
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
	"testing"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestWithTerminateReasonUsesExistingOutputField(t *testing.T) {
	const reason = "operator requested termination"
	request := &protos.TerminateRequest{
		InstanceId: "workflow-1",
		Recursive:  true,
	}
	require.NoError(t, WithTerminateReason(reason)(request))

	assert.Equal(t, "workflow-1", request.GetInstanceId())
	assert.True(t, request.GetRecursive())
	assert.Equal(t, wrapperspb.String(reason), request.GetOutput())

	encoded, err := proto.Marshal(request)
	require.NoError(t, err)
	decoded := new(protos.TerminateRequest)
	require.NoError(t, proto.Unmarshal(encoded, decoded))
	assert.Equal(t, wrapperspb.String(reason), decoded.GetOutput())
}


func TestWithTerminateReasonRejectsExistingOutput(t *testing.T) {
	request := &protos.TerminateRequest{
		Output: wrapperspb.String("existing output"),
	}

	err := WithTerminateReason("operator requested termination")(request)
	require.ErrorContains(t, err, "termination output and reason cannot be supplied together")
	assert.Equal(t, "existing output", request.GetOutput().GetValue())
}

func TestWithTerminateReasonPreservesExplicitEmptyValue(t *testing.T) {
	request := new(protos.TerminateRequest)
	require.NoError(t, WithTerminateReason("")(request))
	require.NotNil(t, request.GetOutput())
	assert.Empty(t, request.GetOutput().GetValue())
}

func TestWithTerminateReasonRejectsNilRequest(t *testing.T) {
	require.ErrorContains(t, WithTerminateReason("reason")(nil), "terminate request is nil")
}

func terminationEvent(eventID int32, input *wrapperspb.StringValue) *protos.HistoryEvent {
	return &protos.HistoryEvent{
		EventId: eventID,
		EventType: &protos.HistoryEvent_ExecutionTerminated{
			ExecutionTerminated: &protos.ExecutionTerminatedEvent{Input: input},
		},
	}
}

func TestGetWorkflowTerminationReason(t *testing.T) {
	const reason = "operator requested termination"
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			{EventId: 1},
			terminationEvent(2, wrapperspb.String(reason)),
		},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, reason, got)
}

func TestGetWorkflowTerminationReasonEmptyValueIsPresent(t *testing.T) {
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{terminationEvent(1, wrapperspb.String(""))},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	require.True(t, found)
	assert.Empty(t, got)
}

func TestGetWorkflowTerminationReasonUsesFirstTerminationEvent(t *testing.T) {
	laterReason := wrapperspb.String("later reason")
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			terminationEvent(1, nil),
			terminationEvent(2, laterReason),
		},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, got)
}

func TestGetWorkflowTerminationReasonWithoutTerminationEvent(t *testing.T) {
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{{EventId: 1}, nil},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, got)
}

func TestGetWorkflowTerminationReasonNilHistory(t *testing.T) {
	_, found, err := GetWorkflowTerminationReason(nil)
	require.ErrorContains(t, err, "workflow history is nil")
	assert.False(t, found)
}
