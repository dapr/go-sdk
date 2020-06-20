package client

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

const (
	daprPortDefault    = "50001"
	daprPortEnvVarName = "DAPR_GRPC_PORT"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// NewClient instantiates Dapr client using DAPR_GRPC_PORT environment variable as port.
func NewClient() (client *Client, err error) {
	port := os.Getenv(daprPortEnvVarName)
	if port == "" {
		port = daprPortDefault
	}
	return NewClientWithPort(port)
}

// NewClientWithPort instantiates Dapr using specific port.
func NewClientWithPort(port string) (client *Client, err error) {
	if port == "" {
		return nil, errors.New("nil port")
	}
	return NewClientWithAddress(net.JoinHostPort("127.0.0.1", port))
}

// NewClientWithAddress instantiates Dapr using specific address (inclding port).
func NewClientWithAddress(address string) (client *Client, err error) {
	if address == "" {
		return nil, errors.New("nil address")
	}
	logger.Printf("dapr client initializing for: %s", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "error creating connection to '%s': %v", address, err)
	}
	return NewClientWithConnection(conn), nil
}

// NewClientWithConnection instantiates Dapr client using specific connection.
func NewClientWithConnection(conn *grpc.ClientConn) *Client {
	return &Client{
		connection:  conn,
		protoClient: pb.NewDaprClient(conn),
	}
}

// Client is the Dapr client.
type Client struct {
	connection  *grpc.ClientConn
	protoClient pb.DaprClient
}

// Close cleans up all resources created by the client.
func (c *Client) Close() {
	if c.connection != nil {
		c.connection.Close()
	}
}

func authContext(ctx context.Context) context.Context {
	token := os.Getenv("DAPR_API_TOKEN")
	if token == "" {
		return ctx
	}
	md := metadata.Pairs("dapr-api-token", token)
	return metadata.NewOutgoingContext(ctx, md)
}
