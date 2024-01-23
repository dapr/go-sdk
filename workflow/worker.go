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

type WorkflowWorker struct {
	tasks  *task.TaskRegistry
	client *durabletaskclient.TaskHubGrpcClient

	mutex  sync.Mutex // TODO: implement
	quit   chan bool
	close  func()
	cancel context.CancelFunc
}

type Workflow func(ctx *WorkflowContext) (any, error)

type Activity func(ctx ActivityContext) (any, error)

func NewWorker() (*WorkflowWorker, error) {
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, err
	}

	return &WorkflowWorker{
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
		wfCtx := &WorkflowContext{orchestrationContext: ctx}
		return w(wfCtx)
	}
}

func (ww *WorkflowWorker) RegisterWorkflow(w Workflow) error {
	wrappedOrchestration := wrapWorkflow(w)

	// get the function name for the passed workflow
	name, err := getFunctionName(w)
	if err != nil {
		return fmt.Errorf("failed to get workflow decorator: %v", err)
	}

	err = ww.tasks.AddOrchestratorN(name, wrappedOrchestration)
	return err
}

func wrapActivity(a Activity) task.Activity {
	return func(ctx task.ActivityContext) (any, error) {
		aCtx := ActivityContext{ctx: ctx}

		return a(aCtx)
	}
}

func (ww *WorkflowWorker) RegisterActivity(a Activity) error {
	wrappedActivity := wrapActivity(a)

	// get the function name for the passed activity
	name, err := getFunctionName(a)
	if err != nil {
		return fmt.Errorf("failed to get activity decorator: %v", err)
	}

	err = ww.tasks.AddActivityN(name, wrappedActivity)
	return err
}

func (ww *WorkflowWorker) Start() error {
	// go func start
	errChan := make(chan error)
	go func() {
		defer ww.close()
		ctx, cancel := context.WithCancel(context.Background())
		err := ww.client.StartWorkItemListener(ctx, ww.tasks)
		if err != nil {
			cancel()
			errChan <- fmt.Errorf("failed to start work stream: %v", err)
			return
		}
		ww.cancel = cancel
		log.Println("work item listener started")
		errChan <- nil
		<-ww.quit
		log.Println("work item listener shutdown")
	}()
	err := <-errChan
	return err
}

func (ww *WorkflowWorker) Shutdown() error {
	ww.cancel()
	// send close signal
	ww.quit <- true
	log.Println("work item listener shutdown signal sent")
	return nil
}
