package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

// BindingEvent represents the input and output of binding invocation
type BindingEvent struct {

	// Name of the input binding component.
	Name string

	// Data is the payload that the input bindings sent.
	Data []byte

	// Metadata is set by the input binging components.
	Metadata map[string]string
}

// AddBindingEventHandler add the provided handler to the server binding halder collection
func (s *ServiceImp) AddBindingEventHandler(name string, fn func(ctx context.Context, in *BindingEvent) (out []byte, err error)) {
	s.bindingHandlers[name] = fn
}

// ListInputBindings is called by Dapr to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage
func (s *ServiceImp) ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error) {
	list := make([]string, 0)
	for k := range s.bindingHandlers {
		list = append(list, k)
	}

	return &pb.ListInputBindingsResponse{
		Bindings: list,
	}, nil
}

// OnBindingEvent gets invoked every time a new event is fired from a registered binding. The message carries the binding name, a payload and optional metadata
func (s *ServiceImp) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	if val, ok := s.bindingHandlers[in.Name]; ok {
		e := &BindingEvent{
			Name:     in.Name,
			Data:     in.Data,
			Metadata: in.Metadata,
		}
		data, err := val(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing %s binding", in.Name)
		}
		return &pb.BindingEventResponse{
			Data: data,
		}, nil
	}

	return nil, fmt.Errorf("binding not implemented: %s", in.Name)
}
