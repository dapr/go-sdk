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
	logger        = log.New(os.Stdout, "", 0)
	_      Client = (*GRPCClient)(nil)
)

// Client is the interface for Dapr client implementation.
type Client interface {
	// InvokeBinding invokes specific operation on the configured Dapr binding.
	// This method covers input, output, and bi-directional bindings.
	InvokeBinding(ctx context.Context, name, op string, in []byte, min map[string]string) (out []byte, mout map[string]string, err error)

	// InvokeOutputBinding invokes configured Dapr binding with data (allows nil).InvokeOutputBinding
	// This method differs from InvokeBinding in that it doesn't expect any content being returned from the invoked method.
	InvokeOutputBinding(ctx context.Context, name, operation string, data []byte) error

	// InvokeService invokes service without raw data ([]byte).
	InvokeService(ctx context.Context, serviceID, method string) (out []byte, err error)

	// InvokeServiceWithContent invokes service without content (data + content type).
	InvokeServiceWithContent(ctx context.Context, serviceID, method, contentType string, data []byte) (out []byte, err error)

	// PublishEvent pubishes data onto specific pubsub topic.
	PublishEvent(ctx context.Context, topic string, in []byte) error

	// GetSecret retreaves preconfigred secret from specified store using key.
	GetSecret(ctx context.Context, store, key string, meta map[string]string) (out map[string]string, err error)

	// SaveState saves the fully loaded state to store.
	SaveState(ctx context.Context, s *State) error

	// SaveStateData saves the raw data into store using default state options.
	SaveStateData(ctx context.Context, store, key, etag string, data []byte) error

	// SaveStateItem saves the single state item to store.
	SaveStateItem(ctx context.Context, store string, item *StateItem) error

	// GetState retreaves state from specific store using default consistency option.
	GetState(ctx context.Context, store, key string) (out []byte, etag string, err error)

	// GetStateWithConsistency retreaves state from specific store using provided state consistency.
	GetStateWithConsistency(ctx context.Context, store, key string, sc StateConsistency) (out []byte, etag string, err error)

	// DeleteState deletes content from store using default state options.
	DeleteState(ctx context.Context, store, key string) error

	// DeleteStateVersion deletes content from store using provided state options and etag.
	DeleteStateVersion(ctx context.Context, store, key, etag string, opts *StateOptions) error

	// Close cleans up all resources created by the client.
	Close()
}

// NewClient instantiates Dapr client using DAPR_GRPC_PORT environment variable as port.
func NewClient() (client Client, err error) {
	port := os.Getenv(daprPortEnvVarName)
	if port == "" {
		port = daprPortDefault
	}
	return NewClientWithPort(port)
}

// NewClientWithPort instantiates Dapr using specific port.
func NewClientWithPort(port string) (client Client, err error) {
	if port == "" {
		return nil, errors.New("nil port")
	}
	return NewClientWithAddress(net.JoinHostPort("127.0.0.1", port))
}

// NewClientWithAddress instantiates Dapr using specific address (inclding port).
func NewClientWithAddress(address string) (client Client, err error) {
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
func NewClientWithConnection(conn *grpc.ClientConn) Client {
	return &GRPCClient{
		connection:  conn,
		protoClient: pb.NewDaprClient(conn),
	}
}

// GRPCClient is the gRPC implementation of Dapr client.
type GRPCClient struct {
	connection  *grpc.ClientConn
	protoClient pb.DaprClient
}

// Close cleans up all resources created by the client.
func (c *GRPCClient) Close() {
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
