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

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/task"
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

func convertMetadata(orchestrationMetadata *protos.OrchestrationMetadata) *Metadata {
	metadata := Metadata{
		InstanceID:             orchestrationMetadata.GetInstanceId(),
		Name:                   orchestrationMetadata.GetName(),
		RuntimeStatus:          Status(orchestrationMetadata.GetRuntimeStatus().Number()),
		CreatedAt:              orchestrationMetadata.GetCreatedAt().AsTime(),
		LastUpdatedAt:          orchestrationMetadata.GetLastUpdatedAt().AsTime(),
		SerializedInput:        orchestrationMetadata.GetInput().GetValue(),
		SerializedOutput:       orchestrationMetadata.GetOutput().GetValue(),
		SerializedCustomStatus: orchestrationMetadata.GetCustomStatus().GetValue(),
	}
	if orchestrationMetadata.GetFailureDetails() != nil {
		metadata.FailureDetails = &FailureDetails{
			Type:           orchestrationMetadata.GetFailureDetails().GetErrorType(),
			Message:        orchestrationMetadata.GetFailureDetails().GetErrorMessage(),
			StackTrace:     orchestrationMetadata.GetFailureDetails().GetStackTrace().GetValue(),
			IsNonRetriable: orchestrationMetadata.GetFailureDetails().GetIsNonRetriable(),
		}

		if orchestrationMetadata.GetFailureDetails().GetInnerFailure() != nil {
			var root *FailureDetails
			current := root
			failure := orchestrationMetadata.GetFailureDetails().GetInnerFailure()
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
	instanceID  string
	rawInput    *wrapperspb.StringValue
	retryPolicy *RetryPolicy
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

func ChildWorkflowRetryPolicy(policy RetryPolicy) callChildWorkflowOption {
	return func(opts *callChildWorkflowOptions) error {
		opts.retryPolicy = &policy
		return nil
	}
}

func (opts *callChildWorkflowOptions) getRetryPolicy() *task.RetryPolicy {
	if opts.retryPolicy == nil {
		return nil
	}
	return &task.RetryPolicy{
		MaxAttempts:          opts.retryPolicy.MaxAttempts,
		InitialRetryInterval: opts.retryPolicy.InitialRetryInterval,
		BackoffCoefficient:   opts.retryPolicy.BackoffCoefficient,
		MaxRetryInterval:     opts.retryPolicy.MaxRetryInterval,
		RetryTimeout:         opts.retryPolicy.RetryTimeout,
	}
}

// NewTaskSlice returns a slice of tasks which can be executed in parallel
func NewTaskSlice(length int) []task.Task {
	taskSlice := make([]task.Task, length)
	return taskSlice
}
