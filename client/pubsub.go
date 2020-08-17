package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEvent pubishes data onto topic in specific pubsub component.
func (c *GRPCClient) PublishEvent(ctx context.Context, component, topic string, in []byte) error {
	if component == "" {
		return errors.New("nil component")
	}
	if topic == "" {
		return errors.New("nil topic")
	}
	if in == nil {
		return errors.New("nil in")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName: component,
		Topic:      topic,
		Data:       in,
	}

	_, err := c.protoClient.PublishEvent(authContext(ctx), envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topic)
	}

	return nil
}
