package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// InvokeBinding invokes specific operation on the configured Dapr binding
func (c *Client) InvokeBinding(ctx context.Context, name, op string, in []byte, meta map[string]string) error {
	if name == "" {
		return errors.New("nil topic")
	}

	envelop := &pb.InvokeBindingRequest{
		Name:      name,
		Operation: op,
		Data:      in,
		Metadata:  meta,
	}

	_, err := c.protoClient.InvokeBinding(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error invoking binding %s", name)
	}

	return nil
}

// InvokeBindingJSON invokes configured Dapr binding with an instance
func (c *Client) InvokeBindingJSON(ctx context.Context, name, operation string, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.InvokeBinding(ctx, name, operation, b, nil)
}
