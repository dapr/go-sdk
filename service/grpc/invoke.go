package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"

	cpb "github.com/dapr/go-sdk/dapr/proto/common/v1"
)

// AddServiceInvocationHandler appends provided service invocation handler with its method to the service
func (s *ServiceImp) AddServiceInvocationHandler(method string, fn func(ctx context.Context, in *InvocationEvent) (our *Content, err error)) error {
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
		e := &InvocationEvent{}
		if in != nil {
			e.ContentType = in.ContentType

			if in.Data != nil {
				e.Data = in.Data.Value
				e.DataTypeURL = in.Data.TypeUrl
			}

			if in.HttpExtension != nil {
				e.Verb = in.HttpExtension.Verb.String()
				e.QueryString = in.HttpExtension.Querystring
			}
		}

		ct, er := fn(ctx, e)
		if er != nil {
			return nil, errors.Wrap(er, "error executing handler")
		}

		if ct == nil {
			return &cpb.InvokeResponse{}, nil
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
