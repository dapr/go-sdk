package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"

	cpb "github.com/dapr/go-sdk/dapr/proto/common/v1"
)

// AddInvocationHandler adds provided handler to the local collection before server start
func (s *ServerImp) AddInvocationHandler(method string, fn func(ctx context.Context, contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte)) {
	s.invokeHandlers[method] = fn
}

// OnInvoke gets invoked when a remote service has called the app through Dapr
func (s *ServerImp) OnInvoke(ctx context.Context, in *cpb.InvokeRequest) (*cpb.InvokeResponse, error) {
	if val, ok := s.invokeHandlers[in.Method]; ok {
		ct, d := val(ctx, in.ContentType, in.Data.Value)
		return &cpb.InvokeResponse{
			ContentType: ct,
			Data: &any.Any{
				Value: d,
			},
		}, nil
	}
	return nil, fmt.Errorf("method not implemented: %s", in.Method)
}
