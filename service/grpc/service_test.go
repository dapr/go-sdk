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
	assert.NotNil(t, server)
	err := server.Stop()
	assert.Nilf(t, err, "error stopping server")
}
