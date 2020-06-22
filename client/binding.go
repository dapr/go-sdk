package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// InvokeBinding invokes specific operation on the configured Dapr binding.
// This method covers input, output, and bi-directional bindings.
func (c *GRPCClient) InvokeBinding(ctx context.Context, name, op string, in []byte, min map[string]string) (out []byte, mout map[string]string, err error) {
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

// InvokeOutputBinding invokes configured Dapr binding with data (allows nil).InvokeOutputBinding
// This method differs from InvokeBinding in that it doesn't expect any content being returned from the invoked method.
func (c *GRPCClient) InvokeOutputBinding(ctx context.Context, name, operation string, data []byte) error {
	_, _, err := c.InvokeBinding(ctx, name, operation, data, nil)
	if err != nil {
		return errors.Wrap(err, "error invoking output binding")
	}
	return nil
}
