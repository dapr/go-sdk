package dapr

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/dapr/go-sdk/dapr"
	"github.com/dapr/go-sdk/daprclient"
)

// NewClient creates a new Dapr Client object for consuming Dapr.
func NewClient(opts ...grpc.DialOption) (*Client, error) {
	daprPort := os.Getenv("DAPR_GRPC_PORT")
	daprAddress := fmt.Sprintf("localhost:%s", daprPort)
	conn, err := grpc.Dial(daprAddress, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: dapr.NewDaprClient(conn),
		conn:   conn,
	}, nil
}

// Client ... TODO
type Client struct {
	client dapr.DaprClient
	conn   io.Closer
}

// Invoke ... TODO
func (c *Client) Invoke(ctx context.Context, service, method string, arguments, result proto.Message) error {
	args, err := ptypes.MarshalAny(arguments)
	if err != nil {
		return nil
	}
	res, err := c.client.InvokeService(ctx, &dapr.InvokeServiceEnvelope{
		Id:     service,
		Method: method,
		Data:   args,
	})
	if err != nil {
		return err
	}
	return ptypes.UnmarshalAny(res.Data, result)
}

// Publish ... TODO
func (c *Client) Publish(ctx context.Context, topic string, data proto.Message) error {
	d, err := ptypes.MarshalAny(data)
	if err != nil {
		return err
	}
	_, err = c.client.PublishEvent(context.Background(), &dapr.PublishEventEnvelope{
		Topic: topic,
		Data:  d,
	})
	return err
}

// Binding ... TODO
func (c *Client) Binding(ctx context.Context, name string, data proto.Message) error {
	d, err := ptypes.MarshalAny(data)
	if err != nil {
		return err
	}
	_, err = c.client.InvokeBinding(context.Background(), &dapr.InvokeBindingEnvelope{
		Name: name,
		Data: d,
	})
	return err
}

// SaveState ... TODO
func (c *Client) SaveState(ctx context.Context, requests ...*State) error {
	reqs := make([]*dapr.StateRequest, len(requests))
	for i, request := range requests {
		req, err := ptypes.MarshalAny(request.Value)
		if err != nil {
			return err
		}
		reqs[i] = &dapr.StateRequest{
			Key:   request.Key,
			Value: req,
		}
	}
	_, err := c.client.SaveState(ctx, &dapr.SaveStateEnvelope{Requests: reqs})
	return err
}

// GetState ... TODO
func (c *Client) GetState(ctx context.Context, key string, result proto.Message) error {
	r, err := c.client.GetState(ctx, &dapr.GetStateEnvelope{
		Key: key,
	})
	if err != nil {
		return err
	}
	return ptypes.UnmarshalAny(r.Data, result)
}

// DeleteState ... TODO
func (c *Client) DeleteState(ctx context.Context, key string) error {
	_, err := c.client.DeleteState(ctx, &dapr.DeleteStateEnvelope{
		Key: key,
	})
	return err
}

// Close ... TODO
func (c *Client) Close() error {
	return c.conn.Close()
}

// State ... TODO
type State struct {
	Key   string
	Value proto.Message
}

type wrapper struct{}

// Serve ... TODO
func Serve(port string) error {
	// TODO: read port as env var DAPR_PORT
	// https://github.com/dapr/dapr/issues/102
	lis, err := net.Listen(`tcp`, port)
	if err != nil {
		return err
	}

	svr := grpc.NewServer()
	daprclient.RegisterDaprClientServer(svr, &wrapper{})
	return svr.Serve(lis)
}

// --------- INVOCATIONS ---------

var handlers = make(map[string]InvokeHandler, 16) // TODO: concurrent writes

// InvokeHandler ...
type InvokeHandler func(ctx context.Context, args proto.Message) (result proto.Message, err error)

// AddInvokeHandler ...
func AddInvokeHandler(name string, handler InvokeHandler) {
	handlers[name] = handler
}

// This method gets invoked when a remote service has called the app through Dapr
// The payload carries a Method to identify the method, a set of metadata properties and an optional payload
func (*wrapper) OnInvoke(ctx context.Context, in *daprclient.InvokeEnvelope) (*any.Any, error) {
	handler, ok := handlers[in.Method]
	if !ok {
		return nil, fmt.Errorf(`handler not available: %v`, in.Method)
	}
	res, err := handler(ctx, in.Data)
	if err != nil {
		return nil, err
	}
	return ptypes.MarshalAny(res)
}

// --------- BINDINGS ---------

var bindings = make(map[string]BindingHandler, 16) // TODO: concurrent writes

// BindingHandler ...
type BindingHandler func(ctx context.Context, args proto.Message) (result proto.Message, err error)

// AddBindingHandler ...
func AddBindingHandler(name string, handler BindingHandler) {
	bindings[name] = handler
}

// This method gets invoked every time a new event is fired from a registerd binding. The message carries the binding name, a payload and optional metadata
func (w *wrapper) OnBindingEvent(ctx context.Context, in *daprclient.BindingEventEnvelope) (*daprclient.BindingResponseEnvelope, error) {
	handler, ok := bindings[in.Name]
	if !ok {
		return nil, fmt.Errorf(`binding not handled: %v`, in.Name)
	}
	res, err := handler(ctx, in.Data)
	if err != nil {
		return nil, err
	}
	var out *any.Any
	if res != nil {
		out, err = ptypes.MarshalAny(res)
	}
	if err != nil {
		return nil, err
	}
	return &daprclient.BindingResponseEnvelope{Data: out}, nil
}

// GetBindingsSubscriptions will be called by Dapr to get the list of bindings the app will get invoked by.
func (w *wrapper) GetBindingsSubscriptions(ctx context.Context, in *empty.Empty) (*daprclient.GetBindingsSubscriptionsEnvelope, error) {
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	return &daprclient.GetBindingsSubscriptionsEnvelope{
		Bindings: names,
	}, nil
}

// --------- TOPICS ---------

var topics = make(map[string]TopicHandler, 16) // TODO: concurrent writes

// TopicHandler ...
type TopicHandler func(ctx context.Context, data proto.Message) error

// AddTopicHandler ...
func AddTopicHandler(topic string, handler TopicHandler) {
	topics[topic] = handler
}

// This method is fired whenever a message has been published to a topic that has been subscribed. Dapr sends published messages in a CloudEvents 0.3 envelope.
func (w *wrapper) OnTopicEvent(ctx context.Context, in *daprclient.CloudEventEnvelope) (*empty.Empty, error) {
	handler, ok := topics[in.Topic]
	if !ok {
		return nil, fmt.Errorf(`topic not handled: %v`, in.Topic)
	}
	err := handler(ctx, in.Data)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// GetTopicSubscriptions will be called by Dapr to get the list of topics the app wants to subscribe to.
func (w *wrapper) GetTopicSubscriptions(ctx context.Context, in *empty.Empty) (*daprclient.GetTopicSubscriptionsEnvelope, error) {
	names := make([]string, 0, len(topics))
	for name := range topics {
		names = append(names, name)
	}
	return &daprclient.GetTopicSubscriptionsEnvelope{
		Topics: names,
	}, nil
}
