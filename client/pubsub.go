package client

import (
	"context"
	"encoding/json"

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

// PublishEventfromStruct serializes an struct onto raw and pubishes its contents as data onto topic in specific pubsub component.
func (c *GRPCClient) PublishEventfromStruct(ctx context.Context, component, topic string, in interface{}) error {

	bytes, err := json.Marshal(in)

	if err != nil {
		return errors.WithMessage(err, "error serializing input struct")
	}

	err = c.PublishEvent(ctx, component, topic, bytes)
	if err != nil {
		return errors.Wrapf(err, "error publishing serialized data as event unto %s topic", topic)
	}

	return nil
}
