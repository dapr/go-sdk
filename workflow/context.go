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

func (wfc *WorkflowContext) GetInput(v interface{}) error {
	return wfc.orchestrationContext.GetInput(&v)
}

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

func (wfc *WorkflowContext) IsReplaying() bool {
	return wfc.orchestrationContext.IsReplaying
}

func (wfc *WorkflowContext) CallActivity(activity interface{}, opts ...callActivityOption) task.Task {
	options := new(callActivityOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return nil
		}
	}

	return wfc.orchestrationContext.CallActivity(activity, task.WithRawActivityInput(options.rawInput.GetValue()))
}

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

func (wfc *WorkflowContext) CreateTimer(duration time.Duration) task.Task {
	return wfc.orchestrationContext.CreateTimer(duration)
}

func (wfc *WorkflowContext) WaitForExternalEvent(eventName string, timeout time.Duration) task.Task {
	if eventName == "" {
		return nil
	}
	if timeout == 0 {
		// default to 10 seconds
		timeout = time.Second * 10
	}
	return wfc.orchestrationContext.WaitForSingleEvent(eventName, timeout)
}

func (wfc *WorkflowContext) ContinueAsNew(newInput any, keepEvents bool) {
	if !keepEvents {
		wfc.orchestrationContext.ContinueAsNew(newInput)
	}
	wfc.orchestrationContext.ContinueAsNew(newInput, task.WithKeepUnprocessedEvents())
}
