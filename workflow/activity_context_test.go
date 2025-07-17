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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/durabletask-go/task"
)

type testingTaskActivityContext struct {
	inputBytes      []byte
	ctx             context.Context
	taskExecutionID string
}

func (t *testingTaskActivityContext) GetTaskID() int32 {
	return 0
}

func (t *testingTaskActivityContext) GetTaskExecutionID() string {
	return t.taskExecutionID
}

func (t *testingTaskActivityContext) GetInput(v any) error {
	return json.Unmarshal(t.inputBytes, &v)
}

func (t *testingTaskActivityContext) Context() context.Context {
	return t.ctx
}

func TestActivityContext(t *testing.T) {
	inputString := "testInputString"
	inputBytes, err := json.Marshal(inputString)
	require.NoErrorf(t, err, "required no error, but got %v", err)

	ac := ActivityContext{ctx: &testingTaskActivityContext{inputBytes: inputBytes, ctx: t.Context()}}
	t.Run("test getinput", func(t *testing.T) {
		var inputReturn string
		err := ac.GetInput(&inputReturn)
		require.NoError(t, err)
		assert.Equal(t, inputString, inputReturn)
	})

	t.Run("test context", func(t *testing.T) {
		assert.Equal(t, t.Context(), ac.Context())
	})
}

func TestCallActivityOptions(t *testing.T) {
	t.Run("activity input - valid", func(t *testing.T) {
		opts := returnCallActivityOptions(ActivityInput("test"))
		assert.Equal(t, "\"test\"", opts.rawInput.GetValue())
	})

	t.Run("activity input - invalid", func(t *testing.T) {
		opts := returnCallActivityOptions(ActivityInput(make(chan int)))
		assert.Empty(t, opts.rawInput.GetValue())
	})

	t.Run("activity raw input - valid", func(t *testing.T) {
		opts := returnCallActivityOptions(ActivityRawInput("test"))
		assert.Equal(t, "test", opts.rawInput.GetValue())
	})

	t.Run("activity retry policy - set", func(t *testing.T) {
		opts := returnCallActivityOptions(ActivityRetryPolicy(RetryPolicy{
			MaxAttempts:          3,
			InitialRetryInterval: 100 * time.Millisecond,
			BackoffCoefficient:   2,
			MaxRetryInterval:     2 * time.Second,
		}))
		assert.Equal(t, &task.RetryPolicy{
			MaxAttempts:          3,
			InitialRetryInterval: 100 * time.Millisecond,
			BackoffCoefficient:   2,
			MaxRetryInterval:     2 * time.Second,
		}, opts.getRetryPolicy())
	})

	t.Run("activity retry policy - empty", func(t *testing.T) {
		opts := returnCallActivityOptions()
		assert.Empty(t, opts.getRetryPolicy())
	})
}

func returnCallActivityOptions(opts ...callActivityOption) callActivityOptions {
	options := new(callActivityOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return *options
		}
	}
	return *options
}

func TestMarshalData(t *testing.T) {
	t.Run("test nil input", func(t *testing.T) {
		out, err := marshalData(nil)
		require.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("test string input", func(t *testing.T) {
		out, err := marshalData("testString")
		require.NoError(t, err)
		fmt.Println(out)
		assert.Equal(t, []byte{0x22, 0x74, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22}, out)
	})
}

func TestTaskExecutionID(t *testing.T) {
	ac := ActivityContext{ctx: &testingTaskActivityContext{ctx: t.Context(), taskExecutionID: "testTaskExecutionID"}}

	t.Run("test getTaskExecutionID", func(t *testing.T) {
		assert.Equal(t, "testTaskExecutionID", ac.GetTaskExecutionID())
	})
}

func TestTaskID(t *testing.T) {
	ac := ActivityContext{ctx: &testingTaskActivityContext{ctx: t.Context(), taskExecutionID: "testTaskExecutionID"}}

	t.Run("test getTaskID", func(t *testing.T) {
		assert.EqualValues(t, 0, ac.ctx.GetTaskID())
	})
}
