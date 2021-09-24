package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

const (
	daprPortDefault    = "50001"
	daprPortEnvVarName = "DAPR_GRPC_PORT"
	traceparentKey     = "traceparent"
	apiTokenKey        = "dapr-api-token" /* #nosec */
	apiTokenEnvVarName = "DAPR_API_TOKEN" /* #nosec */
)

var (
	logger               = log.New(os.Stdout, "", 0)
	_             Client = &GRPCClient{}
	defaultClient Client
	doOnce        sync.Once
)

type Config struct {
	// addr is the hostport ie. 127.0.0.1:50001
	addr string
	conn grpc.ClientConn
}

// Option implements functional options for configuring the client
type Option func(*GRPCClient) error

// WithConnection instantiates Dapr client using specific connection.
func WithConnection(conn *grpc.ClientConn) Option {
	return func(g *GRPCClient) error {
		g.connection = conn
		g.protoClient = pb.NewDaprClient(conn)
		return nil
	}
}

// WithAddress allows customizing the address and dialoptions of grpc
// connection to dapr
func WithAddress(addr string, opts ...grpc.DialOption) Option {
	return func(g *GRPCClient) error {
		if addr == "" {
			return errors.New("nil address")
		}

		if opts == nil {
			opts = []grpc.DialOption{grpc.WithInsecure()}
		}

		logger.Printf("dapr client initializing for: %s", addr)
		conn, err := grpc.Dial(addr, opts...)
		if err != nil {
			return errors.Wrapf(err, "error creating connection to '%s': %v", addr, err)
		}

		return WithConnection(conn)(g)
	}
}

// New instantiates Dapr client using DAPR_GRPC_PORT environment variable as port.
// Note, this default factory function creates Dapr client only once. All subsequent invocations
// will return the already created instance. To create multiple instances of the Dapr client,
func New(opts ...Option) (*GRPCClient, error) {

	if hasToken := os.Getenv(apiTokenEnvVarName); hasToken != "" {
		logger.Println("client uses API token")
	}

	g := &GRPCClient{
		// set defaults
		authToken: os.Getenv(apiTokenEnvVarName),
	}

	addr := fmt.Sprintf("127.0.0.1:%s", os.Getenv(daprPortEnvVarName))

	// If no options are defined, then set up a default connection with
	// default dial options
	if len(opts) == 0 {
		if err := WithAddress(addr)(g); err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}
	return g, nil
}

// NewClient is deprecated, see New()
func NewClient() (client Client, err error) {
	log.Print("NewClient() is deprecated, see New()")
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

// NewClientWithPort is deprecated, see WithAddress()
func NewClientWithPort(port string) (client Client, err error) {
	log.Print("NewClientWithPort() is deprecated, see WithAddress()")
	if port == "" {
		return nil, errors.New("nil port")
	}
	return NewClientWithAddress(net.JoinHostPort("127.0.0.1", port))
}

// NewClientWithAddress is deprecated, see WithAddress()
func NewClientWithAddress(address string) (client Client, err error) {
	return New(WithAddress(address))
}

// NewClientWithConnection is deprecated, see WithConnection()
func NewClientWithConnection(conn *grpc.ClientConn) Client {
	client, _ := New(WithConnection(conn))
	return client
}

// GRPCClient is the gRPC implementation of Dapr client.
type GRPCClient struct {
	connection *grpc.ClientConn
	// native grpc client used to connect to dapr
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
// Allows empty string to reset token on existing client.
func (c *GRPCClient) WithAuthToken(token string) {
	c.mux.Lock()
	c.authToken = token
	c.mux.Unlock()
}

// WithTraceID adds existing trace ID to the outgoing context.
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
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(apiTokenKey, c.authToken))
}

// Shutdown the sidecar.
func (c *GRPCClient) Shutdown(ctx context.Context) error {
	_, err := c.protoClient.Shutdown(c.withAuthToken(ctx), &emptypb.Empty{})
	if err != nil {
		return errors.Wrap(err, "error shutting down the sidecar")
	}
	return nil
}
