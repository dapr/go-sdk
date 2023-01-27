/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc/metadata"

	cpb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	cc "github.com/dapr/go-sdk/service/common"
)

// AddServiceInvocationHandler appends provided service invocation handler with its method to the service.
func (s *Server) AddServiceInvocationHandler(method string, fn cc.ServiceInvocationHandler) error {
	if method == "" || method == "/" {
		return fmt.Errorf("servie name required")
	}

	if method[0] == '/' {
		method = method[1:]
	}

	if fn == nil {
		return fmt.Errorf("invocation handler required")
	}
	s.invokeHandlers[method] = fn
	return nil
}

// OnInvoke gets invoked when a remote service has called the app through Dapr.
func (s *Server) OnInvoke(ctx context.Context, in *cpb.InvokeRequest) (*cpb.InvokeResponse, error) {
	if in == nil {
		return nil, errors.New("nil invoke request")
	}
	if s.authToken != "" {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			return nil, errors.New("authentication failed")
		} else if vals := md.Get(cc.APITokenKey); len(vals) > 0 {
			if vals[0] != s.authToken {
				return nil, errors.New("authentication failed: app token mismatch")
			}
		} else {
			return nil, errors.New("authentication failed. app token key not exist")
		}
	}
	if fn, ok := s.invokeHandlers[in.Method]; ok {
		e := &cc.InvocationEvent{}
		e.ContentType = in.ContentType

		if in.Data != nil {
			e.Data = in.Data.Value
			e.DataTypeURL = in.Data.TypeUrl
		}

		if in.HttpExtension != nil {
			e.Verb = in.HttpExtension.Verb.String()
			e.QueryString = in.HttpExtension.Querystring
		}

		ct, er := fn(ctx, e)
		if er != nil {
			return nil, er
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
