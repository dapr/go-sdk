package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"

	cpb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	"github.com/dapr/go-sdk/service"
)

// AddServiceInvocationHandler appends provided service invocation handler with its name to the service
func (s *ServiceImp) AddServiceInvocationHandler(method string, fn func(ctx context.Context, in *service.InvocationEvent) (our *service.InvocationEvent, err error)) error {
	if method == "" {
		return fmt.Errorf("servie name required")
	}
	s.invokeHandlers[method] = fn
	return nil
}

// OnInvoke gets invoked when a remote service has called the app through Dapr
func (s *ServiceImp) OnInvoke(ctx context.Context, in *cpb.InvokeRequest) (*cpb.InvokeResponse, error) {
	if in == nil {
		return nil, errors.New("nil invoke request")
	}
	if fn, ok := s.invokeHandlers[in.Method]; ok {
		var e *service.InvocationEvent
		if in.Data != nil {
			e = &service.InvocationEvent{
				ContentType: in.ContentType,
				Data:        in.Data.Value,
				DataTypeURL: in.Data.TypeUrl,
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
				Value:   ct.Data,
				TypeUrl: ct.DataTypeURL,
			},
		}, nil
	}
	return nil, fmt.Errorf("method not implemented: %s", in.Method)
}
