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
	"net"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// NewService creates new Service.
func NewService(address string) (s common.Service, err error) {
	if address == "" {
		return nil, errors.New("nil address")
	}
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = errors.Wrapf(err, "failed to TCP listen on: %s", address)
		return
	}
	s = newService(lis)
	return
}

// NewServiceWithListener creates new Service with specific listener.
func NewServiceWithListener(lis net.Listener) common.Service {
	return newService(lis)
}

func newService(lis net.Listener) *Server {
	return &Server{
		listener:        lis,
		invokeHandlers:  make(map[string]common.ServiceInvocationHandler),
		topicRegistrar:  make(internal.TopicRegistrar),
		bindingHandlers: make(map[string]common.BindingInvocationHandler),
		authToken:       os.Getenv(common.AppAPITokenEnvVar),
	}
}

// Server is the gRPC service implementation for Dapr.
type Server struct {
	pb.UnimplementedAppCallbackServer
	listener        net.Listener
	invokeHandlers  map[string]common.ServiceInvocationHandler
	topicRegistrar  internal.TopicRegistrar
	bindingHandlers map[string]common.BindingInvocationHandler
	authToken       string
}

func (s *Server) RegisterActorImplFactory(f actor.Factory, opts ...config.Option) {
	panic("Actor is not supported by gRPC API")
}

// Start registers the server and starts it.
func (s *Server) Start() error {
	gs := grpc.NewServer()
	pb.RegisterAppCallbackServer(gs, s)
	return gs.Serve(s.listener)
}

// Stop stops the previously started service.
func (s *Server) Stop() error {
	return s.listener.Close()
}
