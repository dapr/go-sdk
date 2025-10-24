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

package connectrpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	runtimev1 "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/dapr/go-sdk/service/common"
)

// AddHealthCheckHandler appends provided app health check handler.
func (s *Server) AddHealthCheckHandler(_ string, fn common.HealthCheckHandler) error {
	if fn == nil {
		return errors.New("health check handler required")
	}

	s.healthCheckHandler = fn

	return nil
}

// HealthCheck check app health status.
func (s *Server) HealthCheck(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[runtimev1.HealthCheckResponse], error) {
	if s.healthCheckHandler != nil {
		if err := s.healthCheckHandler(ctx); err != nil {
			return connect.NewResponse(&runtimev1.HealthCheckResponse{}), err
		}

		return connect.NewResponse(&runtimev1.HealthCheckResponse{}), nil
	}

	return nil, errors.New("health check handler not implemented")
}
