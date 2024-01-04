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

func (wfs *WorkflowState) RuntimeStatus() Status {
	s := Status(wfs.Metadata.RuntimeStatus.Number())
	return s
}
