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
func NewServiceFromCallbackChannel(client DaprClienter, grpcOpts ...grpc.ServerOption) (common.Service, error) {
	clientConn := client.GrpcClientConn()

	// Establish a connection using the callback channel
	lis := newlistenerFromCallbackChannel(clientConn)
	err := lis.Connect()
	if err != nil {
		return nil, err
	}

	// Create a server on the connection and return it
	srv := newService(lis, grpcOpts...)
	pb.RegisterDaprAppChannelCallbackServer(srv.grpcServer, srv)
	return srv, nil
}

// HealthCheck check app health status.
func (s *Server) Ping(stream pb.DaprAppChannelCallback_PingServer) error {
	// Send a ping every 5s
	const pingInterval = 5 * time.Second

	t := time.NewTicker(pingInterval)
	defer t.Stop()
	ctxDoneCh := stream.Context().Done()
	// Send a ping as soon as the stream is up too
	// Dapr expects a first ping to validate the server is up
	firstMsg := make(chan struct{})
	close(firstMsg)
loop:
	for {
		select {
		case <-ctxDoneCh:
			// Force a reconnection
			break loop
		case <-firstMsg:
			firstMsg = nil
			in := emptyPbPool.Get()
			err := stream.SendMsg(in)
			emptyPbPool.Put(in)

			// If there's an error, we can assume that the channel is down, so force a reconnection
			if err != nil {
				break loop
			}
		case <-t.C:
			// On the interval, send a ping
			if stream.Context().Err() != nil {
				// Check for context errors again
				break loop
			}
			in := emptyPbPool.Get()
			err := stream.SendMsg(in)
			emptyPbPool.Put(in)

			// If there's an error, we can assume that the channel is down, so force a reconnection
			if err != nil {
				break loop
			}
		}
	}

	if lis, ok := s.listener.(*listenerFromCallbackChannel); ok && lis != nil {
		fmt.Println("recreating callback channel connection")
		go lis.Connect()
	}

	return nil
}

func newlistenerFromCallbackChannel(grpcConn *grpc.ClientConn) *listenerFromCallbackChannel {
	return &listenerFromCallbackChannel{
		grpcConn: grpcConn,
		connCh:   make(chan net.Conn, 1),
	}
}

// listenerFromCallbackChannel implements net.Listener that uses connections established through the app callback channel
type listenerFromCallbackChannel struct {
	grpcConn *grpc.ClientConn
	connCh   chan net.Conn
	lock     sync.Mutex
}

// Connect creates a new connection.
func (l *listenerFromCallbackChannel) Connect() error {
	conn, err := l.doConnect()
	if err != nil {
		return err
	}
	l.AddConn(conn)
	return nil
}

func (l *listenerFromCallbackChannel) doConnect() (net.Conn, error) {
	// Invoke ConnectAppCallback to get the port we should connect to
	// We use WaitForReady and a background context to block until the connection is up
	appCallbackClient := pb.NewDaprAppChannelClient(l.grpcConn)
	res, err := appCallbackClient.ConnectAppCallback(
		context.Background(),
		&pb.ConnectAppCallbackRequest{},
		grpc.WaitForReady(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke ConnectAppCallback: %w", err)
	}

	if res == nil || res.Port < 0 {
		return nil, fmt.Errorf("response from ConnectAppCallback does not contain a port")
	}

	// Determine the host from the target of the gRPC connection, if present
	host := "127.0.0.1"
	target := l.grpcConn.Target()
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

	return conn, nil
}

// AddConn adds a connection so it's the next one to be accepted
func (l *listenerFromCallbackChannel) AddConn(conn net.Conn) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Drain the channel first
	l.drain()

	l.connCh <- conn
}

func (l *listenerFromCallbackChannel) drain() {
	for {
		select {
		case c := <-l.connCh:
			_ = c.Close()
		default:
			return
		}
	}
}

func (l *listenerFromCallbackChannel) Accept() (net.Conn, error) {
	// This blocks until a connection is added
	return <-l.connCh, nil
}

func (l *listenerFromCallbackChannel) Close() error {
	return nil
}

func (l *listenerFromCallbackChannel) Addr() net.Addr {
	return &net.TCPAddr{}
}
