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
	"errors"
	"fmt"
	"net"
	"os"
	"sync/atomic"

	"google.golang.org/grpc"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// NewService creates new Service.
func NewService(address string) (s common.Service, err error) {
	if address == "" {
		return nil, errors.New("empty address")
	}
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = fmt.Errorf("failed to TCP listen on %s: %w", address, err)
		return
	}
	s = newService(lis)
	return
}

// NewServiceWithListener creates new Service with specific listener.
func NewServiceWithListener(lis net.Listener, opts ...grpc.ServerOption) common.Service {
	return newService(lis, opts...)
}

func newService(lis net.Listener, opts ...grpc.ServerOption) *Server {
	s := &Server{
		listener:        lis,
		invokeHandlers:  make(map[string]common.ServiceInvocationHandler),
		topicRegistrar:  make(internal.TopicRegistrar),
		bindingHandlers: make(map[string]common.BindingInvocationHandler),
		authToken:       os.Getenv(common.AppAPITokenEnvVar),
	}

	gs := grpc.NewServer(opts...)
	pb.RegisterAppCallbackServer(gs, s)
	pb.RegisterAppCallbackHealthCheckServer(gs, s)
	s.grpcServer = gs

	return s
}

// Server is the gRPC service implementation for Dapr.
type Server struct {
	pb.UnimplementedAppCallbackServer
	pb.UnimplementedAppCallbackHealthCheckServer
	listener           net.Listener
	invokeHandlers     map[string]common.ServiceInvocationHandler
	topicRegistrar     internal.TopicRegistrar
	bindingHandlers    map[string]common.BindingInvocationHandler
	healthCheckHandler common.HealthCheckHandler
	authToken          string
	grpcServer         *grpc.Server
	started            uint32
}

func (s *Server) RegisterActorImplFactory(f actor.Factory, opts ...config.Option) {
	panic("Actor is not supported by gRPC API")
}

// Start registers the server and starts it.
func (s *Server) Start() error {
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New("a gRPC server can only be started once")
	}
	return s.grpcServer.Serve(s.listener)
}

// Stop stops the previously-started service.
func (s *Server) Stop() error {
	if atomic.LoadUint32(&s.started) == 0 {
		return nil
	}
	s.grpcServer.Stop()
	s.grpcServer = nil
	return nil
}

// GrecefulStop stops the previously-started service gracefully.
func (s *Server) GracefulStop() error {
	if atomic.LoadUint32(&s.started) == 0 {
		return nil
	}
	s.grpcServer.GracefulStop()
	s.grpcServer = nil
	return nil
}

// GrpcServer returns the grpc.Server object managed by the server.
func (s *Server) GrpcServer() *grpc.Server {
	return s.grpcServer
}
