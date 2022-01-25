package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// GetSecret retrieves preconfigred secret from specified store using key.
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

// GetBulkSecret retrieves all preconfigred secrets for this application.
func (c *GRPCClient) GetBulkSecret(ctx context.Context, storeName string, meta map[string]string) (data map[string]map[string]string, err error) {
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

	data = map[string]map[string]string{}

	if resp != nil {
		d := resp.GetData()
		for k, v := range d {
			data[k] = v.Secrets
		}
	}

	return
}
