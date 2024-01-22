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
	"log"
	"time"

	"github.com/microsoft/durabletask-go/task"
)

type Context struct {
	orchestrationContext *task.OrchestrationContext
}

func (wfc *Context) GetInput(v interface{}) error {
	return wfc.orchestrationContext.GetInput(&v)
}

func (wfc *Context) Name() string {
	return wfc.orchestrationContext.Name
}

// InstanceID returns the ID of the currently executing workflow
func (wfc *Context) InstanceID() string {
	return fmt.Sprintf("%v", wfc.orchestrationContext.ID)
}

// CurrentUTCDateTime returns the current workflow time as UTC. Note that this should be used instead of `time.Now()`, which is not compatible with workflow replays.
func (wfc *Context) CurrentUTCDateTime() time.Time {
	return wfc.orchestrationContext.CurrentTimeUtc
}

func (wfc *Context) IsReplaying() bool {
	return wfc.orchestrationContext.IsReplaying
}

func (wfc *Context) CallActivity(activity interface{}, opts ...callActivityOption) task.Task {
	var inp any
	if err := wfc.GetInput(&inp); err != nil {
		log.Printf("unable to get activity input: %v", err)
	}
	// the call should continue despite being unable to obtain an input

	return wfc.orchestrationContext.CallActivity(activity, task.WithActivityInput(inp))
}

func (wfc *Context) CallChildWorkflow(workflow interface{}) task.Task {
	return wfc.orchestrationContext.CallSubOrchestrator(workflow)
}

func (wfc *Context) CreateTimer(duration time.Duration) task.Task {
	return wfc.orchestrationContext.CreateTimer(duration)
}

func (wfc *Context) WaitForExternalEvent(eventName string, timeout time.Duration) task.Task {
	if eventName == "" {
		return nil
	}
	if timeout == 0 {
		// default to 10 seconds
		timeout = time.Second * 10
	}
	return wfc.orchestrationContext.WaitForSingleEvent(eventName, timeout)
}

func (wfc *Context) ContinueAsNew(newInput any, keepEvents bool) {
	if !keepEvents {
		wfc.orchestrationContext.ContinueAsNew(newInput)
	}
	wfc.orchestrationContext.ContinueAsNew(newInput, task.WithKeepUnprocessedEvents())
}
