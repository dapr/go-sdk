package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"

	cpb "github.com/dapr/go-sdk/dapr/proto/common/v1"
)

// InvocationEvent represents the input and output of binding invocation
type InvocationEvent struct {

	// ContentType of the Data
	ContentType string

	// Data is the payload that the input bindings sent.
	Data []byte
}

// AddInvocationHandler adds provided handler to the local collection before server start
func (s *ServiceImp) AddInvocationHandler(method string, fn func(ctx context.Context, in *InvocationEvent) (our *InvocationEvent, err error)) {
	s.invokeHandlers[method] = fn
}

// OnInvoke gets invoked when a remote service has called the app through Dapr
func (s *ServiceImp) OnInvoke(ctx context.Context, in *cpb.InvokeRequest) (*cpb.InvokeResponse, error) {
	if in == nil {
		return nil, errors.New("nil invoke request")
	}
	if fn, ok := s.invokeHandlers[in.Method]; ok {
		var e *InvocationEvent
		if in.Data != nil {
			e = &InvocationEvent{
				ContentType: in.ContentType,
				Data:        in.Data.Value,
			}
		}

		ct, er := fn(ctx, e)
		if er != nil {
			return nil, errors.Wrap(er, "error executing handler")
		}

		if ct == nil {
			return nil, nil
		}

		return &cpb.InvokeResponse{
			ContentType: ct.ContentType,
			Data: &any.Any{
				Value: ct.Data,
			},
		}, nil
	}
	return nil, fmt.Errorf("method not implemented: %s", in.Method)
}
