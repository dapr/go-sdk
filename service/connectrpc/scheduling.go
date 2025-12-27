package connectrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	runtimev1 "github.com/dapr/dapr/pkg/proto/runtime/v1"

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

// OnJobEventAlpha1 is invoked by the sidecar following a scheduled job registered in
// the scheduler
func (s *Server) OnJobEventAlpha1(ctx context.Context, in *connect.Request[runtimev1.JobEventRequest]) (*connect.Response[runtimev1.JobEventResponse], error) {
	// parse the job type from the method or name
	jobType, found := strings.CutPrefix(in.Msg.GetMethod(), "job/")
	if !found {
		if in.Msg.GetName() == "" {
			return connect.NewResponse(&runtimev1.JobEventResponse{}), errors.New("unsupported invocation")
		}
		jobType = in.Msg.GetName()
	}

	if fn, ok := s.jobEventHandlers[jobType]; ok {
		e := &common.JobEvent{
			JobType: jobType,
			Data:    in.Msg.GetData().GetValue(),
		}
		if err := fn(ctx, e); err != nil {
			return nil, fmt.Errorf("error executing %s binding: %w", in.Msg.GetName(), err)
		}
		return connect.NewResponse(&runtimev1.JobEventResponse{}), nil
	}
	return connect.NewResponse(&runtimev1.JobEventResponse{}), errors.New("job event handler not found")
}
