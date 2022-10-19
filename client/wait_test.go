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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	unresponsiveServerHost         = "127.0.0.1"
	unresponsiveTcpPort            = "0" // Port set to 0 so OS auto-selects one
	unresponsiveUnixSocketFilePath = "/tmp/unresponsive-server.socket"
	autoCloseTimeout               = 1 * time.Minute
)

func listenButKeepSilent(serverListener net.Listener, serverAddr string) {
	for {
		conn, err := serverListener.Accept() // Accept connections but that's it!
		if err == nil {
			break
		} else {
			// logger.Printf("Server on address %s got a new connection", serverAddr)
			go func(conn net.Conn) {
				time.Sleep(autoCloseTimeout)
				conn.Close()
			}(conn)
		}
	}
}

func createUnresponsiveTcpServer() (serverAddr string, serverListener net.Listener, err error) {
	serverAddr = "" // default
	serverListener, err = net.Listen("tcp", net.JoinHostPort(unresponsiveServerHost, unresponsiveTcpPort))
	if err != nil {
		logger.Fatal(err)
		return "", nil, err
	}

	serverAddr = serverListener.Addr().String()
	logger.Println("Created TCP server on address", serverAddr)

	go listenButKeepSilent(serverListener, serverAddr)

	return serverAddr, serverListener, nil
}

func createUnresponsiveUnixServer() (serverAddr string, serverListener net.Listener, err error) {
	serverListener, err = net.Listen("unix", unresponsiveUnixSocketFilePath)
	if err != nil {
		logger.Fatalf("socket test server created with error: %v", err)
		return "", nil, err
	}

	serverAddr = serverListener.Addr().String()

	go listenButKeepSilent(serverListener, serverAddr)

	return serverAddr, serverListener, nil
}

func createClientConnection(ctx context.Context, serverAddr string) (client Client, err error) {
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

func TestGrpcWait(t *testing.T) {
	const (
		waitTimeout       = 5 * time.Second
		connectionTimeout = 4 * waitTimeout // Larger than waitTimeout but still bounded
	)
	ctx := context.Background()

	t.Run("Happy Case Client test", func(t *testing.T) {
		err := testClient.Wait(ctx, waitTimeout)
		assert.NoError(t, err)
	})

	t.Run("Non-responding TCP server times out", func(t *testing.T) {
		serverAddr, serverListener, err := createUnresponsiveTcpServer()
		assert.NoError(t, err)
		defer serverListener.Close()

		clientConnectionTimeoutCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
		defer cancel()

		client, err := createClientConnection(clientConnectionTimeoutCtx, serverAddr)
		assert.NoError(t, err)

		err = client.Wait(ctx, waitTimeout)
		assert.Error(t, err)
		assert.Equal(t, errWaitTimedOut, err)
	})

	t.Run("Non-responding Unix Domain Socket server times out", func(t *testing.T) {
		serverAddr, serverListener, err := createUnresponsiveUnixServer()
		assert.NoError(t, err)
		defer serverListener.Close()
		defer os.Remove(unresponsiveUnixSocketFilePath)

		clientConnectionTimeoutCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
		defer cancel()

		client, err := createClientConnection(clientConnectionTimeoutCtx, serverAddr)
		assert.NoError(t, err)

		err = client.Wait(ctx, waitTimeout)
		assert.Error(t, err)
		assert.Equal(t, errWaitTimedOut, err)
	})
}
