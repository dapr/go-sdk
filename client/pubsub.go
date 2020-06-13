package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEvent is the message to publish event data to pubsub topic
func (c *Client) PublishEvent(ctx context.Context, topic string, in []byte) error {
	if topic == "" {
		return errors.New("nil topic")
	}
	if in == nil {
		return errors.New("nil in")
	}

	envelop := &pb.PublishEventRequest{
		Topic: topic,
		Data:  in,
	}

	_, err := c.protoClient.PublishEvent(authContext(ctx), envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topic)
	}

	return nil
}

// PublishEventJSON is the message to publish event data to pubsub topic with identity
func (c *Client) PublishEventJSON(ctx context.Context, topic string, in interface{}) error {
	if in == nil {
		return errors.New("nil in")
	}
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.PublishEvent(ctx, topic, b)
}
