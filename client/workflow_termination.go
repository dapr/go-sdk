/*
Copyright 2026 The Dapr Authors
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

package client

import (
	"errors"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/workflow"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// WithTerminateReason configures the value persisted in the existing
// termination input field. The runtime exposes this value as workflow
// custom status for a terminated instance.
func WithTerminateReason(reason string) workflow.TerminateOptions {
	return workflow.TerminateOptions(func(req *protos.TerminateRequest) error {
		if req == nil {
			return errors.New("terminate request is nil")
		}

		req.Output = wrapperspb.String(reason)
		return nil
	})
}

// GetWorkflowTerminationReason returns the input from the first
// ExecutionTerminated event in the workflow history. The bool is false when
// the first termination event has no input. An explicitly empty input is
// present and therefore returns true with an empty string.
func GetWorkflowTerminationReason(history *workflow.GetInstanceHistoryResponse) (string, bool, error) {
	if history == nil {
		return "", false, errors.New("workflow history is nil")
	}

	for _, event := range history.Events {
		if event == nil {
			continue
		}

		terminated := event.GetExecutionTerminated()
		if terminated == nil {
			continue
		}

		input := terminated.GetInput()
		if input == nil {
			return "", false, nil
		}

		return input.GetValue(), true, nil
	}

	return "", false, nil
}
