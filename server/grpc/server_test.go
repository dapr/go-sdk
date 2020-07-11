package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

// go test -v -count=1 -run TestServer ./server/grpc
func TestServer(t *testing.T) {
	t.Parallel()
	server := getTestServer()
	startTestServer(server)
	stopTestServer(t, server)
}

func getTestServer() Server {
	return NewServerWithListener(bufconn.Listen(1024 * 1024))
}

func startTestServer(server Server) {
	go func() {
		if err := server.Start(); err != nil && err.Error() != "closed" {
			panic(err)
		}
	}()
}

func stopTestServer(t *testing.T, server Server) {
	err := server.Stop()
	assert.Nilf(t, err, "error stopping server")
}
