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

func (s Status) String() string {
	status := [...]string{
		"running",
		"completed",
		"continued_as_new",
		"failed",
		"canceled",
		"terminated",
		"pending",
		"suspended",
	}
	if s > StatusSuspended || s < StatusRunning {
		return "unknown"
	}
	return status[s]
}

type WorkflowState struct {
	Metadata api.OrchestrationMetadata
}

func (wfs *WorkflowState) RuntimeStatus() Status {
	s := Status(wfs.Metadata.RuntimeStatus.Number())
	return s
}
