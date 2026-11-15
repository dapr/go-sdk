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
	"bytes"
	"testing"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

func TestWithTerminateReasonPreservesUnknownFields(t *testing.T) {
	const reason = "operator requested termination"
	const unrelatedField protowire.Number = 100

	request := &protos.TerminateRequest{InstanceId: "workflow-1"}
	unrelated := protowire.AppendTag(nil, unrelatedField, protowire.VarintType)
	unrelated = protowire.AppendVarint(unrelated, 7)
	request.ProtoReflect().SetUnknown(unrelated)

	require.NoError(t, WithTerminateReason(reason)(request))

	encoded, err := proto.Marshal(request)
	require.NoError(t, err)

	decoded := new(protos.TerminateRequest)
	require.NoError(t, proto.Unmarshal(encoded, decoded))

	got, found, err := getStringWrapperField(
		decoded.ProtoReflect(),
		terminateRequestReasonFieldNumber,
		"reason",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, reason, got)
	assert.True(t, bytes.HasPrefix(decoded.ProtoReflect().GetUnknown(), unrelated))
}

func TestWithTerminateReasonEmptyValueIsPresent(t *testing.T) {
	request := new(protos.TerminateRequest)

	require.NoError(t, WithTerminateReason("")(request))

	got, found, err := getStringWrapperField(
		request.ProtoReflect(),
		terminateRequestReasonFieldNumber,
		"reason",
	)
	require.NoError(t, err)
	require.True(t, found)
	assert.Empty(t, got)
}

func TestGetWorkflowTerminationReason(t *testing.T) {
	const reason = "operator requested termination"

	encoded, err := proto.Marshal(terminationEvent(2, &reason))
	require.NoError(t, err)

	decoded := new(protos.HistoryEvent)
	require.NoError(t, proto.Unmarshal(encoded, decoded))

	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			{EventId: 1},
			decoded,
		},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, reason, got)
}

func TestGetWorkflowTerminationReasonEmptyValueIsPresent(t *testing.T) {
	empty := ""
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{terminationEvent(1, &empty)},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	require.True(t, found)
	assert.Empty(t, got)
}

func TestGetWorkflowTerminationReasonUsesFirstTerminationEvent(t *testing.T) {
	laterReason := "later reason"
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			terminationEvent(1, nil),
			terminationEvent(2, &laterReason),
		},
	}

	got, found, err := GetWorkflowTerminationReason(&history)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, got)
}

func TestGetWorkflowTerminationReasonWithoutTerminationEvent(t *testing.T) {
	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			{EventId: 1},
			nil,
		},
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

func TestGetWorkflowTerminationReasonRejectsMalformedUnknownField(t *testing.T) {
	terminated := &protos.ExecutionTerminatedEvent{}
	terminated.ProtoReflect().SetUnknown([]byte{0x1a, 0x02, 0x0a})

	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			{
				EventId: 42,
				EventType: &protos.HistoryEvent_ExecutionTerminated{
					ExecutionTerminated: terminated,
				},
			},
		},
	}

	_, found, err := GetWorkflowTerminationReason(&history)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "history event 42")
	assert.False(t, found)
}


func TestGetWorkflowTerminationReasonRejectsUnexpectedEndGroup(t *testing.T) {
	terminated := &protos.ExecutionTerminatedEvent{}
	// Field 3 with the end-group wire type is invalid at this message level.
	terminated.ProtoReflect().SetUnknown([]byte{0x1c})

	history := workflow.GetInstanceHistoryResponse{
		Events: []*protos.HistoryEvent{
			{
				EventId: 43,
				EventType: &protos.HistoryEvent_ExecutionTerminated{
					ExecutionTerminated: terminated,
				},
			},
		},
	}

	_, found, err := GetWorkflowTerminationReason(&history)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "history event 43")
	assert.False(t, found)
}

func terminationEvent(eventID int32, reason *string) *protos.HistoryEvent {
	terminated := &protos.ExecutionTerminatedEvent{}
	if reason != nil {
		wrapper := protowire.AppendString(nil, 1, *reason)
		unknown := protowire.AppendTag(nil, executionTerminatedReasonFieldNumber, protowire.BytesType)
		unknown = protowire.AppendBytes(unknown, wrapper)
		terminated.ProtoReflect().SetUnknown(unknown)
	}

	return &protos.HistoryEvent{
		EventId: eventID,
		EventType: &protos.HistoryEvent_ExecutionTerminated{
			ExecutionTerminated: terminated,
		},
	}
}
