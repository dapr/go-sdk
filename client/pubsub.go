package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEvent pubishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, component, topic string, in []byte) error {
	if topic == "" {
		return errors.New("topic name required")
	}
	if component == "" {
		return errors.New("component name required")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName: component,
		Topic:      topic,
		Data:       in,
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topic)
	}

	return nil
}
