package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/dapr/v1"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

func (c *Client) InvokeBinding(ctx context.Context, name string, in []byte) error {
	if name == "" {
		return errors.New("nil topic")
	}

	envelop := &pb.InvokeBindingEnvelope{
		Name: name,
		Data: &any.Any{
			Value: in,
		},
	}

	_, err := c.protoClient.InvokeBinding(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error invoking binding %s", name)
	}

	return nil
}

func (c *Client) InvokeBindingJSON(ctx context.Context, name string, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.InvokeBinding(ctx, name, b)
}
