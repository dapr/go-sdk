package client

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/dapr/go-sdk/dapr/proto/dapr/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	daprPortDefault    = "50005"
	daprPortEnvVarName = "DAPR_GRPC_PORT"
)

// NewClientWithAddress instantiates dapr client locally using port from DAPR_GRPC_PORT env var
// When DAPR_GRPC_PORT client defaults to 50005
func NewClient() (client *Client, err error) {
	port := os.Getenv(daprPortEnvVarName)
	if port == "" {
		port = daprPortDefault
	}
	return NewClientWithPort(port)
}

// NewClientWithAddress instantiates dapr client locally for the specific port
func NewClientWithPort(port string) (client *Client, err error) {
	address := net.JoinHostPort("127.0.0.1", port)
	return NewClientWithAddress(address)
}

// NewClientWithAddress instantiates dapr client configured for the specific address
func NewClientWithAddress(address string) (client *Client, err error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "error creating connection to '%s': %v", address, err)
	}
	client = &Client{
		Logger:      log.New(os.Stdout, "", 0),
		Connection:  conn,
		ProtoClient: pb.NewDaprClient(conn),
	}
	return
}

// Client is the dapr client
type Client struct {
	Logger      *log.Logger
	Connection  *grpc.ClientConn
	ProtoClient pb.DaprClient
}

// Close cleans up all resources created by the client
func (c *Client) Close(ctx context.Context) {
	if c.Connection != nil {
		c.Connection.Close()
	}
}
