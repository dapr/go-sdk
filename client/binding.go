package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// InvokeBinding invokes specific operation on the configured Dapr binding
func (c *Client) InvokeBinding(ctx context.Context, name, op string, in []byte, min map[string]string) (out []byte, mout map[string]string, err error) {
	if name == "" {
		return nil, nil, errors.New("nil topic")
	}

	req := &pb.InvokeBindingRequest{
		Name:      name,
		Operation: op,
		Data:      in,
		Metadata:  min,
	}

	resp, err := c.protoClient.InvokeBinding(authContext(ctx), req)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error invoking binding %s", name)
	}

	if resp != nil {
		return resp.Data, resp.Metadata, nil
	}

	return nil, nil, nil
}

// InvokeBindingJSON invokes configured Dapr binding with an instance
func (c *Client) InvokeBindingJSON(ctx context.Context, name, operation string, in interface{}) (out []byte, outm map[string]string, err error) {
	if in == nil {
		return nil, nil, errors.New("nil in")
	}
	b, err := json.Marshal(in)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error marshaling content")
	}
	return c.InvokeBinding(ctx, name, operation, b, nil)
}

// InvokeOutputBinding invokes configured Dapr binding with data (allows nil)
func (c *Client) InvokeOutputBinding(ctx context.Context, name, operation string, data []byte) error {
	_, _, err := c.InvokeBinding(ctx, name, operation, data, nil)
	if err != nil {
		return errors.Wrap(err, "error invoking output binding")
	}
	return nil
}
