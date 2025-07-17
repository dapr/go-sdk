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

	dapr "github.com/dapr/go-sdk/client"

	"github.com/dapr/durabletask-go/backend"
	durabletaskclient "github.com/dapr/durabletask-go/client"
	"github.com/dapr/durabletask-go/task"
)

type WorkflowWorker struct {
	tasks  *task.TaskRegistry
	client *durabletaskclient.TaskHubGrpcClient

	close  func()
	cancel context.CancelFunc
}

type Workflow func(ctx *WorkflowContext) (any, error)

type Activity func(ctx ActivityContext) (any, error)

type workerOption func(*workerOptions) error

type workerOptions struct {
	daprClient dapr.Client
}

// WorkerWithDaprClient allows you to specify a custom dapr.Client for the worker.
func WorkerWithDaprClient(input dapr.Client) workerOption {
	return func(opts *workerOptions) error {
		opts.daprClient = input
		return nil
	}
}

// NewWorker returns a worker that can interface with the workflow engine
func NewWorker(opts ...workerOption) (*WorkflowWorker, error) {
	options := new(workerOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return nil, errors.New("failed to load options")
		}
	}
	var daprClient dapr.Client
	var err error
	if options.daprClient == nil {
		daprClient, err = dapr.NewClient()
	} else {
		daprClient = options.daprClient
	}
	if err != nil {
		return nil, err
	}
	grpcConn := daprClient.GrpcClientConn()

	return &WorkflowWorker{
		tasks:  task.NewTaskRegistry(),
		client: durabletaskclient.NewTaskHubGrpcClient(grpcConn, backend.DefaultLogger()),
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

type registerOptions struct {
	Name string
}

type registerOption func(*registerOptions) error

// WithName allows you to specify a custom name for the workflow or activity being registered.
// Activities and Workflows registered without an explicit name will use the function name as the name.
func WithName(name string) registerOption {
	return func(opts *registerOptions) error {
		opts.Name = name
		return nil
	}
}

func processRegisterOptions(defaultOptions registerOptions, opts ...registerOption) (registerOptions, error) {
	options := defaultOptions
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return options, fmt.Errorf("failed processing options: %w", err)
		}
	}
	return options, nil
}

// RegisterWorkflow adds a workflow function to the registry
func (ww *WorkflowWorker) RegisterWorkflow(w Workflow, opts ...registerOption) error {
	wrappedOrchestration := wrapWorkflow(w)

	options, err := processRegisterOptions(registerOptions{}, opts...)
	if err != nil {
		return err
	}

	if options.Name == "" {
		// get the function name for the passed workflow if there's
		// no explicit name provided.
		name, err := getFunctionName(w)
		if err != nil {
			return fmt.Errorf("failed to get workflow decorator: %v", err)
		}
		options.Name = name
	}

	return ww.tasks.AddOrchestratorN(options.Name, wrappedOrchestration)
}

func wrapActivity(a Activity) task.Activity {
	return func(ctx task.ActivityContext) (any, error) {
		aCtx := ActivityContext{ctx: ctx}

		result, err := a(aCtx)
		if err != nil {
			activityName, _ := getFunctionName(a) // Get the activity name for context
			return nil, fmt.Errorf("activity %s failed: %w", activityName, err)
		}

		return result, nil
	}
}

// RegisterActivity adds an activity function to the registry
func (ww *WorkflowWorker) RegisterActivity(a Activity, opts ...registerOption) error {
	wrappedActivity := wrapActivity(a)

	options, err := processRegisterOptions(registerOptions{}, opts...)
	if err != nil {
		return err
	}

	if options.Name == "" {
		// get the function name for the passed workflow if there's
		// no explicit name provided.
		name, err := getFunctionName(a)
		if err != nil {
			return fmt.Errorf("failed to get activity decorator: %v", err)
		}
		options.Name = name
	}

	return ww.tasks.AddActivityN(options.Name, wrappedActivity)
}

// Start initialises a non-blocking worker to handle workflows and activities registered
// prior to this being called.
func (ww *WorkflowWorker) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	ww.cancel = cancel
	if err := ww.client.StartWorkItemListener(ctx, ww.tasks); err != nil {
		return fmt.Errorf("failed to start work stream: %v", err)
	}
	log.Println("work item listener started")
	return nil
}

// Shutdown stops the worker
func (ww *WorkflowWorker) Shutdown() error {
	ww.cancel()
	ww.close()
	log.Println("work item listener shutdown")
	return nil
}
