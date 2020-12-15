package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// GetSecret retreaves preconfigred secret from specified store using key.
func (c *GRPCClient) GetSecret(ctx context.Context, storeName, key string, meta map[string]string) (data map[string]string, err error) {
	if storeName == "" {
		return nil, errors.New("nil storeName")
	}
	if key == "" {
		return nil, errors.New("nil key")
	}

	req := &pb.GetSecretRequest{
		Key:       key,
		StoreName: storeName,
		Metadata:  meta,
	}

	resp, err := c.protoClient.GetSecret(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking service")
	}

	if resp != nil {
		data = resp.GetData()
	}

	return
}

// GetBulkSecret retreaves all preconfigred secrets for this application.
func (c *GRPCClient) GetBulkSecret(ctx context.Context, storeName string, meta map[string]string) (data map[string]string, err error) {
	if storeName == "" {
		return nil, errors.New("nil storeName")
	}

	req := &pb.GetBulkSecretRequest{
		StoreName: storeName,
		Metadata:  meta,
	}

	resp, err := c.protoClient.GetBulkSecret(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking service")
	}

	if resp != nil {
		data = resp.GetData()
	}

	return
}
