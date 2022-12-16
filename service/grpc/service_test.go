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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

func TestServer(t *testing.T) {
	server := getTestServer()
	startTestServer(server)
	stopTestServer(t, server)
}

func TestServerWithListener(t *testing.T) {
	server := NewServiceWithListener(bufconn.Listen(1024 * 1024))
	assert.NotNil(t, server)
}

func TestService(t *testing.T) {
	_, err := NewService("")
	assert.Errorf(t, err, "expected error from lack of address")
}

func getTestServer() *Server {
	return newService(bufconn.Listen(1024 * 1024))
}

func startTestServer(server *Server) {
	go func() {
		if err := server.Start(); err != nil && err.Error() != "closed" {
			panic(err)
		}
	}()
}

func stopTestServer(t *testing.T, server *Server) {
	t.Helper()

	assert.NotNil(t, server)
	err := server.Stop()
	assert.Nilf(t, err, "error stopping server")
}
