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

import "github.com/microsoft/durabletask-go/api"

type Status int

const (
	StatusRunning Status = iota
	StatusCompleted
	StatusContinuedAsNew
	StatusFailed
	StatusCanceled
	StatusTerminated
	StatusPending
	StatusSuspended
	StatusUnknown
)

// String returns the runtime status as a string.
func (s Status) String() string {
	status := [...]string{
		"RUNNING",
		"COMPLETED",
		"CONTINUED_AS_NEW",
		"FAILED",
		"CANCELED",
		"TERMINATED",
		"PENDING",
		"SUSPENDED",
	}
	if s > StatusSuspended || s < StatusRunning {
		return "UNKNOWN"
	}
	return status[s]
}

type WorkflowState struct {
	Metadata api.OrchestrationMetadata
}

// RuntimeStatus returns the status from a workflow state.
func (wfs *WorkflowState) RuntimeStatus() Status {
	s := Status(wfs.Metadata.RuntimeStatus.Number())
	return s
}
