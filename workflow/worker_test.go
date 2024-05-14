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
	"testing"

	daprClient "github.com/dapr/go-sdk/client"

	"github.com/microsoft/durabletask-go/task"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntime(t *testing.T) {
	t.Run("failure to create newruntime without dapr", func(t *testing.T) {
		wr, err := NewWorker()
		require.Error(t, err)
		assert.Empty(t, wr)
	})
}

func TestWorkflowRuntime(t *testing.T) {
	testWorker := WorkflowWorker{
		tasks:  task.NewTaskRegistry(),
		client: nil,
	}

	// TODO: Mock grpc conn - currently requires dapr to be available
	t.Run("register workflow", func(t *testing.T) {
		err := testWorker.RegisterWorkflow(testWorkflow)
		require.NoError(t, err)
	})
	t.Run("register workflow - anonymous func", func(t *testing.T) {
		err := testWorker.RegisterWorkflow(func(ctx *WorkflowContext) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
	})
	t.Run("register activity", func(t *testing.T) {
		err := testWorker.RegisterActivity(testActivity)
		require.NoError(t, err)
	})
	t.Run("register activity - anonymous func", func(t *testing.T) {
		err := testWorker.RegisterActivity(func(ctx ActivityContext) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
	})
}

func TestWorkerOptions(t *testing.T) {
	t.Run("worker client option", func(t *testing.T) {
		options := returnWorkerOptions(WorkerWithDaprClient(&daprClient.GRPCClient{}))
		assert.NotNil(t, options.daprClient)
	})
}

func returnWorkerOptions(opts ...workerOption) workerOptions {
	options := new(workerOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return *options
		}
	}
	return *options
}

func TestWrapWorkflow(t *testing.T) {
	t.Run("wrap workflow", func(t *testing.T) {
		orchestrator := wrapWorkflow(testWorkflow)
		assert.NotNil(t, orchestrator)
	})
}

func TestWrapActivity(t *testing.T) {
	t.Run("wrap activity", func(t *testing.T) {
		activity := wrapActivity(testActivity)
		assert.NotNil(t, activity)
	})
}

func TestGetFunctionName(t *testing.T) {
	t.Run("get function name", func(t *testing.T) {
		name, err := getFunctionName(testWorkflow)
		require.NoError(t, err)
		assert.Equal(t, "testWorkflow", name)
	})
	t.Run("get function name - nil", func(t *testing.T) {
		name, err := getFunctionName(nil)
		require.Error(t, err)
		assert.Equal(t, "", name)
	})
}

func testWorkflow(ctx *WorkflowContext) (any, error) {
	_ = ctx
	return nil, nil
}

func testActivity(ctx ActivityContext) (any, error) {
	_ = ctx
	return nil, nil
}
