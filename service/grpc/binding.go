package grpc

import (
	"context"
	"fmt"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

// AddBindingInvocationHandler appends provided binding invocation handler with its name to the service
func (s *Server) AddBindingInvocationHandler(name string, fn func(ctx context.Context, in *common.BindingEvent) (out []byte, err error)) error {
	if name == "" {
		return fmt.Errorf("binding name required")
	}
	if fn == nil {
		return fmt.Errorf("binding handler required")
	}
	s.bindingHandlers[name] = fn
	return nil
}

// ListInputBindings is called by Dapr to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage
func (s *Server) ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error) {
	list := make([]string, 0)
	for k := range s.bindingHandlers {
		list = append(list, k)
	}

	return &pb.ListInputBindingsResponse{
		Bindings: list,
	}, nil
}

// OnBindingEvent gets invoked every time a new event is fired from a registered binding. The message carries the binding name, a payload and optional metadata
func (s *Server) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	if in == nil {
		return nil, errors.New("nil binding event request")
	}
	if fn, ok := s.bindingHandlers[in.Name]; ok {
		e := &common.BindingEvent{
			Data:     in.Data,
			Metadata: in.Metadata,
		}
		data, err := fn(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing %s binding", in.Name)
		}
		return &pb.BindingEventResponse{
			Data: data,
		}, nil
	}

	return nil, fmt.Errorf("binding not implemented: %s", in.Name)
}
