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

	"github.com/microsoft/durabletask-go/task"
)

type WorkflowContext struct {
	orchestrationContext *task.OrchestrationContext
}

// GetInput casts the input from the context to a specified interface.
func (wfc *WorkflowContext) GetInput(v interface{}) error {
	return wfc.orchestrationContext.GetInput(&v)
}

// Name returns the name string from the workflow context.
func (wfc *WorkflowContext) Name() string {
	return wfc.orchestrationContext.Name
}

// InstanceID returns the ID of the currently executing workflow
func (wfc *WorkflowContext) InstanceID() string {
	return fmt.Sprintf("%v", wfc.orchestrationContext.ID)
}

// CurrentUTCDateTime returns the current workflow time as UTC. Note that this should be used instead of `time.Now()`, which is not compatible with workflow replays.
func (wfc *WorkflowContext) CurrentUTCDateTime() time.Time {
	return wfc.orchestrationContext.CurrentTimeUtc
}

// IsReplaying returns whether the workflow is replaying.
func (wfc *WorkflowContext) IsReplaying() bool {
	return wfc.orchestrationContext.IsReplaying
}

// CallActivity returns a completable task for a given activity.
// You must call Await(output any) on the returned Task to block the workflow and wait for the task to complete.
// The value passed to the Await method must be a pointer or can be nil to ignore the returned value.
// Alternatively, tasks can be awaited using the task.WhenAll or task.WhenAny methods, allowing the workflow
// to block and wait for multiple tasks at the same time.
func (wfc *WorkflowContext) CallActivity(activity interface{}, opts ...callActivityOption) task.Task {
	options := new(callActivityOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return nil
		}
	}

	return wfc.orchestrationContext.CallActivity(activity, task.WithRawActivityInput(options.rawInput.GetValue()))
}

// CallChildWorkflow returns a completable task for a given workflow.
// You must call Await(output any) on the returned Task to block the workflow and wait for the task to complete.
// The value passed to the Await method must be a pointer or can be nil to ignore the returned value.
// Alternatively, tasks can be awaited using the task.WhenAll or task.WhenAny methods, allowing the workflow
// to block and wait for multiple tasks at the same time.
func (wfc *WorkflowContext) CallChildWorkflow(workflow interface{}, opts ...callChildWorkflowOption) task.Task {
	options := new(callChildWorkflowOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return nil
		}
	}
	if options.instanceID != "" {
		return wfc.orchestrationContext.CallSubOrchestrator(workflow, task.WithRawSubOrchestratorInput(options.rawInput.GetValue()), task.WithSubOrchestrationInstanceID(options.instanceID))
	}
	return wfc.orchestrationContext.CallSubOrchestrator(workflow, task.WithRawSubOrchestratorInput(options.rawInput.GetValue()))
}

// CreateTimer returns a completable task that blocks for a given duration.
// You must call Await(output any) on the returned Task to block the workflow and wait for the task to complete.
// The value passed to the Await method must be a pointer or can be nil to ignore the returned value.
// Alternatively, tasks can be awaited using the task.WhenAll or task.WhenAny methods, allowing the workflow
// to block and wait for multiple tasks at the same time.
func (wfc *WorkflowContext) CreateTimer(duration time.Duration) task.Task {
	return wfc.orchestrationContext.CreateTimer(duration)
}

// WaitForExternalEvent returns a completabel task that waits for a given event to be received.
// You must call Await(output any) on the returned Task to block the workflow and wait for the task to complete.
// The value passed to the Await method must be a pointer or can be nil to ignore the returned value.
// Alternatively, tasks can be awaited using the task.WhenAll or task.WhenAny methods, allowing the workflow
// to block and wait for multiple tasks at the same time.
func (wfc *WorkflowContext) WaitForExternalEvent(eventName string, timeout time.Duration) task.Task {
	if eventName == "" {
		return nil
	}
	return wfc.orchestrationContext.WaitForSingleEvent(eventName, timeout)
}

// ContinueAsNew configures the workflow.
func (wfc *WorkflowContext) ContinueAsNew(newInput any, keepEvents bool) {
	if !keepEvents {
		wfc.orchestrationContext.ContinueAsNew(newInput)
	}
	wfc.orchestrationContext.ContinueAsNew(newInput, task.WithKeepUnprocessedEvents())
}
