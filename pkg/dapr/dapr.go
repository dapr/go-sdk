package dapr

import (
	"context"

	"github.com/golang/protobuf/proto"
)

// NewClient ... TODO
func NewClient(opts ...interface{}) (Client, error) {
	// TODO: dial the connection and create a NewDaprClient (or accept a conn)
	return nil, nil
}

// Client ... TODO
type Client interface {
	Invoke(ctx context.Context, service, method string, arguments, result proto.Message) error
	Publish(ctx context.Context, topic string, data proto.Message) error
	Binding(ctx context.Context, name string, data proto.Message) error

	// State Methods
	SaveState(ctx context.Context, requests ...*State) error
	GetState(ctx context.Context, key string) (*State, error)
	DeleteState(ctx context.Context, key string) (*State, error)
}

// State ... TODO
type State struct {
	Name  string
	Value proto.Message
}

// Register ... TODO
func Register(server Server) error {
	// TODO: do a server registration
	return nil
}

// Server ... TODO
type Server interface {
	Subscriptions() []string
	Bindings() []string
	OnBinding(ctx context.Context, args interface{}) (interface{}, error) // TODO: stronger types
	OnTopic(ctx context.Context, msg proto.Message) error
	// TODO: use reflection to call additional methods when getting OnInvoke messages
}
