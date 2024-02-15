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
	"fmt"
	"time"

	"github.com/microsoft/durabletask-go/api"
	"github.com/microsoft/durabletask-go/task"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Metadata struct {
	InstanceID             string          `json:"id"`
	Name                   string          `json:"name"`
	RuntimeStatus          Status          `json:"status"`
	CreatedAt              time.Time       `json:"createdAt"`
	LastUpdatedAt          time.Time       `json:"lastUpdatedAt"`
	SerializedInput        string          `json:"serializedInput"`
	SerializedOutput       string          `json:"serializedOutput"`
	SerializedCustomStatus string          `json:"serializedCustomStatus"`
	FailureDetails         *FailureDetails `json:"failureDetails"`
}

type FailureDetails struct {
	Type           string          `json:"type"`
	Message        string          `json:"message"`
	StackTrace     string          `json:"stackTrace"`
	InnerFailure   *FailureDetails `json:"innerFailure"`
	IsNonRetriable bool            `json:"IsNonRetriable"`
}

func convertMetadata(orchestrationMetadata *api.OrchestrationMetadata) *Metadata {
	metadata := Metadata{
		InstanceID:             string(orchestrationMetadata.InstanceID),
		Name:                   orchestrationMetadata.Name,
		RuntimeStatus:          Status(orchestrationMetadata.RuntimeStatus.Number()),
		CreatedAt:              orchestrationMetadata.CreatedAt,
		LastUpdatedAt:          orchestrationMetadata.LastUpdatedAt,
		SerializedInput:        orchestrationMetadata.SerializedInput,
		SerializedOutput:       orchestrationMetadata.SerializedOutput,
		SerializedCustomStatus: orchestrationMetadata.SerializedCustomStatus,
	}
	if orchestrationMetadata.FailureDetails != nil {
		metadata.FailureDetails = &FailureDetails{
			Type:           orchestrationMetadata.FailureDetails.GetErrorType(),
			Message:        orchestrationMetadata.FailureDetails.GetErrorMessage(),
			StackTrace:     orchestrationMetadata.FailureDetails.GetStackTrace().GetValue(),
			IsNonRetriable: orchestrationMetadata.FailureDetails.GetIsNonRetriable(),
		}

		if orchestrationMetadata.FailureDetails.GetInnerFailure() != nil {
			var root *FailureDetails
			current := root
			failure := orchestrationMetadata.FailureDetails.GetInnerFailure()
			for {
				current.Type = failure.GetErrorType()
				current.Message = failure.GetErrorMessage()
				if failure.GetStackTrace() != nil {
					current.StackTrace = failure.GetStackTrace().GetValue()
				}
				if failure.GetInnerFailure() == nil {
					break
				}
				failure = failure.GetInnerFailure()
				var inner *FailureDetails
				current.InnerFailure = inner
				current = inner
			}
			metadata.FailureDetails.InnerFailure = root
		}
	}
	return &metadata
}

type callChildWorkflowOptions struct {
	instanceID string
	rawInput   *wrapperspb.StringValue
}

type callChildWorkflowOption func(*callChildWorkflowOptions) error

// ChildWorkflowInput is an option to provide a JSON-serializable input when calling a child workflow.
func ChildWorkflowInput(input any) callChildWorkflowOption {
	return func(opts *callChildWorkflowOptions) error {
		bytes, err := marshalData(input)
		if err != nil {
			return fmt.Errorf("failed to marshal input data to JSON: %v", err)
		}
		opts.rawInput = wrapperspb.String(string(bytes))
		return nil
	}
}

// ChildWorkflowRawInput is an option to provide a byte slice input when calling a child workflow.
func ChildWorkflowRawInput(input string) callChildWorkflowOption {
	return func(opts *callChildWorkflowOptions) error {
		opts.rawInput = wrapperspb.String(input)
		return nil
	}
}

// ChildWorkflowInstanceID is an option to provide an instance id when calling a child workflow.
func ChildWorkflowInstanceID(instanceID string) callChildWorkflowOption {
	return func(opts *callChildWorkflowOptions) error {
		opts.instanceID = instanceID
		return nil
	}
}

// NewTaskSlice returns a slice of tasks which can be executed in parallel
func NewTaskSlice(length int) []task.Task {
	taskSlice := make([]task.Task, length)
	return taskSlice
}
