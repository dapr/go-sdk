package client

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

const (
	daprPortDefault    = "50001"
	daprPortEnvVarName = "DAPR_GRPC_PORT"
	traceparentKey     = "traceparent"
	apiTokenKey        = "dapr-api-token"
	apiTokenEnvVarName = "DAPR_API_TOKEN"
)

var (
	logger               = log.New(os.Stdout, "", 0)
	_             Client = (*GRPCClient)(nil)
	defaultClient Client
	doOnce        sync.Once
)

// Client is the interface for Dapr client implementation.
type Client interface {
	// InvokeBinding invokes specific operation on the configured Dapr binding.
	// This method covers input, output, and bi-directional bindings.
	InvokeBinding(ctx context.Context, in *BindingInvocation) (out *BindingEvent, err error)

	// InvokeOutputBinding invokes configured Dapr binding with data.InvokeOutputBinding
	// This method differs from InvokeBinding in that it doesn't expect any content being returned from the invoked method.
	InvokeOutputBinding(ctx context.Context, in *BindingInvocation) error

	// InvokeService invokes service without raw data
	InvokeService(ctx context.Context, serviceID, method string) (out []byte, err error)

	// InvokeServiceWithContent invokes service with content
	InvokeServiceWithContent(ctx context.Context, serviceID, method string, content *DataContent) (out []byte, err error)

	// PublishEvent pubishes data onto topic in specific pubsub component.
	PublishEvent(ctx context.Context, component, topic string, in []byte) error

	// GetSecret retreaves preconfigred secret from specified store using key.
	GetSecret(ctx context.Context, store, key string, meta map[string]string) (out map[string]string, err error)

	// SaveState saves the raw data into store using default state options.
	SaveState(ctx context.Context, store, key string, data []byte) error

	// SaveStateItems saves multiple state item to store with specified options.
	SaveStateItems(ctx context.Context, store string, items ...*SetStateItem) error

	// GetState retreaves state from specific store using default consistency option.
	GetState(ctx context.Context, store, key string) (item *StateItem, err error)

	// GetStateWithConsistency retreaves state from specific store using provided state consistency.
	GetStateWithConsistency(ctx context.Context, store, key string, meta map[string]string, sc StateConsistency) (item *StateItem, err error)

	// GetBulkItems retreaves state for multiple keys from specific store.
	GetBulkItems(ctx context.Context, store string, keys []string, parallelism int32) ([]*StateItem, error)

	// DeleteState deletes content from store using default state options.
	DeleteState(ctx context.Context, store, key string) error

	// DeleteStateWithETag deletes content from store using provided state options and etag.
	DeleteStateWithETag(ctx context.Context, store, key, etag string, meta map[string]string, opts *StateOptions) error

	// ExecuteStateTransaction provides way to execute multiple operations on a specified store.
	ExecuteStateTransaction(ctx context.Context, store string, meta map[string]string, ops []*StateOperation) error

	// WithTraceID adds existing trace ID to the outgoing context.
	WithTraceID(ctx context.Context, id string) context.Context

	// WithAuthToken sets Dapr API token on the instantiated client.
	WithAuthToken(token string)

	// Close cleans up all resources created by the client.
	Close()
}

// NewClient instantiates Dapr client using DAPR_GRPC_PORT environment variable as port.
// Note, this default factory function creates Dapr client only once. All subsequent invocations
// will return the already created instance. To create multiple instances of the Dapr client,
// use one of the parameterized factory functions:
//   NewClientWithPort(port string) (client Client, err error)
//   NewClientWithAddress(address string) (client Client, err error)
//   NewClientWithConnection(conn *grpc.ClientConn) Client
func NewClient() (client Client, err error) {
	port := os.Getenv(daprPortEnvVarName)
	if port == "" {
		port = daprPortDefault
	}
	var onceErr error
	doOnce.Do(func() {
		c, err := NewClientWithPort(port)
		onceErr = errors.Wrap(err, "error creating default client")
		defaultClient = c
	})

	return defaultClient, onceErr
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
	if hasToken := os.Getenv(apiTokenEnvVarName); hasToken != "" {
		logger.Println("client uses API token")
	}
	return NewClientWithConnection(conn), nil
}

// NewClientWithConnection instantiates Dapr client using specific connection.
func NewClientWithConnection(conn *grpc.ClientConn) Client {
	return &GRPCClient{
		connection:  conn,
		protoClient: pb.NewDaprClient(conn),
		authToken:   os.Getenv(apiTokenEnvVarName),
	}
}

// GRPCClient is the gRPC implementation of Dapr client.
type GRPCClient struct {
	connection  *grpc.ClientConn
	protoClient pb.DaprClient
	authToken   string
	mux         sync.Mutex
}

// Close cleans up all resources created by the client.
func (c *GRPCClient) Close() {
	if c.connection != nil {
		c.connection.Close()
	}
}

// WithAuthToken sets Dapr API token on the instantiated client.
// Allows empty string to reset token on existing client
func (c *GRPCClient) WithAuthToken(token string) {
	c.mux.Lock()
	c.authToken = token
	c.mux.Unlock()
}

// WithTraceID adds existing trace ID to the outgoing context
func (c *GRPCClient) WithTraceID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	logger.Printf("using trace parent ID: %s", id)
	md := metadata.Pairs(traceparentKey, id)
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *GRPCClient) withAuthToken(ctx context.Context) context.Context {
	if c.authToken == "" {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(apiTokenKey, string(c.authToken)))
}
