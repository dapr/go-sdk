package dapr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"

	"github.com/dapr/go-sdk/dapr"
)

// NewClient ... TODO
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
	args, err := toAny(arguments)
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
	return fromAny(result, res.Data)
}

// Publish ... TODO
func (c *Client) Publish(ctx context.Context, topic string, data proto.Message) error {
	d, err := toAny(data)
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
	d, err := toAny(data)
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
		req, err := toAny(request.Value)
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
	return fromAny(result, r.Data)
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

// Serve ... TODO
func Serve(port int, server Server) error {
	// TODO: read port as env var DAPR_PORT
	// https://github.com/dapr/dapr/issues/102
	return errors.New(`TODO`)
}

// Server ... TODO
type Server interface {
	Subscriptions() []string
	Bindings() []string
	OnBinding(ctx context.Context, args interface{}) (interface{}, error) // TODO: stronger types
	OnTopic(ctx context.Context, msg proto.Message) error
	// TODO: use reflection to call additional methods when getting OnInvoke messages
}

// TODO!!!! These methods are some helpers to get from any.Any to user defined proto.Message

func toAny(in proto.Message) (*any.Any, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}
	return &any.Any{
		Value: data,
	}, nil
}

func fromAny(obj proto.Message, stuff *any.Any) error {
	return proto.Unmarshal(stuff.Value, obj)
}
