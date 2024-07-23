package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimepb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
)

// AddJobEventHandler registers a job handler
func (s *Server) AddJobEventHandler(name string, fn common.JobEventHandler) error {
	if name == "" {
		return errors.New("job event name cannot be empty")
	}

	if fn == nil {
		return errors.New("job event handler not supplied")
	}

	s.jobEventHandlers[name] = fn
	return nil
}

// OnJobEvent is invoked by the sidecar following a scheduled job registered in
// the scheduler
func (s *Server) OnJobEventAlpha1(ctx context.Context, in *runtimepb.JobEventRequest) (*runtimepb.JobEventResponse, error) {
	// parse the job type from the method or name
	jobType, found := strings.CutPrefix(in.GetMethod(), "job/")
	if !found {
		if in.GetName() == "" {
			return &runtimepb.JobEventResponse{}, errors.New("unsupported invocation")
		}
		jobType = in.GetName()
	}

	if fn, ok := s.jobEventHandlers[jobType]; ok {
		e := &common.JobEvent{
			JobType: jobType,
			Data:    in.GetData().GetValue(),
		}
		if err := fn(ctx, e); err != nil {
			return nil, fmt.Errorf("error executing %s binding: %w", in.GetName(), err)
		}
		return &runtimepb.JobEventResponse{}, nil
	}
	return &runtimepb.JobEventResponse{}, errors.New("job event handler not found")
}
