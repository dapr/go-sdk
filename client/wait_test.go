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

package client

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	unresponsiveServerHost         = "127.0.0.1"
	unresponsiveTCPPort            = "0" // Port set to 0 so O.S. auto-selects one for us
	unresponsiveUnixSocketFilePath = "/tmp/unresponsive-server.socket"

	waitTimeout       = 5 * time.Second
	connectionTimeout = 4 * waitTimeout       // Larger than waitTimeout but still bounded
	autoCloseTimeout  = 2 * connectionTimeout // Server will close connections after this
)

type Server struct {
	listener     net.Listener
	address      string
	done         chan bool
	nClientsSeen uint64
}

func (s *Server) Close() {
	close(s.done)
	if err := s.listener.Close(); err != nil {
		logger.Fatal(err)
	}
	os.Remove(unresponsiveUnixSocketFilePath)
}

func (s *Server) listenButKeepSilent() {
	for {
		conn, err := s.listener.Accept() // Accept connections but that's it!
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				logger.Fatal(err)
				break
			}
		} else {
			go func(conn net.Conn) {
				atomic.AddUint64(&s.nClientsSeen, 1)
				time.Sleep(autoCloseTimeout)
				conn.Close()
			}(conn)
		}
	}
}

func createUnresponsiveTCPServer() (*Server, error) {
	return createUnresponsiveServer("tcp", net.JoinHostPort(unresponsiveServerHost, unresponsiveTCPPort))
}

func createUnresponsiveUnixServer() (*Server, error) {
	return createUnresponsiveServer("unix", unresponsiveUnixSocketFilePath)
}

func createUnresponsiveServer(network string, unresponsiveServerAddress string) (*Server, error) {
	serverListener, err := net.Listen(network, unresponsiveServerAddress)
	if err != nil {
		logger.Fatalf("Creation of test server on network %s and address %s failed with error: %v",
			network, unresponsiveServerAddress, err)
		return nil, err
	}

	server := &Server{
		listener:     serverListener,
		address:      serverListener.Addr().String(),
		done:         make(chan bool),
		nClientsSeen: 0,
	}

	go server.listenButKeepSilent()

	return server, nil
}

func createNonBlockingClient(ctx context.Context, serverAddr string) (client Client, err error) {
	conn, err := grpc.DialContext(
		ctx,
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	return NewClientWithConnection(conn), nil
}

func TestGrpcWaitHappyCase(t *testing.T) {
	ctx := context.Background()

	err := testClient.Wait(ctx, waitTimeout)
	assert.NoError(t, err)
}

func TestGrpcWaitUnresponsiveTcpServer(t *testing.T) {
	ctx := context.Background()

	server, err := createUnresponsiveTCPServer()
	assert.NoError(t, err)
	defer server.Close()

	clientConnectionTimeoutCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()
	client, err := createNonBlockingClient(clientConnectionTimeoutCtx, server.address)
	assert.NoError(t, err)

	err = client.Wait(ctx, waitTimeout)
	assert.Error(t, err)
	assert.Equal(t, errWaitTimedOut, err)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&server.nClientsSeen))
}

func TestGrpcWaitUnresponsiveUnixServer(t *testing.T) {
	ctx := context.Background()

	server, err := createUnresponsiveUnixServer()
	assert.NoError(t, err)
	defer server.Close()

	clientConnectionTimeoutCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()
	client, err := createNonBlockingClient(clientConnectionTimeoutCtx, "unix://"+server.address)
	assert.NoError(t, err)

	err = client.Wait(ctx, waitTimeout)
	assert.Error(t, err)
	assert.Equal(t, errWaitTimedOut, err)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&server.nClientsSeen))
}
