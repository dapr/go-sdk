package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/dapr/v1"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

func (c *Client) SaveState(ctx context.Context, store, key string, in []byte) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}

	envelop := &pb.SaveStateEnvelope{
		StoreName: store,
		Requests: []*pb.StateRequest{
			{
				Key: key,
				Value: &any.Any{
					Value: in,
				},
			},
		},
	}

	_, err := c.protoClient.SaveState(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error saving state into %s", store)
	}

	return nil
}

func (c *Client) SaveStateJSON(ctx context.Context, store, key string, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.SaveState(ctx, store, key, b)
}

func (c *Client) GetState(ctx context.Context, store, key string) (out []byte, err error) {
	if store == "" {
		return nil, errors.New("nil store")
	}
	if key == "" {
		return nil, errors.New("nil key")
	}
	envelop := &pb.GetStateEnvelope{
		StoreName: store,
		Key:       key,
	}

	result, err := c.protoClient.GetState(ctx, envelop)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting state from %s for %s key", store, key)
	}

	return result.Data.Value, nil
}

func (c *Client) DeleteState(ctx context.Context, store, key string) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}
	envelop := &pb.DeleteStateEnvelope{
		StoreName: store,
		Key:       key,
	}

	_, err := c.protoClient.DeleteState(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error deleting state from %s for %s key", store, key)
	}

	return nil
}
