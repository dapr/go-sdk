package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

func TestServer(t *testing.T) {
	t.Parallel()
	server := getTestServer()
	startTestServer(server)
	stopTestServer(t, server)
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
