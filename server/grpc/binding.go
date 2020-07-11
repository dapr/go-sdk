package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/server/event"
)

// AddBindingEventHandler add the provided handler to the server binding halder collection
func (s *ServerImp) AddBindingEventHandler(name string, fn func(ctx context.Context, in *event.BindingEvent) error) {
	s.bindingHandlers[name] = fn
}

// ListInputBindings is called by Dapr to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage
func (s *ServerImp) ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error) {
	list := make([]string, 0)
	for k := range s.bindingHandlers {
		list = append(list, k)
	}

	return &pb.ListInputBindingsResponse{
		Bindings: list,
	}, nil
}

// OnBindingEvent gets invoked every time a new event is fired from a registered binding. The message carries the binding name, a payload and optional metadata
func (s *ServerImp) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	if val, ok := s.bindingHandlers[in.Name]; ok {
		e := &event.BindingEvent{
			Name:     in.Name,
			Data:     in.Data,
			Metadata: in.Metadata,
		}
		err := val(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing %s binding", in.Name)
		}
		return &pb.BindingEventResponse{}, nil
	}

	return nil, fmt.Errorf("binding not implemented: %s", in.Name)
}
