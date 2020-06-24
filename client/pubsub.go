package client

import (
	"context"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEvent pubishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, topic string, in []byte) error {
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
