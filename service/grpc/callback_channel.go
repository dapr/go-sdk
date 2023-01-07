/*
Copyright 2023 The Dapr Authors
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
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
)

var emptyPbPool = sync.Pool{
	New: func() any {
		return &emptypb.Empty{}
	},
}

// DaprClienter is an interface implemented by the gRPC client of this SDK.
type DaprClienter interface {
	GrpcClientConn() *grpc.ClientConn
}

// NewServiceFromCallbackChannel creates a new Service by using the callback channel.
// This makes an outbound connection to Dapr, without creating a listener.
// It requires an existing gRPC client connection to Dapr.
func NewServiceFromCallbackChannel(ctx context.Context, client DaprClienter, grpcOpts ...grpc.ServerOption) (common.Service, error) {
	clientConn := client.GrpcClientConn()

	// Invoke ConnectAppCallback to get the port we should connect to
	appCallbackClient := pb.NewDaprAppChannelClient(clientConn)
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

	// Do not use TCP keepalives since we have a health channel in the app
	err = conn.SetKeepAlive(false)
	if err != nil {
		return nil, fmt.Errorf("failed to disable keep-alives in the TCP connection with Dapr at address %v", addr)
	}

	// Use the established connection to create a new common.Service
	lis := newListenerFromConn()
	lis.conn <- conn
	srv := newService(lis, grpcOpts...)
	pb.RegisterDaprAppChannelCallbackServer(srv.grpcServer, srv)
	return srv, nil
}

// HealthCheck check app health status.
func (s *Server) Ping(stream pb.DaprAppChannelCallback_PingServer) error {
	// Send a ping every 5s, including as soon as it's connected (Dapr expects a first ping to validate the server is up)
	const pingInterval = 5 * time.Second
	for stream.Context().Err() == nil {
		in := emptyPbPool.Get()
		err := stream.SendMsg(in)
		emptyPbPool.Put(in)
		if err != nil {
			// TODO: CLOSE THE CHANNEL
			break
		}
		time.Sleep(pingInterval)
	}
	return nil
}

// NewServiceWithConnection creates a new Service based on an already-established TCP connection.
func NewServiceWithConnection(conn net.Conn, grpcOpts ...grpc.ServerOption) common.Service {
	lis := newListenerFromConn()
	return newService(lis, grpcOpts...)
}

func newListenerFromConn() *listenerFromConn {
	return &listenerFromConn{
		conn: make(chan net.Conn, 1),
	}
}

// listenerFromConn implements net.Listener returning an existing net.Conn
type listenerFromConn struct {
	conn chan net.Conn
	lock sync.Mutex
}

// AddConn adds a conection so it's the next one to be accepted
func (l *listenerFromConn) AddConn(conn net.Conn) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Drain the channel first
	l.drain()

	l.conn <- conn
}

func (l *listenerFromConn) drain() {
	for {
		select {
		case c := <-l.conn:
			_ = c.Close()
		default:
			break
		}
	}
}

func (l *listenerFromConn) Accept() (net.Conn, error) {
	// This blocks until a connection is added
	return <-l.conn, nil
}

func (l *listenerFromConn) Close() error {
	return nil
}

func (l *listenerFromConn) Addr() net.Addr {
	return &net.TCPAddr{}
}
