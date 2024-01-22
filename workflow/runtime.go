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
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/microsoft/durabletask-go/backend"
	durabletaskclient "github.com/microsoft/durabletask-go/client"
	"github.com/microsoft/durabletask-go/task"
)

type WorkflowRuntime struct {
	tasks  *task.TaskRegistry
	client *durabletaskclient.TaskHubGrpcClient

	mutex sync.Mutex // TODO: implement
	quit  chan bool
	close func()
}

type Workflow func(ctx *Context) (any, error)

type Activity func(ctx ActivityContext) (any, error)

func NewRuntime() (*WorkflowRuntime, error) {
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, err
	}

	return &WorkflowRuntime{
		tasks:  task.NewTaskRegistry(),
		client: durabletaskclient.NewTaskHubGrpcClient(daprClient.GrpcClientConn(), backend.DefaultLogger()),
		quit:   make(chan bool),
		close:  daprClient.Close,
	}, nil
}

// getFunctionName returns the function name as a string
func getFunctionName(f interface{}) (string, error) {
	if f == nil {
		return "", errors.New("nil function name")
	}

	callSplit := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), ".")

	funcName := callSplit[len(callSplit)-1]

	if funcName == "1" {
		return "", errors.New("anonymous function name")
	}

	return funcName, nil
}

func wrapWorkflow(w Workflow) task.Orchestrator {
	return func(ctx *task.OrchestrationContext) (any, error) {
		wfCtx := &Context{orchestrationContext: ctx}
		return w(wfCtx)
	}
}

func (wr *WorkflowRuntime) RegisterWorkflow(w Workflow) error {
	wrappedOrchestration := wrapWorkflow(w)

	// get decorator for workflow
	name, err := getFunctionName(w)
	if err != nil {
		return fmt.Errorf("failed to get workflow decorator: %v", err)
	}

	err = wr.tasks.AddOrchestratorN(name, wrappedOrchestration)
	return err
}

func wrapActivity(a Activity) task.Activity {
	return func(ctx task.ActivityContext) (any, error) {
		aCtx := ActivityContext{ctx: ctx}

		return a(aCtx)
	}
}

func (wr *WorkflowRuntime) RegisterActivity(a Activity) error {
	wrappedActivity := wrapActivity(a)

	// get decorator for activity
	name, err := getFunctionName(a)
	if err != nil {
		return fmt.Errorf("failed to get activity decorator: %v", err)
	}

	err = wr.tasks.AddActivityN(name, wrappedActivity)
	return err
}

func (wr *WorkflowRuntime) Start() error {
	// go func start
	go func() {
		defer wr.close()
		err := wr.client.StartWorkItemListener(context.Background(), wr.tasks)
		if err != nil {
			log.Fatalf("failed to start work stream: %v", err)
		}
		log.Println("work item listener started")
		<-wr.quit
		log.Println("work item listener shutdown")
	}()
	return nil
}

func (wr *WorkflowRuntime) Shutdown() error {
	// send close signal
	wr.quit <- true
	log.Println("work item listener shutdown signal sent")
	return nil
}
