package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// GetSecret retreaves preconfigred secret from specified store using key.
func (c *GRPCClient) GetSecret(ctx context.Context, store, key string, meta map[string]string) (out map[string]string, err error) {
	if store == "" {
		return nil, errors.New("nil store")
	}
	if key == "" {
		return nil, errors.New("nil key")
	}

	req := &pb.GetSecretRequest{
		Key:       key,
		StoreName: store,
		Metadata:  meta,
	}

	resp, err := c.protoClient.GetSecret(authContext(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking service")
	}

	if resp != nil {
		out = resp.GetData()
	}

	return
}
