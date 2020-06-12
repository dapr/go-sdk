package client

import (
	"context"
	"encoding/json"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// SaveState is the message to save multiple states into state store
func (c *Client) SaveState(ctx context.Context, store, key string, in []byte) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}

	envelop := &pb.SaveStateRequest{
		StoreName: store,
		States: []*v1.StateItem{
			{
				Key:   key,
				Value: in,
			},
		},
	}

	_, err := c.protoClient.SaveState(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error saving state into %s", store)
	}

	return nil
}

// SaveStateJSON is the message to save multiple states into state store with identity
func (c *Client) SaveStateJSON(ctx context.Context, store, key string, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.SaveState(ctx, store, key, b)
}

// GetState is the message to get key-value states from specific state store
func (c *Client) GetState(ctx context.Context, store, key string) (out []byte, err error) {
	if store == "" {
		return nil, errors.New("nil store")
	}
	if key == "" {
		return nil, errors.New("nil key")
	}
	envelop := &pb.GetStateRequest{
		StoreName: store,
		Key:       key,
	}

	result, err := c.protoClient.GetState(ctx, envelop)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting state from %s for %s key", store, key)
	}

	return result.Data, nil
}

// DeleteState is the message to delete key-value states from specific state store
func (c *Client) DeleteState(ctx context.Context, store, key string) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}
	envelop := &pb.DeleteStateRequest{
		StoreName: store,
		Key:       key,
	}

	_, err := c.protoClient.DeleteState(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error deleting state from %s for %s key", store, key)
	}

	return nil
}
