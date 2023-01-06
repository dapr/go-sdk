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
	"net"
	"os"
	"strconv"
	"sync/atomic"

	"google.golang.org/grpc"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// DaprClienter is an interface implemented by the gRPC client of this SDK.
type DaprClienter interface {
	GrpcClientConn() *grpc.ClientConn
}

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

// NewServiceWithListener creates a new Service with specific listener.
func NewServiceWithListener(lis net.Listener) common.Service {
	return newService(lis)
}

// NewServiceFromCallbackChannel creates a new Service by using the callback channel.
// This makes an outbound connection to Dapr, without creating a listener.
// It requires an existing gRPC client connection to Dapr.
func NewServiceFromCallbackChannel(ctx context.Context, client DaprClienter) (common.Service, error) {
	clientConn := client.GrpcClientConn()

	// Invoke ConnectAppCallback to get the port we should connect to
	appCallbackClient := pb.NewDaprAppCallbackClient(clientConn)
	res, err := appCallbackClient.ConnectAppCallback(ctx, &pb.ConnectAppCallbackRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke ConnectAppCallback: %w", err)
	}

	if res == nil || res.Port < 0 {
		return nil, fmt.Errorf("response from ConnectAppCallback does not contain a port")
	}

	// Determine the host from the target of the gRPC connection, if present
	host := "127.0.0.1"
	target := client.GrpcClientConn().Target()
	if target != "" {
		var h string
		h, _, err = net.SplitHostPort(target)
		if err == nil && h != "" {
			host = h
		}
	}

	// Establish the TCP connection to daprd
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, strconv.Itoa(int(res.Port))))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP address for Dapr at port %d", res.Port)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP connection with Dapr at address %v", addr)
	}

	err = conn.SetKeepAlive(true)
	if err != nil {
		return nil, fmt.Errorf("failed to enable keep-alives in the TCP connection with Dapr at address %v", addr)
	}

	// Use the established connection to create a new common.Service
	return NewServiceWithConnection(conn), nil
}

// NewServiceWithConnection creates a new Service based on an already-established TCP connection.
func NewServiceWithConnection(conn net.Conn) common.Service {
	lis := newListenerFromConn(conn)
	return newService(lis)
}

func newService(lis net.Listener) *Server {
	s := &Server{
		listener:        lis,
		invokeHandlers:  make(map[string]common.ServiceInvocationHandler),
		topicRegistrar:  make(internal.TopicRegistrar),
		bindingHandlers: make(map[string]common.BindingInvocationHandler),
		authToken:       os.Getenv(common.AppAPITokenEnvVar),
	}

	gs := grpc.NewServer()
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

func newListenerFromConn(conn net.Conn) *listenerFromConn {
	usedCh := make(chan struct{}, 1)
	usedCh <- struct{}{}
	return &listenerFromConn{
		conn:   conn,
		usedCh: usedCh,
	}
}

// listenerFromConn implements net.Listener returning an existing net.Conn
type listenerFromConn struct {
	conn   net.Conn
	usedCh chan struct{}
}

func (l listenerFromConn) Accept() (net.Conn, error) {
	// If the connection has already been used, this will block forever
	<-l.usedCh
	return l.conn, nil
}

func (l listenerFromConn) Close() error {
	return l.conn.Close()
}

func (l listenerFromConn) Addr() net.Addr {
	return l.conn.LocalAddr()
}
