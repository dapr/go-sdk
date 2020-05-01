package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/dapr/v1"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

func (c *Client) PublishEvent(ctx context.Context, topic string, in []byte) error {
	if topic == "" {
		return errors.New("nil topic")
	}

	envelop := &pb.PublishEventEnvelope{
		Topic: topic,
		Data: &any.Any{
			Value: in,
		},
	}

	_, err := c.ProtoClient.PublishEvent(ctx, envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topic)
	}

	return nil
}

func (c *Client) PublishEventJSON(ctx context.Context, topic string, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.PublishEvent(ctx, topic, b)
}
